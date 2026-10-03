package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/util"
)

type FinanceRepo struct {
	pool *pgxpool.Pool
}

func NewFinanceRepo(pool *pgxpool.Pool) *FinanceRepo {
	return &FinanceRepo{pool: pool}
}

const cashSelectCols = `cash_id, group_id, cash_type, date, account_name, description,
	debit, credit, created_by, created_at, updated_at`

func scanCash(row pgx.Row) (model.CashTransaction, error) {
	var k model.CashTransaction
	err := row.Scan(&k.CashID, &k.GroupID, &k.CashType, &k.Date, &k.AccountName,
		&k.Description, &k.Debit, &k.Credit, &k.CreatedBy, &k.CreatedAt, &k.UpdatedAt)
	return k, err
}

func (r *FinanceRepo) CashList(ctx context.Context, groupID, kasType string) ([]model.CashTransaction, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+cashSelectCols+` FROM cash_transactions
		 WHERE group_id = $1 AND cash_type = $2
		 ORDER BY date ASC, created_at ASC, cash_id ASC`,
		groupID, kasType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.CashTransaction, 0)
	for rows.Next() {
		k, err := scanCash(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

func (r *FinanceRepo) CashInsert(ctx context.Context, k *model.CashTransaction) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO cash_transactions
		   (cash_id, group_id, cash_type, date, account_name, description,
		    debit, credit, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4::date, $5, $6, $7, $8, $9, now(), now())`,
		k.CashID, k.GroupID, k.CashType, k.Date.Format("2006-01-02"),
		k.AccountName, k.Description, k.Debit, k.Credit, k.CreatedBy)
	return err
}

func (r *FinanceRepo) CashUpdate(ctx context.Context, k *model.CashTransaction) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE cash_transactions SET date=$2::date, account_name=$3, description=$4,
		  debit=$5, credit=$6, updated_at=now()
		 WHERE cash_id=$1 AND group_id=$7`,
		k.CashID, k.Date.Format("2006-01-02"), k.AccountName,
		k.Description, k.Debit, k.Credit, ptrStr(k.GroupID))
	return err
}

func (r *FinanceRepo) CashDelete(ctx context.Context, groupID, kasID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM cash_transactions WHERE cash_id=$1 AND group_id=$2`,
		kasID, groupID)
	return err
}

const dueMemberCols = `due_member_id, group_id, member_name, monthly_target, status, created_at, updated_at`

func scanDueMember(row pgx.Row) (model.DueMember, error) {
	var m model.DueMember
	err := row.Scan(&m.MemberID, &m.GroupID, &m.MemberName, &m.MonthlyTarget,
		&m.Status, &m.CreatedAt, &m.UpdatedAt)
	return m, err
}

func (r *FinanceRepo) DueMembers(ctx context.Context, groupID string) ([]model.DueMember, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+dueMemberCols+` FROM due_members
		 WHERE group_id = $1 ORDER BY member_name ASC`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.DueMember, 0)
	for rows.Next() {
		m, err := scanDueMember(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *FinanceRepo) DueMemberUpsert(ctx context.Context, m *model.DueMember) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO due_members (due_member_id, group_id, member_name, monthly_target, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, COALESCE(NULLIF($5,''), 'ACTIVE'), now(), now())
		 ON CONFLICT (due_member_id) DO UPDATE SET
		   member_name = EXCLUDED.member_name,
		   monthly_target = EXCLUDED.monthly_target,
		   status = EXCLUDED.status,
		   updated_at = now()`,
		m.MemberID, m.GroupID, m.MemberName, m.MonthlyTarget, m.Status)
	return err
}

