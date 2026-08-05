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

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/julialeu/clinic-booking-system/clinic-core/internal/application/command"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/application/query"
	grpcadapter "github.com/julialeu/clinic-booking-system/clinic-core/internal/infrastructure/grpc"
	appointmentv1 "github.com/julialeu/clinic-booking-system/clinic-core/internal/infrastructure/grpc/gen/appointment/v1"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/infrastructure/persistence"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/platform/clock"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/platform/postgres"
)

const (
	defaultDSN  = "postgres://clinic:clinic_dev_password@localhost:5432/clinic_core?sslmode=disable"
	defaultPort = ":50051"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("clinic-core: %v", err)
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

	repository := persistence.NewAppointmentRepository(pool)
	transactions := postgres.NewTransactionManager(pool)
	systemClock := clock.NewSystem()

	server := grpcadapter.NewAppointmentServer(
		command.NewReserveAppointmentHandler(repository, transactions, systemClock),
		command.NewConfirmAppointmentHandler(repository, transactions, systemClock),
		command.NewCancelAppointmentHandler(repository, transactions, systemClock),
		query.NewWeeklyAgendaHandler(pool),
	)

	grpcServer := grpc.NewServer()
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
		return shutdown(grpcServer)
	}
}

// shutdown espera a que terminen las peticiones en curso,
// forzando el cierre si tardan demasiado.
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
