package service

import (
	"context"
	"strings"
	"time"

	"google.golang.org/api/sheets/v4"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/util"
)

// Sync tab zakat legacy kas-latukan-web (READ-ONLY, tidak pernah di-push balik):
//   - Zakat          -> zakat_records (header)
//   - Zakat_Muzaki   -> zakat_payers (anak, group by zakat_id)
//   - Zakat_Mustahik -> zakat_recipients (anak, group by zakat_id)
//   - Master_Muzaki / Master_Mustahik -> master_payers / master_recipients
//
// Tab legacy TANPA group_id: group diisi dari parameter (satu spreadsheet
// = satu group_label, ditentukan operator saat import/sync).
// Semua pembacaan toleran header lama/baru via sheetGet; tidak ada tulis
// ke tab legacy dalam file ini.

// normZakatCategory didelegasikan ke util agar ejaan seragam (kanonis MAL).
// "ZAKAT MAAL" -> "MAAL" -> "MAL"; "TERNAK" -> "LIVESTOCK".
func normZakatCategory(v string) string {
	return util.NormZakatCategory(v)
}

func (s *FinanceSyncService) pullLegacyMasters(ctx context.Context, cli *sheets.Service, groupID string) error {
	if rows, err := s.readTab(ctx, cli, sheetTabLegacyMasterMuzaki); err == nil {
		for _, r := range rows {
			name := sheetGet(r, "muzakki_name", "name", "full_name", "member_name", "nama", "nama_lengkap")
			if strings.TrimSpace(name) == "" {
				continue
			}
			_, _ = s.repo.MasterUpsert(ctx, "payer", groupID, strings.TrimSpace(name))
		}
	}
	if rows, err := s.readTab(ctx, cli, sheetTabLegacyMasterMustahik); err == nil {
		for _, r := range rows {
			name := sheetGet(r, "mustahik_name", "name", "full_name", "member_name", "nama", "nama_lengkap")
			if strings.TrimSpace(name) == "" {
				continue
			}
			_, _ = s.repo.MasterUpsert(ctx, "recipient", groupID, strings.TrimSpace(name))
		}
	}
	return nil
}

func (s *FinanceSyncService) pullLegacyZakat(ctx context.Context, cli *sheets.Service, groupID string) error {
	// Master dulu agar MasterUpsert idempotent berjalan sebelum anak me-referensi name.
	_ = s.pullLegacyMasters(ctx, cli, groupID)

	headerRows, err := s.readTab(ctx, cli, sheetTabLegacyZakat)
	if err != nil {
		return err
	}
	muzakiRows, _ := s.readTab(ctx, cli, sheetTabLegacyZakatMuzaki)
	mustahikRows, _ := s.readTab(ctx, cli, sheetTabLegacyZakatMustahik)

	// Kelompokkan anak per zakat_id (abaikan DELETED).
	payersByZakat := map[string][]model.ZakatPayer{}
	for _, r := range muzakiRows {
		zid := strings.TrimSpace(sheetGet(r, "zakat_id", "id"))
		if zid == "" {
			continue
		}
		if strings.ToUpper(strings.TrimSpace(r["status"])) == "DELETED" {
			continue
		}
		name := sheetGet(r, "muzakki_name", "name", "full_name", "nama", "nama_lengkap")
		if strings.TrimSpace(name) == "" {
			continue
		}
		mid, _ := s.repo.MasterUpsert(ctx, "payer", groupID, strings.TrimSpace(name))
		var midPtr *string
		if mid != "" {
			midPtr = &mid
		}
		payersByZakat[zid] = append(payersByZakat[zid], model.ZakatPayer{
			MasterID: midPtr, Name: strings.TrimSpace(name),
			Amount:             parseNum(sheetGet(r, "amount", "nominal")),
			ZakatCategory:      normZakatCategory(sheetGet(r, "zakat_type", "jenis_zakat")),
			FamilyMembersCount: atoi(sheetGet(r, "soul_count", "jumlah_anggota_keluarga"), 0),
		})
	}
	recipsByZakat := map[string][]model.ZakatRecipient{}
	for _, r := range mustahikRows {
		zid := strings.TrimSpace(sheetGet(r, "zakat_id", "id"))
		if zid == "" {
			continue
		}
		if strings.ToUpper(strings.TrimSpace(r["status"])) == "DELETED" {
			continue
		}
		name := sheetGet(r, "mustahik_name", "name", "full_name", "nama", "nama_lengkap")
		if strings.TrimSpace(name) == "" {
			continue
		}
		mid, _ := s.repo.MasterUpsert(ctx, "recipient", groupID, strings.TrimSpace(name))
		var midPtr *string
		if mid != "" {
			midPtr = &mid
		}
		recipsByZakat[zid] = append(recipsByZakat[zid], model.ZakatRecipient{
			MasterID: midPtr, Name: strings.TrimSpace(name),
			Amount:        parseNum(sheetGet(r, "amount", "nominal")),
			ZakatCategory: normZakatCategory(sheetGet(r, "zakat_type", "jenis_zakat")),
		})
	}

	tombs, _ := s.repo.Tombstones(ctx, groupID, "zakat")
	for _, r := range headerRows {
		zid := strings.TrimSpace(sheetGet(r, "zakat_id", "id"))
		if zid == "" || tombs[zid] {
			continue
		}
		title := sheetGet(r, "title")
		muzakkiFallback := ""
		if len(payersByZakat[zid]) > 0 {
			muzakkiFallback = payersByZakat[zid][0].Name
		}
		if title == "" {
			title = muzakkiFallback
		}
		if title == "" {
			continue
		}
		var tgl *time.Time
		if ds := sheetGet(r, "transaction_date", "date", "tanggal"); ds != "" {
			if t, err := util.ParseSheetDate(strings.TrimSpace(ds)); err == nil {
				tgl = &t
			}
		}
		status := strings.ToUpper(strings.TrimSpace(r["status"]))
		if status == "" {
			status = "ACTIVE"
		}
		if status == "DELETED" {
			continue
		}
		gid := groupID
		total := parseNum(sheetGet(r, "total_amount", "total"))
		if total == 0 {
			for _, p := range payersByZakat[zid] {
				total += p.Amount
			}
		}
		z := &model.ZakatRecord{
			ZakatID: zid, GroupID: &gid, Title: title,
			Description:  sheetGet(r, "description", "notes", "keterangan"),
			Location:     sheetGet(r, "location", "tempat"),
			TotalMoneyRp: total,
			Status:       status, TransactionDate: tgl,
		}
		if err := s.repo.ZakatUpsert(ctx, z); err != nil {
			s.recordError(ctx, groupID, "zakat", zid, "legacy->db", err.Error())
			continue
		}
		if items, ok := payersByZakat[zid]; ok && len(items) > 0 {
			if err := s.repo.ReplaceZakatPayers(ctx, zid, items); err != nil {
				s.recordError(ctx, groupID, "zakat_payers", zid, "legacy->db", err.Error())
			}
		}
		if items, ok := recipsByZakat[zid]; ok && len(items) > 0 {
			if err := s.repo.ReplaceZakatRecipients(ctx, zid, items); err != nil {
				s.recordError(ctx, groupID, "zakat_recipients", zid, "legacy->db", err.Error())
			}
		}
		_ = s.repo.MarkSynced(ctx, "zakat_records", "zakat_id", zid, "legacy", 0)
	}
	return nil
}