func (r *FinanceRepo) DueMemberDelete(ctx context.Context, groupID, memberID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM due_members WHERE due_member_id=$1 AND group_id=$2`,
		memberID, groupID)
	return err
}

const duePaymentCols = `payment_id, group_id, member_id, payment_date, total_amount,
	carryover_ir, connecting_fund, community_dues,
	outreach_fund, thousand_fund, funeral_fund, ukhro_mt, notes, status, created_by,
	created_at, updated_at`

func scanDuePayment(row pgx.Row) (model.DuePayment, error) {
	var p model.DuePayment
	err := row.Scan(&p.PaymentID, &p.GroupID, &p.MemberID, &p.PaymentDate, &p.TotalAmount,
		&p.CarryoverIR, &p.ConnectingFund,
		&p.CommunityDues, &p.OutreachFund, &p.ThousandFund, &p.FuneralFund, &p.UkhroMT,
		&p.Notes, &p.Status, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

// CarryoversByPayment mengambil rincian susulan untuk banyak payment sekaligus.
func (r *FinanceRepo) CarryoversByPayment(ctx context.Context, paymentIDs []string) (map[string][]model.DuePaymentCarryover, error) {
	out := map[string][]model.DuePaymentCarryover{}
	if len(paymentIDs) == 0 {
		return out, nil
	}
	rows, err := r.pool.Query(ctx,
		`SELECT carryover_id, payment_id, month, amount, created_at
		 FROM due_payment_carryovers WHERE payment_id = ANY($1) ORDER BY month ASC`, paymentIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var c model.DuePaymentCarryover
		if err := rows.Scan(&c.CarryoverID, &c.PaymentID, &c.Month, &c.Amount, &c.CreatedAt); err != nil {
			return nil, err
		}
		out[c.PaymentID] = append(out[c.PaymentID], c)
	}
	return out, rows.Err()
}

// ReplaceCarryovers menimpa seluruh rincian susulan satu payment (delete + insert).
func (r *FinanceRepo) ReplaceCarryovers(ctx context.Context, paymentID string, items []model.DuePaymentCarryover) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `DELETE FROM due_payment_carryovers WHERE payment_id = $1`, paymentID); err != nil {
		return err
	}
	for _, it := range items {
		if it.CarryoverID == "" {
			it.CarryoverID = util.NewID("CRY")
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO due_payment_carryovers (carryover_id, payment_id, month, amount)
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (payment_id, month) DO UPDATE SET amount = EXCLUDED.amount`,
			it.CarryoverID, paymentID, it.Month, it.Amount); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *FinanceRepo) DuePayments(ctx context.Context, groupID, month string) ([]model.DuePayment, error) {
	q := `SELECT ` + duePaymentCols + ` FROM due_payments WHERE group_id = $1`
	args := []interface{}{groupID}
	if month != "" {
		q += ` AND to_char(payment_date, 'YYYY-MM') = $2`
		args = append(args, month)
	}
	q += ` ORDER BY payment_date ASC, created_at ASC`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.DuePayment, 0)
	for rows.Next() {
		p, err := scanDuePayment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *FinanceRepo) DuePaymentUpsert(ctx context.Context, p *model.DuePayment) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO due_payments
		   (payment_id, group_id, member_id, payment_date, total_amount, carryover_ir,
		    connecting_fund, community_dues,
		    outreach_fund, thousand_fund, funeral_fund, ukhro_mt, notes, status,
		    created_by, created_at, updated_at)
		 VALUES ($1,$2,$3,$4::date,$5,$6,$7,$8,$9,$10,$11,$12,$13,
		         COALESCE(NULLIF($14,''),'ACTIVE'),$15,now(),now())
		 ON CONFLICT (payment_id) DO UPDATE SET
		   member_id = EXCLUDED.member_id,
		   payment_date = EXCLUDED.payment_date,
		   total_amount = EXCLUDED.total_amount,
		   carryover_ir = EXCLUDED.carryover_ir,
		   connecting_fund = EXCLUDED.connecting_fund,
		   community_dues = EXCLUDED.community_dues,
		   outreach_fund = EXCLUDED.outreach_fund,
		   thousand_fund = EXCLUDED.thousand_fund,
		   funeral_fund = EXCLUDED.funeral_fund,
		   ukhro_mt = EXCLUDED.ukhro_mt,
		   notes = EXCLUDED.notes,
		   updated_at = now()`,
		p.PaymentID, p.GroupID, p.MemberID, p.PaymentDate.Format("2006-01-02"),
		p.TotalAmount, p.CarryoverIR,
		p.ConnectingFund, p.CommunityDues, p.OutreachFund, p.ThousandFund,
		p.FuneralFund, p.UkhroMT, p.Notes, p.Status, p.CreatedBy)
	return err
}

func (r *FinanceRepo) DuePaymentReverse(ctx context.Context, groupID, paymentID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE due_payments SET status='REVERSED', updated_at=now()
		 WHERE payment_id=$1 AND group_id=$2`, paymentID, groupID)
	return err
}

