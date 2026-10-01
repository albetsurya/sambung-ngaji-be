// Scratch runner re-sync finance: backup sheet, apply migrasi, pull/sync per grup.
// Dijalankan manual oleh operator. HAPUS SEBELUM COMMIT bila tidak diperlukan.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/service"
)

func mustPool(ctx context.Context) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Println("pool gagal:", err)
		os.Exit(1)
	}
	if err := pool.Ping(ctx); err != nil {
		fmt.Println("ping gagal:", err)
		os.Exit(1)
	}
	return pool
}

func sheetClient(ctx context.Context) *sheets.Service {
	creds := os.Getenv("GOOGLE_SHEETS_CREDENTIALS")
	if creds == "" {
		fmt.Println("GOOGLE_SHEETS_CREDENTIALS kosong")
		os.Exit(1)
	}
	cli, err := sheets.NewService(ctx, option.WithCredentialsJSON([]byte(creds)))
	if err != nil {
		fmt.Println("sheets client gagal:", err)
		os.Exit(1)
	}
	return cli
}

func main() {
	ctx := context.Background()
	mode := os.Getenv("MODE")
	pool := mustPool(ctx)
	defer pool.Close()

	switch mode {
	case "tabs":
		cli := sheetClient(ctx)
		id := os.Getenv("FINANCE_SPREADSHEET_ID")
		meta, err := cli.Spreadsheets.Get(id).Fields("sheets.properties.title").Context(ctx).Do()
		if err != nil {
			fmt.Println("list tabs gagal:", err)
			os.Exit(1)
		}
		for _, sh := range meta.Sheets {
			fmt.Println("tab:", sh.Properties.Title)
		}
	case "dump":
		// MODE=dump TAB PATH — simpan seluruh isi tab ke JSON (analisis/backfill).
		cli := sheetClient(ctx)
		id := os.Getenv("FINANCE_SPREADSHEET_ID")
		tab := "Shodaqoh_Payments"
		if len(os.Args) > 1 {
			tab = os.Args[1]
		}
		path := os.Getenv("BACKUP_PATH")
		if path == "" {
			path = "/tmp/tab-dump.json"
		}
		resp, err := cli.Spreadsheets.Values.Get(id, tab+"!A:ZZ").Context(ctx).Do()
		if err != nil {
			fmt.Println("baca tab gagal:", err)
			os.Exit(1)
		}
		raw, _ := json.Marshal(resp.Values)
		if err := os.WriteFile(path, raw, 0600); err != nil {
			fmt.Println("tulis gagal:", err)
			os.Exit(1)
		}
		fmt.Printf("tab %s: %d baris -> %s\n", tab, len(resp.Values), path)
	case "read":
		cli := sheetClient(ctx)
		id := os.Getenv("FINANCE_SPREADSHEET_ID")
		tab := "Shodaqoh_Payments"
		if len(os.Args) > 1 {
			tab = os.Args[1]
		}
		resp, err := cli.Spreadsheets.Values.Get(id, tab+"!A:ZZ").Context(ctx).Do()
		if err != nil {
			fmt.Println("baca tab gagal:", err)
			os.Exit(1)
		}
		fmt.Printf("tab %s: %d baris\n", tab, len(resp.Values))
		for i, r := range resp.Values {
			if i > 6 {
				break
			}
			raw, _ := json.Marshal(r)
			fmt.Printf("row%d: %s\n", i+1, raw)
		}
	case "migrate":
		files := []string{
			"db/migrations/000023_backfill_group_id.up.sql",
			"db/migrations/000024_normalize_carryover.up.sql",
			"db/migrations/000025_zakat_relational.up.sql",
			"db/migrations/000026_finance_standard.up.sql",
		}
		for _, f := range files {
			sql, err := os.ReadFile(filepath.Join(".", f))
			if err != nil {
				// coba dari workdir backend/go sudah benar; laporkan saja
				fmt.Println("baca gagal:", f, err)
				os.Exit(1)
			}
			if _, err := pool.Exec(ctx, string(sql)); err != nil {
				fmt.Println("migrasi gagal:", f, err)
				os.Exit(1)
			}
			fmt.Println("OK:", f)
		}
	case "backup":
		cli := sheetClient(ctx)
		id := os.Getenv("FINANCE_SPREADSHEET_ID")
		out := map[string]interface{}{}
		for _, tab := range []string{"CASH", "DUE_MEMBERS", "DUE_PAYMENTS", "ZAKAT"} {
			resp, err := cli.Spreadsheets.Values.Get(id, tab+"!A:ZZ").Context(ctx).Do()
			if err != nil {
				fmt.Println("baca tab gagal:", tab, err)
				os.Exit(1)
			}
			out[tab] = resp.Values
			fmt.Printf("tab %s: %d baris\n", tab, len(resp.Values))
		}
		raw, _ := json.MarshalIndent(out, "", " ")
		path := os.Getenv("BACKUP_PATH")
		if path == "" {
			path = "/tmp/sheet-backup.json"
		}
		if err := os.WriteFile(path, raw, 0600); err != nil {
			fmt.Println("tulis backup gagal:", err)
			os.Exit(1)
		}
		fmt.Println("backup tersimpan:", path)
	case "reset":
		// MODE=reset GROUP_ID + CONFIRM=RESET — hapus SELURUH data finance
		// satu grup (payments, members, carryovers, tombstones, errors)
		// untuk import ulang bersih. TIDAK menyentuh tabel non-finance.
		if os.Getenv("CONFIRM") != "RESET" {
			fmt.Println("tambah CONFIRM=RESET untuk eksekusi")
			os.Exit(1)
		}
		gid := ""
		if len(os.Args) > 1 {
			gid = os.Args[1]
		}
		if gid == "" {
			fmt.Println("GROUP_ID wajib diisi")
			os.Exit(1)
		}
		for _, q := range []string{
			`DELETE FROM due_payment_carryovers WHERE payment_id IN (SELECT payment_id FROM due_payments WHERE group_id=$1)`,
			`DELETE FROM due_payments WHERE group_id=$1`,
			`DELETE FROM due_members WHERE group_id=$1`,
			`DELETE FROM cash_transactions WHERE group_id=$1`,
			`DELETE FROM zakat_records WHERE group_id=$1`,
			`DELETE FROM finance_sync_deleted WHERE group_id=$1`,
			`DELETE FROM finance_sync_errors WHERE group_id=$1`,
		} {
			if _, err := pool.Exec(ctx, q, gid); err != nil {
				fmt.Println("reset gagal:", err)
				os.Exit(1)
			}
		}
		fmt.Println("reset OK:", gid)
	case "errors":
		// MODE=errors GROUP_ID — 10 error sync terakhir per grup.
		gid := ""
		if len(os.Args) > 1 {
			gid = os.Args[1]
		}
		rows, _ := pool.Query(ctx, `SELECT entity, entity_id, direction, message, created_at::text
			FROM finance_sync_errors WHERE group_id=$1 ORDER BY created_at DESC LIMIT 10`, gid)
		for rows.Next() {
			var e, id, d, m, t string
			rows.Scan(&e, &id, &d, &m, &t)
			fmt.Printf("[%s] %s %s: %s\n", d, e, id, m)
		}
		rows.Close()
	case "stats":
		// MODE=stats GROUP_ID — ringkasan hasil import per grup.
		gid := ""
		if len(os.Args) > 1 {
			gid = os.Args[1]
		}
		var npay, nmem, ncarry, nerr int
		_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM due_payments WHERE group_id=$1`, gid).Scan(&npay)
		_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM due_members WHERE group_id=$1`, gid).Scan(&nmem)
		_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM due_payment_carryovers c JOIN due_payments p USING (payment_id) WHERE p.group_id=$1`, gid).Scan(&ncarry)
		_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM finance_sync_errors WHERE group_id=$1`, gid).Scan(&nerr)
		fmt.Printf("group=%s payments=%d members=%d carryovers=%d sync_errors=%d", gid, npay, nmem, ncarry, nerr)
		var ncash, nzakat int
		_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM cash_transactions WHERE group_id=$1`, gid).Scan(&ncash)
		_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM zakat_records WHERE group_id=$1`, gid).Scan(&nzakat)
		fmt.Printf(" cash=%d zakat=%d\n", ncash, nzakat)
		rows, _ := pool.Query(ctx, `SELECT payment_date::text, COUNT(*) FROM due_payments WHERE group_id=$1 GROUP BY 1 ORDER BY 1 LIMIT 3`, gid)
		for rows.Next() {
			var d string
			var n int
			rows.Scan(&d, &n)
			fmt.Println("  oldest:", d, n)
		}
		rows.Close()
		rows, _ = pool.Query(ctx, `SELECT payment_date::text, COUNT(*) FROM due_payments WHERE group_id=$1 GROUP BY 1 ORDER BY 1 DESC LIMIT 3`, gid)
		for rows.Next() {
			var d string
			var n int
			rows.Scan(&d, &n)
			fmt.Println("  newest:", d, n)
		}
		rows.Close()
		rows, _ = pool.Query(ctx, `SELECT status, COUNT(*) FROM due_payments WHERE group_id=$1 GROUP BY 1`, gid)
		for rows.Next() {
			var s string
			var n int
			rows.Scan(&s, &n)
			fmt.Println("  status:", s, n)
		}
		rows.Close()
		// Sampel carryover + cek konsistensi SUM(anak) = carryover_ir
		var mismatch int
		_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM due_payments p WHERE group_id=$1 AND carryover_ir <>
			COALESCE((SELECT SUM(amount) FROM due_payment_carryovers WHERE payment_id=p.payment_id),0)`, gid).Scan(&mismatch)
		fmt.Println("  carryover_ir != sum(anak):", mismatch)
	case "get":
		// MODE=get PAYID... — tampilkan tanggal + total + carryover satu payment.
		// MODE=get cash:CASHID — tampilkan satu transaksi kas.
		for _, pid := range os.Args[1:] {
			if strings.HasPrefix(pid, "cash:") {
				id := strings.TrimPrefix(pid, "cash:")
				var d, acc, desc string
				var deb, cred float64
				_ = pool.QueryRow(ctx, `SELECT tanggal::text, account_name, description, debit, credit
					FROM cash_transactions WHERE cash_id=$1`, id).Scan(&d, &acc, &desc, &deb, &cred)
				fmt.Printf("%s date=%s acc=%s desc=%s deb=%v cred=%v\n", id, d, acc, desc, deb, cred)
				continue
			}
			var pdate, status string
			var total, ir float64
			_ = pool.QueryRow(ctx, `SELECT payment_date::text, total_amount, carryover_ir, status
				FROM due_payments WHERE payment_id=$1`, pid).Scan(&pdate, &total, &ir, &status)
			fmt.Printf("%s date=%s total=%v ir=%v status=%s\n", pid, pdate, total, ir, status)
			crows, _ := pool.Query(ctx, `SELECT month, amount FROM due_payment_carryovers WHERE payment_id=$1 ORDER BY month`, pid)
			for crows.Next() {
				var m string
				var a float64
				crows.Scan(&m, &a)
				fmt.Printf("    %s: %v\n", m, a)
			}
			crows.Close()
		}
	case "jul31":
		// MODE=jul31 [GROUP] — audit baris tanggal 31 Juli yang deskripsi/
		// notes-nya menyebut Agustus (indikasi geser -1 hari, harusnya 1 Agu).
		// Bagian 1: semua baris 31 Juli. Bagian 2: semua baris ber-kata
		// Agustus dengan tanggal 25 Jul – 5 Agu (zona rawan geser).
		gid := ""
		if len(os.Args) > 1 {
			gid = os.Args[1]
		}
		gf := ""
		args := []interface{}{}
		if gid != "" {
			gf = " AND group_id=$1"
			args = append(args, gid)
		}
		fmt.Println("== semua baris 31 Juli ==")
		rows, _ := pool.Query(ctx, `SELECT 'cash' t, cash_id id, tanggal::text d, account_name || ' | ' || description txt
			FROM cash_transactions WHERE EXTRACT(DAY FROM tanggal)=31 AND EXTRACT(MONTH FROM tanggal)=7`+gf+` ORDER BY tanggal`, args...)
		for rows.Next() {
			var t, id, d, txt string
			rows.Scan(&t, &id, &d, &txt)
			fmt.Printf("[%s] %s date=%s %s\n", t, id, d, txt)
		}
		rows.Close()
		rows, _ = pool.Query(ctx, `SELECT 'due' t, payment_id id, payment_date::text d, notes txt
			FROM due_payments WHERE EXTRACT(DAY FROM payment_date)=31 AND EXTRACT(MONTH FROM payment_date)=7`+gf+` ORDER BY payment_date`, args...)
		for rows.Next() {
			var t, id, d, txt string
			rows.Scan(&t, &id, &d, &txt)
			fmt.Printf("[%s] %s date=%s notes=%s\n", t, id, d, txt)
		}
		rows.Close()
		fmt.Println("== baris ber-kata Agustus, tanggal 25 Jul–5 Agu ==")
		rows, _ = pool.Query(ctx, `SELECT 'cash' t, cash_id id, tanggal::text d, account_name || ' | ' || description txt
			FROM cash_transactions WHERE (description ILIKE '%agus%' OR account_name ILIKE '%agus%')
			AND ((EXTRACT(MONTH FROM tanggal)=7 AND EXTRACT(DAY FROM tanggal)>=25) OR (EXTRACT(MONTH FROM tanggal)=8 AND EXTRACT(DAY FROM tanggal)<=5))`+gf+` ORDER BY tanggal`, args...)
		for rows.Next() {
			var t, id, d, txt string
			rows.Scan(&t, &id, &d, &txt)
			fmt.Printf("[%s] %s date=%s %s\n", t, id, d, txt)
		}
		rows.Close()
		rows, _ = pool.Query(ctx, `SELECT 'due' t, payment_id id, payment_date::text d, notes txt
			FROM due_payments WHERE notes ILIKE '%agus%'
			AND ((EXTRACT(MONTH FROM payment_date)=7 AND EXTRACT(DAY FROM payment_date)>=25) OR (EXTRACT(MONTH FROM payment_date)=8 AND EXTRACT(DAY FROM payment_date)<=5))`+gf+` ORDER BY payment_date`, args...)
		for rows.Next() {
			var t, id, d, txt string
			rows.Scan(&t, &id, &d, &txt)
			fmt.Printf("[%s] %s date=%s notes=%s\n", t, id, d, txt)
		}
		rows.Close()
	case "pull", "sync", "dupes":
		// MODE=pull GROUP... — sheet->DB saja (aman, tanpa push balik).
		// MODE=sync GROUP... — dua arah + tulis ulang tab mirror.
		// MODE=dupes [PAYID...] — cek tanggal DB untuk payment tertentu.
		finRepo := repository.NewFinanceRepo(pool)
		groups := repository.NewGroupRepo(pool)
		finSvc := service.NewFinanceService(finRepo)
		syncSvc := service.NewFinanceSyncService(finRepo, groups, finSvc)
		if !syncSvc.IsConfigured() {
			fmt.Println("sheet sync belum dikonfigurasi")
			os.Exit(1)
		}
		target := os.Args[1:]
		var groupIDs []string
		if len(target) == 0 || target[0] == "all" {
			gs, err := groups.FindAll(ctx, false)
			if err != nil {
				fmt.Println("list grup gagal:", err)
				os.Exit(1)
			}
			for _, g := range gs {
				groupIDs = append(groupIDs, g.GroupID)
			}
		} else {
			groupIDs = target
		}
		for _, gid := range groupIDs {
			var err error
			if mode == "pull" {
				err = syncSvc.PullGroup(ctx, gid)
			} else {
				err = syncSvc.SyncGroup(ctx, gid)
			}
			fmt.Printf("%s %s: err=%v\n", mode, gid, err)
		}
	default:
		fmt.Println("MODE=migrate|backup|pull|sync (arg: group_id...|all)")
		os.Exit(1)
	}
}
