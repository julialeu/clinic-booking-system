package grpc

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/julialeu/clinic-booking-system/clinic-core/internal/application/command"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/application/query"
	"github.com/julialeu/clinic-booking-system/clinic-core/internal/domain/appointment"
	appointmentv1 "github.com/julialeu/clinic-booking-system/clinic-core/internal/infrastructure/grpc/gen/appointment/v1"
)

type AppointmentServer struct {
	appointmentv1.UnimplementedAppointmentServiceServer

	reserve *command.ReserveAppointmentHandler
	confirm *command.ConfirmAppointmentHandler
	cancel  *command.CancelAppointmentHandler
	agenda  *query.WeeklyAgendaHandler
}

func NewAppointmentServer(
	reserve *command.ReserveAppointmentHandler,
	confirm *command.ConfirmAppointmentHandler,
	cancel *command.CancelAppointmentHandler,
	agenda *query.WeeklyAgendaHandler,
) *AppointmentServer {
	return &AppointmentServer{
		reserve: reserve,
		confirm: confirm,
		cancel:  cancel,
		agenda:  agenda,
	}
}

func (s *AppointmentServer) ReserveAppointment(
	ctx context.Context,
	req *appointmentv1.ReserveAppointmentRequest,
) (*appointmentv1.ReserveAppointmentResponse, error) {
	if req.GetType() == nil {
		return nil, status.Error(codes.InvalidArgument, "appointment type is required")
	}
	if req.GetStartsAt() == nil {
		return nil, status.Error(codes.InvalidArgument, "starts_at is required")
	}

	id, err := s.reserve.Handle(ctx, command.ReserveAppointment{
		PatientId:     req.GetPatientId(),
		StartsAt:      req.GetStartsAt().AsTime(),
		TypeName:      req.GetType().GetName(),
		TypeDuration:  time.Duration(req.GetType().GetDurationMinutes()) * time.Minute,
		TypeColor:     req.GetType().GetColor(),
		PriceCents:    int(req.GetType().GetPriceCents()),
		PriceCurrency: req.GetType().GetPriceCurrency(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &appointmentv1.ReserveAppointmentResponse{
		AppointmentId: id.Value(),
	}, nil
}

func (s *AppointmentServer) ConfirmAppointment(
	ctx context.Context,
	req *appointmentv1.ConfirmAppointmentRequest,
) (*appointmentv1.ConfirmAppointmentResponse, error) {
	err := s.confirm.Handle(ctx, command.ConfirmAppointment{
		AppointmentId: req.GetAppointmentId(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &appointmentv1.ConfirmAppointmentResponse{}, nil
}

func (s *AppointmentServer) CancelAppointment(
	ctx context.Context,
	req *appointmentv1.CancelAppointmentRequest,
) (*appointmentv1.CancelAppointmentResponse, error) {
	err := s.cancel.Handle(ctx, command.CancelAppointment{
		AppointmentId: req.GetAppointmentId(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &appointmentv1.CancelAppointmentResponse{}, nil
}

func (s *AppointmentServer) GetWeeklyAgenda(
	ctx context.Context,
	req *appointmentv1.GetWeeklyAgendaRequest,
) (*appointmentv1.GetWeeklyAgendaResponse, error) {
	if req.GetAnyDayOfWeek() == nil {
		return nil, status.Error(codes.InvalidArgument, "any_day_of_week is required")
	}

	result, err := s.agenda.Handle(ctx, query.WeeklyAgenda{
		AnyDayOfWeek: req.GetAnyDayOfWeek().AsTime(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	days := make([]*appointmentv1.AgendaDay, 0, len(result.Days))
	for _, day := range result.Days {
		appointments := make([]*appointmentv1.AgendaAppointment, 0, len(day.Appointments))
		for _, item := range day.Appointments {
			appointments = append(appointments, &appointmentv1.AgendaAppointment{
				AppointmentId: item.AppointmentId,
				PatientId:     item.PatientId,
				StartsAt:      timestamppb.New(item.StartsAt),
				EndsAt:        timestamppb.New(item.EndsAt),
				TypeName:      item.TypeName,
				TypeColor:     item.TypeColor,
				Status:        toProtoStatus(item.Status),
			})
		}

		days = append(days, &appointmentv1.AgendaDay{
			Date:         timestamppb.New(day.Date),
			Appointments: appointments,
		})
	}

	return &appointmentv1.GetWeeklyAgendaResponse{
		WeekStart: timestamppb.New(result.WeekStart),
		WeekEnd:   timestamppb.New(result.WeekEnd),
		Days:      days,
	}, nil
}

// toGRPCError traduce errores de dominio y aplicación a códigos gRPC.
func toGRPCError(err error) error {
	switch {
	case errors.Is(err, appointment.ErrAppointmentNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, command.ErrSlotNotAvailable):
		return status.Error(codes.AlreadyExists, err.Error())

	case errors.Is(err, appointment.ErrPastAppointment),
		errors.Is(err, appointment.ErrDurationMismatch),
		errors.Is(err, appointment.ErrInvalidPatientId),
		errors.Is(err, appointment.ErrInvalidAppointmentId):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, appointment.ErrNotReserved),
		errors.Is(err, appointment.ErrReservationLapsed),
		errors.Is(err, appointment.ErrAlreadyCancelled),
		errors.Is(err, appointment.ErrAlreadyCompleted),
		errors.Is(err, appointment.ErrAlreadyStarted),
		errors.Is(err, appointment.ErrNotConfirmed),
		errors.Is(err, appointment.ErrNotFinished):
		return status.Error(codes.FailedPrecondition, err.Error())

	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func toProtoStatus(label string) appointmentv1.AppointmentStatus {
	switch label {
	case "reserved":
		return appointmentv1.AppointmentStatus_APPOINTMENT_STATUS_RESERVED
	case "confirmed":
		return appointmentv1.AppointmentStatus_APPOINTMENT_STATUS_CONFIRMED
	case "cancelled":
		return appointmentv1.AppointmentStatus_APPOINTMENT_STATUS_CANCELLED
	case "completed":
		return appointmentv1.AppointmentStatus_APPOINTMENT_STATUS_COMPLETED
	default:
		return appointmentv1.AppointmentStatus_APPOINTMENT_STATUS_UNSPECIFIED
	}
}
