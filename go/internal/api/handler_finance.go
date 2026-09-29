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

func handleFinanceKasList(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	res, err := svc.KasList(c.Context(), groupID, BodyString(c, "kas_type"))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleFinanceKasSave(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	res, err := svc.KasSave(c.Context(), service.KasSaveInput{
		KasID:       BodyString(c, "kas_id"),
		GroupID:     groupID,
		KasType:     BodyString(c, "kas_type"),
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

func handleFinanceKasDelete(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	if err := svc.KasDelete(c.Context(), groupID, BodyString(c, "kas_id")); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"deleted": true})
}

func handleFinanceKasDuplicate(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	res, err := svc.KasDuplicate(c.Context(), groupID, BodyString(c, "kas_type"), BodyString(c, "kas_id"), financeUser(c))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleFinanceKasCarryForward(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	res, err := svc.KasCarryForward(c.Context(), groupID, BodyString(c, "kas_type"), BodyString(c, "month"), financeUser(c))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleFinanceShodaqohData(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	res, err := svc.ShodaqohData(c.Context(), groupID, BodyString(c, "month"))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleFinanceShodaqohMemberSave(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	res, err := svc.ShodaqohMemberSave(c.Context(), groupID,
		BodyString(c, "member_id"), BodyString(c, "member_name"), BodyFloat(c, "monthly_target"))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleFinanceShodaqohMemberDelete(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	if err := svc.ShodaqohMemberDelete(c.Context(), groupID, BodyString(c, "member_id")); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"deleted": true})
}

func handleFinanceShodaqohPaymentSave(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	res, err := svc.ShodaqohPaymentSave(c.Context(), service.ShodaqohPaymentInput{
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

func handleFinanceShodaqohPaymentReverse(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	if err := svc.ShodaqohPaymentReverse(c.Context(), groupID, BodyString(c, "payment_id")); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"reversed": true})
}

func handleFinanceShodaqohLastNominals(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	res, err := svc.ShodaqohLastNominals(c.Context(), groupID, BodyString(c, "member_id"))
	if err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, res)
}

func handleFinanceZakatList(c *fiber.Ctx, svc *service.FinanceService) error {
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

func handleFinanceZakatSave(c *fiber.Ctx, svc *service.FinanceService) error {
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

func handleFinanceZakatStatus(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	if err := svc.ZakatSetStatus(c.Context(), groupID, BodyString(c, "zakat_id"), BodyString(c, "status")); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"updated": true})
}

func handleFinanceZakatDelete(c *fiber.Ctx, svc *service.FinanceService) error {
	groupID, ok := financeGroup(c)
	if !ok {
		return nil
	}
	if err := svc.ZakatDelete(c.Context(), groupID, BodyString(c, "zakat_id")); err != nil {
		return Fail(c, err.Error())
	}
	return Ok(c, fiber.Map{"deleted": true})
}
