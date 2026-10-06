package api

import (
	"github.com/gofiber/fiber/v2"

	"pengajian-backend/internal/service"
)

func financeGroup(c *fiber.Ctx) (string, bool) {
	groupID := BodyString(c, "group_id")
	if g, isSuper := ActorOf(c); !isSuper {
		if g == "" || g == UnassignedGroup {
			_ = Fail(c, "Akun Anda belum dipetakan ke kelompok")
			return "", false
		}
		return g, true
	}
	if groupID == "" {
		_ = Fail(c, "Pilih group_label dulu (group_id wajib diisi)")
		return "", false
	}
	return groupID, true
}

func financeUser(c *fiber.Ctx) string {
	if u := UserOf(c); u != nil {
		return u.Username
	}
	return ""
}

// bodyCarryoverItems membaca rincian susulan terstruktur dari body:
// carryover_items: [{month: "YYYY-MM", amount: number}].
// Melewatkan validasi ringan; validasi penuh di service.normalizeCarryovers.
func bodyCarryoverItems(c *fiber.Ctx) []service.DuePaymentCarryoverInput {
	body := BodyOf(c)
	raw, ok := body["carryover_items"]
	if !ok || raw == nil {
		return nil
	}
	items, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	out := make([]service.DuePaymentCarryoverInput, 0, len(items))
	for _, it := range items {
		m, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		month, _ := m["month"].(string)
		var amount float64
		switch v := m["amount"].(type) {
		case float64:
			amount = v
		case float32:
			amount = float64(v)
		case int:
			amount = float64(v)
		case int64:
			amount = float64(v)
		}
		if month == "" && amount == 0 {
			continue
		}
		out = append(out, service.DuePaymentCarryoverInput{Month: month, Amount: amount})
	}
	return out
}

func handleCashLedger(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	res, err := svc.CashList(c.Context(), groupID, BodyString(c, "cash_type"))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleCashSave(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	res, err := svc.CashSave(c.Context(), service.CashSaveInput{
		CashID:      BodyString(c, "cash_id"),
		GroupID:     groupID,
		CashType:    BodyString(c, "cash_type"),
		Date:        BodyString(c, "date"),
		AccountName: BodyString(c, "account_name"),
		Description: BodyString(c, "description"),
		Debit:       BodyFloat(c, "debit"),
		Credit:      BodyFloat(c, "credit"),
		CreatedBy:   financeUser(c),
	})
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleCashDelete(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	if err := svc.CashDelete(c.Context(), groupID, BodyString(c, "cash_id")); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"deleted": true})
}

func handleCashDuplicate(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	res, err := svc.CashDuplicate(c.Context(), groupID, BodyString(c, "cash_type"), BodyString(c, "cash_id"), financeUser(c))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleCashCarryForward(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	res, err := svc.CashCarryForward(c.Context(), groupID, BodyString(c, "cash_type"), BodyString(c, "month"), financeUser(c))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleDuesData(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	res, err := svc.DuesData(c.Context(), groupID, BodyString(c, "month"))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleDueMemberSave(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	res, err := svc.DueMemberSave(c.Context(), groupID,
		BodyString(c, "member_id"), BodyString(c, "member_name"), BodyFloat(c, "monthly_target"))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleDueMemberDelete(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	if err := svc.DueMemberDelete(c.Context(), groupID, BodyString(c, "member_id")); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"deleted": true})
}

func handleDuePaymentSave(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	res, err := svc.DuePaymentSave(c.Context(), service.DuePaymentInput{
		PaymentID:       BodyString(c, "payment_id"),
		GroupID:         groupID,
		MemberID:        BodyString(c, "member_id"),
		PaymentDate:     BodyString(c, "payment_date"),
		CarryoverItems:  bodyCarryoverItems(c),
		CarryoverMonths: BodyString(c, "carryover_months"),
		CarryoverIR:     BodyFloat(c, "carryover_ir"),
		ConnectingFund:  BodyFloat(c, "connecting_fund"),
		CommunityDues:   BodyFloat(c, "community_dues"),
		OutreachFund:    BodyFloat(c, "outreach_fund"),
		ThousandFund:    BodyFloat(c, "thousand_fund"),
		FuneralFund:     BodyFloat(c, "funeral_fund"),
		UkhroMT:         BodyFloat(c, "ukhro_mt"),
		Notes:           BodyString(c, "notes"),
		CreatedBy:       financeUser(c),
	})
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleDuePaymentReverse(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	if err := svc.DuePaymentReverse(c.Context(), groupID, BodyString(c, "payment_id")); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"reversed": true})
}

func handleDueLastNominals(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	res, err := svc.DueLastNominals(c.Context(), groupID, BodyString(c, "member_id"))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleDuePostToCash(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	res, err := svc.DuePostToCash(c.Context(), groupID, BodyString(c, "month"), financeUser(c))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"posted": len(res), "items": res})
}

func handleDueCancelPostToCash(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	n, err := svc.DueCancelPostToCash(c.Context(), groupID, BodyString(c, "month"))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"cancelled": n})
}

