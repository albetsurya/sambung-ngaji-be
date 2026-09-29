package service

import (
	"context"
	"strings"
	"time"

	"google.golang.org/api/sheets/v4"

	"pengajian-backend/internal/model"
)

func newerThan(sheetTS string, dbTS string) bool {
	st := parseSheetTime(sheetTS)
	if st.IsZero() {
		return false
	}
	dt := parseSheetTime(dbTS)
	if dt.IsZero() {
		return true
	}
	return st.After(dt)
}

func (s *FinanceSyncService) pullCash(ctx context.Context, cli *sheets.Service, groupID string) error {
	rows, err := s.readTab(ctx, cli, sheetTabCash)
	if err != nil {
		return err
	}
	states, err := s.repo.CashStates(ctx, groupID)
	if err != nil {
		return err
	}
	tombs, _ := s.repo.Tombstones(ctx, groupID, "cash")
	for i, r := range rows {
		if r["group_id"] != groupID || r["cash_id"] == "" || tombs[r["cash_id"]] {
			continue
		}
		sheetRow := i + 2
		dbTS, exists := states[r["cash_id"]]
		if exists && !newerThan(r["updated_at"], dbTS) {
			continue
		}
		tgl, err := time.Parse("2006-01-02", r["tanggal"])
		if err != nil {
			s.recordError(ctx, groupID, "cash", r["cash_id"], "sheet->db", "tanggal tidak valid: "+r["tanggal"])
			continue
		}
		gid := groupID
		k := &model.CashTransaction{
			CashID: r["cash_id"], GroupID: &gid, CashType: normCashType(r["cash_type"]),
			Tanggal: tgl, AccountName: r["account_name"], Description: r["description"],
			Debit: parseNum(r["debit"]), Credit: parseNum(r["credit"]),
			CreatedBy: r["created_by"],
		}
		if !exists {
			if err := s.repo.CashInsert(ctx, k); err != nil {
				s.recordError(ctx, groupID, "cash", r["cash_id"], "sheet->db", err.Error())
				continue
			}
		} else if err := s.repo.CashUpdate(ctx, k); err != nil {
			s.recordError(ctx, groupID, "cash", r["cash_id"], "sheet->db", err.Error())
			continue
		}
		_ = s.repo.MarkSynced(ctx, "cash_transactions", "cash_id", r["cash_id"], "sheet", sheetRow)
	}
	return nil
}

func (s *FinanceSyncService) pullDueMembers(ctx context.Context, cli *sheets.Service, groupID string) error {
	rows, err := s.readTab(ctx, cli, sheetTabDueMembers)
	if err != nil {
		return err
	}
	states, err := s.repo.DueMemberStates(ctx, groupID)
	if err != nil {
		return err
	}
	tombs, _ := s.repo.Tombstones(ctx, groupID, "due_members")
	for i, r := range rows {
		if r["group_id"] != groupID || r["member_id"] == "" || tombs[r["member_id"]] {
			continue
		}
		sheetRow := i + 2
		dbTS, exists := states[r["member_id"]]
		if exists && !newerThan(r["updated_at"], dbTS) {
			continue
		}
		if strings.TrimSpace(r["member_name"]) == "" {
			s.recordError(ctx, groupID, "due_members", r["member_id"], "sheet->db", "member_name kosong")
			continue
		}
		status := strings.ToUpper(strings.TrimSpace(r["status"]))
		if status == "" {
			status = "ACTIVE"
		}
		gid := groupID
		m := &model.DueMember{
			MemberID: r["member_id"], GroupID: &gid,
			MemberName:    strings.TrimSpace(r["member_name"]),
			MonthlyTarget: parseNum(r["monthly_target"]), Status: status,
		}
		if err := s.repo.DueMemberUpsert(ctx, m); err != nil {
			s.recordError(ctx, groupID, "due_members", r["member_id"], "sheet->db", err.Error())
			continue
		}
		_ = s.repo.MarkSynced(ctx, "due_members", "due_member_id", r["member_id"], "sheet", sheetRow)
	}
	return nil
}

