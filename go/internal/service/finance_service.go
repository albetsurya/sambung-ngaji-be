package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

type FinanceService struct {
	repo *repository.FinanceRepo
	// AfterWrite fires after each successful write (e.g. async sheet push).
	AfterWrite func(groupID string)
}

func (s *FinanceService) notify(groupID string) {
	if s.AfterWrite != nil && groupID != "" {
		go s.AfterWrite(groupID)
	}
}

func NewFinanceService(repo *repository.FinanceRepo) *FinanceService {
	return &FinanceService{repo: repo}
}

func normCashType(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "amil", "kas_amil":
		return "amil"
	default:
		return "main"
	}
}

func toCashDTO(no int, k model.CashTransaction, balance float64) model.CashTransactionDTO {
	gid := ""
	if k.GroupID != nil {
		gid = *k.GroupID
	}
	return model.CashTransactionDTO{
		No:              no,
		CashID:          k.CashID,
		GroupID:         gid,
		CashType:        k.CashType,
		TransactionDate: k.Date.Format("2006-01-02"),
		AccountName:     k.AccountName,
		Description:     k.Description,
		Debit:           k.Debit,
		Credit:          k.Credit,
		Balance:         balance,
		CreatedBy:       k.CreatedBy,
		UpdatedAt:       k.UpdatedAt.Format(time.RFC3339),
	}
}

func isOpeningBalance(account string) bool {
	return strings.TrimSpace(strings.ToUpper(account)) == "SALDO AWAL"
}

func indonesianMonthLabel(monthKey string) string {
	names := []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	if len(monthKey) != 7 {
		return monthKey
	}
	var y, m int
	if _, err := fmt.Sscanf(monthKey, "%d-%d", &y, &m); err != nil || m < 1 || m > 12 {
		return monthKey
	}
	return fmt.Sprintf("%s %d", names[m], y)
}

func (s *FinanceService) CashList(ctx context.Context, groupID, cashType string) (*model.CashSummaryDTO, error) {
	cashType = normCashType(cashType)
	rows, err := s.repo.CashList(ctx, groupID, cashType)
	if err != nil {
		return nil, err
	}
	items := make([]model.CashTransactionDTO, 0, len(rows))
	var totalDebit, totalCredit, balance float64
	initial := 0.0
	initialSet := false
	for i, k := range rows {
		if isOpeningBalance(k.AccountName) {
			balance = k.Debit - k.Credit
			if !initialSet {
				initial = balance
				initialSet = true
			}
		} else {
			totalDebit += k.Debit
			totalCredit += k.Credit
			balance += k.Debit - k.Credit
		}
		items = append(items, toCashDTO(i+1, k, balance))
	}
	ending := balance
	if len(rows) == 0 {
		ending = initial
	}
	return &model.CashSummaryDTO{
		CashType:       cashType,
		Transactions:   items,
		InitialBalance: initial,
		TotalDebit:     totalDebit,
		TotalCredit:    totalCredit,
		EndingBalance:  ending,
	}, nil
}

type CashSaveInput struct {
	CashID      string
	GroupID     string
	CashType    string
	Date        string
	AccountName string
	Description string
	Debit       float64
	Credit      float64
	CreatedBy   string
}

func (s *FinanceService) CashSave(ctx context.Context, in CashSaveInput) (*model.CashTransactionDTO, error) {
	if in.GroupID == "" {
		return nil, errors.New("group_id wajib diisi")
	}
	if in.Description == "" {
		return nil, errors.New("keterangan wajib diisi")
	}
	tgl, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		return nil, errors.New("tanggal tidak valid (YYYY-MM-DD)")
	}
	if in.Debit < 0 || in.Credit < 0 {
		return nil, errors.New("nominal tidak boleh negatif")
	}
	gid := in.GroupID
	k := &model.CashTransaction{
		CashID:      in.CashID,
		GroupID:     &gid,
		CashType:    normCashType(in.CashType),
		Date:        tgl,
		AccountName: in.AccountName,
		Description: in.Description,
		Debit:       in.Debit,
		Credit:      in.Credit,
		CreatedBy:   in.CreatedBy,
	}
	if k.CashID == "" {
		k.CashID = util.NewID("KAS")
		if err := s.repo.CashInsert(ctx, k); err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.CashUpdate(ctx, k); err != nil {
			return nil, err
		}
	}
	_ = s.repo.MarkSynced(ctx, "cash_transactions", "cash_id", k.CashID, "app", 0)
	sum, err := s.CashList(ctx, in.GroupID, k.CashType)
	if err != nil {
		return nil, err
	}
	for _, it := range sum.Transactions {
		if it.CashID == k.CashID {
			s.notify(in.GroupID)
			return &it, nil
		}
	}
	return &model.CashTransactionDTO{CashID: k.CashID, GroupID: in.GroupID}, nil
}

