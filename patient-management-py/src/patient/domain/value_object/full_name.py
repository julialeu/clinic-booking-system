from __future__ import annotations

from dataclasses import dataclass


class InvalidFullName(ValueError):
    pass


@dataclass(frozen=True, slots=True)
class FullName:
    """
    Nombre de un paciente siguiendo la convención española:
    nombre, primer apellido, y segundo apellido opcional.
    """

    first_name: str
    first_surname: str
    second_surname: str | None = None

    def __post_init__(self) -> None:
        first_name = self.first_name.strip() if self.first_name else ""
        first_surname = self.first_surname.strip() if self.first_surname else ""
        second_surname = self.second_surname.strip() if self.second_surname else None

        if not first_name:
            raise InvalidFullName("First name cannot be empty")

        if not first_surname:
            raise InvalidFullName("First surname cannot be empty")

        if not second_surname:
            second_surname = None

        object.__setattr__(self, "first_name", first_name)
        object.__setattr__(self, "first_surname", first_surname)
        object.__setattr__(self, "second_surname", second_surname)

    @property
    def surnames(self) -> str:
        if self.second_surname is None:
            return self.first_surname
        return f"{self.first_surname} {self.second_surname}"

    @property
    def full(self) -> str:
        return f"{self.first_name} {self.surnames}"

    @property
    def informal(self) -> str:
        """Nombre para saludar en mensajes: en España se usa el de pila."""
        return self.first_name