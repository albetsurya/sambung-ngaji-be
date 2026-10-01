package service

import (
	"context"
	"strings"

	"google.golang.org/api/sheets/v4"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/util"
)

// Sync tab legacy bendahara (READ-ONLY, tidak pernah di-push balik):
// bendahara tetap mengisi sheet ini; aplikasi hanya menarik.
//   - Shodaqoh_Members  -> due_members (upsert)
//   - Shodaqoh_Payments -> due_payments + due_payment_carryovers
//
// Perbedaan format legacy vs tab app (DUE_*):
//   - Header: transaction_date (bukan payment_date), full_name, kas_transaction_no.
//   - Nominal format Rupiah Indonesia: "Rp1.286.000", "Rp -".
//   - carryover_months: "YYYY-MM" (kadang "2026-1", multi "2025-08,2025-09").
//   - carryover_breakdown: JSON map {"2025-08":83333} atau "{}" (=kosong).
//   - total_amount sheet DIPERTAHANKAN apa adanya (tidak dihitung ulang).
//   - status ACTIVE/INACTIVE dipertahankan (INACTIVE = riwayat nonaktif).
//   - Tab legacy TIDAK punya group_id: group diisi dari parameter (satu
//     sheet = satu kelompok, ditentukan operator saat import/sync).
const (
	sheetTabLegacyMembers  = "Shodaqoh_Members"
	sheetTabLegacyPayments = "Shodaqoh_Payments"
)

func (s *FinanceSyncService) pullLegacyMembers(ctx context.Context, cli *sheets.Service, groupID string) error {
	rows, err := s.readTab(ctx, cli, sheetTabLegacyMembers)
	if err != nil {
		return err
	}
	states, err := s.repo.DueMemberStates(ctx, groupID)
	if err != nil {
		return err
	}
	tombs, _ := s.repo.Tombstones(ctx, groupID, "due_members")
	for i, r := range rows {
		if r["member_id"] == "" || tombs[r["member_id"]] {
			continue
		}
		sheetRow := i + 2
		dbTS, exists := states[r["member_id"]]
		if exists && !newerThan(r["updated_at"], dbTS) {
			continue
		}
		if strings.TrimSpace(r["member_name"]) == "" {
			s.recordError(ctx, groupID, "due_members", r["member_id"], "legacy->db", "member_name kosong")
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
			s.recordError(ctx, groupID, "due_members", r["member_id"], "legacy->db", err.Error())
			continue
		}
		_ = s.repo.MarkSynced(ctx, "due_members", "due_member_id", r["member_id"], "legacy", sheetRow)
	}
	return nil
}

func (s *FinanceSyncService) pullLegacyPayments(ctx context.Context, cli *sheets.Service, groupID string) error {
	rows, err := s.readTab(ctx, cli, sheetTabLegacyPayments)
	if err != nil {
		return err
	}
	states, err := s.repo.DuePaymentStates(ctx, groupID)
	if err != nil {
		return err
	}
	tombs, _ := s.repo.Tombstones(ctx, groupID, "due_payments")
	for i, r := range rows {
		if r["payment_id"] == "" || r["member_id"] == "" || tombs[r["payment_id"]] {
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
			s.recordError(ctx, groupID, "due_payments", r["payment_id"], "legacy->db", "transaction_date tidak valid: "+dateStr)
			continue
		}
		// Rincian susulan: breakdown JSON (eksak) > teks bulan (bagi rata).
		var carryItems []model.DuePaymentCarryover
		var carryIR float64
		if items, sum, ok := util.ParseCarryoverBreakdown(r["carryover_breakdown"]); ok {
			for _, it := range items {
				carryItems = append(carryItems, model.DuePaymentCarryover{
					CarryoverID: util.NewID("CRY"), Month: it.Month, Amount: it.Amount,
				})
			}
			carryIR = sum
		} else {
			months, unknowns := util.SplitSheetMonths(r["carryover_months"])
			if len(unknowns) > 0 {
				s.recordError(ctx, groupID, "due_payments", r["payment_id"], "legacy->db", "bulan susulan tidak dikenal: "+strings.Join(unknowns, ", "))
				continue
			}
			sheetIR := parseNum(r["carryover_ir"])
			carryItems, carryIR, _ = buildSheetCarryovers(months, sheetIR, "", "")
		}
		// Total legacy DIPERTAHANKAN dari sheet (tidak dihitung ulang):
		// total sheet adalah angka setor riil, bukan sekadar jumlah kolom.
		total := parseNum(r["total_amount"])
		status := strings.ToUpper(strings.TrimSpace(r["status"]))
		if status == "" {
			status = "ACTIVE"
		}
		notes := strings.TrimSpace(r["notes"])
		if b := strings.TrimSpace(r["carryover_breakdown"]); b != "" && b != "{}" {
			if _, _, ok := util.ParseCarryoverBreakdown(b); !ok {
				if notes == "" {
					notes = "[Rincian lama] " + b
				} else {
					notes = notes + " | [Rincian lama] " + b
				}
			}
		}
		gid := groupID
		createdBy := strings.TrimSpace(r["created_by"])
		if createdBy == "" {
			createdBy = "legacy-sheet"
		}
		p := &model.DuePayment{
			PaymentID: r["payment_id"], GroupID: &gid, MemberID: r["member_id"],
			PaymentDate: tgl, TotalAmount: total,
			CarryoverIR: carryIR, Carryovers: carryItems,
			ConnectingFund: parseNum(r["connecting_fund"]),
			CommunityDues:  parseNum(r["community_dues"]),
			OutreachFund:   parseNum(r["outreach_fund"]),
			ThousandFund:   parseNum(r["thousand_fund"]),
			FuneralFund:    parseNum(r["funeral_fund"]),
			UkhroMT:        parseNum(r["ukhro_mt"]),
			Notes:          notes, Status: status, CreatedBy: createdBy,
		}
		if err := s.repo.DuePaymentUpsert(ctx, p); err != nil {
			s.recordError(ctx, groupID, "due_payments", r["payment_id"], "legacy->db", err.Error())
			continue
		}
		if err := s.repo.ReplaceCarryovers(ctx, p.PaymentID, carryItems); err != nil {
			s.recordError(ctx, groupID, "due_payments", r["payment_id"], "legacy->db", "gagal simpan rincian susulan: "+err.Error())
			continue
		}
		_ = s.repo.MarkSynced(ctx, "due_payments", "payment_id", r["payment_id"], "legacy", sheetRow)
	}
	return nil
}
