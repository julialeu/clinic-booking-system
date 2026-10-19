from __future__ import annotations

from collections.abc import Iterator

from fastapi import Depends
from sqlalchemy.orm import Session

from src.patient.application.command.register_patient_handler import (
    RegisterPatientHandler,
)
from src.patient.infrastructure.persistence.sqlalchemy_patient_repository import (
    SqlAlchemyPatientRepository,
)
from src.shared.infrastructure.database import SessionFactory
from src.shared.infrastructure.sqlalchemy_event_store import SqlAlchemyEventStore
from src.shared.infrastructure.system_clock import SystemClock


def get_session() -> Iterator[Session]:
    """Una sesión por petición: commit al terminar, rollback si falla."""
    session = SessionFactory()
    try:
        yield session
        session.commit()
    except Exception:
        session.rollback()
        raise
    finally:
        session.close()


def get_repository(
    session: Session = Depends(get_session),
) -> SqlAlchemyPatientRepository:
    return SqlAlchemyPatientRepository(session)


def get_register_handler(
    session: Session = Depends(get_session),
) -> RegisterPatientHandler:
    return RegisterPatientHandler(
        SqlAlchemyPatientRepository(session),
        SqlAlchemyEventStore(session),
        SystemClock(),
    )