func (r *FinanceRepo) ZakatListWithCounts(ctx context.Context, groupID string) ([]model.ZakatRecordDTO, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT z.zakat_id, z.group_id, z.title, z.description, z.location,
		        z.soul_count, z.total_rice_kg, z.total_money_rp, z.status, z.transaction_date, z.completed_at, z.version, z.created_by, z.updated_by, z.created_at, z.updated_at,
		        COALESCE(p.cnt, 0) AS payer_count,
		        COALESCE(rc.cnt, 0) AS recipient_count,
		        COALESCE(c.cats, '') AS cats
		 FROM zakat_records z
		 LEFT JOIN (SELECT zakat_id, COUNT(*) cnt FROM zakat_payers GROUP BY zakat_id) p ON p.zakat_id = z.zakat_id
		 LEFT JOIN (SELECT zakat_id, COUNT(*) cnt FROM zakat_recipients GROUP BY zakat_id) rc ON rc.zakat_id = z.zakat_id
		 LEFT JOIN (
		   SELECT zakat_id, string_agg(zakat_category, ',' ORDER BY zakat_category) AS cats
		   FROM (SELECT DISTINCT zakat_id, zakat_category FROM zakat_payers
		         UNION
		         SELECT DISTINCT zakat_id, zakat_category FROM zakat_recipients) u
		   GROUP BY zakat_id
		 ) c ON c.zakat_id = z.zakat_id
		 WHERE z.group_id = $1 AND z.deleted_at IS NULL
		 ORDER BY z.transaction_date DESC NULLS LAST, z.created_at DESC`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.ZakatRecordDTO
	for rows.Next() {
		var z model.ZakatRecord
		var payerCount, recipientCount int
		var cats string
		if err := rows.Scan(&z.ZakatID, &z.GroupID, &z.Title, &z.Description, &z.Location,
			&z.SoulCount, &z.TotalRiceKg, &z.TotalMoneyRp, &z.Status, &z.TransactionDate, &z.CompletedAt, &z.Version, &z.CreatedBy, &z.UpdatedBy, &z.CreatedAt, &z.UpdatedAt,
			&payerCount, &recipientCount, &cats); err != nil {
			return nil, err
		}
		var tgl, comp string
		if z.TransactionDate != nil {
			tgl = z.TransactionDate.Format("2006-01-02")
		}
		if z.CompletedAt != nil {
			comp = z.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		categories := make([]string, 0)
		for _, c := range strings.Split(cats, ",") {
			if c = strings.TrimSpace(c); c != "" {
				categories = append(categories, c)
			}
		}
		out = append(out, model.ZakatRecordDTO{
			ZakatID:     z.ZakatID,
			GroupID:     ptrStr(z.GroupID),
			Title:       z.Title,
			Description: z.Description,
			Location:    z.Location,
			Categories:  categories,
			SoulCount:   z.SoulCount,
			TotalRiceKg:     z.TotalRiceKg,
			TotalMoneyRp:    z.TotalMoneyRp,
			Status:          z.Status,
			TransactionDate: tgl,
			CompletedAt:     comp,
			Version:         z.Version,
			UpdatedBy:       z.UpdatedBy,
			UpdatedAt:       z.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			PayerCount:      payerCount,
			RecipientCount:  recipientCount,
		})
	}
	return out, rows.Err()
}

func (r *FinanceRepo) ZakatList(ctx context.Context, groupID string) ([]model.ZakatRecord, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT zakat_id, group_id, title, description, location, soul_count, total_rice_kg,
		  total_money_rp, status, transaction_date, completed_at, deleted_at, version, created_by, updated_by, created_at, updated_at
		 FROM zakat_records WHERE group_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.ZakatRecord, 0)
	for rows.Next() {
		var z model.ZakatRecord
		if err := rows.Scan(&z.ZakatID, &z.GroupID, &z.Title, &z.Description, &z.Location,
			&z.SoulCount, &z.TotalRiceKg, &z.TotalMoneyRp, &z.Status,
			&z.TransactionDate, &z.CompletedAt, &z.DeletedAt, &z.Version, &z.CreatedBy, &z.UpdatedBy, &z.CreatedAt, &z.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, z)
	}
	return out, rows.Err()
}

func (r *FinanceRepo) ZakatUpsert(ctx context.Context, z *model.ZakatRecord) error {
	var tgl interface{}
	if z.TransactionDate != nil {
		tgl = z.TransactionDate.Format("2006-01-02")
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO zakat_records
		   (zakat_id, group_id, title, description, location, soul_count, total_rice_kg,
		    total_money_rp, status, transaction_date, version, created_by, updated_by, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,COALESCE(NULLIF($9,''),'PENDING'),$10::date,$11,$12,$13,now(),now())
		 ON CONFLICT (zakat_id) DO UPDATE SET
		   title = EXCLUDED.title,
		   description = EXCLUDED.description,
		   location = EXCLUDED.location,
		   soul_count = EXCLUDED.soul_count,
		   total_rice_kg = EXCLUDED.total_rice_kg,
		   total_money_rp = EXCLUDED.total_money_rp,
		   status = EXCLUDED.status,
		   transaction_date = EXCLUDED.transaction_date,
		   version = EXCLUDED.version,
		   updated_by = EXCLUDED.updated_by,
		   updated_at = now()`,
		z.ZakatID, z.GroupID, z.Title, z.Description, z.Location, z.SoulCount,
		z.TotalRiceKg, z.TotalMoneyRp, z.Status, tgl, z.Version, z.CreatedBy, z.UpdatedBy)
	return err
}

