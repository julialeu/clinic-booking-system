from __future__ import annotations

from src.patient.application.command.register_patient import RegisterPatient
from src.patient.domain.model.patient import Patient
from src.patient.domain.model.patient_repository import PatientRepository
from src.patient.domain.value_object import FullName, PatientId, PhoneNumber
from src.shared.domain.clock import Clock
from src.shared.domain.event_store import EventStore


class PhoneAlreadyRegistered(Exception):
    pass


class RegisterPatientHandler:
    def __init__(
        self,
        patients: PatientRepository,
        events: EventStore,
        clock: Clock,
    ) -> None:
        self._patients = patients
        self._events = events
        self._clock = clock

    def __call__(self, command: RegisterPatient) -> PatientId:
        phone = PhoneNumber(command.phone)

        if self._patients.find_by_phone(phone) is not None:
            raise PhoneAlreadyRegistered(
                "A patient with this phone number already exists"
            )

        patient = Patient.register(
            name=FullName(
                command.first_name,
                command.first_surname,
                command.second_surname,
            ),
            phone=phone,
            now=self._clock.now(),
            email=command.email,
        )

        self._patients.save(patient)
        self._events.append(patient.pull_domain_events())

        return patient.id