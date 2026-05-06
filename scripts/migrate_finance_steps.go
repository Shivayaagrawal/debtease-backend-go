// Run with: DB_URL=postgres://postgres:postgres@localhost:5432/debtease?sslmode=disable go run scripts/migrate_finance_steps.go
// Creates the finance_steps table if it does not exist. Requires calculation_sessions to exist (run other migrations first).
package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

const migration = `
CREATE TABLE IF NOT EXISTS finance_steps (
    session_id UUID PRIMARY KEY,
    monthly_income NUMERIC(18,2) NULL,
    fixed_obligations NUMERIC(18,2) NULL,
    loans JSONB NOT NULL DEFAULT '[]'::jsonb,
    engine_version TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_finance_steps_session
        FOREIGN KEY (session_id)
        REFERENCES calculation_sessions (session_id)
        ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_finance_steps_session_id ON finance_steps (session_id);
`

func main() {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		fmt.Fprintln(os.Stderr, "DB_URL is required")
		os.Exit(1)
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open:", err)
		os.Exit(1)
	}
	defer db.Close()
	if _, err := db.Exec(migration); err != nil {
		fmt.Fprintln(os.Stderr, "migrate:", err)
		os.Exit(1)
	}
	fmt.Println("finance_steps migration applied successfully")
}
