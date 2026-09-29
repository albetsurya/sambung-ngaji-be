package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

type FinanceService struct {
	repo *repository.FinanceRepo
}

func NewFinanceService(repo *repository.FinanceRepo) *FinanceService {
	return &FinanceService{repo: repo}
}

func normKasType(s string) string {
	if strings.ToLower(s) == "kas_amil" {
		return "kas_amil"
	}
	return "main"
}

func toKasDTO(no int, k model.KasTransaction, balance float64) model.KasTransactionDTO {
	gid := ""
	if k.GroupID != nil {
		gid = *k.GroupID
	}
	return model.KasTransactionDTO{
		No:              no,
		KasID:           k.KasID,
		GroupID:         gid,
		KasType:         k.KasType,
		TransactionDate: k.Tanggal.Format("2006-01-02"),
		AccountName:     k.AccountName,
		Description:     k.Description,
		Debit:           k.Debit,
		Credit:          k.Credit,
		Balance:         balance,
		CreatedBy:       k.CreatedBy,
	}
}

func (s *FinanceService) KasList(ctx context.Context, groupID, kasType string) (*model.KasSummaryDTO, error) {
	kasType = normKasType(kasType)
	rows, err := s.repo.KasList(ctx, groupID, kasType)
	if err != nil {
		return nil, err
	}
	items := make([]model.KasTransactionDTO, 0, len(rows))
	var totalDebit, totalCredit, balance float64
	for i, k := range rows {
		totalDebit += k.Debit
		totalCredit += k.Credit
		balance += k.Debit - k.Credit
		items = append(items, toKasDTO(i+1, k, balance))
	}
	return &model.KasSummaryDTO{
		KasType:        kasType,
		Transactions:   items,
		InitialBalance: 0,
		TotalDebit:     totalDebit,
		TotalCredit:    totalCredit,
		EndingBalance:  balance,
	}, nil
}

type KasSaveInput struct {
	KasID       string
	GroupID     string
	KasType     string
	Tanggal     string
	AccountName string
	Description string
	Debit       float64
	Credit      float64
	CreatedBy   string
}

func (s *FinanceService) KasSave(ctx context.Context, in KasSaveInput) (*model.KasTransactionDTO, error) {
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
	k := &model.KasTransaction{
		KasID:       in.KasID,
		GroupID:     &gid,
		KasType:     normKasType(in.KasType),
		Tanggal:     tgl,
		AccountName: in.AccountName,
		Description: in.Description,
		Debit:       in.Debit,
		Credit:      in.Credit,
		CreatedBy:   in.CreatedBy,
	}
	if k.KasID == "" {
		k.KasID = util.NewID("KAS")
		if err := s.repo.KasInsert(ctx, k); err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.KasUpdate(ctx, k); err != nil {
			return nil, err
		}
	}
	sum, err := s.KasList(ctx, in.GroupID, k.KasType)
	if err != nil {
		return nil, err
	}
	for _, it := range sum.Transactions {
		if it.KasID == k.KasID {
			return &it, nil
		}
	}
	return &model.KasTransactionDTO{KasID: k.KasID, GroupID: in.GroupID}, nil
}

func (s *FinanceService) KasDelete(ctx context.Context, groupID, kasID string) error {
	if kasID == "" {
		return errors.New("kas_id wajib diisi")
	}
	return s.repo.KasDelete(ctx, groupID, kasID)
}

func (s *FinanceService) KasDuplicate(ctx context.Context, groupID, kasType, kasID, createdBy string) (*model.KasTransactionDTO, error) {
	rows, err := s.repo.KasList(ctx, groupID, normKasType(kasType))
	if err != nil {
		return nil, err
	}
	for _, k := range rows {
		if k.KasID == kasID {
			gid := groupID
			cp := &model.KasTransaction{
				KasID:       util.NewID("KAS"),
				GroupID:     &gid,
				KasType:     k.KasType,
				Tanggal:     k.Tanggal,
				AccountName: k.AccountName,
				Description: k.Description,
				Debit:       k.Debit,
				Credit:      k.Credit,
				CreatedBy:   createdBy,
			}
			if err := s.repo.KasInsert(ctx, cp); err != nil {
				return nil, err
			}
			return s.KasSave(ctx, KasSaveInput{
				KasID: cp.KasID, GroupID: groupID, KasType: cp.KasType,
				Tanggal:     cp.Tanggal.Format("2006-01-02"),
				AccountName: cp.AccountName, Description: cp.Description,
				Debit: cp.Debit, Credit: cp.Credit, CreatedBy: createdBy,
			})
		}
	}
	return nil, errors.New("transaksi tidak ditemukan")
}

