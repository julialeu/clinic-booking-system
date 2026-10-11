from __future__ import annotations

from abc import ABC, abstractmethod
from collections.abc import Sequence

from src.shared.domain.event import DomainEvent


class EventStore(ABC):
    @abstractmethod
    def append(self, events: Sequence[DomainEvent]) -> None: ...