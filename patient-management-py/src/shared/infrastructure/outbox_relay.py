from __future__ import annotations

import json
import logging
import time
from collections.abc import Sequence
from typing import Any

from sqlalchemy import Engine, text

from src.shared.domain.event_publisher import EventPublisher, PublishableMessage

logger = logging.getLogger(__name__)

SELECT_PENDING = text(
    """
    SELECT id, aggregate_id, event_name, payload, occurred_on
    FROM outbox_events
    WHERE published_at IS NULL
    ORDER BY id
    LIMIT :limit
    FOR UPDATE SKIP LOCKED
    """
)

MARK_PUBLISHED = text(
    "UPDATE outbox_events SET published_at = now() WHERE id = ANY(:ids)"
)

RECORD_FAILURE = text(
    """
    UPDATE outbox_events
    SET attempts = attempts + 1, last_error = :error
    WHERE id = ANY(:ids)
    """
)


class OutboxRelay:
    def __init__(
        self,
        engine: Engine,
        publisher: EventPublisher,
        topic: str,
        batch_size: int = 100,
        poll_interval: float = 2.0,
    ) -> None:
        self._engine = engine
        self._publisher = publisher
        self._topic = topic
        self._batch_size = batch_size
        self._poll_interval = poll_interval

    def run_forever(self) -> None:
        logger.info("outbox relay started, topic: %s", self._topic)

        while True:
            try:
                published = self.process_batch()
                if published:
                    logger.info("published %d events", published)
            except Exception:
                logger.exception("outbox relay batch failed")

            time.sleep(self._poll_interval)

    def process_batch(self) -> int:
        with self._engine.begin() as connection:
            rows = connection.execute(
                SELECT_PENDING, {"limit": self._batch_size}
            ).mappings().all()

            if not rows:
                return 0

            messages = [self._to_message(row) for row in rows]
            ids = [row["id"] for row in rows]

            try:
                self._publisher.publish(messages)
            except Exception as error:
                self._record_failure(ids, str(error))
                raise

            connection.execute(MARK_PUBLISHED, {"ids": ids})

        return len(rows)

    def _to_message(self, row: Any) -> PublishableMessage:
        return PublishableMessage(
            topic=self._topic,
            key=row["aggregate_id"],
            payload=json.dumps(row["payload"]).encode(),
            headers={
                "event_type": row["event_name"],
                "aggregate_type": "patient",
                "occurred_on": row["occurred_on"].isoformat(),
            },
        )

    def _record_failure(self, ids: Sequence[int], error: str) -> None:
        with self._engine.begin() as connection:
            connection.execute(RECORD_FAILURE, {"ids": list(ids), "error": error})