func (s *FinanceService) CashDelete(ctx context.Context, groupID, kasID string) error {
	if kasID == "" {
		return errors.New("cash_id wajib diisi")
	}
	if err := s.repo.CashDelete(ctx, groupID, kasID); err != nil {
		return err
	}
	_ = s.repo.Tombstone(ctx, groupID, "cash", kasID)
	s.notify(groupID)
	return nil
}

func (s *FinanceService) CashDuplicate(ctx context.Context, groupID, cashType, kasID, createdBy string) (*model.CashTransactionDTO, error) {
	rows, err := s.repo.CashList(ctx, groupID, normCashType(cashType))
	if err != nil {
		return nil, err
	}
	for _, k := range rows {
		if k.CashID == kasID {
			gid := groupID
			cp := &model.CashTransaction{
				CashID:      util.NewID("KAS"),
				GroupID:     &gid,
				CashType:    k.CashType,
				Date:        k.Date,
				AccountName: k.AccountName,
				Description: k.Description,
				Debit:       k.Debit,
				Credit:      k.Credit,
				CreatedBy:   createdBy,
			}
			if err := s.repo.CashInsert(ctx, cp); err != nil {
				return nil, err
			}
			return s.CashSave(ctx, CashSaveInput{
				CashID: cp.CashID, GroupID: groupID, CashType: cp.CashType,
				Date:        cp.Date.Format("2006-01-02"),
				AccountName: cp.AccountName, Description: cp.Description,
				Debit: cp.Debit, Credit: cp.Credit, CreatedBy: createdBy,
			})
		}
	}
	return nil, errors.New("transaksi tidak ditemukan")
}

func (s *FinanceService) CashCarryForward(ctx context.Context, groupID, cashType, monthKey, createdBy string) (*model.CashTransactionDTO, error) {
	cashType = normCashType(cashType)
	if len(monthKey) != 7 {
		return nil, errors.New("Bulan tidak valid. Gunakan format YYYY-MM.")
	}
	next, err := time.Parse("2006-01", monthKey)
	if err != nil || next.Month() < 1 || next.Month() > 12 {
		return nil, errors.New("Bulan tidak valid. Gunakan format YYYY-MM.")
	}
	rows, err := s.repo.CashList(ctx, groupID, cashType)
	if err != nil {
		return nil, err
	}
	var balance float64
	monthEnding := 0.0
	foundMonth := false
	for _, k := range rows {
		if isOpeningBalance(k.AccountName) {
			balance = k.Debit - k.Credit
		} else {
			balance += k.Debit - k.Credit
		}
		if k.Date.Format("2006-01") == monthKey {
			monthEnding = balance
			foundMonth = true
		}
	}
	if !foundMonth {
		return nil, errors.New("Tidak ditemukan transaksi pada bulan " + monthKey + ".")
	}
	ending := monthEnding
	next = next.AddDate(0, 1, 0)
	nextKey := next.Format("2006-01")
	for _, k := range rows {
		if isOpeningBalance(k.AccountName) && k.Date.Format("2006-01") == nextKey {
			return nil, errors.New("SALDO AWAL untuk " + nextKey + " sudah ada.")
		}
	}
	gid := groupID
	k := &model.CashTransaction{
		CashID:      util.NewID("KAS"),
		GroupID:     &gid,
		CashType:    cashType,
		Date:        time.Date(next.Year(), next.Month(), 1, 0, 0, 0, 0, time.UTC),
		AccountName: "SALDO AWAL",
		Description: "Saldo awal dari " + indonesianMonthLabel(monthKey),
		Debit:       ending,
		Credit:      0,
		CreatedBy:   createdBy,
	}
	if err := s.repo.CashInsert(ctx, k); err != nil {
		return nil, err
	}
	return s.CashSave(ctx, CashSaveInput{
		CashID: k.CashID, GroupID: groupID, CashType: cashType,
		Date:        k.Date.Format("2006-01-02"),
		AccountName: k.AccountName, Description: k.Description,
		Debit: k.Debit, CreatedBy: createdBy,
	})
}

func toDueMemberDTO(m model.DueMember) model.DueMemberDTO {
	gid := ""
	if m.GroupID != nil {
		gid = *m.GroupID
	}
	return model.DueMemberDTO{
		MemberID: m.MemberID, GroupID: gid, MemberName: m.MemberName,
		MonthlyTarget: m.MonthlyTarget, Status: m.Status,
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
	}
}

