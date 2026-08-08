package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/julialeu/clinic-booking-system/clinic-core/internal/infrastructure/messaging"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/platform/postgres"
)

const (
	defaultDSN     = "postgres://clinic:clinic_dev_password@localhost:5432/clinic_core?sslmode=disable"
	defaultBrokers = "localhost:9092"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("outbox-relay: %v", err)
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

	brokers := strings.Split(envOr("KAFKA_BROKERS", defaultBrokers), ",")
	publisher, err := messaging.NewKafkaPublisher(brokers)
	if err != nil {
		return err
	}
	defer publisher.Close()

	relay := messaging.NewOutboxRelay(pool, publisher, messaging.DefaultRelayConfig())

	log.Printf("outbox-relay started, brokers: %s", strings.Join(brokers, ","))
	if err := relay.Run(ctx); err != nil {
		return err
	}

	log.Println("outbox-relay stopped")
	return nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
