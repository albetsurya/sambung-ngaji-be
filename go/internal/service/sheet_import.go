package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/util"
)

type gasImportResult struct {
	Imported map[string]int `json:"imported"`
}

// ImportFromGAS pulls legacy Apps Script data into one group (one-time migration).
func (s *FinanceSyncService) ImportFromGAS(ctx context.Context, groupID string) (*gasImportResult, error) {
	base := strings.TrimRight(os.Getenv("GAS_IMPORT_URL"), "/")
	if base == "" {
		return nil, fmt.Errorf("GAS_IMPORT_URL belum dikonfigurasi")
	}
	res := &gasImportResult{Imported: map[string]int{}}
	get := func(action, kasType string) (map[string]interface{}, error) {
		url := base + "?action=" + action
		if kasType != "" {
			url += "&kasType=" + kasType
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		httpClient := &http.Client{Timeout: 60 * time.Second}
		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		raw, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		var out map[string]interface{}
		if err := json.Unmarshal(raw, &out); err != nil {
			return nil, fmt.Errorf("respons GAS bukan JSON valid untuk %s", action)
		}
		return out, nil
	}

	gid := groupID
	str := func(m map[string]interface{}, keys ...string) string {
		for _, k := range keys {
			if v, ok := m[k]; ok && v != nil {
				if s, ok := v.(string); ok && s != "" {
					return s
				}
			}
		}
		return ""
	}
	num := func(m map[string]interface{}, keys ...string) float64 {
		for _, k := range keys {
			if v, ok := m[k]; ok && v != nil {
				switch t := v.(type) {
				case float64:
					return t
				case string:
					return parseNum(t)
				}
			}
		}
		return 0
	}
	mark := func(table, idCol, id string) {
		_ = s.repo.MarkSynced(ctx, table, idCol, id, "sheet-import", 0)
	}

	for _, kt := range []string{"main", "kas_amil"} {
		out, err := get("getData", kt)
		if err != nil {
			s.recordError(ctx, groupID, "cash", kt, "gas-import", err.Error())
			continue
		}
		txs, _ := out["transactions"].([]interface{})
		for _, raw := range txs {
			m, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			tgl, err := util.ParseSheetDate(str(m, "tanggal", "transaction_date"))
			if err != nil {
				continue
			}
			k := &model.CashTransaction{
				CashID: util.NewID("KAS"), GroupID: &gid, CashType: normCashType(kt),
				Tanggal:     tgl,
				AccountName: str(m, "account", "account_name"),
				Description: str(m, "keterangan", "description"),
				Debit:       num(m, "debet", "debit"),
				Credit:      num(m, "kredit", "credit"),
				CreatedBy:   str(m, "createdBy", "created_by"),
			}
			if err := s.repo.CashInsert(ctx, k); err != nil {
				s.recordError(ctx, groupID, "cash", k.CashID, "gas-import", err.Error())
				continue
			}
			mark("cash_transactions", "cash_id", k.CashID)
			res.Imported["cash_"+kt]++
		}
	}

	sho, err := get("getShodaqohData", "")
	if err != nil {
		s.recordError(ctx, groupID, "due", "", "gas-import", err.Error())
	} else {
		if arr, ok := sho["members"].([]interface{}); ok {
			for _, raw := range arr {
				m, ok := raw.(map[string]interface{})
				if !ok {
					continue
				}
				id := str(m, "member_id", "memberId")
				name := str(m, "member_name", "nama")
				if id == "" || name == "" {
					continue
				}
				dm := &model.DueMember{
					MemberID: id, GroupID: &gid, MemberName: name,
					MonthlyTarget: num(m, "monthly_target", "nominal_bulanan", "nominalBulanan"),
					Status:        "ACTIVE",
				}
				if err := s.repo.DueMemberUpsert(ctx, dm); err != nil {
					s.recordError(ctx, groupID, "due_members", id, "gas-import", err.Error())
					continue
				}
				mark("due_members", "due_member_id", id)
				res.Imported["due_members"]++
			}
		}
		if arr, ok := sho["payments"].([]interface{}); ok {
			for _, raw := range arr {
				m, ok := raw.(map[string]interface{})
				if !ok {
					continue
				}
				id := str(m, "payment_id", "paymentId")
				mid := str(m, "member_id", "memberId")
				if id == "" || mid == "" {
					continue
				}
				tgl, err := util.ParseSheetDate(str(m, "payment_date", "tanggal"))
				if err != nil {
					continue
				}
				status := strings.ToUpper(str(m, "status"))
				if status == "" {
					status = "ACTIVE"
				}
				months, unknowns := util.SplitSheetMonths(str(m, "carryover_months", "susulan_bulan"))
				if len(unknowns) > 0 {
					s.recordError(ctx, groupID, "due_payments", id, "gas-import", "bulan susulan tidak dikenal: "+strings.Join(unknowns, ", "))
					continue
				}
				sheetIR := num(m, "carryover_ir", "susulan_ir")
				carryItems, carryIR, notes := buildSheetCarryovers(months, sheetIR,
					str(m, "carryover_breakdown", "susulan_rincian"), str(m, "notes", "keterangan"))
				total := carryIR + num(m, "connecting_fund", "uang_sambung", "uang") +
					num(m, "community_dues", "jimpitan") + num(m, "outreach_fund", "siar_siar") +
					num(m, "thousand_fund", "seribuan") + num(m, "funeral_fund", "kafan") +
					num(m, "ukhro_mt")
				if total == 0 {
					total = num(m, "total_amount", "total")
				}
				p := &model.DuePayment{
					PaymentID: id, GroupID: &gid, MemberID: mid, PaymentDate: tgl,
					TotalAmount:    total,
					CarryoverIR:    carryIR,
					Carryovers:     carryItems,
					ConnectingFund: num(m, "connecting_fund", "uang_sambung", "uang"),
					CommunityDues:  num(m, "community_dues", "jimpitan"),
					OutreachFund:   num(m, "outreach_fund", "siar_siar"),
					ThousandFund:   num(m, "thousand_fund", "seribuan"),
					FuneralFund:    num(m, "funeral_fund", "kafan"),
					UkhroMT:        num(m, "ukhro_mt"),
					Notes:          notes, Status: status,
				}
				if err := s.repo.DuePaymentUpsert(ctx, p); err != nil {
					s.recordError(ctx, groupID, "due_payments", id, "gas-import", err.Error())
					continue
				}
				if err := s.repo.ReplaceCarryovers(ctx, id, carryItems); err != nil {
					s.recordError(ctx, groupID, "due_payments", id, "gas-import", "gagal simpan rincian susulan: "+err.Error())
					continue
				}
				mark("due_payments", "payment_id", id)
				res.Imported["due_payments"]++
			}
		}
	}

	zak, err := get("getZakatList", "")
	if err != nil {
		s.recordError(ctx, groupID, "zakat", "", "gas-import", err.Error())
	} else {
		if arr, ok := zak["data"].([]interface{}); ok {
			for _, raw := range arr {
				m, ok := raw.(map[string]interface{})
				if !ok {
					continue
				}
				id := str(m, "id")
				if id == "" {
					id = util.NewID("ZKT")
				}
				det, _ := json.Marshal(m)
				status := strings.ToUpper(str(m, "status"))
				if status == "" {
					status = "PENDING"
				}
				var tgl *time.Time
				if ds := str(m, "transaction_date"); ds != "" {
					if t, err := util.ParseSheetDate(ds); err == nil {
						tgl = &t
					}
				}
				z := &model.ZakatRecord{
					ZakatID: id, GroupID: &gid, ZakatType: "FITRAH",
					MuzakkiName: str(m, "title"), SoulCount: 1,
					TotalRiceKg: 0, TotalMoneyRp: num(m, "total"),
					Status: status, TransactionDate: tgl, Details: string(det),
				}
				if err := s.repo.ZakatUpsert(ctx, z); err != nil {
					s.recordError(ctx, groupID, "zakat", id, "gas-import", err.Error())
					continue
				}
				mark("zakat_records", "zakat_id", id)
				res.Imported["zakat"]++
			}
		}
	}
	return res, nil
}
