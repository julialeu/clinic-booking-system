package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/julialeu/clinic-booking-system/clinic-core/internal/application/command"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/application/query"
	grpcadapter "github.com/julialeu/clinic-booking-system/clinic-core/internal/infrastructure/grpc"
	appointmentv1 "github.com/julialeu/clinic-booking-system/clinic-core/internal/infrastructure/grpc/gen/appointment/v1"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/infrastructure/persistence"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/platform/clock"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/platform/postgres"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/platform/telemetry"
)

const (
	defaultDSN     = "postgres://clinic:clinic_dev_password@localhost:5432/clinic_core?sslmode=disable"
	defaultPort    = ":50051"
	defaultMetrics = ":9091"
	defaultOTLP    = "localhost:4317"
	serviceName    = "clinic-core"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("clinic-core: %v", err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownTracing, err := telemetry.Setup(ctx, telemetry.Config{
		ServiceName:  serviceName,
		Environment:  envOr("ENVIRONMENT", "development"),
		OTLPEndpoint: envOr("OTLP_ENDPOINT", defaultOTLP),
	})
	if err != nil {
		return err
	}

	shutdownMetrics := telemetry.ServeMetrics(ctx, envOr("METRICS_ADDRESS", defaultMetrics))

	pool, err := postgres.NewPool(ctx, postgres.DefaultConfig(envOr("DATABASE_URL", defaultDSN)))
	if err != nil {
		return err
	}
	defer pool.Close()

	repository := persistence.NewAppointmentRepository(pool)
	outbox := persistence.NewOutboxRepository(pool)
	transactions := postgres.NewTransactionManager(pool)
	systemClock := clock.NewSystem()

	server := grpcadapter.NewAppointmentServer(
		command.NewReserveAppointmentHandler(repository, outbox, transactions, systemClock),
		command.NewConfirmAppointmentHandler(repository, outbox, transactions, systemClock),
		command.NewCancelAppointmentHandler(repository, outbox, transactions, systemClock),
		query.NewWeeklyAgendaHandler(pool),
	)

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	appointmentv1.RegisterAppointmentServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	address := envOr("GRPC_ADDRESS", defaultPort)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("clinic-core listening on %s", address)
		serverErrors <- grpcServer.Serve(listener)
	}()

	select {
	case err := <-serverErrors:
		return err

	case <-ctx.Done():
		log.Println("shutdown signal received")

		if err := shutdown(grpcServer); err != nil {
			return err
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_ = shutdownMetrics(shutdownCtx)
		return shutdownTracing(shutdownCtx)
	}
}

func shutdown(server *grpc.Server) error {
	stopped := make(chan struct{})

	go func() {
		server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		log.Println("shutdown completed")
		return nil

	case <-time.After(10 * time.Second):
		server.Stop()
		return errors.New("graceful shutdown timed out")
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
