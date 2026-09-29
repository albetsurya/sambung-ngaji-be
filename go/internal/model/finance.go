package model

import "time"

type KasTransaction struct {
	KasID       string
	GroupID     *string
	KasType     string
	Tanggal     time.Time
	AccountName string
	Description string
	Debit       float64
	Credit      float64
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type KasTransactionDTO struct {
	No              int     `json:"no"`
	KasID           string  `json:"kas_id"`
	GroupID         string  `json:"group_id"`
	KasType         string  `json:"kas_type"`
	TransactionDate string  `json:"transaction_date"`
	AccountName     string  `json:"account_name"`
	Description     string  `json:"description"`
	Debit           float64 `json:"debit"`
	Credit          float64 `json:"credit"`
	Balance         float64 `json:"balance"`
	CreatedBy       string  `json:"created_by"`
}

type KasSummaryDTO struct {
	KasType        string              `json:"kas_type"`
	Transactions   []KasTransactionDTO `json:"transactions"`
	InitialBalance float64             `json:"initial_balance"`
	TotalDebit     float64             `json:"total_debit"`
	TotalCredit    float64             `json:"total_credit"`
	EndingBalance  float64             `json:"ending_balance"`
}

type ShodaqohMember struct {
	MemberID      string
	GroupID       *string
	MemberName    string
	MonthlyTarget float64
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ShodaqohMemberDTO struct {
	MemberID      string  `json:"member_id"`
	GroupID       string  `json:"group_id"`
	MemberName    string  `json:"member_name"`
	MonthlyTarget float64 `json:"monthly_target"`
	Status        string  `json:"status"`
}

type ShodaqohPayment struct {
	PaymentID          string
	GroupID            *string
	MemberID           string
	PaymentDate        time.Time
	TotalAmount        float64
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
	Status             string
	CreatedBy          string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type ShodaqohPaymentDTO struct {
	PaymentID          string  `json:"payment_id"`
	GroupID            string  `json:"group_id"`
	MemberID           string  `json:"member_id"`
	PaymentDate        string  `json:"payment_date"`
	TotalAmount        float64 `json:"total_amount"`
	CarryoverIR        float64 `json:"carryover_ir"`
	CarryoverMonths    string  `json:"carryover_months"`
	CarryoverBreakdown string  `json:"carryover_breakdown"`
	ConnectingFund     float64 `json:"connecting_fund"`
	CommunityDues      float64 `json:"community_dues"`
	OutreachFund       float64 `json:"outreach_fund"`
	ThousandFund       float64 `json:"thousand_fund"`
	FuneralFund        float64 `json:"funeral_fund"`
	UkhroMT            float64 `json:"ukhro_mt"`
	Notes              string  `json:"notes"`
	Status             string  `json:"status"`
}

type ShodaqohDashboardDTO struct {
	Target      float64 `json:"target"`
	Received    float64 `json:"received"`
	PaidCount   int     `json:"paidCount"`
	UnpaidCount int     `json:"unpaidCount"`
	MemberCount int     `json:"memberCount"`
}

type ShodaqohDataDTO struct {
	SelectedMonth string               `json:"selected_month"`
	Members       []ShodaqohMemberDTO  `json:"members"`
	Payments      []ShodaqohPaymentDTO `json:"payments"`
	Dashboard     ShodaqohDashboardDTO `json:"dashboard"`
}

type ZakatRecord struct {
	ZakatID         string
	GroupID         *string
	ZakatType       string
	MuzakkiName     string
	SoulCount       int
	TotalRiceKg     float64
	TotalMoneyRp    float64
	Status          string
	TransactionDate *time.Time
	Details         string
	CreatedBy       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ZakatRecordDTO struct {
	ZakatID         string        `json:"zakat_id"`
	GroupID         string        `json:"group_id"`
	ZakatType       string        `json:"zakat_type"`
	MuzakkiName     string        `json:"muzakki_name"`
	SoulCount       int           `json:"soul_count"`
	TotalRiceKg     float64       `json:"total_rice_kg"`
	TotalMoneyRp    float64       `json:"total_money_rp"`
	Status          string        `json:"status"`
	TransactionDate string        `json:"transaction_date"`
	MuzakkiList     []interface{} `json:"muzakki_list"`
	MustahikList    []interface{} `json:"mustahik_list"`
}
