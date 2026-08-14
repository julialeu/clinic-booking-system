package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/julialeu/clinic-booking-system/notification-service/internal/application"
	"github.com/julialeu/clinic-booking-system/notification-service/internal/infrastructure/kafka"
	"github.com/julialeu/clinic-booking-system/notification-service/internal/infrastructure/persistence"
	"github.com/julialeu/clinic-booking-system/notification-service/internal/infrastructure/whatsapp"
	"github.com/julialeu/clinic-booking-system/notification-service/internal/platform/postgres"
)

const (
	defaultDSN     = "postgres://clinic:clinic_dev_password@localhost:5432/notifications?sslmode=disable"
	defaultBrokers = "localhost:9092"
	defaultTopic   = "clinic.appointments"
	defaultGroup   = "notification-service"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("notification-service: %v", err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, envOr("DATABASE_URL", defaultDSN))
	if err != nil {
		return err
	}
	defer pool.Close()

	handler := application.NewAppointmentEventHandler(
		persistence.NewProcessedEventsRepository(pool),
		persistence.NewStaticPatientDirectory(),
		whatsapp.NewLogSender(),
	)

	brokers := strings.Split(envOr("KAFKA_BROKERS", defaultBrokers), ",")
	topic := envOr("KAFKA_TOPIC", defaultTopic)
	group := envOr("KAFKA_GROUP", defaultGroup)

	consumer, err := kafka.NewConsumer(brokers, topic, group, handler)
	if err != nil {
		return err
	}

	defer consumer.Close()

	log.Printf("notification-service consuming %s as group %s", topic, group)
	if err := consumer.Run(ctx); err != nil {
		return err
	}

	log.Println("notification-service stopped")
	return nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
