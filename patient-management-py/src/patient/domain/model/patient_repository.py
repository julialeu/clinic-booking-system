from __future__ import annotations

from abc import ABC, abstractmethod

from src.patient.domain.model.patient import Patient
from src.patient.domain.value_object import PatientId, PhoneNumber


class PatientRepository(ABC):

    @abstractmethod
    def save(self, patient: Patient) -> None: ...

    @abstractmethod
    def find_by_id(self, patient_id: PatientId) -> Patient | None: ...

    @abstractmethod
    def find_by_phone(self, phone: PhoneNumber) -> Patient | None: ...

    @abstractmethod
    def search_by_name(self, term: str, limit: int = 20) -> list[Patient]: ...