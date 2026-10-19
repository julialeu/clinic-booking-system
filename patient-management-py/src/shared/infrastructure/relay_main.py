from __future__ import annotations

import logging
import signal
import sys
from types import FrameType

from src.patient.infrastructure.persistence.mapping import configure_mappings
from src.shared.infrastructure.database import engine
from src.shared.infrastructure.kafka_event_publisher import KafkaEventPublisher
from src.shared.infrastructure.outbox_relay import OutboxRelay
from src.shared.infrastructure.settings import settings

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(message)s",
)


def main() -> None:
    configure_mappings()

    publisher = KafkaEventPublisher(settings.kafka_brokers)
    relay = OutboxRelay(engine, publisher, settings.patients_topic)

    def shutdown(_signum: int, _frame: FrameType | None) -> None:
        logging.info("shutdown signal received")
        publisher.close()
        sys.exit(0)

    signal.signal(signal.SIGINT, shutdown)
    signal.signal(signal.SIGTERM, shutdown)

    relay.run_forever()


if __name__ == "__main__":
    main()