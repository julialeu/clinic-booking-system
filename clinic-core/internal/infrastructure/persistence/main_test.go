package persistence

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("clinic_core_test"),
		tcpostgres.WithUsername("clinic"),
		tcpostgres.WithPassword("test_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("starting postgres container: %v", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("getting connection string: %v", err)
	}

	testPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("creating test pool: %v", err)
	}

	if err := applyMigrations(ctx); err != nil {
		log.Fatalf("applying migrations: %v", err)
	}

	code := m.Run()

	testPool.Close()
	if err := testcontainers.TerminateContainer(container); err != nil {
		log.Printf("terminating container: %v", err)
	}

	os.Exit(code)
}

func applyMigrations(ctx context.Context) error {
	path := filepath.Join("..", "..", "..", "migrations", "000001_create_appointments_table.up.sql")

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading migration file: %w", err)
	}

	if _, err := testPool.Exec(ctx, string(content)); err != nil {
		return fmt.Errorf("executing migration: %w", err)
	}
	return nil
}

func truncateAppointments(t *testing.T) {
	t.Helper()
	if _, err := testPool.Exec(context.Background(), "TRUNCATE appointments"); err != nil {
		t.Fatalf("truncating appointments: %v", err)
	}
}
