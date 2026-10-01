package model

import "time"

type CashTransaction struct {
	CashID      string
	GroupID     *string
	CashType    string
	Tanggal     time.Time
	AccountName string
	Description string
	Debit       float64
	Credit      float64
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CashTransactionDTO struct {
	No              int     `json:"no"`
	CashID          string  `json:"cash_id"`
	GroupID         string  `json:"group_id"`
	CashType        string  `json:"cash_type"`
	TransactionDate string  `json:"transaction_date"`
	AccountName     string  `json:"account_name"`
	Description     string  `json:"description"`
	Debit           float64 `json:"debit"`
	Credit          float64 `json:"credit"`
	Balance         float64 `json:"balance"`
	CreatedBy       string  `json:"created_by"`
	UpdatedAt       string  `json:"updated_at"`
}

type CashSummaryDTO struct {
	CashType       string               `json:"cash_type"`
	Transactions   []CashTransactionDTO `json:"transactions"`
	InitialBalance float64              `json:"initial_balance"`
	TotalDebit     float64              `json:"total_debit"`
	TotalCredit    float64              `json:"total_credit"`
	EndingBalance  float64              `json:"ending_balance"`
}

type DueMember struct {
	MemberID      string
	GroupID       *string
	MemberName    string
	MonthlyTarget float64
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type DueMemberDTO struct {
	MemberID      string  `json:"member_id"`
	GroupID       string  `json:"group_id"`
	MemberName    string  `json:"member_name"`
	MonthlyTarget float64 `json:"monthly_target"`
	Status        string  `json:"status"`
	UpdatedAt     string  `json:"updated_at"`
}

// DuePaymentCarryover = rincian susulan IR per bulan (1 baris = 1 bulan).
// Menggantikan kolom multi-nilai carryover_months + teks bebas
// carryover_breakdown (dihapus di migrasi 000024). Aturan:
//   - month selalu "YYYY-MM", amount >= 0, unik per payment.
//   - DuePayment.CarryoverIR adalah CACHE = SUM(amount), dihitung server.
type DuePaymentCarryover struct {
	CarryoverID string
	PaymentID   string
	Month       string
	Amount      float64
	CreatedAt   time.Time
}

type DuePaymentCarryoverDTO struct {
	Month  string  `json:"month"`
	Amount float64 `json:"amount"`
}

type DuePayment struct {
	PaymentID      string
	GroupID        *string
	MemberID       string
	PaymentDate    time.Time
	TotalAmount    float64
	CarryoverIR    float64
	Carryovers     []DuePaymentCarryover
	ConnectingFund float64
	CommunityDues  float64
	OutreachFund   float64
	ThousandFund   float64
	FuneralFund    float64
	UkhroMT        float64
	Notes          string
	Status         string
	CreatedBy      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type DuePaymentDTO struct {
	PaymentID       string                   `json:"payment_id"`
	GroupID         string                   `json:"group_id"`
	MemberID        string                   `json:"member_id"`
	PaymentDate     string                   `json:"payment_date"`
	TotalAmount     float64                  `json:"total_amount"`
	CarryoverIR     float64                  `json:"carryover_ir"`
	CarryoverMonths []string                 `json:"carryover_months"`
	CarryoverItems  []DuePaymentCarryoverDTO `json:"carryover_items"`
	ConnectingFund  float64                  `json:"connecting_fund"`
	CommunityDues   float64                  `json:"community_dues"`
	OutreachFund    float64                  `json:"outreach_fund"`
	ThousandFund    float64                  `json:"thousand_fund"`
	FuneralFund     float64                  `json:"funeral_fund"`
	UkhroMT         float64                  `json:"ukhro_mt"`
	Notes           string                   `json:"notes"`
	Status          string                   `json:"status"`
	UpdatedAt       string                   `json:"updated_at"`
}

type DuesDashboardDTO struct {
	Target      float64 `json:"target"`
	Received    float64 `json:"received"`
	PaidCount   int     `json:"paidCount"`
	UnpaidCount int     `json:"unpaidCount"`
	MemberCount int     `json:"memberCount"`
}

type DuesDataDTO struct {
	SelectedMonth string           `json:"selected_month"`
	Members       []DueMemberDTO   `json:"members"`
	Payments      []DuePaymentDTO  `json:"payments"`
	Dashboard     DuesDashboardDTO `json:"dashboard"`
}

type ZakatRecord struct {
	ZakatID   string
	GroupID   *string
	Title     string
	Description string
	Location    string
	// Tipe zakat tidak lagi di header: tiap muzakki/mustahik/alokasi
	// membawa zakat_category sendiri (multi-tipe per record).
	SoulCount       int
	TotalRiceKg     float64
	TotalMoneyRp    float64
	Status          string
	TransactionDate *time.Time
	CompletedAt     *time.Time
	DeletedAt       *time.Time
	Version         int
	CreatedBy       string
	UpdatedBy       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ZakatPayer struct {
	PayerID            string
	ZakatID            string
	MasterID           *string
	Name               string
	Amount             float64
	ZakatCategory      string
	FamilyMembersCount int
	SortOrder          int
	CreatedAt          time.Time
}

type ZakatPayerDTO struct {
	PayerID            string  `json:"payer_id"`
	MasterID           string  `json:"master_id,omitempty"`
	Name               string  `json:"name"`
	Amount             float64 `json:"amount"`
	ZakatCategory      string  `json:"zakat_category"`
	FamilyMembersCount int     `json:"family_members_count"`
	SortOrder          int     `json:"sort_order"`
}

type ZakatRecipient struct {
	RecipientID   string
	ZakatID       string
	MasterID      *string
	Name          string
	Amount        float64
	ZakatCategory string
	SortOrder     int
	CreatedAt     time.Time
}

type ZakatRecipientDTO struct {
	RecipientID   string  `json:"recipient_id"`
	MasterID      string  `json:"master_id,omitempty"`
	Name          string  `json:"name"`
	Amount        float64 `json:"amount"`
	ZakatCategory string  `json:"zakat_category"`
	SortOrder     int     `json:"sort_order"`
}

type ZakatAllocation struct {
	ZakatID                string
	Category               string
	RecipientPercent       int
	RecipientAmount        float64
	RecipientGroupPercent  int
	RecipientGroupAmount   float64
	RecipientRegionPercent int
	RecipientRegionAmount  float64
	SabilillahPercent      int
	SabilillahAmount       float64
	AmilPercent            int
	AmilAmount             float64
	AmilGroupPercent       int
	AmilGroupAmount        float64
	AmilVillagePercent     int
	AmilVillageAmount      float64
	AmilRegionPercent      int
	AmilRegionAmount       float64
}

type ZakatAllocationGroupDTO struct {
	Percent int                      `json:"percent"`
	Amount  float64                  `json:"amount"`
	Group   *ZakatAllocationGroupDTO `json:"group,omitempty"`
	Region  *ZakatAllocationGroupDTO `json:"region,omitempty"`
	Village *ZakatAllocationGroupDTO `json:"village,omitempty"`
}

type ZakatAllocationCategoryDTO struct {
	Total      float64                 `json:"total"`
	Recipient  ZakatAllocationGroupDTO `json:"recipient"`
	Sabilillah ZakatAllocationGroupDTO `json:"sabilillah"`
	Amil       ZakatAllocationGroupDTO `json:"amil"`
}

type ZakatAllocationsDTO struct {
	Fitrah *ZakatAllocationCategoryDTO `json:"fitrah"`
	Maal   *ZakatAllocationCategoryDTO `json:"maal"`
	// Rincian per kategori untuk semua tipe (FITRAH/MAL/TIJAROH/ZURU/...).
	ByCategory map[string]*ZakatAllocationCategoryDTO `json:"by_category,omitempty"`
}

type MasterEntry struct {
	MasterID string `json:"master_id"`
	GroupID  string `json:"group_id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
}

type ZakatRecordDTO struct {
	ZakatID     string `json:"zakat_id"`
	GroupID     string `json:"group_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location"`
	// Kategori yang hadir di record (dari muzakki/mustahik), mis. ["FITRAH","MAL"].
	Categories    []string `json:"categories,omitempty"`
	SoulCount     int      `json:"soul_count"`     // cache
	TotalRiceKg   float64  `json:"total_rice_kg"`  // atribut header
	TotalMoneyRp  float64  `json:"total_money_rp"` // cache
	Status          string  `json:"status"`
	TransactionDate string  `json:"transaction_date"`
	CompletedAt     string  `json:"completed_at,omitempty"`
	Version         int     `json:"version"`
	UpdatedAt       string  `json:"updated_at"`
	UpdatedBy       string  `json:"updated_by"`

	PayerCount     int `json:"payer_count,omitempty"`
	RecipientCount int `json:"recipient_count,omitempty"`

	PayerList     []ZakatPayerDTO      `json:"payer_list,omitempty"`
	RecipientList []ZakatRecipientDTO  `json:"recipient_list,omitempty"`
	Allocations   *ZakatAllocationsDTO `json:"allocations,omitempty"`

	// Backward compat fields if needed
	MuzakkiList  []interface{} `json:"muzakki_list,omitempty"`
	MustahikList []interface{} `json:"mustahik_list,omitempty"`
}
