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
	defaultDSN        = "postgres://clinic:clinic_dev_password@localhost:5432/notifications?sslmode=disable"
	defaultBrokers    = "localhost:9092"
	appointmentsTopic = "clinic.appointments"
	patientsTopic     = "clinic.patients"
	defaultGroup      = "notification-service"
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

	patients := persistence.NewPatientRepository(pool)

	router := application.NewTopicRouter(map[string]application.Handler{
		appointmentsTopic: application.NewAppointmentEventHandler(
			persistence.NewProcessedEventsRepository(pool),
			patients,
			whatsapp.NewLogSender(),
		),
		patientsTopic: application.NewPatientEventHandler(patients),
	})

	brokers := strings.Split(envOr("KAFKA_BROKERS", defaultBrokers), ",")
	topics := []string{appointmentsTopic, patientsTopic}
	group := envOr("KAFKA_GROUP", defaultGroup)

	consumer, err := kafka.NewConsumer(brokers, topics, group, router)
	if err != nil {
		return err
	}
	defer consumer.Close()

	log.Printf("notification-service consuming %v as group %s", topics, group)
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
