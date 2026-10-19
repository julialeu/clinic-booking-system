from __future__ import annotations

from datetime import UTC, datetime

import pytest
from sqlalchemy import create_engine, delete, select
from sqlalchemy.orm import Session, sessionmaker

from src.patient.application.command.register_patient import RegisterPatient
from src.patient.application.command.register_patient_handler import (
    PhoneAlreadyRegistered,
    RegisterPatientHandler,
)
from src.patient.infrastructure.persistence.mapping import (
    outbox_events_table,
    patients_table,
)
from src.patient.infrastructure.persistence.sqlalchemy_patient_repository import (
    SqlAlchemyPatientRepository,
)
from src.shared.infrastructure.settings import settings
from src.shared.infrastructure.sqlalchemy_event_store import SqlAlchemyEventStore


engine = create_engine(settings.database_url)
TestSession = sessionmaker(bind=engine, expire_on_commit=False)


class FixedClock:
    def now(self) -> datetime:
        return datetime(2026, 9, 1, 10, 0, tzinfo=UTC)


@pytest.fixture
def session() -> Session:
    with TestSession() as session:
        session.execute(delete(outbox_events_table))
        session.execute(delete(patients_table))
        session.commit()
        yield session


def build_handler(session: Session) -> RegisterPatientHandler:
    return RegisterPatientHandler(
        SqlAlchemyPatientRepository(session),
        SqlAlchemyEventStore(session),
        FixedClock(),
    )


def test_registers_a_patient(session: Session) -> None:
    handler = build_handler(session)

    patient_id = handler(
        RegisterPatient(first_name="Marta", first_surname="Sáez", phone="+34611222333")
    )
    session.commit()

    rows = session.execute(select(patients_table)).all()

    assert len(rows) == 1
    assert patient_id.value is not None


def test_writes_the_event_to_the_outbox(session: Session) -> None:
    handler = build_handler(session)

    patient_id = handler(
        RegisterPatient(first_name="Marta", first_surname="Sáez", phone="+34611222333")
    )
    session.commit()

    event = session.execute(select(outbox_events_table)).one()

    assert event.event_name == "patient.registered"
    assert event.aggregate_id == patient_id.value
    assert event.payload["firstName"] == "Marta"
    assert event.published_at is None


def test_rejects_a_duplicate_phone(session: Session) -> None:
    handler = build_handler(session)

    handler(
        RegisterPatient(first_name="Marta", first_surname="Sáez", phone="+34611222333")
    )
    session.commit()

    with pytest.raises(PhoneAlreadyRegistered):
        handler(
            RegisterPatient(
                first_name="Otro", first_surname="Paciente", phone="+34 611 222 333"
            )
        )


def test_a_failed_registration_leaves_no_orphan_event(session: Session) -> None:
    handler = build_handler(session)

    handler(
        RegisterPatient(first_name="Marta", first_surname="Sáez", phone="+34611222333")
    )
    session.commit()

    with pytest.raises(PhoneAlreadyRegistered):
        handler(
            RegisterPatient(
                first_name="Otro", first_surname="Paciente", phone="+34611222333"
            )
        )
    session.rollback()

    events = session.execute(select(outbox_events_table)).all()

    assert len(events) == 1