func (r *FinanceRepo) ZakatDetail(ctx context.Context, groupID, zakatID string) (*model.ZakatRecord, []model.ZakatPayer, []model.ZakatRecipient, []model.ZakatAllocation, error) {
	var z model.ZakatRecord
	err := r.pool.QueryRow(ctx,
		`SELECT zakat_id, group_id, title, description, location, soul_count, total_rice_kg,
		  total_money_rp, status, transaction_date, completed_at, deleted_at, version, created_by, updated_by, created_at, updated_at
		 FROM zakat_records WHERE zakat_id = $1 AND group_id = $2 AND deleted_at IS NULL`, zakatID, groupID).Scan(
		&z.ZakatID, &z.GroupID, &z.Title, &z.Description, &z.Location,
		&z.SoulCount, &z.TotalRiceKg, &z.TotalMoneyRp, &z.Status,
		&z.TransactionDate, &z.CompletedAt, &z.DeletedAt, &z.Version, &z.CreatedBy, &z.UpdatedBy, &z.CreatedAt, &z.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, nil, nil, nil
		}
		return nil, nil, nil, nil, err
	}

	payersRows, err := r.pool.Query(ctx, `SELECT payer_id, zakat_id, master_id, name, amount, zakat_category, family_members_count, sort_order, created_at FROM zakat_payers WHERE zakat_id = $1 ORDER BY sort_order ASC`, zakatID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer payersRows.Close()
	var payers []model.ZakatPayer
	for payersRows.Next() {
		var p model.ZakatPayer
		if err := payersRows.Scan(&p.PayerID, &p.ZakatID, &p.MasterID, &p.Name, &p.Amount, &p.ZakatCategory, &p.FamilyMembersCount, &p.SortOrder, &p.CreatedAt); err != nil {
			return nil, nil, nil, nil, err
		}
		payers = append(payers, p)
	}

	recipRows, err := r.pool.Query(ctx, `SELECT recipient_id, zakat_id, master_id, name, amount, zakat_category, sort_order, created_at FROM zakat_recipients WHERE zakat_id = $1 ORDER BY sort_order ASC`, zakatID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer recipRows.Close()
	var recipients []model.ZakatRecipient
	for recipRows.Next() {
		var p model.ZakatRecipient
		if err := recipRows.Scan(&p.RecipientID, &p.ZakatID, &p.MasterID, &p.Name, &p.Amount, &p.ZakatCategory, &p.SortOrder, &p.CreatedAt); err != nil {
			return nil, nil, nil, nil, err
		}
		recipients = append(recipients, p)
	}

	allocRows, err := r.pool.Query(ctx, `SELECT zakat_id, category, recipient_percent, recipient_amount, recipient_group_percent, recipient_group_amount, recipient_region_percent, recipient_region_amount, sabilillah_percent, sabilillah_amount, amil_percent, amil_amount, amil_group_percent, amil_group_amount, amil_village_percent, amil_village_amount, amil_region_percent, amil_region_amount FROM zakat_allocations WHERE zakat_id = $1`, zakatID)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer allocRows.Close()
	var allocs []model.ZakatAllocation
	for allocRows.Next() {
		var a model.ZakatAllocation
		if err := allocRows.Scan(&a.ZakatID, &a.Category, &a.RecipientPercent, &a.RecipientAmount, &a.RecipientGroupPercent, &a.RecipientGroupAmount, &a.RecipientRegionPercent, &a.RecipientRegionAmount, &a.SabilillahPercent, &a.SabilillahAmount, &a.AmilPercent, &a.AmilAmount, &a.AmilGroupPercent, &a.AmilGroupAmount, &a.AmilVillagePercent, &a.AmilVillageAmount, &a.AmilRegionPercent, &a.AmilRegionAmount); err != nil {
			return nil, nil, nil, nil, err
		}
		allocs = append(allocs, a)
	}

	return &z, payers, recipients, allocs, nil
}

func (r *FinanceRepo) ReplaceZakatPayers(ctx context.Context, zakatID string, items []model.ZakatPayer) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM zakat_payers WHERE zakat_id = $1`, zakatID); err != nil {
		return err
	}
	for i, p := range items {
		if p.PayerID == "" { p.PayerID = util.NewID("PYR") }
		var masterID *string
		if p.MasterID != nil && *p.MasterID != "" { masterID = p.MasterID }
		if _, err := tx.Exec(ctx, `
			INSERT INTO zakat_payers
			  (payer_id, zakat_id, master_id, name, amount, zakat_category,
			   family_members_count, sort_order)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			p.PayerID, zakatID, masterID, p.Name, p.Amount, p.ZakatCategory,
			p.FamilyMembersCount, i); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *FinanceRepo) ReplaceZakatRecipients(ctx context.Context, zakatID string, items []model.ZakatRecipient) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM zakat_recipients WHERE zakat_id = $1`, zakatID); err != nil {
		return err
	}
	for i, p := range items {
		if p.RecipientID == "" { p.RecipientID = util.NewID("RCP") }
		var masterID *string
		if p.MasterID != nil && *p.MasterID != "" { masterID = p.MasterID }
		if _, err := tx.Exec(ctx, `
			INSERT INTO zakat_recipients
			  (recipient_id, zakat_id, master_id, name, amount, zakat_category, sort_order)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			p.RecipientID, zakatID, masterID, p.Name, p.Amount, p.ZakatCategory, i); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *FinanceRepo) ReplaceZakatAllocations(ctx context.Context, zakatID string, items []model.ZakatAllocation) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil { return err }
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM zakat_allocations WHERE zakat_id = $1`, zakatID); err != nil {
		return err
	}
	for _, a := range items {
		if _, err := tx.Exec(ctx, `
			INSERT INTO zakat_allocations
			  (zakat_id, category, recipient_percent, recipient_amount, recipient_group_percent, recipient_group_amount, recipient_region_percent, recipient_region_amount, sabilillah_percent, sabilillah_amount, amil_percent, amil_amount, amil_group_percent, amil_group_amount, amil_village_percent, amil_village_amount, amil_region_percent, amil_region_amount)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
			zakatID, a.Category, a.RecipientPercent, a.RecipientAmount, a.RecipientGroupPercent, a.RecipientGroupAmount, a.RecipientRegionPercent, a.RecipientRegionAmount, a.SabilillahPercent, a.SabilillahAmount, a.AmilPercent, a.AmilAmount, a.AmilGroupPercent, a.AmilGroupAmount, a.AmilVillagePercent, a.AmilVillageAmount, a.AmilRegionPercent, a.AmilRegionAmount); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *FinanceRepo) MastersList(ctx context.Context, groupID string) ([]model.MasterEntry, []model.MasterEntry, error) {
	payersRows, err := r.pool.Query(ctx, `SELECT master_id, group_id, name, status FROM master_payers WHERE group_id = $1 ORDER BY name ASC`, groupID)
	if err != nil { return nil, nil, err }
	defer payersRows.Close()
	var payers []model.MasterEntry
	for payersRows.Next() {
		var e model.MasterEntry
		if err := payersRows.Scan(&e.MasterID, &e.GroupID, &e.Name, &e.Status); err != nil { return nil, nil, err }
		payers = append(payers, e)
	}

	recipRows, err := r.pool.Query(ctx, `SELECT master_id, group_id, name, status FROM master_recipients WHERE group_id = $1 ORDER BY name ASC`, groupID)
	if err != nil { return nil, nil, err }
	defer recipRows.Close()
	var recipients []model.MasterEntry
	for recipRows.Next() {
		var e model.MasterEntry
		if err := recipRows.Scan(&e.MasterID, &e.GroupID, &e.Name, &e.Status); err != nil { return nil, nil, err }
		recipients = append(recipients, e)
	}
	return payers, recipients, nil
}