func (s *FinanceService) KasCarryForward(ctx context.Context, groupID, kasType, monthKey, createdBy string) (*model.KasTransactionDTO, error) {
	kasType = normKasType(kasType)
	if len(monthKey) != 7 {
		return nil, errors.New("monthKey tidak valid (YYYY-MM)")
	}
	rows, err := s.repo.KasList(ctx, groupID, kasType)
	if err != nil {
		return nil, err
	}
	var balance float64
	for _, k := range rows {
		balance += k.Debit - k.Credit
	}
	next, err := time.Parse("2006-01", monthKey)
	if err != nil {
		return nil, errors.New("monthKey tidak valid (YYYY-MM)")
	}
	next = next.AddDate(0, 1, 0)
	gid := groupID
	k := &model.KasTransaction{
		KasID:       util.NewID("KAS"),
		GroupID:     &gid,
		KasType:     kasType,
		Tanggal:     time.Date(next.Year(), next.Month(), 1, 0, 0, 0, 0, time.UTC),
		AccountName: "SALDO AWAL",
		Description: "Saldo awal pindahan " + monthKey,
		Debit:       balance,
		Credit:      0,
		CreatedBy:   createdBy,
	}
	if err := s.repo.KasInsert(ctx, k); err != nil {
		return nil, err
	}
	return s.KasSave(ctx, KasSaveInput{
		KasID: k.KasID, GroupID: groupID, KasType: kasType,
		Tanggal:     k.Tanggal.Format("2006-01-02"),
		AccountName: k.AccountName, Description: k.Description,
		Debit: k.Debit, CreatedBy: createdBy,
	})
}

func toShodaqohMemberDTO(m model.ShodaqohMember) model.ShodaqohMemberDTO {
	gid := ""
	if m.GroupID != nil {
		gid = *m.GroupID
	}
	return model.ShodaqohMemberDTO{
		MemberID: m.MemberID, GroupID: gid, MemberName: m.MemberName,
		MonthlyTarget: m.MonthlyTarget, Status: m.Status,
	}
}

func toShodaqohPaymentDTO(p model.ShodaqohPayment) model.ShodaqohPaymentDTO {
	gid := ""
	if p.GroupID != nil {
		gid = *p.GroupID
	}
	return model.ShodaqohPaymentDTO{
		PaymentID: p.PaymentID, GroupID: gid, MemberID: p.MemberID,
		PaymentDate: p.PaymentDate.Format("2006-01-02"), TotalAmount: p.TotalAmount,
		CarryoverIR: p.CarryoverIR, CarryoverMonths: p.CarryoverMonths,
		CarryoverBreakdown: p.CarryoverBreakdown, ConnectingFund: p.ConnectingFund,
		CommunityDues: p.CommunityDues, OutreachFund: p.OutreachFund,
		ThousandFund: p.ThousandFund, FuneralFund: p.FuneralFund,
		UkhroMT: p.UkhroMT, Notes: p.Notes, Status: p.Status,
	}
}

