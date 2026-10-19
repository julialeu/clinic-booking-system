from __future__ import annotations

from abc import ABC, abstractmethod
from collections.abc import Sequence
from dataclasses import dataclass, field


@dataclass(frozen=True, slots=True)
class PublishableMessage:
    topic: str
    key: str
    payload: bytes
    headers: dict[str, str] = field(default_factory=dict)


class EventPublisher(ABC):
    @abstractmethod
    def publish(self, messages: Sequence[PublishableMessage]) -> None: ...

    @abstractmethod
    def close(self) -> None: ...