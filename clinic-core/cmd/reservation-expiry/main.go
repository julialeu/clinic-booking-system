package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/julialeu/clinic-booking-system/clinic-core/internal/application/command"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/infrastructure/persistence"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/platform/clock"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/platform/postgres"
)

const (
	defaultDSN      = "postgres://clinic:clinic_dev_password@localhost:5432/clinic_core?sslmode=disable"
	defaultInterval = 30 * time.Second
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("reservation-expiry: %v", err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, postgres.DefaultConfig(envOr("DATABASE_URL", defaultDSN)))
	if err != nil {
		return err
	}
	defer pool.Close()

	handler := command.NewExpireReservationsHandler(
		persistence.NewAppointmentRepository(pool),
		persistence.NewOutboxRepository(pool),
		postgres.NewTransactionManager(pool),
		clock.NewSystem(),
	)

	ticker := time.NewTicker(defaultInterval)
	defer ticker.Stop()

	log.Printf("reservation-expiry started, checking every %s", defaultInterval)

	for {
		select {
		case <-ctx.Done():
			log.Println("reservation-expiry stopped")
			return nil

		case <-ticker.C:
			expired, err := handler.Handle(ctx)
			if err != nil {
				log.Printf("reservation-expiry: %v", err)
				continue
			}
			if expired > 0 {
				log.Printf("reservation-expiry: released %d slots", expired)
			}
		}
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
