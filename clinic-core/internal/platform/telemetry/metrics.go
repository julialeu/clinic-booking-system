package telemetry

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	AppointmentsReserved = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "clinic_appointments_reserved_total",
		Help: "Appointments successfully reserved",
	})

	AppointmentsRejected = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "clinic_appointments_rejected_total",
		Help: "Reservation attempts rejected, by reason",
	}, []string{"reason"})

	ReservationsExpired = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "clinic_reservations_expired_total",
		Help: "Reservation holds released after expiry",
	})

	OutboxPending = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "clinic_outbox_pending_events",
		Help: "Events waiting to be published",
	})

	CommandDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "clinic_command_duration_seconds",
		Help:    "Time spent handling application commands",
		Buckets: prometheus.DefBuckets,
	}, []string{"command", "outcome"})
)

func init() {
	prometheus.MustRegister(
		AppointmentsReserved,
		AppointmentsRejected,
		ReservationsExpired,
		OutboxPending,
		CommandDuration,
	)
}

func ServeMetrics(ctx context.Context, address string) func(context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return
		}
	}()

	return server.Shutdown
}
