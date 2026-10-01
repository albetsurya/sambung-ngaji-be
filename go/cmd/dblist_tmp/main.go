// Temporary lister (hapus setelah dipakai).
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Println("pool gagal:", err)
		os.Exit(1)
	}
	defer pool.Close()
	rows, err := pool.Query(ctx, `SELECT group_id, group_name FROM groups ORDER BY group_name`)
	if err != nil {
		fmt.Println("query gagal:", err)
		os.Exit(1)
	}
	defer rows.Close()
	for rows.Next() {
		var id, name string
		_ = rows.Scan(&id, &name)
		fmt.Printf("%s | %s\n", id, name)
	}
}