func toDuePaymentDTO(p model.DuePayment) model.DuePaymentDTO {
	gid := ""
	if p.GroupID != nil {
		gid = *p.GroupID
	}
	months := make([]string, 0, len(p.Carryovers))
	items := make([]model.DuePaymentCarryoverDTO, 0, len(p.Carryovers))
	for _, c := range p.Carryovers {
		months = append(months, c.Month)
		items = append(items, model.DuePaymentCarryoverDTO{Month: c.Month, Amount: c.Amount})
	}
	return model.DuePaymentDTO{
		PaymentID: p.PaymentID, GroupID: gid, MemberID: p.MemberID,
		PaymentDate: p.PaymentDate.Format("2006-01-02"), TotalAmount: p.TotalAmount,
		CarryoverIR: p.CarryoverIR, CarryoverMonths: months, CarryoverItems: items,
		ConnectingFund: p.ConnectingFund,
		CommunityDues:  p.CommunityDues, OutreachFund: p.OutreachFund,
		ThousandFund: p.ThousandFund, FuneralFund: p.FuneralFund,
		UkhroMT: p.UkhroMT, Notes: p.Notes, Status: p.Status,
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *FinanceService) DuesData(ctx context.Context, groupID, month string) (*model.DuesDataDTO, error) {
	members, err := s.repo.DueMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}
	payments, err := s.repo.DuePayments(ctx, groupID, month)
	if err != nil {
		return nil, err
	}
	mDTO := make([]model.DueMemberDTO, 0, len(members))
	var target float64
	for _, m := range members {
		if m.Status == "ACTIVE" {
			target += m.MonthlyTarget
		}
		mDTO = append(mDTO, toDueMemberDTO(m))
	}
	pDTO := make([]model.DuePaymentDTO, 0, len(payments))
	var received float64
	paid := map[string]bool{}
	ids := make([]string, 0, len(payments))
	for _, p := range payments {
		ids = append(ids, p.PaymentID)
	}
	carryMap, err := s.repo.CarryoversByPayment(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, p := range payments {
		if p.Status == "REVERSED" {
			continue
		}
		// INACTIVE = riwayat nonaktif dari sheet (mis. batal/koreksi):
		// tetap ditampilkan sebagai arsip, tapi tidak dihitung
		// sebagai penerimaan maupun pelunasan.
		if p.Status != "INACTIVE" {
			received += p.TotalAmount
			paid[p.MemberID] = true
		}
		p.Carryovers = carryMap[p.PaymentID]
		pDTO = append(pDTO, toDuePaymentDTO(p))
	}
	active := 0
	for _, m := range members {
		if m.Status == "ACTIVE" {
			active++
		}
	}
	return &model.DuesDataDTO{
		SelectedMonth: month,
		Members:       mDTO,
		Payments:      pDTO,
		Dashboard: model.DuesDashboardDTO{
			Target: target, Received: received,
			PaidCount: len(paid), UnpaidCount: active - len(paid),
			MemberCount: len(members),
		},
	}, nil
}

func (s *FinanceService) DueMemberSave(ctx context.Context, groupID, memberID, name string, target float64) (*model.DueMemberDTO, error) {
	if name == "" {
		return nil, errors.New("name anggota wajib diisi")
	}
	if memberID == "" {
		memberID = util.NewID("SHM")
	}
	gid := groupID
	m := &model.DueMember{
		MemberID: memberID, GroupID: &gid,
		MemberName: name, MonthlyTarget: target, Status: "ACTIVE",
	}
	if err := s.repo.DueMemberUpsert(ctx, m); err != nil {
		return nil, err
	}
	_ = s.repo.MarkSynced(ctx, "due_members", "due_member_id", m.MemberID, "app", 0)
	dto := toDueMemberDTO(*m)
	s.notify(groupID)
	return &dto, nil
}

func (s *FinanceService) DueMemberDelete(ctx context.Context, groupID, memberID string) error {
	if err := s.repo.DueMemberDelete(ctx, groupID, memberID); err != nil {
		return err
	}
	_ = s.repo.Tombstone(ctx, groupID, "due_members", memberID)
	s.notify(groupID)
	return nil
}

type DuePaymentCarryoverInput struct {
	Month  string
	Amount float64
}

type DuePaymentInput struct {
	PaymentID   string
	GroupID     string
	MemberID    string
	PaymentDate string
	// Cara baru (disarankan): rincian per bulan. carryover_ir dihitung = SUM.
	CarryoverItems []DuePaymentCarryoverInput
	// Cara lama (kompatibel): teks bulan + nominal total → dibagi rata.
	// Diabaikan bila CarryoverItems diisi.
	CarryoverMonths string
	CarryoverIR     float64
	ConnectingFund  float64
	CommunityDues   float64
	OutreachFund    float64
	ThousandFund    float64
	FuneralFund     float64
	UkhroMT         float64
	Notes           string
	CreatedBy       string
}

// normalizeCarryovers mengubah input (baru/lama/sheet) menjadi rincian
// per bulan yang tervalidasi + total susulan. Aturan:
//   - items: bulan harus YYYY-MM, unik, amount >= 0.
//   - legacy (months teks + ir): token diparse fleksibel, nominal dibagi
//     rata (sisa ke bulan terakhir). Token tak dikenal → error agar
//     pemanggil tahu, bukan hilang diam-diam.
func normalizeCarryovers(items []DuePaymentCarryoverInput, legacyMonths string, legacyIR float64) ([]model.DuePaymentCarryover, float64, error) {
	if len(items) > 0 {
		seen := map[string]bool{}
		out := make([]model.DuePaymentCarryover, 0, len(items))
		var sum float64
		for _, it := range items {
			m, err := util.ParseSheetMonth(it.Month)
			if err != nil {
				return nil, 0, errors.New("bulan susulan tidak valid: " + it.Month)
			}
			if seen[m] {
				return nil, 0, errors.New("bulan susulan duplikat: " + m)
			}
			seen[m] = true
			if it.Amount < 0 {
				return nil, 0, errors.New("nominal susulan tidak boleh negatif")
			}
			sum += it.Amount
			out = append(out, model.DuePaymentCarryover{
				CarryoverID: util.NewID("CRY"), Month: m, Amount: it.Amount,
			})
		}
		sort.Slice(out, func(i, j int) bool { return out[i].Month < out[j].Month })
		return out, sum, nil
	}
	months, unknowns := util.SplitSheetMonths(legacyMonths)
	if len(unknowns) > 0 {
		return nil, 0, errors.New("bulan susulan tidak dikenal: " + strings.Join(unknowns, ", "))
	}
	if len(months) == 0 || legacyIR <= 0 {
		return nil, 0, nil
	}
	// Bagi rata (sisa pembulatan ke bulan terakhir) agar SUM = carryover_ir.
	per := math.Floor(legacyIR/float64(len(months))*100) / 100
	out := make([]model.DuePaymentCarryover, 0, len(months))
	var acc float64
	for i, m := range months {
		amt := per
		if i == len(months)-1 {
			amt = legacyIR - acc
		}
		acc += amt
		out = append(out, model.DuePaymentCarryover{
			CarryoverID: util.NewID("CRY"), Month: m, Amount: amt,
		})
	}
	return out, legacyIR, nil
}

func (s *FinanceService) DuePaymentSave(ctx context.Context, in DuePaymentInput) (*model.DuePaymentDTO, error) {
	if in.MemberID == "" {
		return nil, errors.New("member_id wajib diisi")
	}
	tgl, err := util.ParseSheetDate(in.PaymentDate)
	if err != nil {
		return nil, errors.New("payment_tanggal tidak valid (YYYY-MM-DD)")
	}
	carryovers, carryIR, err := normalizeCarryovers(in.CarryoverItems, in.CarryoverMonths, in.CarryoverIR)
	if err != nil {
		return nil, err
	}
	total := carryIR + in.ConnectingFund + in.CommunityDues +
		in.OutreachFund + in.ThousandFund + in.FuneralFund + in.UkhroMT
	if total <= 0 {
		return nil, errors.New("total pembayaran harus lebih besar dari 0")
	}
	if in.PaymentID == "" {
		in.PaymentID = util.NewID("SHP")
	}
	gid := in.GroupID
	p := &model.DuePayment{
		PaymentID: in.PaymentID, GroupID: &gid, MemberID: in.MemberID,
		PaymentDate: tgl, TotalAmount: total, CarryoverIR: carryIR,
		Carryovers:     carryovers,
		ConnectingFund: in.ConnectingFund, CommunityDues: in.CommunityDues,
		OutreachFund: in.OutreachFund, ThousandFund: in.ThousandFund,
		FuneralFund: in.FuneralFund, UkhroMT: in.UkhroMT,
		Notes: in.Notes, Status: "ACTIVE", CreatedBy: in.CreatedBy,
	}
	if err := s.repo.DuePaymentUpsert(ctx, p); err != nil {
		return nil, err
	}
	for i := range p.Carryovers {
		p.Carryovers[i].PaymentID = p.PaymentID
	}
	if err := s.repo.ReplaceCarryovers(ctx, p.PaymentID, p.Carryovers); err != nil {
		return nil, err
	}
	_ = s.repo.MarkSynced(ctx, "due_payments", "payment_id", p.PaymentID, "app", 0)
	dto := toDuePaymentDTO(*p)
	s.notify(in.GroupID)
	return &dto, nil
}

func (s *FinanceService) DuePaymentReverse(ctx context.Context, groupID, paymentID string) error {
	if err := s.repo.DuePaymentReverse(ctx, groupID, paymentID); err != nil {
		return err
	}
	_ = s.repo.MarkSynced(ctx, "due_payments", "payment_id", paymentID, "app", 0)
	s.notify(groupID)
	return nil
}

// DuePostMarker menandai jurnal kas hasil posting shodaqoh agar idempoten.
// Pola diadaptasi dari Post_Shodaqoh.js kas-latukan-web (POSTED_SHODAQOH_YYYY-MM).
func DuePostMarker(monthKey string) string {
	return "POSTED_SHODAQOH_" + monthKey
}

// DuePostToCash mengagregat pembayaran shodaqoh bulan tertentu menjadi jurnal kas.
// 7 pos ala kas-latukan-web: susulan IR, uang sambung, jimpitan, siar-siar,
// seribuan, kafan, ukhro MT. Satu baris DEBIT per pos yang totalnya > 0.
func (s *FinanceService) DuePostToCash(ctx context.Context, groupID, monthKey, createdBy string) ([]model.CashTransactionDTO, error) {
	if len(monthKey) != 7 {
		return nil, errors.New("bulan tidak valid. Gunakan format YYYY-MM.")
	}
	tgl, err := time.Parse("2006-01-02", monthKey+"-01")
	if err != nil {
		return nil, errors.New("bulan tidak valid. Gunakan format YYYY-MM.")
	}
	payments, err := s.repo.DuePayments(ctx, groupID, monthKey)
	if err != nil {
		return nil, err
	}
	marker := DuePostMarker(monthKey)
	existing, err := s.repo.CashList(ctx, groupID, "main")
	if err != nil {
		return nil, err
	}
	for _, k := range existing {
		if strings.Contains(k.Description, marker) {
			return nil, errors.New("bulan " + monthKey + " sudah pernah diposting ke kas.")
		}
	}
	var totIR, totSambung, totJimpitan, totSiar, totSeribu, totKafan, totUkhro float64
	count := 0
	for _, p := range payments {
		if p.Status == "REVERSED" || p.Status == "INACTIVE" || p.TotalAmount <= 0 {
			continue
		}
		count++
		totIR += p.CarryoverIR
		totSambung += p.ConnectingFund
		totJimpitan += p.CommunityDues
		totSiar += p.OutreachFund
		totSeribu += p.ThousandFund
		totKafan += p.FuneralFund
		totUkhro += p.UkhroMT
	}
	if count == 0 {
		return nil, errors.New("tidak ada pembayaran shodaqoh pada bulan " + monthKey + ".")
	}
	label := indonesianMonthLabel(monthKey)
	pos := []struct {
		account string
		amount  float64
	}{
		{"INFAQ SHODAQOH IR", totIR},
		{"UANG SAMBUNG", totSambung},
		{"JIMPITAN", totJimpitan},
		{"SIAR-SIAR", totSiar},
		{"SERIBUAN", totSeribu},
		{"KAFAN", totKafan},
		{"UKHRO MT", totUkhro},
	}
	out := make([]model.CashTransactionDTO, 0, len(pos))
	gid := groupID
	for _, pp := range pos {
		if pp.amount <= 0 {
			continue
		}
		k := &model.CashTransaction{
			CashID:      util.NewID("KAS"),
			GroupID:     &gid,
			CashType:    "main",
			Date:        tgl,
			AccountName: pp.account,
			Description: "Shodaqoh " + label + " - " + pp.account + " [" + marker + "]",
			Debit:       pp.amount,
			Credit:      0,
			CreatedBy:   createdBy,
		}
		if err := s.repo.CashInsert(ctx, k); err != nil {
			return nil, err
		}
		_ = s.repo.MarkSynced(ctx, "cash_transactions", "cash_id", k.CashID, "app", 0)
		out = append(out, model.CashTransactionDTO{
			CashID: k.CashID, GroupID: groupID, CashType: "main",
			TransactionDate: tgl.Format("2006-01-02"),
			AccountName:     k.AccountName, Description: k.Description,
			Debit: k.Debit, Credit: 0, CreatedBy: createdBy,
		})
	}
	if len(out) == 0 {
		return nil, errors.New("total agregat shodaqoh nol, tidak ada yang diposting.")
	}
	s.notify(groupID)
	return out, nil
}

// DueCancelPostToCash menghapus jurnal hasil posting bulan tertentu.
func (s *FinanceService) DueCancelPostToCash(ctx context.Context, groupID, monthKey string) (int, error) {
	if len(monthKey) != 7 {
		return 0, errors.New("bulan tidak valid. Gunakan format YYYY-MM.")
	}
	marker := DuePostMarker(monthKey)
	rows, err := s.repo.CashList(ctx, groupID, "main")
	if err != nil {
		return 0, err
	}
	deleted := 0
	for _, k := range rows {
		if !strings.Contains(k.Description, marker) {
			continue
		}
		if err := s.repo.CashDelete(ctx, groupID, k.CashID); err != nil {
			return deleted, err
		}
		_ = s.repo.Tombstone(ctx, groupID, "cash", k.CashID)
		deleted++
	}
	if deleted == 0 {
		return 0, errors.New("tidak ditemukan hasil posting bulan " + monthKey + ".")
	}
	s.notify(groupID)
	return deleted, nil
}

func (s *FinanceService) DueLastNominals(ctx context.Context, groupID, memberID string) (*model.DuePaymentDTO, error) {
	payments, err := s.repo.DuePayments(ctx, groupID, "")
	if err != nil {
		return nil, err
	}
	var latest *model.DuePayment
	for i := range payments {
		p := payments[i]
		if p.MemberID != memberID || p.Status == "REVERSED" {
			continue
		}
		if latest == nil || p.PaymentDate.After(latest.PaymentDate) {
			cp := p
			latest = &cp
		}
	}
	if latest == nil {
		return nil, errors.New("belum ada pembayaran anggota ini")
	}
	carryMap, err := s.repo.CarryoversByPayment(ctx, []string{latest.PaymentID})
	if err != nil {
		return nil, err
	}
	latest.Carryovers = carryMap[latest.PaymentID]
	dto := toDuePaymentDTO(*latest)
	return &dto, nil
}

type ZakatPayerInput struct {
	PayerID            string  `json:"payer_id"`
	MasterID           string  `json:"master_id"`
	Name               string  `json:"name"`
	Amount             float64 `json:"amount"`
	ZakatCategory      string  `json:"zakat_category"`
	FamilyMembersCount int     `json:"family_members_count"`
}

type ZakatRecipientInput struct {
	RecipientID   string  `json:"recipient_id"`
	MasterID      string  `json:"master_id"`
	Name          string  `json:"name"`
	Amount        float64 `json:"amount"`
	ZakatCategory string  `json:"zakat_category"`
}

type ZakatAllocationInput struct {
	Category               string  `json:"category"` // kanonis: FITRAH/MAL/TIJAROH/ZURU/LIVESTOCK/OTHER
	RecipientPercent       int     `json:"recipient_percent"`
	RecipientAmount        float64 `json:"recipient_amount"`
	RecipientGroupPercent  int     `json:"recipient_group_percent"`
	RecipientGroupAmount   float64 `json:"recipient_group_amount"`
	RecipientRegionPercent int     `json:"recipient_region_percent"`
	RecipientRegionAmount  float64 `json:"recipient_region_amount"`
	SabilillahPercent      int     `json:"sabilillah_percent"`
	SabilillahAmount       float64 `json:"sabilillah_amount"`
	AmilPercent            int     `json:"amil_percent"`
	AmilAmount             float64 `json:"amil_amount"`
	AmilGroupPercent       int     `json:"amil_group_percent"`
	AmilGroupAmount        float64 `json:"amil_group_amount"`
	AmilVillagePercent     int     `json:"amil_village_percent"`
	AmilVillageAmount      float64 `json:"amil_village_amount"`
	AmilRegionPercent      int     `json:"amil_region_percent"`
	AmilRegionAmount       float64 `json:"amil_region_amount"`
}

func toZakatRecordDTO(z model.ZakatRecord, payers []model.ZakatPayer, recips []model.ZakatRecipient, allocs []model.ZakatAllocation) model.ZakatRecordDTO {
	gid := ""
	if z.GroupID != nil {
		gid = *z.GroupID
	}
	tgl := ""
	if z.TransactionDate != nil {
		tgl = z.TransactionDate.Format("2006-01-02")
	}
	comp := ""
	if z.CompletedAt != nil {
		comp = z.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	pList := make([]model.ZakatPayerDTO, 0, len(payers))
	for _, p := range payers {
		mid := ""
		if p.MasterID != nil {
			mid = *p.MasterID
		}
		pList = append(pList, model.ZakatPayerDTO{
			PayerID: p.PayerID, MasterID: mid, Name: p.Name, Amount: p.Amount,
			ZakatCategory: p.ZakatCategory, FamilyMembersCount: p.FamilyMembersCount, SortOrder: p.SortOrder,
		})
	}

	rList := make([]model.ZakatRecipientDTO, 0, len(recips))
	for _, p := range recips {
		mid := ""
		if p.MasterID != nil {
			mid = *p.MasterID
		}
		rList = append(rList, model.ZakatRecipientDTO{
			RecipientID: p.RecipientID, MasterID: mid, Name: p.Name, Amount: p.Amount,
			ZakatCategory: p.ZakatCategory, SortOrder: p.SortOrder,
		})
	}

	aDTO := model.ZakatAllocationsDTO{ByCategory: map[string]*model.ZakatAllocationCategoryDTO{}}
	for _, a := range allocs {
		cat := &model.ZakatAllocationCategoryDTO{
			Total: a.RecipientAmount + a.SabilillahAmount + a.AmilAmount,
			Recipient: model.ZakatAllocationGroupDTO{
				Percent: a.RecipientPercent, Amount: a.RecipientAmount,
				Group:  &model.ZakatAllocationGroupDTO{Percent: a.RecipientGroupPercent, Amount: a.RecipientGroupAmount},
				Region: &model.ZakatAllocationGroupDTO{Percent: a.RecipientRegionPercent, Amount: a.RecipientRegionAmount},
			},
			Sabilillah: model.ZakatAllocationGroupDTO{Percent: a.SabilillahPercent, Amount: a.SabilillahAmount},
			Amil: model.ZakatAllocationGroupDTO{
				Percent: a.AmilPercent, Amount: a.AmilAmount,
				Group:   &model.ZakatAllocationGroupDTO{Percent: a.AmilGroupPercent, Amount: a.AmilGroupAmount},
				Village: &model.ZakatAllocationGroupDTO{Percent: a.AmilVillagePercent, Amount: a.AmilVillageAmount},
				Region:  &model.ZakatAllocationGroupDTO{Percent: a.AmilRegionPercent, Amount: a.AmilRegionAmount},
			},
		}
		normCat := util.NormZakatCategory(a.Category)
		aDTO.ByCategory[normCat] = cat
		if normCat == "FITRAH" {
			aDTO.Fitrah = cat
		} else if aDTO.Maal == nil {
			aDTO.Maal = cat
		}
	}

	catSet := map[string]struct{}{}
	for _, p := range payers {
		catSet[util.NormZakatCategory(p.ZakatCategory)] = struct{}{}
	}
	for _, p := range recips {
		catSet[util.NormZakatCategory(p.ZakatCategory)] = struct{}{}
	}
	categories := make([]string, 0, len(catSet))
	for c := range catSet {
		categories = append(categories, c)
	}
	sort.Strings(categories)

	return model.ZakatRecordDTO{
		ZakatID: z.ZakatID, GroupID: gid,
		Title: z.Title, Description: z.Description, Location: z.Location,
		Categories: categories,
		SoulCount:  z.SoulCount, TotalRiceKg: z.TotalRiceKg, TotalMoneyRp: z.TotalMoneyRp,
		Status: z.Status, TransactionDate: tgl, CompletedAt: comp, Version: z.Version,
		UpdatedAt: z.UpdatedAt.Format(time.RFC3339), UpdatedBy: z.UpdatedBy,
		PayerList: pList, RecipientList: rList, Allocations: &aDTO,
	}
}

func (s *FinanceService) ZakatList(ctx context.Context, groupID string) ([]model.ZakatRecordDTO, error) {
	return s.repo.ZakatListWithCounts(ctx, groupID)
}

func (s *FinanceService) ZakatDetail(ctx context.Context, groupID, zakatID string) (*model.ZakatRecordDTO, error) {
	z, payers, recips, allocs, err := s.repo.ZakatDetail(ctx, groupID, zakatID)
	if err != nil {
		return nil, err
	}
	if z == nil {
		return nil, errors.New("zakat tidak ditemukan")
	}
	dto := toZakatRecordDTO(*z, payers, recips, allocs)
	return &dto, nil
}

type ZakatHeaderInput struct {
	ZakatID         string  `json:"zakat_id"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	Location        string  `json:"location"`
	SoulCount       int     `json:"soul_count"`
	TotalRiceKg     float64 `json:"total_rice_kg"`
	TotalMoneyRp    float64 `json:"total_money_rp"`
	TransactionDate string  `json:"transaction_date"`
}

func (s *FinanceService) ZakatSave(ctx context.Context, groupID string, in ZakatHeaderInput, actor string) (*model.ZakatRecordDTO, error) {
	if in.ZakatID != "" {
		old, _, _, _, _ := s.repo.ZakatDetail(ctx, groupID, in.ZakatID)
		if old != nil && old.Status == "COMPLETED" {
			return nil, errors.New("zakat sudah selesai, tidak bisa diubah")
		}
	}
	if in.ZakatID == "" {
		in.ZakatID = util.NewID("ZKT")
	}
	tgl, _ := time.Parse("2006-01-02", in.TransactionDate)
	z := &model.ZakatRecord{
		ZakatID: in.ZakatID, GroupID: &groupID,
		Title: in.Title, Description: in.Description, Location: in.Location,
		SoulCount: in.SoulCount, TotalRiceKg: in.TotalRiceKg, TotalMoneyRp: in.TotalMoneyRp,
		TransactionDate: &tgl, CreatedBy: actor, UpdatedBy: actor, Status: "ACTIVE", Version: 1,
	}
	if err := s.repo.ZakatUpsert(ctx, z); err != nil {
		return nil, err
	}
	_ = s.repo.MarkSynced(ctx, "zakat_records", "zakat_id", z.ZakatID, "app", 0)
	s.notify(groupID)
	return s.ZakatDetail(ctx, groupID, z.ZakatID)
}

func (s *FinanceService) ZakatSavePayers(ctx context.Context, groupID, zakatID string, items []ZakatPayerInput) error {
	z, _, _, _, _ := s.repo.ZakatDetail(ctx, groupID, zakatID)
	if z == nil {
		return errors.New("zakat tidak ditemukan")
	}
	if z.Status == "COMPLETED" {
		return errors.New("zakat sudah selesai")
	}
	payers := make([]model.ZakatPayer, 0, len(items))
	for _, it := range items {
		var mid *string
		if it.MasterID != "" {
			mid = &it.MasterID
		}
		cat := strings.TrimSpace(it.ZakatCategory)
		if cat == "" {
			cat = "FITRAH"
		} else {
			cat = util.NormZakatCategory(cat)
		}
		payers = append(payers, model.ZakatPayer{
			PayerID: it.PayerID, MasterID: mid, Name: it.Name, Amount: it.Amount,
			ZakatCategory: cat, FamilyMembersCount: it.FamilyMembersCount,
		})
	}
	if err := s.repo.ReplaceZakatPayers(ctx, zakatID, payers); err != nil {
		return err
	}
	s.notify(groupID)
	return nil
}

func (s *FinanceService) ZakatSaveRecipients(ctx context.Context, groupID, zakatID string, items []ZakatRecipientInput) error {
	z, _, _, _, _ := s.repo.ZakatDetail(ctx, groupID, zakatID)
	if z == nil {
		return errors.New("zakat tidak ditemukan")
	}
	if z.Status == "COMPLETED" {
		return errors.New("zakat sudah selesai")
	}
	recips := make([]model.ZakatRecipient, 0, len(items))
	for _, it := range items {
		var mid *string
		if it.MasterID != "" {
			mid = &it.MasterID
		}
		cat := strings.TrimSpace(it.ZakatCategory)
		if cat == "" {
			cat = "FITRAH"
		} else {
			cat = util.NormZakatCategory(cat)
		}
		recips = append(recips, model.ZakatRecipient{
			RecipientID: it.RecipientID, MasterID: mid, Name: it.Name, Amount: it.Amount,
			ZakatCategory: cat,
		})
	}
	if err := s.repo.ReplaceZakatRecipients(ctx, zakatID, recips); err != nil {
		return err
	}
	s.notify(groupID)
	return nil
}

func (s *FinanceService) ZakatSaveAllocations(ctx context.Context, groupID, zakatID string, items []ZakatAllocationInput) error {
	z, _, _, _, _ := s.repo.ZakatDetail(ctx, groupID, zakatID)
	if z == nil {
		return errors.New("zakat tidak ditemukan")
	}
	if z.Status == "COMPLETED" {
		return errors.New("zakat sudah selesai")
	}
	allocs := make([]model.ZakatAllocation, 0, len(items))
	for _, it := range items {
		if it.RecipientPercent+it.SabilillahPercent+it.AmilPercent != 100 {
			return errors.New("total persen rincian harus 100%")
		}
		allocs = append(allocs, model.ZakatAllocation{
			ZakatID: zakatID, Category: util.NormZakatCategory(it.Category),
			RecipientPercent: it.RecipientPercent, RecipientAmount: it.RecipientAmount,
			RecipientGroupPercent: it.RecipientGroupPercent, RecipientGroupAmount: it.RecipientGroupAmount,
			RecipientRegionPercent: it.RecipientRegionPercent, RecipientRegionAmount: it.RecipientRegionAmount,
			SabilillahPercent: it.SabilillahPercent, SabilillahAmount: it.SabilillahAmount,
			AmilPercent: it.AmilPercent, AmilAmount: it.AmilAmount,
			AmilGroupPercent: it.AmilGroupPercent, AmilGroupAmount: it.AmilGroupAmount,
			AmilVillagePercent: it.AmilVillagePercent, AmilVillageAmount: it.AmilVillageAmount,
			AmilRegionPercent: it.AmilRegionPercent, AmilRegionAmount: it.AmilRegionAmount,
		})
	}
	if err := s.repo.ReplaceZakatAllocations(ctx, zakatID, allocs); err != nil {
		return err
	}
	s.notify(groupID)
	return nil
}

func (s *FinanceService) ZakatSetStatus(ctx context.Context, groupID, zakatID, status string) error {
	z, _, recips, allocs, _ := s.repo.ZakatDetail(ctx, groupID, zakatID)
	if z == nil {
		return errors.New("zakat tidak ditemukan")
	}
	if strings.ToUpper(status) == "COMPLETED" {
		var danaRecipient float64
		for _, a := range allocs {
			if util.NormZakatCategory(a.Category) == "MAL" {
				danaRecipient += a.RecipientGroupAmount + a.RecipientRegionAmount
			} else {
				danaRecipient += a.RecipientAmount
			}
		}
		var totalRecip float64
		for _, r := range recips {
			totalRecip += r.Amount
		}
		if totalRecip < danaRecipient-1 {
			return fmt.Errorf("recipient belum teralokasi semua! Kurang Rp %.0f", danaRecipient-totalRecip)
		}
	}
	var comp *time.Time
	if strings.ToUpper(status) == "COMPLETED" {
		now := time.Now()
		comp = &now
	}
	if err := s.repo.ZakatSetStatus(ctx, groupID, zakatID, strings.ToUpper(status), comp); err != nil {
		return err
	}
	_ = s.repo.MarkSynced(ctx, "zakat_records", "zakat_id", zakatID, "app", 0)
	s.notify(groupID)
	return nil
}

func (s *FinanceService) ZakatMasters(ctx context.Context, groupID string) (map[string]interface{}, error) {
	p, r, err := s.repo.MastersList(ctx, groupID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"payers": p, "recipients": r}, nil
}

func (s *FinanceService) ZakatAddMaster(ctx context.Context, groupID, kind, name string) (string, error) {
	return s.repo.MasterUpsert(ctx, kind, groupID, name)
}

func (s *FinanceService) ZakatDelete(ctx context.Context, groupID, zakatID string) error {
	if err := s.repo.ZakatDelete(ctx, groupID, zakatID); err != nil {
		return err
	}
	_ = s.repo.Tombstone(ctx, groupID, "zakat", zakatID)
	s.notify(groupID)
	return nil
}
