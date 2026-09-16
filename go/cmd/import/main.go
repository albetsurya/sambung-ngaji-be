package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var (
	dryRun   = flag.Bool("dry-run", true, "Validasi tanpa insert ke DB")
	only     = flag.String("only", "", "Hanya import sheet tertentu")
	truncate = flag.Bool("truncate", false, "TRUNCATE table sebelum insert")
	dataDir  = flag.String("data-dir", "data-csv", "Folder berisi CSV")
)

type sheetStep struct {
	name string
	fn   func(context.Context, *pgxpool.Pool, []map[string]string) error
}

func main() {
	flag.Parse()
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL tidak diset")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	mode := "DRY-RUN"
	if !*dryRun {
		mode = "IMPORT"
	}
	fmt.Printf("Mode: %s\nData dir: %s\n\n", mode, *dataDir)

	steps := []sheetStep{
		{"settings", importSettings},
		{"announcement_templates", importAnnouncementTemplates},
		{"members", importMembers},
		{"groups", importGroups},
		{"users", importUsers},
		{"meetings", importMeetings},
		{"attendance", importAttendance},
		{"monitoring", importMonitoring},
		{"announcements", importAnnouncements},
		{"pending_members", importPendingMembers},
		{"audit_logs", importAuditLogs},
		{"ai_usage", importAiUsage},
		{"wa_queue", importWaQueue},
	}

	var totalRows int
	for _, step := range steps {
		if *only != "" && *only != step.name {
			continue
		}
		path := dataDir2(*dataDir, step.name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("[SKIP] %-25s (file tidak ada)\n", step.name)
			continue
		}
		rows, err := readCSV(path)
		if err != nil {
			log.Fatalf("[%s] baca CSV: %v", step.name, err)
		}
		if len(rows) == 0 {
			fmt.Printf("[SKIP] %-25s (kosong)\n", step.name)
			continue
		}
		fmt.Printf("[RUN ] %-25s %d baris\n", step.name, len(rows))
		totalRows += len(rows)

		if *dryRun {
			continue
		}
		if *truncate {
			if _, err := pool.Exec(ctx, "TRUNCATE TABLE "+step.name+" CASCADE"); err != nil {
				log.Fatalf("[%s] truncate: %v", step.name, err)
			}
		}
		if err := step.fn(ctx, pool, rows); err != nil {
			log.Fatalf("[%s] %v", step.name, err)
		}
		fmt.Printf("[OK  ] %-25s\n", step.name)
	}

	fmt.Printf("\nTotal: %d baris\n", totalRows)
	if *dryRun {
		fmt.Println("\nDRY-RUN. Untuk insert: ulangi dengan -dry-run=false")
	}
}
