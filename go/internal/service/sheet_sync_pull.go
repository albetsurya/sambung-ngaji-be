package service

import (
	"context"
	"math"
	"strings"
	"time"

	"google.golang.org/api/sheets/v4"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/util"
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
		dateStr := sheetGet(r, "transaction_date", "tanggal")
		tgl, err := util.ParseSheetDate(dateStr)
		if err != nil {
			s.recordError(ctx, groupID, "cash", r["cash_id"], "sheet->db", "transaction_date tidak valid: "+dateStr)
			continue
		}
		gid := groupID
		k := &model.CashTransaction{
			CashID: r["cash_id"], GroupID: &gid, CashType: normCashType(r["cash_type"]),
			Tanggal: tgl, AccountName: sheetGet(r, "account_name", "account"),
			Description: sheetGet(r, "description", "notes", "keterangan"),
			Debit: parseNum(sheetGet(r, "debit", "debet")), Credit: parseNum(sheetGet(r, "credit", "kredit")),
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
		memberID := sheetGet(r, "member_id", "due_member_id")
		memberName := sheetGet(r, "member_name", "nama", "full_name")
		if r["group_id"] != groupID || memberID == "" || tombs[memberID] {
			continue
		}
		sheetRow := i + 2
		dbTS, exists := states[memberID]
		if exists && !newerThan(r["updated_at"], dbTS) {
			continue
		}
		if strings.TrimSpace(memberName) == "" {
			s.recordError(ctx, groupID, "due_members", memberID, "sheet->db", "member_name kosong")
			continue
		}
		status := strings.ToUpper(strings.TrimSpace(r["status"]))
		if status == "" {
			status = "ACTIVE"
		}
		gid := groupID
		m := &model.DueMember{
			MemberID: memberID, GroupID: &gid,
			MemberName:    strings.TrimSpace(memberName),
			MonthlyTarget: parseNum(sheetGet(r, "monthly_target", "nominal_bulanan")), Status: status,
		}
		if err := s.repo.DueMemberUpsert(ctx, m); err != nil {
			s.recordError(ctx, groupID, "due_members", memberID, "sheet->db", err.Error())
			continue
		}
		_ = s.repo.MarkSynced(ctx, "due_members", "due_member_id", memberID, "sheet", sheetRow)
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
		dateStr := sheetGet(r, "payment_date", "transaction_date", "tanggal")
		tgl, err := util.ParseSheetDate(dateStr)
		if err != nil {
			s.recordError(ctx, groupID, "due_payments", r["payment_id"], "sheet->db", "payment_date tidak valid: "+dateStr)
			continue
		}
		// Susulan: teks bulan sheet → rincian per bulan (bagi rata dari carryover_ir).
		// carryover_ir sheet TIDAK dipercaya mentah: total selalu dihitung ulang
		// dari rincian agar SUM(anak) = carryover_ir (aturan 000024).
		months, unknowns := util.SplitSheetMonths(sheetGet(r, "carryover_months", "susulan_bulan"))
		if len(unknowns) > 0 {
			s.recordError(ctx, groupID, "due_payments", r["payment_id"], "sheet->db", "bulan susulan tidak dikenal: "+strings.Join(unknowns, ", "))
			continue
		}
		sheetIR := parseNum(sheetGet(r, "carryover_ir", "susulan_ir"))
		carryItems, carryIR, notes := buildSheetCarryovers(months, sheetIR, sheetGet(r, "carryover_breakdown", "susulan_rincian"), sheetGet(r, "notes", "keterangan"))
		total := carryIR + parseNum(sheetGet(r, "connecting_fund", "uang_sambung")) +
			parseNum(sheetGet(r, "community_dues", "jimpitan")) + parseNum(sheetGet(r, "outreach_fund", "siar_siar")) +
			parseNum(sheetGet(r, "thousand_fund", "seribuan")) + parseNum(sheetGet(r, "funeral_fund", "kafan")) + parseNum(r["ukhro_mt"])
		status := strings.ToUpper(strings.TrimSpace(r["status"]))
		if status == "" {
			status = "ACTIVE"
		}
		gid := groupID
		p := &model.DuePayment{
			PaymentID: r["payment_id"], GroupID: &gid, MemberID: r["member_id"],
			PaymentDate: tgl, TotalAmount: total,
			CarryoverIR:    carryIR,
			Carryovers:     carryItems,
			ConnectingFund: parseNum(r["connecting_fund"]),
			CommunityDues:  parseNum(r["community_dues"]),
			OutreachFund:   parseNum(r["outreach_fund"]),
			ThousandFund:   parseNum(r["thousand_fund"]),
			FuneralFund:    parseNum(r["funeral_fund"]),
			UkhroMT:        parseNum(r["ukhro_mt"]),
			Notes:          notes, Status: status,
		}
		if err := s.repo.DuePaymentUpsert(ctx, p); err != nil {
			s.recordError(ctx, groupID, "due_payments", r["payment_id"], "sheet->db", err.Error())
			continue
		}
		if err := s.repo.ReplaceCarryovers(ctx, p.PaymentID, carryItems); err != nil {
			s.recordError(ctx, groupID, "due_payments", r["payment_id"], "sheet->db", "gagal simpan rincian susulan: "+err.Error())
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
	_, err = s.repo.ZakatStates(ctx, groupID)
	if err != nil {
		return err
	}
	tombs, _ := s.repo.Tombstones(ctx, groupID, "zakat")
	for i, r := range rows {
		if r["zakat_id"] == "" || tombs[r["zakat_id"]] {
			continue
		}
		sheetRow := i + 2
		// ZAKAT tab has no group_id column; single-group spreadsheet assumed.
		// Skip newerThan check because trigger auto-updates updated_at on every write.
		title := sheetGet(r, "title")
		muzakkiName := sheetGet(r, "muzakki_name", "nama")
		if strings.TrimSpace(title) == "" && strings.TrimSpace(muzakkiName) == "" {
			s.recordError(ctx, groupID, "zakat", r["zakat_id"], "sheet->db", "title/muzakki_name kosong")
			continue
		}
		if title == "" {
			title = muzakkiName
		}
		var tgl *time.Time
		if ds := sheetGet(r, "transaction_date", "tanggal"); ds != "" {
			if t, err := util.ParseSheetDate(strings.TrimSpace(ds)); err == nil {
				tgl = &t
			}
		}
		status := strings.ToUpper(strings.TrimSpace(r["status"]))
		if status == "" {
			status = "ACTIVE"
		}
		gid := groupID
		z := &model.ZakatRecord{
			ZakatID: r["zakat_id"], GroupID: &gid,
			Title: title, Description: sheetGet(r, "description", "notes", "keterangan"),
			Location: sheetGet(r, "location", "tempat"),
			SoulCount: atoi(sheetGet(r, "soul_count", "jumlah_anggota_keluarga"), 0),
			TotalRiceKg: parseNum(r["total_rice_kg"]), TotalMoneyRp: parseNum(sheetGet(r, "total_money_rp", "total_amount", "total", "nominal")),
			Status: status, TransactionDate: tgl,
		}
		if err := s.repo.ZakatUpsert(ctx, z); err != nil {
			s.recordError(ctx, groupID, "zakat", r["zakat_id"], "sheet->db", err.Error())
			continue
		}
		allocs := buildZakatAllocationsFromSheet(r, r["zakat_id"])
		if len(allocs) > 0 {
			if err := s.repo.ReplaceZakatAllocations(ctx, r["zakat_id"], allocs); err != nil {
				s.recordError(ctx, groupID, "zakat_allocations", r["zakat_id"], "sheet->db", err.Error())
			}
		}
		_ = s.repo.MarkSynced(ctx, "zakat_records", "zakat_id", r["zakat_id"], "sheet", sheetRow)
	}
	return nil
}

func buildZakatAllocationsFromSheet(r map[string]string, zakatID string) []model.ZakatAllocation {
	var allocs []model.ZakatAllocation
	fitrah := model.ZakatAllocation{
		ZakatID: zakatID, Category: "FITRAH",
		RecipientPercent:       atoi(sheetGet(r, "fitrah_mustahik_persen", "fitrah_mustahik_percent"), 0),
		RecipientAmount:        parseNum(sheetGet(r, "fitrah_mustahik_nominal", "fitrah_mustahik_amount")),
		RecipientGroupPercent:  atoi(sheetGet(r, "fitrah_amil_kelompok_persen", "fitrah_amil_kelompok_percent"), 0),
		RecipientGroupAmount:   parseNum(sheetGet(r, "fitrah_amil_kelompok_nominal", "fitrah_amil_kelompok_amount")),
		RecipientRegionPercent: atoi(sheetGet(r, "fitrah_amil_daerah_persen", "fitrah_amil_daerah_percent"), 0),
		RecipientRegionAmount:  parseNum(sheetGet(r, "fitrah_amil_daerah_nominal", "fitrah_amil_daerah_amount")),
		SabilillahPercent:      atoi(sheetGet(r, "fitrah_sabilillah_persen", "fitrah_sabilillah_percent"), 0),
		SabilillahAmount:       parseNum(sheetGet(r, "fitrah_sabilillah_nominal", "fitrah_sabilillah_amount")),
		AmilPercent:            atoi(sheetGet(r, "fitrah_amil_persen", "fitrah_amil_percent"), 0),
		AmilAmount:             parseNum(sheetGet(r, "fitrah_amil_nominal", "fitrah_amil_amount")),
		AmilGroupPercent:       atoi(sheetGet(r, "fitrah_amil_kelompok_persen", "fitrah_amil_kelompok_percent"), 0),
		AmilGroupAmount:        parseNum(sheetGet(r, "fitrah_amil_kelompok_nominal", "fitrah_amil_kelompok_amount")),
		AmilVillagePercent:     atoi(sheetGet(r, "fitrah_amil_desa_persen", "fitrah_amil_desa_percent"), 0),
		AmilVillageAmount:      parseNum(sheetGet(r, "fitrah_amil_desa_nominal", "fitrah_amil_desa_amount")),
		AmilRegionPercent:      atoi(sheetGet(r, "fitrah_amil_daerah_persen", "fitrah_amil_daerah_percent"), 0),
		AmilRegionAmount:       parseNum(sheetGet(r, "fitrah_amil_daerah_nominal", "fitrah_amil_daerah_amount")),
	}
	maal := model.ZakatAllocation{
		ZakatID: zakatID, Category: "MAAL",
		RecipientPercent:       atoi(sheetGet(r, "maal_mustahik_persen", "maal_mustahik_percent"), 0),
		RecipientAmount:        parseNum(sheetGet(r, "maal_mustahik_nominal", "maal_mustahik_amount")),
		RecipientGroupPercent:  atoi(sheetGet(r, "maal_mustahik_kelompok_persen", "maal_mustahik_kelompok_percent"), 0),
		RecipientGroupAmount:   parseNum(sheetGet(r, "maal_mustahik_kelompok_nominal", "maal_mustahik_kelompok_amount")),
		RecipientRegionPercent: atoi(sheetGet(r, "maal_mustahik_daerah_persen", "maal_mustahik_daerah_percent"), 0),
		RecipientRegionAmount:  parseNum(sheetGet(r, "maal_mustahik_daerah_nominal", "maal_mustahik_daerah_amount")),
		SabilillahPercent:      atoi(sheetGet(r, "maal_sabilillah_persen", "maal_sabilillah_percent"), 0),
		SabilillahAmount:       parseNum(sheetGet(r, "maal_sabilillah_nominal", "maal_sabilillah_amount")),
		AmilPercent:            atoi(sheetGet(r, "maal_amil_persen", "maal_amil_percent"), 0),
		AmilAmount:             parseNum(sheetGet(r, "maal_amil_nominal", "maal_amil_amount")),
		AmilGroupPercent:       atoi(sheetGet(r, "maal_amil_kelompok_persen", "maal_amil_kelompok_percent"), 0),
		AmilGroupAmount:        parseNum(sheetGet(r, "maal_amil_kelompok_nominal", "maal_amil_kelompok_amount")),
		AmilVillagePercent:     atoi(sheetGet(r, "maal_amil_desa_persen", "maal_amil_desa_percent"), 0),
		AmilVillageAmount:      parseNum(sheetGet(r, "maal_amil_desa_nominal", "maal_amil_desa_amount")),
		AmilRegionPercent:      atoi(sheetGet(r, "maal_amil_daerah_persen", "maal_amil_daerah_percent"), 0),
		AmilRegionAmount:       parseNum(sheetGet(r, "maal_amil_daerah_nominal", "maal_amil_daerah_amount")),
	}
	if hasNonZeroAllocation(fitrah) {
		allocs = append(allocs, fitrah)
	}
	if hasNonZeroAllocation(maal) {
		allocs = append(allocs, maal)
	}
	return allocs
}

func hasNonZeroAllocation(a model.ZakatAllocation) bool {
	return a.RecipientAmount != 0 || a.SabilillahAmount != 0 || a.AmilAmount != 0 ||
		a.RecipientGroupAmount != 0 || a.RecipientRegionAmount != 0 ||
		a.AmilGroupAmount != 0 || a.AmilVillageAmount != 0 || a.AmilRegionAmount != 0
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
	if err := s.mergeWriteTab(ctx, cli, sheetTabCash, sheetCashHeaders, groupID, cashRows); err != nil {
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
	if err := s.mergeWriteTab(ctx, cli, sheetTabDueMembers, sheetDueMemberHeaders, groupID, memberRows); err != nil {
		return err
	}
	paymentRows := make([][]interface{}, 0, len(dues.Payments))
	for _, p := range dues.Payments {
		months := make([]string, 0, len(p.CarryoverItems))
		for _, it := range p.CarryoverItems {
			months = append(months, it.Month)
		}
		paymentRows = append(paymentRows, []interface{}{
			p.PaymentID, groupID, p.MemberID, p.PaymentDate, p.TotalAmount,
			p.CarryoverIR, strings.Join(months, ", "),
			p.ConnectingFund, p.CommunityDues, p.OutreachFund, p.ThousandFund,
			p.FuneralFund, p.UkhroMT, p.Notes, p.Status, now,
		})
	}
	if err := s.mergeWriteTab(ctx, cli, sheetTabDuePayments, sheetDuePaymentHeaders, groupID, paymentRows); err != nil {
		return err
	}

	zakats, err := s.svc.ZakatList(ctx, groupID)
	if err != nil {
		return err
	}
	zakatRows := make([][]interface{}, 0, len(zakats))
	for _, z := range zakats {
		zakatRows = append(zakatRows, []interface{}{
			z.ZakatID, groupID, strings.Join(z.Categories, ","), z.Title, z.Description,
			z.Location, z.SoulCount, z.TotalRiceKg, z.TotalMoneyRp,
			z.Status, z.TransactionDate, z.CompletedAt, z.Version, now,
		})
	}
	if err := s.mergeWriteTab(ctx, cli, sheetTabZakat, sheetZakatHeaders, groupID, zakatRows); err != nil {
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

// buildSheetCarryovers mengubah kolom sheet (carryover_ir + teks bulan +
// sisa kolom breakdown lama) menjadi rincian per bulan + total + notes.
// Aturan: nominal dibagi rata ke tiap bulan (sisa ke bulan terakhir) agar
// SUM = carryover_ir. Isi breakdown lama yang masih ada disambung ke notes
// supaya tidak hilang.
func buildSheetCarryovers(months []string, sheetIR float64, breakdown, notes string) ([]model.DuePaymentCarryover, float64, string) {
	notes = strings.TrimSpace(notes)
	if b := strings.TrimSpace(breakdown); b != "" {
		if notes == "" {
			notes = "[Susulan lama] " + b
		} else {
			notes = notes + " | [Susulan lama] " + b
		}
	}
	if len(months) == 0 || sheetIR <= 0 {
		return nil, 0, notes
	}
	per := math.Floor(sheetIR/float64(len(months))*100) / 100
	out := make([]model.DuePaymentCarryover, 0, len(months))
	var acc float64
	for i, m := range months {
		amt := per
		if i == len(months)-1 {
			amt = sheetIR - acc
		}
		acc += amt
		out = append(out, model.DuePaymentCarryover{CarryoverID: util.NewID("CRY"), Month: m, Amount: amt})
	}
	return out, sheetIR, notes
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
