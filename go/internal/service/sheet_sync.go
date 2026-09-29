package service

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

const (
	sheetTabCash        = "CASH"
	sheetTabDueMembers  = "DUE_MEMBERS"
	sheetTabDuePayments = "DUE_PAYMENTS"
	sheetTabZakat       = "ZAKAT"
)

var (
	sheetCashHeaders       = []string{"cash_id", "group_id", "cash_type", "tanggal", "account_name", "description", "debit", "credit", "created_by", "updated_at"}
	sheetDueMemberHeaders  = []string{"member_id", "group_id", "member_name", "monthly_target", "status", "updated_at"}
	sheetDuePaymentHeaders = []string{"payment_id", "group_id", "member_id", "payment_date", "total_amount", "carryover_ir", "carryover_months", "connecting_fund", "community_dues", "outreach_fund", "thousand_fund", "funeral_fund", "ukhro_mt", "notes", "status", "updated_at"}
	sheetZakatHeaders      = []string{"zakat_id", "group_id", "zakat_type", "muzakki_name", "soul_count", "total_rice_kg", "total_money_rp", "status", "transaction_date", "updated_at"}
)

type FinanceSyncService struct {
	repo      *repository.FinanceRepo
	groups    *repository.GroupRepo
	svc       *FinanceService
	sheetID   string
	credsJSON string
	// mu menyerikan seluruh operasi sync: mencegah dua SyncGroup/PullGroup
	// berjalan bersamaan (ketuk tombol berkali-kali + auto-sync AfterWrite),
	// yang bisa membuat tulis-baca tab saling timpa.
	mu sync.Mutex
}

func NewFinanceSyncService(repo *repository.FinanceRepo, groups *repository.GroupRepo, svc *FinanceService) *FinanceSyncService {
	return &FinanceSyncService{
		repo:      repo,
		groups:    groups,
		svc:       svc,
		sheetID:   os.Getenv("FINANCE_SPREADSHEET_ID"),
		credsJSON: os.Getenv("GOOGLE_SHEETS_CREDENTIALS"),
	}
}

func (s *FinanceSyncService) configured() bool {
	return s.sheetID != "" && s.credsJSON != ""
}

func (s *FinanceSyncService) IsConfigured() bool {
	return s.configured()
}

func (s *FinanceSyncService) client(ctx context.Context) (*sheets.Service, error) {
	if !s.configured() {
		return nil, fmt.Errorf("sheet sync belum dikonfigurasi (FINANCE_SPREADSHEET_ID / GOOGLE_SHEETS_CREDENTIALS)")
	}
	return sheets.NewService(ctx, option.WithCredentialsJSON([]byte(s.credsJSON)))
}

