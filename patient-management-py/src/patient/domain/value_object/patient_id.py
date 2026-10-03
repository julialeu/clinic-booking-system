from __future__ import annotations

import uuid
from dataclasses import dataclass


class InvalidPatientId(ValueError):
    pass


@dataclass(frozen=True, slots=True)
class PatientId:
    """Identificador de un paciente. Inmutable y comparado por valor."""

    value: str

    def __post_init__(self) -> None:
        try:
            uuid.UUID(self.value)
        except (ValueError, AttributeError, TypeError) as error:
            raise InvalidPatientId(f"Invalid patient id: {self.value}") from error

    @classmethod
    def generate(cls) -> PatientId:
        return cls(str(uuid.uuid4()))

    def __str__(self) -> str:
        return self.value