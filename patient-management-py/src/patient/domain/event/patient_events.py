from __future__ import annotations

from dataclasses import dataclass
from datetime import datetime
from typing import Any

from src.shared.domain.event import DomainEvent


@dataclass(frozen=True, slots=True)
class PatientRegistered(DomainEvent):
    patient_id: str
    first_name: str
    full_name: str
    phone: str
    registered_at: datetime

    @property
    def aggregate_id(self) -> str:
        return self.patient_id

    @property
    def event_name(self) -> str:
        return "patient.registered"

    @property
    def occurred_on(self) -> datetime:
        return self.registered_at

    def payload(self) -> dict[str, Any]:
        return {
            "patientId": self.patient_id,
            "firstName": self.first_name,
            "fullName": self.full_name,
            "phone": self.phone,
            "occurredOn": self.registered_at.isoformat(),
        }


@dataclass(frozen=True, slots=True)
class PatientContactChanged(DomainEvent):
    patient_id: str
    first_name: str
    full_name: str
    phone: str
    changed_at: datetime

    @property
    def aggregate_id(self) -> str:
        return self.patient_id

    @property
    def event_name(self) -> str:
        return "patient.contact_changed"

    @property
    def occurred_on(self) -> datetime:
        return self.changed_at

    def payload(self) -> dict[str, Any]:
        return {
            "patientId": self.patient_id,
            "firstName": self.first_name,
            "fullName": self.full_name,
            "phone": self.phone,
            "occurredOn": self.changed_at.isoformat(),
        }