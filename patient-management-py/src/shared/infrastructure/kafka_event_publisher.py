from __future__ import annotations

import logging
from collections.abc import Sequence

from confluent_kafka import KafkaException, Producer

from src.shared.domain.event_publisher import EventPublisher, PublishableMessage

logger = logging.getLogger(__name__)


class KafkaEventPublisher(EventPublisher):
    def __init__(self, brokers: str) -> None:
        self._producer = Producer(
            {
                "bootstrap.servers": brokers,
                # Espera a que todas las réplicas confirmen: el relay
                # no puede marcar como publicado algo que quizá se perdió.
                "acks": "all",
                "enable.idempotence": True,
            }
        )

    def publish(self, messages: Sequence[PublishableMessage]) -> None:
        if not messages:
            return

        for message in messages:
            self._producer.produce(
                topic=message.topic,
                key=message.key.encode(),
                value=message.payload,
                headers=list(message.headers.items()),
            )

        # produce() solo encola; flush espera la confirmación del broker.
        pending = self._producer.flush(timeout=10)

        if pending > 0:
            raise KafkaException(f"{pending} messages were not delivered")

    def close(self) -> None:
        self._producer.flush(timeout=5)