func (r *FinanceRepo) MasterUpsert(ctx context.Context, kind, groupID, name string) (string, error) {
	table  := "master_payers"
	prefix := "MPY"
	if kind == "recipient" {
		table  = "master_recipients"
		prefix = "MRS"
	}
	var id string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO `+table+` (master_id, group_id, name)
		VALUES ($1, $2, $3)
		ON CONFLICT (group_id, name) DO UPDATE SET updated_at = now()
		RETURNING master_id`,
		util.NewID(prefix), groupID, name,
	).Scan(&id)
	return id, err
}

func (r *FinanceRepo) ZakatSetStatus(ctx context.Context, groupID, zakatID, status string, completedAt *time.Time) error {
	var comp sql.NullTime
	if completedAt != nil {
		comp.Time = *completedAt
		comp.Valid = true
	}
	_, err := r.pool.Exec(ctx,
		`UPDATE zakat_records SET status=$3, completed_at=$4, updated_at=now(), version=version+1
		 WHERE zakat_id=$1 AND group_id=$2`, zakatID, groupID, status, comp)
	return err
}

func (r *FinanceRepo) ZakatDelete(ctx context.Context, groupID, zakatID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE zakat_records SET deleted_at=now(), updated_at=now(), version=version+1
		 WHERE zakat_id=$1 AND group_id=$2`, zakatID, groupID)
	return err
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (r *FinanceRepo) syncStates(ctx context.Context, table, idCol, groupID string) (map[string]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+idCol+`, to_char(updated_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') FROM `+table+` WHERE group_id = $1`,
		groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var id, ts string
		if err := rows.Scan(&id, &ts); err != nil {
			return nil, err
		}
		out[id] = ts
	}
	return out, rows.Err()
}

