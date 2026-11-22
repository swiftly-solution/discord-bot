package store_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/swiftly-solution/discord-bot/store"
)

// Runs against TEST_DATABASE_URL in a TEMP table, so real rows are never touched.
func TestChannelRoles(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.MaxConns = 1 // the TEMP table only exists on one connection
	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	schema, err := os.ReadFile("../../schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, strings.Replace(string(schema), "CREATE TABLE", "CREATE TEMP TABLE", 1)); err != nil {
		t.Fatal(err)
	}
	st := store.New(db)

	step := func(name string, got bool, err error, want bool) {
		t.Helper()
		if err != nil || got != want {
			t.Fatalf("%s = %v, %v; want %v", name, got, err, want)
		}
	}
	added, err := st.Add(ctx, "g", "c", store.Honeypot)
	step("first Add", added, err, true)
	added, err = st.Add(ctx, "g", "c", store.Honeypot)
	step("second Add", added, err, false)
	added, err = st.Add(ctx, "g", "c", store.Forum)
	step("Add other role", added, err, true)

	ids, err := st.Channels(ctx, store.Honeypot)
	if err != nil || len(ids) != 1 || ids[0] != "c" {
		t.Fatalf("Channels = %v, %v", ids, err)
	}

	removed, err := st.Remove(ctx, "g", "c", store.Honeypot)
	step("first Remove", removed, err, true)
	removed, err = st.Remove(ctx, "g", "c", store.Honeypot)
	step("second Remove", removed, err, false)
	if ids, _ := st.Channels(ctx, store.Forum); len(ids) != 1 {
		t.Fatalf("Remove touched another role: %v", ids)
	}
}