func (s *FinanceSyncService) pullDuePayments(ctx context.Context, cli *sheets.Service, groupID string) error {
	rows, err := s.readTab(ctx, cli, sheetTabDuePayments)
	if err != nil {
		return err
	}
	states, err := s.repo.DuePaymentStates(ctx, groupID)
	if err != nil {
		return err
	}
	tombs, _ := s.repo.Tombstones(ctx, groupID, "due_payments")
	for i, r := range rows {
		if r["group_id"] != groupID || r["payment_id"] == "" || r["member_id"] == "" || tombs[r["payment_id"]] {
			continue
		}
		sheetRow := i + 2
		dbTS, exists := states[r["payment_id"]]
		if exists && !newerThan(r["updated_at"], dbTS) {
			continue
		}
		tgl, err := time.Parse("2006-01-02", r["payment_date"])
		if err != nil {
			s.recordError(ctx, groupID, "due_payments", r["payment_id"], "sheet->db", "payment_date tidak valid: "+r["payment_date"])
			continue
		}
		total := parseNum(r["carryover_ir"]) + parseNum(r["connecting_fund"]) +
			parseNum(r["community_dues"]) + parseNum(r["outreach_fund"]) +
			parseNum(r["thousand_fund"]) + parseNum(r["funeral_fund"]) + parseNum(r["ukhro_mt"])
		status := strings.ToUpper(strings.TrimSpace(r["status"]))
		if status == "" {
			status = "ACTIVE"
		}
		gid := groupID
		p := &model.DuePayment{
			PaymentID: r["payment_id"], GroupID: &gid, MemberID: r["member_id"],
			PaymentDate: tgl, TotalAmount: total,
			CarryoverIR:     parseNum(r["carryover_ir"]),
			CarryoverMonths: r["carryover_months"], CarryoverBreakdown: r["carryover_breakdown"],
			ConnectingFund: parseNum(r["connecting_fund"]),
			CommunityDues:  parseNum(r["community_dues"]),
			OutreachFund:   parseNum(r["outreach_fund"]),
			ThousandFund:   parseNum(r["thousand_fund"]),
			FuneralFund:    parseNum(r["funeral_fund"]),
			UkhroMT:        parseNum(r["ukhro_mt"]),
			Notes:          r["notes"], Status: status,
		}
		if err := s.repo.DuePaymentUpsert(ctx, p); err != nil {
			s.recordError(ctx, groupID, "due_payments", r["payment_id"], "sheet->db", err.Error())
			continue
		}
		_ = s.repo.MarkSynced(ctx, "due_payments", "payment_id", r["payment_id"], "sheet", sheetRow)
	}
	return nil
}

func (s *FinanceSyncService) pullZakat(ctx context.Context, cli *sheets.Service, groupID string) error {
	rows, err := s.readTab(ctx, cli, sheetTabZakat)
	if err != nil {
		return err
	}
	states, err := s.repo.ZakatStates(ctx, groupID)
	if err != nil {
		return err
	}
	tombs, _ := s.repo.Tombstones(ctx, groupID, "zakat")
	for i, r := range rows {
		if r["group_id"] != groupID || r["zakat_id"] == "" || tombs[r["zakat_id"]] {
			continue
		}
		sheetRow := i + 2
		dbTS, exists := states[r["zakat_id"]]
		if exists && !newerThan(r["updated_at"], dbTS) {
			continue
		}
		if strings.TrimSpace(r["muzakki_name"]) == "" {
			s.recordError(ctx, groupID, "zakat", r["zakat_id"], "sheet->db", "muzakki_name kosong")
			continue
		}
		zt := strings.ToUpper(strings.TrimSpace(r["zakat_type"]))
		if zt != "MAL" {
			zt = "FITRAH"
		}
		var tgl *time.Time
		if strings.TrimSpace(r["transaction_date"]) != "" {
			if t, err := time.Parse("2006-01-02", strings.TrimSpace(r["transaction_date"])); err == nil {
				tgl = &t
			}
		}
		status := strings.ToUpper(strings.TrimSpace(r["status"]))
		if status == "" {
			status = "PENDING"
		}
		gid := groupID
		z := &model.ZakatRecord{
			ZakatID: r["zakat_id"], GroupID: &gid, ZakatType: zt,
			MuzakkiName:  strings.TrimSpace(r["muzakki_name"]),
			SoulCount:    atoi(r["soul_count"], 1),
			TotalRiceKg:  parseNum(r["total_rice_kg"]),
			TotalMoneyRp: parseNum(r["total_money_rp"]),
			Status:       status, TransactionDate: tgl, Details: "{}",
		}
		if err := s.repo.ZakatUpsert(ctx, z); err != nil {
			s.recordError(ctx, groupID, "zakat", r["zakat_id"], "sheet->db", err.Error())
			continue
		}
		_ = s.repo.MarkSynced(ctx, "zakat_records", "zakat_id", r["zakat_id"], "sheet", sheetRow)
	}
	return nil
}