func (r *FinanceRepo) CashStates(ctx context.Context, groupID string) (map[string]string, error) {
	return r.syncStates(ctx, "cash_transactions", "cash_id", groupID)
}

func (r *FinanceRepo) DueMemberStates(ctx context.Context, groupID string) (map[string]string, error) {
	return r.syncStates(ctx, "due_members", "due_member_id", groupID)
}

func (r *FinanceRepo) DuePaymentStates(ctx context.Context, groupID string) (map[string]string, error) {
	return r.syncStates(ctx, "due_payments", "payment_id", groupID)
}

func (r *FinanceRepo) ZakatStates(ctx context.Context, groupID string) (map[string]string, error) {
	return r.syncStates(ctx, "zakat_records", "zakat_id", groupID)
}

func (r *FinanceRepo) MarkSynced(ctx context.Context, table, idCol, id, source string, sheetRow int) error {
	if source == "app" {
		_, err := r.pool.Exec(ctx,
			`UPDATE `+table+` SET sync_source='app', sheet_row=$3
			 WHERE `+idCol+`=$1 AND COALESCE(sync_source,'app') != 'sheet'`, id, sheetRow)
		return err
	}
	_, err := r.pool.Exec(ctx,
		`UPDATE `+table+` SET sync_source=$2, sheet_row=$3
		 WHERE `+idCol+`=$1`, id, source, sheetRow)
	return err
}

