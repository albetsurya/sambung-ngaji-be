package repository

import (
	"context"
	"strings"

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

const cashSelectCols = `cash_id, group_id, cash_type, tanggal, account_name, description,
	debit, credit, created_by, created_at, updated_at`

func scanCash(row pgx.Row) (model.CashTransaction, error) {
	var k model.CashTransaction
	err := row.Scan(&k.CashID, &k.GroupID, &k.CashType, &k.Tanggal, &k.AccountName,
		&k.Description, &k.Debit, &k.Credit, &k.CreatedBy, &k.CreatedAt, &k.UpdatedAt)
	return k, err
}

func (r *FinanceRepo) CashList(ctx context.Context, groupID, kasType string) ([]model.CashTransaction, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+cashSelectCols+` FROM cash_transactions
		 WHERE group_id = $1 AND cash_type = $2
		 ORDER BY tanggal ASC, created_at ASC, cash_id ASC`,
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
		   (cash_id, group_id, cash_type, tanggal, account_name, description,
		    debit, credit, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4::date, $5, $6, $7, $8, $9, now(), now())`,
		k.CashID, k.GroupID, k.CashType, k.Tanggal.Format("2006-01-02"),
		k.AccountName, k.Description, k.Debit, k.Credit, k.CreatedBy)
	return err
}

func (r *FinanceRepo) CashUpdate(ctx context.Context, k *model.CashTransaction) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE cash_transactions SET tanggal=$2::date, account_name=$3, description=$4,
		  debit=$5, credit=$6, updated_at=now()
		 WHERE cash_id=$1 AND group_id=$7`,
		k.CashID, k.Tanggal.Format("2006-01-02"), k.AccountName,
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

func (r *FinanceRepo) ZakatList(ctx context.Context, groupID string) ([]model.ZakatRecord, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT zakat_id, group_id, zakat_type, muzakki_name, soul_count, total_rice_kg,
		  total_money_rp, status, transaction_date, details, created_by, created_at, updated_at
		 FROM zakat_records WHERE group_id = $1 ORDER BY created_at DESC`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.ZakatRecord, 0)
	for rows.Next() {
		var z model.ZakatRecord
		if err := rows.Scan(&z.ZakatID, &z.GroupID, &z.ZakatType, &z.MuzakkiName,
			&z.SoulCount, &z.TotalRiceKg, &z.TotalMoneyRp, &z.Status,
			&z.TransactionDate, &z.Details, &z.CreatedBy, &z.CreatedAt, &z.UpdatedAt); err != nil {
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
		   (zakat_id, group_id, zakat_type, muzakki_name, soul_count, total_rice_kg,
		    total_money_rp, status, transaction_date, details, created_by, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,COALESCE(NULLIF($8,''),'PENDING'),$9::date,$10,$11,now(),now())
		 ON CONFLICT (zakat_id) DO UPDATE SET
		   zakat_type = EXCLUDED.zakat_type,
		   muzakki_name = EXCLUDED.muzakki_name,
		   soul_count = EXCLUDED.soul_count,
		   total_rice_kg = EXCLUDED.total_rice_kg,
		   total_money_rp = EXCLUDED.total_money_rp,
		   status = EXCLUDED.status,
		   transaction_date = EXCLUDED.transaction_date,
		   details = EXCLUDED.details,
		   updated_at = now()`,
		z.ZakatID, z.GroupID, z.ZakatType, z.MuzakkiName, z.SoulCount,
		z.TotalRiceKg, z.TotalMoneyRp, z.Status, tgl, z.Details, z.CreatedBy)
	return err
}

func (r *FinanceRepo) ZakatSetStatus(ctx context.Context, groupID, zakatID, status string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE zakat_records SET status=$3, updated_at=now()
		 WHERE zakat_id=$1 AND group_id=$2`, zakatID, groupID, status)
	return err
}

func (r *FinanceRepo) ZakatDelete(ctx context.Context, groupID, zakatID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM zakat_records WHERE zakat_id=$1 AND group_id=$2`, zakatID, groupID)
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
