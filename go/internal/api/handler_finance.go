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
		_ = Fail(c, "Pilih kelompok dulu (group_id wajib diisi)")
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
		Tanggal:     BodyString(c, "tanggal"),
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
		PaymentID:          BodyString(c, "payment_id"),
		GroupID:            groupID,
		MemberID:           BodyString(c, "member_id"),
		PaymentDate:        BodyString(c, "payment_date"),
		CarryoverIR:        BodyFloat(c, "carryover_ir"),
		CarryoverMonths:    BodyString(c, "carryover_months"),
		CarryoverBreakdown: BodyString(c, "carryover_breakdown"),
		ConnectingFund:     BodyFloat(c, "connecting_fund"),
		CommunityDues:      BodyFloat(c, "community_dues"),
		OutreachFund:       BodyFloat(c, "outreach_fund"),
		ThousandFund:       BodyFloat(c, "thousand_fund"),
		FuneralFund:        BodyFloat(c, "funeral_fund"),
		UkhroMT:            BodyFloat(c, "ukhro_mt"),
		Notes:              BodyString(c, "notes"),
		CreatedBy:          financeUser(c),
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

func handleZakatSave(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	res, err := svc.ZakatSave(c.Context(), service.ZakatSaveInput{
		ZakatID:         BodyString(c, "zakat_id"),
		GroupID:         groupID,
		ZakatType:       BodyString(c, "zakat_type"),
		MuzakkiName:     BodyString(c, "muzakki_name"),
		SoulCount:       int(BodyFloat(c, "soul_count")),
		TotalRiceKg:     BodyFloat(c, "total_rice_kg"),
		TotalMoneyRp:    BodyFloat(c, "total_money_rp"),
		TransactionDate: BodyString(c, "transaction_date"),
		Details:         BodyString(c, "details"),
		CreatedBy:       financeUser(c),
	})
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
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
