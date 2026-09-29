package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"pengajian-backend/internal/model"
)

type FinanceRepo struct {
	pool *pgxpool.Pool
}

func NewFinanceRepo(pool *pgxpool.Pool) *FinanceRepo {
	return &FinanceRepo{pool: pool}
}

const kasSelectCols = `kas_id, group_id, kas_type, tanggal, account_name, description,
	debit, credit, created_by, created_at, updated_at`

func scanKas(row pgx.Row) (model.KasTransaction, error) {
	var k model.KasTransaction
	err := row.Scan(&k.KasID, &k.GroupID, &k.KasType, &k.Tanggal, &k.AccountName,
		&k.Description, &k.Debit, &k.Credit, &k.CreatedBy, &k.CreatedAt, &k.UpdatedAt)
	return k, err
}

func (r *FinanceRepo) KasList(ctx context.Context, groupID, kasType string) ([]model.KasTransaction, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+kasSelectCols+` FROM kas_transactions
		 WHERE group_id = $1 AND kas_type = $2
		 ORDER BY tanggal ASC, created_at ASC, kas_id ASC`,
		groupID, kasType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.KasTransaction, 0)
	for rows.Next() {
		k, err := scanKas(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

func (r *FinanceRepo) KasInsert(ctx context.Context, k *model.KasTransaction) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO kas_transactions
		   (kas_id, group_id, kas_type, tanggal, account_name, description,
		    debit, credit, created_by, created_at, updated_at)
		 VALUES ($1, $2, $3, $4::date, $5, $6, $7, $8, $9, now(), now())`,
		k.KasID, k.GroupID, k.KasType, k.Tanggal.Format("2006-01-02"),
		k.AccountName, k.Description, k.Debit, k.Credit, k.CreatedBy)
	return err
}

func (r *FinanceRepo) KasUpdate(ctx context.Context, k *model.KasTransaction) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE kas_transactions SET tanggal=$2::date, account_name=$3, description=$4,
		  debit=$5, credit=$6, updated_at=now()
		 WHERE kas_id=$1 AND group_id=$7`,
		k.KasID, k.Tanggal.Format("2006-01-02"), k.AccountName,
		k.Description, k.Debit, k.Credit, ptrStr(k.GroupID))
	return err
}

func (r *FinanceRepo) KasDelete(ctx context.Context, groupID, kasID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM kas_transactions WHERE kas_id=$1 AND group_id=$2`,
		kasID, groupID)
	return err
}

const shodaqohMemberCols = `member_id, group_id, member_name, monthly_target, status, created_at, updated_at`

func scanShodaqohMember(row pgx.Row) (model.ShodaqohMember, error) {
	var m model.ShodaqohMember
	err := row.Scan(&m.MemberID, &m.GroupID, &m.MemberName, &m.MonthlyTarget,
		&m.Status, &m.CreatedAt, &m.UpdatedAt)
	return m, err
}

func (r *FinanceRepo) ShodaqohMembers(ctx context.Context, groupID string) ([]model.ShodaqohMember, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+shodaqohMemberCols+` FROM shodaqoh_members
		 WHERE group_id = $1 ORDER BY member_name ASC`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]model.ShodaqohMember, 0)
	for rows.Next() {
		m, err := scanShodaqohMember(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *FinanceRepo) ShodaqohMemberUpsert(ctx context.Context, m *model.ShodaqohMember) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO shodaqoh_members (member_id, group_id, member_name, monthly_target, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, COALESCE(NULLIF($5,''), 'ACTIVE'), now(), now())
		 ON CONFLICT (member_id) DO UPDATE SET
		   member_name = EXCLUDED.member_name,
		   monthly_target = EXCLUDED.monthly_target,
		   status = EXCLUDED.status,
		   updated_at = now()`,
		m.MemberID, m.GroupID, m.MemberName, m.MonthlyTarget, m.Status)
	return err
}

func (r *FinanceRepo) ShodaqohMemberDelete(ctx context.Context, groupID, memberID string) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM shodaqoh_members WHERE member_id=$1 AND group_id=$2`,
		memberID, groupID)
	return err
}

const shodaqohPaymentCols = `payment_id, group_id, member_id, payment_date, total_amount,
	carryover_ir, carryover_months, carryover_breakdown, connecting_fund, community_dues,
	outreach_fund, thousand_fund, funeral_fund, ukhro_mt, notes, status, created_by,
	created_at, updated_at`

func scanShodaqohPayment(row pgx.Row) (model.ShodaqohPayment, error) {
	var p model.ShodaqohPayment
	err := row.Scan(&p.PaymentID, &p.GroupID, &p.MemberID, &p.PaymentDate, &p.TotalAmount,
		&p.CarryoverIR, &p.CarryoverMonths, &p.CarryoverBreakdown, &p.ConnectingFund,
		&p.CommunityDues, &p.OutreachFund, &p.ThousandFund, &p.FuneralFund, &p.UkhroMT,
		&p.Notes, &p.Status, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *FinanceRepo) ShodaqohPayments(ctx context.Context, groupID, month string) ([]model.ShodaqohPayment, error) {
	q := `SELECT ` + shodaqohPaymentCols + ` FROM shodaqoh_payments WHERE group_id = $1`
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
	out := make([]model.ShodaqohPayment, 0)
	for rows.Next() {
		p, err := scanShodaqohPayment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *FinanceRepo) ShodaqohPaymentUpsert(ctx context.Context, p *model.ShodaqohPayment) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO shodaqoh_payments
		   (payment_id, group_id, member_id, payment_date, total_amount, carryover_ir,
		    carryover_months, carryover_breakdown, connecting_fund, community_dues,
		    outreach_fund, thousand_fund, funeral_fund, ukhro_mt, notes, status,
		    created_by, created_at, updated_at)
		 VALUES ($1,$2,$3,$4::date,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,
		         COALESCE(NULLIF($16,''),'ACTIVE'),$17,now(),now())
		 ON CONFLICT (payment_id) DO UPDATE SET
		   member_id = EXCLUDED.member_id,
		   payment_date = EXCLUDED.payment_date,
		   total_amount = EXCLUDED.total_amount,
		   carryover_ir = EXCLUDED.carryover_ir,
		   carryover_months = EXCLUDED.carryover_months,
		   carryover_breakdown = EXCLUDED.carryover_breakdown,
		   connecting_fund = EXCLUDED.connecting_fund,
		   community_dues = EXCLUDED.community_dues,
		   outreach_fund = EXCLUDED.outreach_fund,
		   thousand_fund = EXCLUDED.thousand_fund,
		   funeral_fund = EXCLUDED.funeral_fund,
		   ukhro_mt = EXCLUDED.ukhro_mt,
		   notes = EXCLUDED.notes,
		   updated_at = now()`,
		p.PaymentID, p.GroupID, p.MemberID, p.PaymentDate.Format("2006-01-02"),
		p.TotalAmount, p.CarryoverIR, p.CarryoverMonths, p.CarryoverBreakdown,
		p.ConnectingFund, p.CommunityDues, p.OutreachFund, p.ThousandFund,
		p.FuneralFund, p.UkhroMT, p.Notes, p.Status, p.CreatedBy)
	return err
}

func (r *FinanceRepo) ShodaqohPaymentReverse(ctx context.Context, groupID, paymentID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE shodaqoh_payments SET status='REVERSED', updated_at=now()
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
