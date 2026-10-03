from __future__ import annotations

from abc import ABC, abstractmethod
from datetime import datetime
from typing import Any


class DomainEvent(ABC):
    """Un hecho ocurrido en el dominio, listo para publicarse."""

    @property
    @abstractmethod
    def aggregate_id(self) -> str: ...

    @property
    @abstractmethod
    def event_name(self) -> str: ...

    @property
    @abstractmethod
    def occurred_on(self) -> datetime: ...

    @abstractmethod
    def payload(self) -> dict[str, Any]: ...