func (r *FinanceRepo) SyncError(ctx context.Context, groupID, entity, entityID, direction, msg string) error {
	var gid *string
	if strings.TrimSpace(groupID) != "" {
		gid = &groupID
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO finance_sync_errors (error_id, group_id, entity, entity_id, direction, message, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, now())`,
		util.NewID("ERR"), gid, entity, entityID, direction, msg)
	return err
}

func (r *FinanceRepo) IDsOf(ctx context.Context, table, idCol, groupID string) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+idCol+` FROM `+table+` WHERE group_id = $1`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (r *FinanceRepo) CashIDs(ctx context.Context, groupID string) ([]string, error) {
	return r.IDsOf(ctx, "cash_transactions", "cash_id", groupID)
}

func (r *FinanceRepo) DueMemberIDs(ctx context.Context, groupID string) ([]string, error) {
	return r.IDsOf(ctx, "due_members", "due_member_id", groupID)
}

func (r *FinanceRepo) DuePaymentIDs(ctx context.Context, groupID string) ([]string, error) {
	return r.IDsOf(ctx, "due_payments", "payment_id", groupID)
}

func (r *FinanceRepo) ZakatIDs(ctx context.Context, groupID string) ([]string, error) {
	return r.IDsOf(ctx, "zakat_records", "zakat_id", groupID)
}

func (r *FinanceRepo) Tombstone(ctx context.Context, groupID, entity, entityID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO finance_sync_deleted (group_id, entity, entity_id, deleted_at)
		 VALUES ($1, $2, $3, now())
		 ON CONFLICT (group_id, entity, entity_id) DO UPDATE SET deleted_at = now()`,
		groupID, entity, entityID)
	return err
}

func (r *FinanceRepo) Tombstones(ctx context.Context, groupID, entity string) (map[string]bool, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT entity_id FROM finance_sync_deleted WHERE group_id = $1 AND entity = $2`,
		groupID, entity)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}
