from __future__ import annotations

import uuid
from datetime import datetime


class InvalidClinicalRecord(ValueError):
    pass


class ClinicalRecord:
    """Entrada del historial clínico. Entidad interna del aggregate."""

    def __init__(
        self,
        record_id: str,
        patient_id: str,
        consultation_reason: str,
        treatment: str,
        recorded_at: datetime,
        appointment_id: str | None = None,
        evolution: str | None = None,
    ) -> None:
        reason = consultation_reason.strip() if consultation_reason else ""
        applied = treatment.strip() if treatment else ""

        if not reason:
            raise InvalidClinicalRecord("Consultation reason cannot be empty")

        if not applied:
            raise InvalidClinicalRecord("Treatment cannot be empty")

        self._id = record_id
        self._patient_id = patient_id
        self._consultation_reason = reason
        self._treatment = applied
        self._recorded_at = recorded_at
        self._appointment_id = appointment_id
        self._evolution = evolution

    @classmethod
    def create(
        cls,
        patient_id: str,
        consultation_reason: str,
        treatment: str,
        recorded_at: datetime,
        appointment_id: str | None = None,
    ) -> ClinicalRecord:
        return cls(
            record_id=str(uuid.uuid4()),
            patient_id=patient_id,
            consultation_reason=consultation_reason,
            treatment=treatment,
            recorded_at=recorded_at,
            appointment_id=appointment_id,
        )

    @property
    def id(self) -> str:
        return self._id

    @property
    def consultation_reason(self) -> str:
        return self._consultation_reason

    @property
    def treatment(self) -> str:
        return self._treatment

    @property
    def evolution(self) -> str | None:
        return self._evolution

    @property
    def recorded_at(self) -> datetime:
        return self._recorded_at

    @property
    def appointment_id(self) -> str | None:
        return self._appointment_id

    def record_evolution(self, evolution: str) -> None:
        """La evolución se añade tras la sesión, al valorar la respuesta."""
        text = evolution.strip() if evolution else ""

        if not text:
            raise InvalidClinicalRecord("Evolution cannot be empty")

        self._evolution = text