func (s *FinanceSyncService) pushAll(ctx context.Context, cli *sheets.Service, groupID string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	mainSum, err := s.svc.CashList(ctx, groupID, "main")
	if err != nil {
		return err
	}
	amilSum, err := s.svc.CashList(ctx, groupID, "amil")
	if err != nil {
		return err
	}
	cashRows := make([][]interface{}, 0)
	for _, it := range append(mainSum.Transactions, amilSum.Transactions...) {
		cashRows = append(cashRows, []interface{}{
			it.CashID, groupID, it.CashType, it.TransactionDate, it.AccountName,
			it.Description, it.Debit, it.Credit, it.CreatedBy, now,
		})
	}
	if err := s.writeTab(ctx, cli, sheetTabCash, sheetCashHeaders, cashRows); err != nil {
		return err
	}

	dues, err := s.svc.DuesData(ctx, groupID, "")
	if err != nil {
		return err
	}
	memberRows := make([][]interface{}, 0, len(dues.Members))
	for _, m := range dues.Members {
		memberRows = append(memberRows, []interface{}{
			m.MemberID, groupID, m.MemberName, m.MonthlyTarget, m.Status, now,
		})
	}
	if err := s.writeTab(ctx, cli, sheetTabDueMembers, sheetDueMemberHeaders, memberRows); err != nil {
		return err
	}
	paymentRows := make([][]interface{}, 0, len(dues.Payments))
	for _, p := range dues.Payments {
		paymentRows = append(paymentRows, []interface{}{
			p.PaymentID, groupID, p.MemberID, p.PaymentDate, p.TotalAmount,
			p.CarryoverIR, p.CarryoverMonths, p.CarryoverBreakdown,
			p.ConnectingFund, p.CommunityDues, p.OutreachFund, p.ThousandFund,
			p.FuneralFund, p.UkhroMT, p.Notes, p.Status, now,
		})
	}
	if err := s.writeTab(ctx, cli, sheetTabDuePayments, sheetDuePaymentHeaders, paymentRows); err != nil {
		return err
	}

	zakats, err := s.svc.ZakatList(ctx, groupID)
	if err != nil {
		return err
	}
	zakatRows := make([][]interface{}, 0, len(zakats))
	for _, z := range zakats {
		zakatRows = append(zakatRows, []interface{}{
			z.ZakatID, groupID, z.ZakatType, z.MuzakkiName, z.SoulCount,
			z.TotalRiceKg, z.TotalMoneyRp, z.Status, z.TransactionDate, now,
		})
	}
	if err := s.writeTab(ctx, cli, sheetTabZakat, sheetZakatHeaders, zakatRows); err != nil {
		return err
	}
	return s.markPushed(ctx, groupID)
}

func (s *FinanceSyncService) markPushed(ctx context.Context, groupID string) error {
	for _, t := range [][2]string{
		{"cash_transactions", "cash_id"},
		{"due_members", "due_member_id"},
		{"due_payments", "payment_id"},
		{"zakat_records", "zakat_id"},
	} {
		ids, err := s.repo.IDsOf(ctx, t[0], t[1], groupID)
		if err != nil {
			return err
		}
		for _, id := range ids {
			_ = s.repo.MarkSynced(ctx, t[0], t[1], id, "app", 0)
		}
	}
	return nil
}

func atoi(v string, def int) int {
	v = strings.TrimSpace(v)
	if v == "" {
		return def
	}
	var n int
	for _, c := range v {
		if c < '0' || c > '9' {
			return def
		}
		n = n*10 + int(c-'0')
	}
	return n
}
