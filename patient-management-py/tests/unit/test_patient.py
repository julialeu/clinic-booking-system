from __future__ import annotations

from datetime import UTC, datetime

import pytest

from src.patient.domain.event.patient_events import (
    PatientContactChanged,
    PatientRegistered,
)
from src.patient.domain.model.clinical_record import InvalidClinicalRecord
from src.patient.domain.model.patient import Patient, PatientAlreadyDischarged
from src.patient.domain.value_object import FullName, PhoneNumber

NOW = datetime(2026, 9, 1, 10, 0, tzinfo=UTC)


def register_patient() -> Patient:
    return Patient.register(
        FullName("María", "García", "López"),
        PhoneNumber("+34600111222"),
        NOW,
    )


def test_a_new_patient_is_active() -> None:
    assert register_patient().is_active


def test_registering_emits_a_domain_event() -> None:
    patient = register_patient()

    events = patient.pull_domain_events()

    assert len(events) == 1
    assert isinstance(events[0], PatientRegistered)
    assert events[0].payload()["firstName"] == "María"


def test_pulling_events_clears_them() -> None:
    patient = register_patient()
    patient.pull_domain_events()

    assert patient.pull_domain_events() == []


def test_changing_the_phone_emits_a_contact_event() -> None:
    patient = register_patient()
    patient.pull_domain_events()

    patient.change_phone(PhoneNumber("+34699888777"), NOW)

    events = patient.pull_domain_events()

    assert len(events) == 1
    assert isinstance(events[0], PatientContactChanged)
    assert events[0].payload()["phone"] == "+34699888777"


def test_records_clinical_history() -> None:
    patient = register_patient()

    record = patient.add_clinical_record("Dolor lumbar", "Masaje", NOW)

    assert len(patient.clinical_records) == 1
    assert record.consultation_reason == "Dolor lumbar"
    assert record.evolution is None


def test_rejects_clinical_records_with_an_empty_reason() -> None:
    patient = register_patient()

    with pytest.raises(InvalidClinicalRecord):
        patient.add_clinical_record("   ", "Tratamiento", NOW)


def test_a_discharged_patient_cannot_receive_clinical_records() -> None:
    patient = register_patient()
    patient.discharge(NOW)

    with pytest.raises(PatientAlreadyDischarged):
        patient.add_clinical_record("Dolor", "Tratamiento", NOW)


def test_cannot_discharge_twice() -> None:
    patient = register_patient()
    patient.discharge(NOW)

    with pytest.raises(PatientAlreadyDischarged):
        patient.discharge(NOW)


def test_evolution_can_be_added_after_the_session() -> None:
    patient = register_patient()
    record = patient.add_clinical_record("Dolor", "Masaje", NOW)

    record.record_evolution("Mejora notable de la movilidad")

    assert record.evolution == "Mejora notable de la movilidad"


def test_clinical_records_cannot_be_modified_from_outside() -> None:
   
    patient = register_patient()

    assert isinstance(patient.clinical_records, tuple)