from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True, slots=True)
class RegisterPatient:
    first_name: str
    first_surname: str
    phone: str
    second_surname: str | None = None
    email: str | None = None