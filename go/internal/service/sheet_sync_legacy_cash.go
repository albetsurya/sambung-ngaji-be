package service

import (
	"context"
	"strings"

	"google.golang.org/api/sheets/v4"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/util"
)

// Sync tab kas legacy bendahara (READ-ONLY, tidak pernah di-push balik):
//   - Transaksi -> cash_transactions (cash_type=main)
//   - Kas_Amil  -> cash_transactions (cash_type=amil)
//
// Format legacy: date ISO atau DD/MM/YYYY, nominal "Rp840.000,00",
// kolom balance DIABAIKAN (saldo berjalan dihitung aplikasi).
// cash_id deterministik ("TRX<n>" / "AML<n>") agar pull idempotent.
// Tab legacy TANPA group_id: group diisi operator (satu sheet = satu group_label).
const (
	sheetTabLegacyCashMain = "Transaksi"
	sheetTabLegacyCashAmil = "Kas_Amil"
)

func (s *FinanceSyncService) pullLegacyCashTab(ctx context.Context, cli *sheets.Service, groupID, tab, idPrefix, cashType string) error {
	rows, err := s.readTab(ctx, cli, tab)
	if err != nil {
		return err
	}
	states, err := s.repo.CashStates(ctx, groupID)
	if err != nil {
		return err
	}
	tombs, _ := s.repo.Tombstones(ctx, groupID, "cash")
	for i, r := range rows {
		if r["transaction_id"] == "" || tombs[idPrefix+r["transaction_id"]] {
			continue
		}
		cashID := idPrefix + strings.TrimSpace(r["transaction_id"])
		sheetRow := i + 2
		dbTS, exists := states[cashID]
		if exists && !newerThan(r["updated_at"], dbTS) {
			continue
		}
		dateStr := sheetGet(r, "transaction_date", "date", "tanggal")
		tgl, err := util.ParseSheetDate(dateStr)
		if err != nil {
			s.recordError(ctx, groupID, "cash", cashID, "legacy->db", "transaction_tanggal tidak valid: "+dateStr)
			continue
		}
		accountName := sheetGet(r, "account_name", "account")
		if strings.TrimSpace(accountName) == "" {
			s.recordError(ctx, groupID, "cash", cashID, "legacy->db", "account_name kosong")
			continue
		}
		createdBy := strings.TrimSpace(r["created_by"])
		if createdBy == "" {
			createdBy = "legacy-sheet"
		}
		gid := groupID
		k := &model.CashTransaction{
			CashID: cashID, GroupID: &gid, CashType: cashType,
			Date: tgl, AccountName: strings.TrimSpace(accountName),
			Description: sheetGet(r, "description", "notes", "keterangan"),
			Debit:       parseNum(sheetGet(r, "debit", "debet")), Credit: parseNum(sheetGet(r, "credit", "kredit")),
			CreatedBy: createdBy,
		}
		if !exists {
			if err := s.repo.CashInsert(ctx, k); err != nil {
				s.recordError(ctx, groupID, "cash", cashID, "legacy->db", err.Error())
				continue
			}
		} else if err := s.repo.CashUpdate(ctx, k); err != nil {
			s.recordError(ctx, groupID, "cash", cashID, "legacy->db", err.Error())
			continue
		}
		_ = s.repo.MarkSynced(ctx, "cash_transactions", "cash_id", cashID, "legacy", sheetRow)
	}
	return nil
}

func (s *FinanceSyncService) pullLegacyCash(ctx context.Context, cli *sheets.Service, groupID string) error {
	if err := s.pullLegacyCashTab(ctx, cli, groupID, sheetTabLegacyCashMain, "TRX", "main"); err != nil {
		return err
	}
	return s.pullLegacyCashTab(ctx, cli, groupID, sheetTabLegacyCashAmil, "AML", "amil")
}