func (s *FinanceService) ShodaqohData(ctx context.Context, groupID, month string) (*model.ShodaqohDataDTO, error) {
	members, err := s.repo.ShodaqohMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}
	payments, err := s.repo.ShodaqohPayments(ctx, groupID, month)
	if err != nil {
		return nil, err
	}
	mDTO := make([]model.ShodaqohMemberDTO, 0, len(members))
	var target float64
	for _, m := range members {
		if m.Status == "ACTIVE" {
			target += m.MonthlyTarget
		}
		mDTO = append(mDTO, toShodaqohMemberDTO(m))
	}
	pDTO := make([]model.ShodaqohPaymentDTO, 0, len(payments))
	var received float64
	paid := map[string]bool{}
	for _, p := range payments {
		if p.Status == "REVERSED" {
			continue
		}
		received += p.TotalAmount
		paid[p.MemberID] = true
		pDTO = append(pDTO, toShodaqohPaymentDTO(p))
	}
	active := 0
	for _, m := range members {
		if m.Status == "ACTIVE" {
			active++
		}
	}
	return &model.ShodaqohDataDTO{
		SelectedMonth: month,
		Members:       mDTO,
		Payments:      pDTO,
		Dashboard: model.ShodaqohDashboardDTO{
			Target: target, Received: received,
			PaidCount: len(paid), UnpaidCount: active - len(paid),
			MemberCount: len(members),
		},
	}, nil
}

func (s *FinanceService) ShodaqohMemberSave(ctx context.Context, groupID, memberID, name string, target float64) (*model.ShodaqohMemberDTO, error) {
	if name == "" {
		return nil, errors.New("nama anggota wajib diisi")
	}
	if memberID == "" {
		memberID = util.NewID("SHM")
	}
	gid := groupID
	m := &model.ShodaqohMember{
		MemberID: memberID, GroupID: &gid,
		MemberName: name, MonthlyTarget: target, Status: "ACTIVE",
	}
	if err := s.repo.ShodaqohMemberUpsert(ctx, m); err != nil {
		return nil, err
	}
	dto := toShodaqohMemberDTO(*m)
	return &dto, nil
}

func (s *FinanceService) ShodaqohMemberDelete(ctx context.Context, groupID, memberID string) error {
	return s.repo.ShodaqohMemberDelete(ctx, groupID, memberID)
}

type ShodaqohPaymentInput struct {
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

func (s *FinanceService) ShodaqohPaymentSave(ctx context.Context, in ShodaqohPaymentInput) (*model.ShodaqohPaymentDTO, error) {
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
	p := &model.ShodaqohPayment{
		PaymentID: in.PaymentID, GroupID: &gid, MemberID: in.MemberID,
		PaymentDate: tgl, TotalAmount: total, CarryoverIR: in.CarryoverIR,
		CarryoverMonths: in.CarryoverMonths, CarryoverBreakdown: in.CarryoverBreakdown,
		ConnectingFund: in.ConnectingFund, CommunityDues: in.CommunityDues,
		OutreachFund: in.OutreachFund, ThousandFund: in.ThousandFund,
		FuneralFund: in.FuneralFund, UkhroMT: in.UkhroMT,
		Notes: in.Notes, Status: "ACTIVE", CreatedBy: in.CreatedBy,
	}
	if err := s.repo.ShodaqohPaymentUpsert(ctx, p); err != nil {
		return nil, err
	}
	dto := toShodaqohPaymentDTO(*p)
	return &dto, nil
}

func (s *FinanceService) ShodaqohPaymentReverse(ctx context.Context, groupID, paymentID string) error {
	return s.repo.ShodaqohPaymentReverse(ctx, groupID, paymentID)
}

func (s *FinanceService) ShodaqohLastNominals(ctx context.Context, groupID, memberID string) (*model.ShodaqohPaymentDTO, error) {
	payments, err := s.repo.ShodaqohPayments(ctx, groupID, "")
	if err != nil {
		return nil, err
	}
	var latest *model.ShodaqohPayment
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
	dto := toShodaqohPaymentDTO(*latest)
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
	dto := toZakatDTO(*z)
	return &dto, nil
}

func (s *FinanceService) ZakatSetStatus(ctx context.Context, groupID, zakatID, status string) error {
	st := strings.ToUpper(status)
	switch st {
	case "PENDING", "ACTIVE", "COMPLETED", "CANCELLED":
	default:
		return errors.New("status zakat tidak valid")
	}
	return s.repo.ZakatSetStatus(ctx, groupID, zakatID, st)
}

func (s *FinanceService) ZakatDelete(ctx context.Context, groupID, zakatID string) error {
	return s.repo.ZakatDelete(ctx, groupID, zakatID)
}
