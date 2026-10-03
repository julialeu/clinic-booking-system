from __future__ import annotations

import re
from dataclasses import dataclass

E164 = re.compile(r"^\+[1-9]\d{7,14}$")
SEPARATORS = re.compile(r"[\s\-()]")


class InvalidPhoneNumber(ValueError):
    pass


@dataclass(frozen=True, slots=True)
class PhoneNumber:
    """Teléfono en formato internacional, normalizado a E.164."""

    value: str

    def __post_init__(self) -> None:
        raw = self.value.strip() if self.value else ""
        normalised = SEPARATORS.sub("", raw)

        if not E164.match(normalised):
            raise InvalidPhoneNumber(f"Invalid phone number: {self.value}")

        object.__setattr__(self, "value", normalised)

    def __str__(self) -> str:
        return self.value