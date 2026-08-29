package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/julialeu/clinic-booking-system/notification-service/internal/domain/notification"
)

type PatientEventHandler struct {
	projection notification.PatientProjection
}

func NewPatientEventHandler(projection notification.PatientProjection) *PatientEventHandler {
	return &PatientEventHandler{projection: projection}
}

func (h *PatientEventHandler) Handle(
	ctx context.Context,
	reference notification.EventReference,
	payload []byte,
) error {
	switch reference.EventType {
	case notification.EventPatientRegistered, notification.EventPatientContactChanged:
	default:
		return nil
	}

	var event notification.PatientEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("decoding %s: %w", reference.EventType, err)
	}

	err := h.projection.Upsert(ctx, notification.ProjectedPatient{
		PatientId: event.PatientId,
		FirstName: event.FirstName,
		FullName:  event.FullName,
		Phone:     event.Phone,
	})
	if err != nil {
		return err
	}

	log.Printf("projected patient %s (%s)", event.FirstName, event.PatientId)
	return nil
}