func handleZakatList(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	res, err := svc.ZakatList(c.Context(), groupID)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleZakatDetail(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	res, err := svc.ZakatDetail(c.Context(), groupID, BodyString(c, "zakat_id"))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleZakatSave(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	body := BodyOf(c)
	res, err := svc.ZakatSave(c.Context(), groupID, service.ZakatHeaderInput{
		ZakatID:         BodyString(c, "zakat_id"),
		Title:           BodyString(c, "title"),
		Description:     BodyString(c, "description"),
		Location:        BodyString(c, "location"),
		SoulCount:       int(numOr(body["soul_count"])),
		TotalRiceKg:     numOr(body["total_rice_kg"]),
		TotalMoneyRp:    numOr(body["total_money_rp"]),
		TransactionDate: BodyString(c, "transaction_date"),
	}, financeUser(c))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func bodyZakatPayers(c *fiber.Ctx) []service.ZakatPayerInput {
	body := BodyOf(c)
	raw, _ := body["payers"].([]interface{})
	out := make([]service.ZakatPayerInput, 0, len(raw))
	for _, it := range raw {
		m, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		out = append(out, service.ZakatPayerInput{
			PayerID:            strOr(m["payer_id"]),
			MasterID:           strOr(m["master_id"]),
			Name:               strOr(m["name"]),
			Amount:             numOr(m["amount"]),
			ZakatCategory:      strOr(m["zakat_category"]),
			FamilyMembersCount: int(numOr(m["family_members_count"])),
		})
	}
	return out
}

func handleZakatSavePayers(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	if err := svc.ZakatSavePayers(c.Context(), groupID, BodyString(c, "zakat_id"), bodyZakatPayers(c)); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"saved": true})
}

func bodyZakatRecipients(c *fiber.Ctx) []service.ZakatRecipientInput {
	body := BodyOf(c)
	raw, _ := body["recipients"].([]interface{})
	out := make([]service.ZakatRecipientInput, 0, len(raw))
	for _, it := range raw {
		m, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		out = append(out, service.ZakatRecipientInput{
			RecipientID:   strOr(m["recipient_id"]),
			MasterID:      strOr(m["master_id"]),
			Name:          strOr(m["name"]),
			Amount:        numOr(m["amount"]),
			ZakatCategory: strOr(m["zakat_category"]),
		})
	}
	return out
}

func handleZakatSaveRecipients(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	if err := svc.ZakatSaveRecipients(c.Context(), groupID, BodyString(c, "zakat_id"), bodyZakatRecipients(c)); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"saved": true})
}

func bodyZakatAllocations(c *fiber.Ctx) []service.ZakatAllocationInput {
	body := BodyOf(c)
	raw, _ := body["allocations"].([]interface{})
	out := make([]service.ZakatAllocationInput, 0, len(raw))
	for _, it := range raw {
		m, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		out = append(out, service.ZakatAllocationInput{
			Category:               strOr(m["category"]),
			RecipientPercent:       int(numOr(m["recipient_percent"])),
			RecipientAmount:        numOr(m["recipient_amount"]),
			RecipientGroupPercent:  int(numOr(m["recipient_group_percent"])),
			RecipientGroupAmount:   numOr(m["recipient_group_amount"]),
			RecipientRegionPercent: int(numOr(m["recipient_region_percent"])),
			RecipientRegionAmount:  numOr(m["recipient_region_amount"]),
			SabilillahPercent:      int(numOr(m["sabilillah_percent"])),
			SabilillahAmount:       numOr(m["sabilillah_amount"]),
			AmilPercent:            int(numOr(m["amil_percent"])),
			AmilAmount:             numOr(m["amil_amount"]),
			AmilGroupPercent:       int(numOr(m["amil_group_percent"])),
			AmilGroupAmount:        numOr(m["amil_group_amount"]),
			AmilVillagePercent:     int(numOr(m["amil_village_percent"])),
			AmilVillageAmount:      numOr(m["amil_village_amount"]),
			AmilRegionPercent:      int(numOr(m["amil_region_percent"])),
			AmilRegionAmount:       numOr(m["amil_region_amount"]),
		})
	}
	return out
}

func handleZakatSaveAllocations(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	if err := svc.ZakatSaveAllocations(c.Context(), groupID, BodyString(c, "zakat_id"), bodyZakatAllocations(c)); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"saved": true})
}

func handleZakatMasters(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	res, err := svc.ZakatMasters(c.Context(), groupID)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleZakatAddMaster(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	id, err := svc.ZakatAddMaster(c.Context(), groupID, BodyString(c, "kind"), BodyString(c, "name"))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"master_id": id})
}

func handleZakatStatus(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	if err := svc.ZakatSetStatus(c.Context(), groupID, BodyString(c, "zakat_id"), BodyString(c, "status")); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"updated": true})
}

func handleZakatDelete(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	if err := svc.ZakatDelete(c.Context(), groupID, BodyString(c, "zakat_id")); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"deleted": true})
}

func strOr(v interface{}) string {
	s, _ := v.(string)
	return s
}

func numOr(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	default:
		return 0
	}
}

func handleFinanceSync(c *fiber.Ctx, svc *service.FinanceSyncService) error {
	groupID := BodyString(c, "group_id")
	if groupID != "" {
		if err := svc.SyncGroup(c.Context(), groupID); err != nil {
			return Fail(c, err.Error())
		}
		return Ok(c, fiber.Map{"synced_group": groupID})
	}
	if err := svc.SyncAllGroups(c.Context()); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"synced": "all"})
}

func handleFinanceImport(c *fiber.Ctx, svc *service.FinanceSyncService) error {
	groupID := BodyString(c, "group_id")
	if groupID == "" {
		return Fail(c, "group_id wajib diisi untuk import")
	}
	res, err := svc.ImportFromGAS(c.Context(), groupID)
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}
