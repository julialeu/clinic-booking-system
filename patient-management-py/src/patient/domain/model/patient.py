from __future__ import annotations

from datetime import datetime

from src.patient.domain.event.patient_events import (
    PatientContactChanged,
    PatientRegistered,
)
from src.patient.domain.model.clinical_record import ClinicalRecord
from src.patient.domain.value_object import FullName, PatientId, PhoneNumber
from src.shared.domain.event import DomainEvent


class PatientAlreadyDischarged(Exception):
    pass


class Patient:
    """Aggregate root del contexto de historial clínico."""

    def __init__(
        self,
        patient_id: PatientId,
        name: FullName,
        phone: PhoneNumber,
        registered_at: datetime,
        email: str | None = None,
        discharged_at: datetime | None = None,
        clinical_records: list[ClinicalRecord] | None = None,
    ) -> None:
        self._id = patient_id
        self._name = name
        self._phone = phone
        self._email = email
        self._registered_at = registered_at
        self._discharged_at = discharged_at
        self._clinical_records = clinical_records or []
        self._domain_events: list[DomainEvent] = []

    @classmethod
    def register(
        cls,
        name: FullName,
        phone: PhoneNumber,
        now: datetime,
        email: str | None = None,
    ) -> Patient:
        patient = cls(
            patient_id=PatientId.generate(),
            name=name,
            phone=phone,
            registered_at=now,
            email=email,
        )

        patient._record_event(
            PatientRegistered(
                patient_id=patient.id.value,
                first_name=name.informal,
                full_name=name.full,
                phone=phone.value,
                registered_at=now,
            )
        )

        return patient

    @property
    def id(self) -> PatientId:
        return self._id

    @property
    def name(self) -> FullName:
        return self._name

    @property
    def phone(self) -> PhoneNumber:
        return self._phone

    @property
    def email(self) -> str | None:
        return self._email

    @property
    def registered_at(self) -> datetime:
        return self._registered_at

    @property
    def is_active(self) -> bool:
        return self._discharged_at is None

    @property
    def clinical_records(self) -> tuple[ClinicalRecord, ...]:
        return tuple(self._clinical_records)

    def change_phone(self, phone: PhoneNumber, now: datetime) -> None:
        self._phone = phone
        self._record_contact_change(now)

    def change_name(self, name: FullName, now: datetime) -> None:
        self._name = name
        self._record_contact_change(now)

    def add_clinical_record(
        self,
        consultation_reason: str,
        treatment: str,
        recorded_at: datetime,
        appointment_id: str | None = None,
    ) -> ClinicalRecord:
        if not self.is_active:
            raise PatientAlreadyDischarged(
                "Cannot add clinical records to a discharged patient"
            )

        record = ClinicalRecord.create(
            patient_id=self._id.value,
            consultation_reason=consultation_reason,
            treatment=treatment,
            recorded_at=recorded_at,
            appointment_id=appointment_id,
        )

        self._clinical_records.append(record)
        return record

    def discharge(self, now: datetime) -> None:
        if not self.is_active:
            raise PatientAlreadyDischarged("Patient is already discharged")

        self._discharged_at = now

    def pull_domain_events(self) -> list[DomainEvent]:
        events = self._domain_events
        self._domain_events = []
        return events

    def _record_contact_change(self, now: datetime) -> None:
        self._record_event(
            PatientContactChanged(
                patient_id=self._id.value,
                first_name=self._name.informal,
                full_name=self._name.full,
                phone=self._phone.value,
                changed_at=now,
            )
        )

    def _record_event(self, event: DomainEvent) -> None:
        self._domain_events.append(event)