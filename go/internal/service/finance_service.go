package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		TransactionDate: k.Tanggal.Format("2006-01-02"),
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
	Tanggal     string
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
	tgl, err := time.Parse("2006-01-02", in.Tanggal)
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
		Tanggal:     tgl,
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
				Tanggal:     k.Tanggal,
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
				Tanggal:     cp.Tanggal.Format("2006-01-02"),
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
		if k.Tanggal.Format("2006-01") == monthKey {
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
		if isOpeningBalance(k.AccountName) && k.Tanggal.Format("2006-01") == nextKey {
			return nil, errors.New("SALDO AWAL untuk " + nextKey + " sudah ada.")
		}
	}
	gid := groupID
	k := &model.CashTransaction{
		CashID:      util.NewID("KAS"),
		GroupID:     &gid,
		CashType:    cashType,
		Tanggal:     time.Date(next.Year(), next.Month(), 1, 0, 0, 0, 0, time.UTC),
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
		Tanggal:     k.Tanggal.Format("2006-01-02"),
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
	return model.DuePaymentDTO{
		PaymentID: p.PaymentID, GroupID: gid, MemberID: p.MemberID,
		PaymentDate: p.PaymentDate.Format("2006-01-02"), TotalAmount: p.TotalAmount,
		CarryoverIR: p.CarryoverIR, CarryoverMonths: p.CarryoverMonths,
		CarryoverBreakdown: p.CarryoverBreakdown, ConnectingFund: p.ConnectingFund,
		CommunityDues: p.CommunityDues, OutreachFund: p.OutreachFund,
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
	for _, p := range payments {
		if p.Status == "REVERSED" {
			continue
		}
		received += p.TotalAmount
		paid[p.MemberID] = true
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
		return nil, errors.New("nama anggota wajib diisi")
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

type DuePaymentInput struct {
	PaymentID          string
	GroupID            string
	MemberID           string
	PaymentDate        string
	CarryoverIR        float64
	CarryoverMonths    string
	CarryoverBreakdown string
	ConnectingFund     float64
	CommunityDues      float64
	OutreachFund       float64
	ThousandFund       float64
	FuneralFund        float64
	UkhroMT            float64
	Notes              string
	CreatedBy          string
}

func (s *FinanceService) DuePaymentSave(ctx context.Context, in DuePaymentInput) (*model.DuePaymentDTO, error) {
	if in.MemberID == "" {
		return nil, errors.New("member_id wajib diisi")
	}
	tgl, err := time.Parse("2006-01-02", in.PaymentDate)
	if err != nil {
		return nil, errors.New("payment_date tidak valid (YYYY-MM-DD)")
	}
	total := in.CarryoverIR + in.ConnectingFund + in.CommunityDues +
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
		PaymentDate: tgl, TotalAmount: total, CarryoverIR: in.CarryoverIR,
		CarryoverMonths: in.CarryoverMonths, CarryoverBreakdown: in.CarryoverBreakdown,
		ConnectingFund: in.ConnectingFund, CommunityDues: in.CommunityDues,
		OutreachFund: in.OutreachFund, ThousandFund: in.ThousandFund,
		FuneralFund: in.FuneralFund, UkhroMT: in.UkhroMT,
		Notes: in.Notes, Status: "ACTIVE", CreatedBy: in.CreatedBy,
	}
	if err := s.repo.DuePaymentUpsert(ctx, p); err != nil {
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
	dto := toDuePaymentDTO(*latest)
	return &dto, nil
}

func toZakatDTO(z model.ZakatRecord) model.ZakatRecordDTO {
	gid := ""
	if z.GroupID != nil {
		gid = *z.GroupID
	}
	tgl := ""
	if z.TransactionDate != nil {
		tgl = z.TransactionDate.Format("2006-01-02")
	}
	var muzaki, mustahik []interface{}
	if z.Details != "" {
		var det map[string]interface{}
		if err := json.Unmarshal([]byte(z.Details), &det); err == nil {
			if v, ok := det["muzakki_list"].([]interface{}); ok {
				muzaki = v
			}
			if v, ok := det["mustahik_list"].([]interface{}); ok {
				mustahik = v
			}
		}
	}
	return model.ZakatRecordDTO{
		ZakatID: z.ZakatID, GroupID: gid, ZakatType: z.ZakatType,
		MuzakkiName: z.MuzakkiName, SoulCount: z.SoulCount,
		TotalRiceKg: z.TotalRiceKg, TotalMoneyRp: z.TotalMoneyRp,
		Status: z.Status, TransactionDate: tgl,
		MuzakkiList: muzaki, MustahikList: mustahik,
		UpdatedAt: z.UpdatedAt.Format(time.RFC3339),
	}
}

func (s *FinanceService) ZakatList(ctx context.Context, groupID string) ([]model.ZakatRecordDTO, error) {
	rows, err := s.repo.ZakatList(ctx, groupID)
	if err != nil {
		return nil, err
	}
	out := make([]model.ZakatRecordDTO, 0, len(rows))
	for _, z := range rows {
		out = append(out, toZakatDTO(z))
	}
	return out, nil
}

type ZakatSaveInput struct {
	ZakatID         string
	GroupID         string
	ZakatType       string
	MuzakkiName     string
	SoulCount       int
	TotalRiceKg     float64
	TotalMoneyRp    float64
	TransactionDate string
	Details         string
	CreatedBy       string
}

func (s *FinanceService) ZakatSave(ctx context.Context, in ZakatSaveInput) (*model.ZakatRecordDTO, error) {
	if in.MuzakkiName == "" {
		return nil, errors.New("nama muzakki wajib diisi")
	}
	if in.ZakatID == "" {
		in.ZakatID = util.NewID("ZKT")
	}
	zt := strings.ToUpper(in.ZakatType)
	if zt != "MAL" {
		zt = "FITRAH"
	}
	var tgl *time.Time
	if in.TransactionDate != "" {
		t, err := time.Parse("2006-01-02", in.TransactionDate)
		if err != nil {
			return nil, errors.New("transaction_date tidak valid (YYYY-MM-DD)")
		}
		tgl = &t
	}
	det := in.Details
	if det == "" {
		det = "{}"
	}
	gid := in.GroupID
	z := &model.ZakatRecord{
		ZakatID: in.ZakatID, GroupID: &gid, ZakatType: zt,
		MuzakkiName: in.MuzakkiName, SoulCount: in.SoulCount,
		TotalRiceKg: in.TotalRiceKg, TotalMoneyRp: in.TotalMoneyRp,
		Status: "PENDING", TransactionDate: tgl, Details: det, CreatedBy: in.CreatedBy,
	}
	if err := s.repo.ZakatUpsert(ctx, z); err != nil {
		return nil, err
	}
	_ = s.repo.MarkSynced(ctx, "zakat_records", "zakat_id", z.ZakatID, "app", 0)
	dto := toZakatDTO(*z)
	s.notify(in.GroupID)
	return &dto, nil
}

func (s *FinanceService) ZakatSetStatus(ctx context.Context, groupID, zakatID, status string) error {
	st := strings.ToUpper(status)
	switch st {
	case "PENDING", "ACTIVE", "COMPLETED", "CANCELLED":
	default:
		return errors.New("status zakat tidak valid")
	}
	if err := s.repo.ZakatSetStatus(ctx, groupID, zakatID, st); err != nil {
		return err
	}
	_ = s.repo.MarkSynced(ctx, "zakat_records", "zakat_id", zakatID, "app", 0)
	s.notify(groupID)
	return nil
}

func (s *FinanceService) ZakatDelete(ctx context.Context, groupID, zakatID string) error {
	if err := s.repo.ZakatDelete(ctx, groupID, zakatID); err != nil {
		return err
	}
	_ = s.repo.Tombstone(ctx, groupID, "zakat", zakatID)
	s.notify(groupID)
	return nil
}