func sheetRows(values [][]interface{}) ([]map[string]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	headers := make([]string, len(values[0]))
	for i, h := range values[0] {
		headers[i] = strings.TrimSpace(fmt.Sprint(h))
	}
	var out []map[string]string
	for _, r := range values[1:] {
		m := map[string]string{}
		for i, h := range headers {
			if h == "" {
				continue
			}
			if i < len(r) {
				m[h] = strings.TrimSpace(fmt.Sprint(r[i]))
			}
		}
		if m["group_id"] == "" && firstCol(m) == "" {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}

func firstCol(m map[string]string) string {
	for _, v := range m {
		if v != "" {
			return v
		}
	}
	return ""
}

func parseSheetTime(v string) time.Time {
	// Kolom tanggal sheet (DATE) harus dibaca sebagai kalender WIB,
	// bukan instant UTC — kalau tidak, 1 Agu 00:00 WIB (31 Jul 17:00 UTC)
	// mundur 1 hari. Lihat util.ParseSheetDate.
	if t, err := util.ParseSheetDate(v); err == nil {
		return t
	}
	// Kolom DATETIME sheet (updated_at/created_at): "21/08/2026 18:54:48"
	// (DD/MM/YYYY, zona WIB) atau sisa format JS Date.toString().
	s := strings.TrimSpace(v)
	if i := strings.Index(s, " ("); i != -1 {
		s = s[:i]
	}
	for _, layout := range []string{
		"02/01/2006 15:04:05",
		"02/01/2006",
		"Mon Jan 02 2006 15:04:05 GMT-0700",
		"Mon Jan 2 2006 15:04:05 GMT-0700",
		time.RFC3339, "2006-01-02T15:04:05Z07:00", "2006-01-02 15:04:05", "2006-01-02",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func parseNum(v string) float64 {
	return util.ParseRpNumber(v)
}

func (s *FinanceSyncService) ensureTabs(ctx context.Context, cli *sheets.Service) error {
	meta, err := cli.Spreadsheets.Get(s.sheetID).Fields("sheets.properties.title").Context(ctx).Do()
	if err != nil {
		return err
	}
	have := map[string]bool{}
	for _, sh := range meta.Sheets {
		if sh.Properties != nil {
			have[sh.Properties.Title] = true
		}
	}
	for _, tab := range []string{sheetTabCash, sheetTabDueMembers, sheetTabDuePayments, sheetTabZakat} {
		if have[tab] {
			continue
		}
		_, err = cli.Spreadsheets.BatchUpdate(s.sheetID, &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{{
				AddSheet: &sheets.AddSheetRequest{
					Properties: &sheets.SheetProperties{Title: tab},
				},
			}},
		}).Context(ctx).Do()
		if err != nil && strings.Contains(err.Error(), "already exists") {
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *FinanceSyncService) readTab(ctx context.Context, cli *sheets.Service, tab string) ([]map[string]string, error) {
	resp, err := cli.Spreadsheets.Values.Get(s.sheetID, tab+"!A:ZZ").Context(ctx).Do()
	if err != nil {
		if strings.Contains(err.Error(), "Unable to parse range") {
			return nil, nil
		}
		return nil, err
	}
	var values [][]interface{}
	for _, r := range resp.Values {
		values = append(values, r)
	}
	return sheetRows(values)
}

func (s *FinanceSyncService) writeTab(ctx context.Context, cli *sheets.Service, tab string, headers []string, rows [][]interface{}) error {
	data := make([][]interface{}, 0, len(rows)+1)
	hdr := make([]interface{}, len(headers))
	for i, h := range headers {
		hdr[i] = h
	}
	data = append(data, hdr)
	data = append(data, rows...)
	if _, err := cli.Spreadsheets.Values.Clear(s.sheetID, tab+"!A:ZZ", &sheets.ClearValuesRequest{}).Context(ctx).Do(); err != nil {
		return err
	}
	_, err := cli.Spreadsheets.Values.Update(s.sheetID, tab+"!A1", &sheets.ValueRange{Values: data}).
		ValueInputOption("RAW").Context(ctx).Do()
	return err
}

// mergeWriteTab menulis ulang tab dengan strategi merge per-grup: baris
// milik grup lain dipertahankan apa adanya, baris grup ini diganti total
// dari DB. Tanpa ini, SyncGroup satu grup akan MENGHAPUS baris mirror
// grup lain (writeTab me-clear seluruh tab).
func (s *FinanceSyncService) mergeWriteTab(ctx context.Context, cli *sheets.Service, tab string, headers []string, groupID string, freshRows [][]interface{}) error {
	existing, err := s.readTab(ctx, cli, tab)
	if err != nil {
		return err
	}
	merged := make([][]interface{}, 0, len(existing)+len(freshRows))
	for _, m := range existing {
		if strings.TrimSpace(m["group_id"]) == groupID {
			continue
		}
		row := make([]interface{}, len(headers))
		for i, h := range headers {
			row[i] = m[h]
		}
		merged = append(merged, row)
	}
	merged = append(merged, freshRows...)
	return s.writeTab(ctx, cli, tab, headers, merged)
}

func (s *FinanceSyncService) recordError(ctx context.Context, groupID, entity, entityID, direction, msg string) {
	_ = s.repo.SyncError(ctx, groupID, entity, entityID, direction, msg)
}

// PullGroup menarik sheet -> DB saja tanpa push balik.
// Dipakai operator untuk re-sync aman: verifikasi dulu, push belakangan.
// Urutan: tab legacy bendahara dulu (sumber utama), lalu tab app.
func (s *FinanceSyncService) PullGroup(ctx context.Context, groupID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cli, err := s.client(ctx)
	if err != nil {
		return err
	}
	if err := s.ensureTabs(ctx, cli); err != nil {
		return err
	}
	if err := s.pullLegacyMembers(ctx, cli, groupID); err != nil {
		s.recordError(ctx, groupID, "due_members", "", "legacy->db", err.Error())
	}
	if err := s.pullLegacyPayments(ctx, cli, groupID); err != nil {
		s.recordError(ctx, groupID, "due_payments", "", "legacy->db", err.Error())
	}
	if err := s.pullLegacyCash(ctx, cli, groupID); err != nil {
		s.recordError(ctx, groupID, "cash", "", "legacy->db", err.Error())
	}
	if err := s.pullCash(ctx, cli, groupID); err != nil {
		s.recordError(ctx, groupID, "cash", "", "sheet->db", err.Error())
	}
	if err := s.pullDueMembers(ctx, cli, groupID); err != nil {
		s.recordError(ctx, groupID, "due_members", "", "sheet->db", err.Error())
	}
	if err := s.pullDuePayments(ctx, cli, groupID); err != nil {
		s.recordError(ctx, groupID, "due_payments", "", "sheet->db", err.Error())
	}
	if err := s.pullZakat(ctx, cli, groupID); err != nil {
		s.recordError(ctx, groupID, "zakat", "", "sheet->db", err.Error())
	}
	return nil
}

// SyncGroup merges one group both ways, then rewrites the four tabs from DB state.
func (s *FinanceSyncService) SyncGroup(ctx context.Context, groupID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cli, err := s.client(ctx)
	if err != nil {
		return err
	}
	if err := s.ensureTabs(ctx, cli); err != nil {
		return err
	}
	if err := s.pullLegacyMembers(ctx, cli, groupID); err != nil {
		s.recordError(ctx, groupID, "due_members", "", "legacy->db", err.Error())
	}
	if err := s.pullLegacyPayments(ctx, cli, groupID); err != nil {
		s.recordError(ctx, groupID, "due_payments", "", "legacy->db", err.Error())
	}
	if err := s.pullLegacyCash(ctx, cli, groupID); err != nil {
		s.recordError(ctx, groupID, "cash", "", "legacy->db", err.Error())
	}
	if err := s.pullCash(ctx, cli, groupID); err != nil {
		s.recordError(ctx, groupID, "cash", "", "sheet->db", err.Error())
	}
	if err := s.pullDueMembers(ctx, cli, groupID); err != nil {
		s.recordError(ctx, groupID, "due_members", "", "sheet->db", err.Error())
	}
	if err := s.pullDuePayments(ctx, cli, groupID); err != nil {
		s.recordError(ctx, groupID, "due_payments", "", "sheet->db", err.Error())
	}
	if err := s.pullZakat(ctx, cli, groupID); err != nil {
		s.recordError(ctx, groupID, "zakat", "", "sheet->db", err.Error())
	}
	if err := s.pushAll(ctx, cli, groupID); err != nil {
		s.recordError(ctx, groupID, "tabs", "", "db->sheet", err.Error())
		return err
	}
	return nil
}

func (s *FinanceSyncService) SyncAllGroups(ctx context.Context) error {
	groups, err := s.groups.FindAll(ctx, false)
	if err != nil {
		return err
	}
	var firstErr error
	for _, g := range groups {
		if err := s.SyncGroup(ctx, g.GroupID); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
