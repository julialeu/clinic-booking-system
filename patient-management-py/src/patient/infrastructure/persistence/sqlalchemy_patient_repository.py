from __future__ import annotations

from sqlalchemy import or_, select
from sqlalchemy.orm import Session

from src.patient.domain.model.patient import Patient
from src.patient.domain.model.patient_repository import PatientRepository
from src.patient.domain.value_object import PatientId, PhoneNumber
from src.patient.infrastructure.persistence.mapping import patients_table


class SqlAlchemyPatientRepository(PatientRepository):
    def __init__(self, session: Session) -> None:
        self._session = session

    def save(self, patient: Patient) -> None:
        self._session.add(patient)

    def find_by_id(self, patient_id: PatientId) -> Patient | None:
        return self._session.get(Patient, patient_id.value)

    def find_by_phone(self, phone: PhoneNumber) -> Patient | None:
        statement = select(Patient).where(patients_table.c.phone == phone.value)
        return self._session.execute(statement).scalar_one_or_none()

    def search_by_name(self, term: str, limit: int = 20) -> list[Patient]:
        pattern = f"%{term.strip().lower()}%"

        statement = (
            select(Patient)
            .where(
                or_(
                    patients_table.c.first_name.ilike(pattern),
                    patients_table.c.first_surname.ilike(pattern),
                    patients_table.c.second_surname.ilike(pattern),
                )
            )
            .limit(limit)
        )

        return list(self._session.execute(statement).scalars().all())