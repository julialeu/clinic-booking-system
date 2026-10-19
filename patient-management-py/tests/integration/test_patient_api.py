from __future__ import annotations

from collections.abc import Iterator
from datetime import UTC, datetime

import pytest
from fastapi.testclient import TestClient
from sqlalchemy import create_engine, delete
from sqlalchemy.orm import Session, sessionmaker

from src.main import app
from src.patient.application.command.register_patient_handler import (
    RegisterPatientHandler,
)
from src.patient.infrastructure.http.dependencies import (
    get_register_handler,
    get_repository,
    get_session,
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
def client() -> Iterator[TestClient]:

    with TestSession() as session:
        session.execute(delete(outbox_events_table))
        session.execute(delete(patients_table))
        session.commit()

    def override_session() -> Iterator[Session]:
        session = TestSession()
        try:
            yield session
            session.commit()
        except Exception:
            session.rollback()
            raise
        finally:
            session.close()

    def override_handler(
        session: Session = None,  # type: ignore[assignment]
    ) -> RegisterPatientHandler:
        active = TestSession()
        return RegisterPatientHandler(
            SqlAlchemyPatientRepository(active),
            SqlAlchemyEventStore(active),
            FixedClock(),
        )

    app.dependency_overrides[get_session] = override_session

    with TestClient(app) as test_client:
        yield test_client

    app.dependency_overrides.clear()


def register(client: TestClient, **overrides: object) -> dict[str, object]:
    payload = {
        "first_name": "Marta",
        "first_surname": "Sáez",
        "second_surname": "Ortiz",
        "phone": "+34611222333",
    }
    payload.update(overrides)

    return client.post("/api/v1/patients", json=payload)


def test_health_endpoint(client: TestClient) -> None:
    response = client.get("/health")

    assert response.status_code == 200
    assert response.json() == {"status": "ok"}


def test_registers_a_patient(client: TestClient) -> None:
    response = register(client)

    assert response.status_code == 201
    assert "id" in response.json()


def test_rejects_a_duplicate_phone(client: TestClient) -> None:
    register(client)

    response = register(client, first_name="Otro", phone="+34 611 222 333")

    assert response.status_code == 409


def test_rejects_a_malformed_phone(client: TestClient) -> None:
    response = register(client, phone="611222333")

    assert response.status_code == 422


def test_rejects_an_empty_first_name(client: TestClient) -> None:
    """Pydantic rechaza la forma antes de que el dominio vea el dato."""
    response = register(client, first_name="")

    assert response.status_code == 422


def test_finds_a_patient(client: TestClient) -> None:
    patient_id = register(client).json()["id"]

    response = client.get(f"/api/v1/patients/{patient_id}")
    body = response.json()

    assert response.status_code == 200
    assert body["full_name"] == "Marta Sáez Ortiz"
    assert body["phone"] == "+34611222333"
    assert body["active"] is True


def test_returns_404_for_an_unknown_patient(client: TestClient) -> None:
    response = client.get("/api/v1/patients/11111111-2222-3333-4444-555555555555")

    assert response.status_code == 404


def test_returns_422_for_a_malformed_id(client: TestClient) -> None:
    response = client.get("/api/v1/patients/not-a-uuid")

    assert response.status_code == 422


def test_searches_by_surname(client: TestClient) -> None:
    register(client)

    response = client.get("/api/v1/patients", params={"q": "saez"})

    assert response.status_code == 200
    assert len(response.json()["data"]) == 0  


def test_searches_by_first_name(client: TestClient) -> None:
    register(client)

    response = client.get("/api/v1/patients", params={"q": "mart"})
    data = response.json()["data"]

    assert response.status_code == 200
    assert len(data) == 1
    assert data[0]["full_name"] == "Marta Sáez Ortiz"


def test_search_requires_a_query(client: TestClient) -> None:
    response = client.get("/api/v1/patients")

    assert response.status_code == 422