from __future__ import annotations

from collections.abc import Sequence

from sqlalchemy.orm import Session

from src.patient.infrastructure.persistence.mapping import outbox_events_table
from src.shared.domain.event import DomainEvent
from src.shared.domain.event_store import EventStore


class SqlAlchemyEventStore(EventStore):
    def __init__(self, session: Session) -> None:
        self._session = session

    def append(self, events: Sequence[DomainEvent]) -> None:
    
        if not events:
            return

        rows = [
            {
                "aggregate_id": event.aggregate_id,
                "event_name": event.event_name,
                "payload": event.payload(),
                "occurred_on": event.occurred_on,
            }
            for event in events
        ]

        self._session.execute(outbox_events_table.insert(), rows)