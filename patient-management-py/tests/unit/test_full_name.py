from __future__ import annotations

import pytest

from src.patient.domain.value_object import FullName, InvalidFullName


def test_builds_a_name_with_two_surnames() -> None:
    name = FullName("María", "García", "López")

    assert name.full == "María García López"
    assert name.surnames == "García López"
    assert name.informal == "María"


def test_second_surname_is_optional() -> None:
    name = FullName("Ana", "Ruiz")

    assert name.full == "Ana Ruiz"
    assert name.second_surname is None


def test_treats_an_empty_second_surname_as_absent() -> None:
    assert FullName("Ana", "Ruiz", "") == FullName("Ana", "Ruiz", None)


def test_trims_surrounding_whitespace() -> None:
    assert FullName("  Ana  ", "  Ruiz  ").full == "Ana Ruiz"


def test_requires_a_first_name() -> None:
    with pytest.raises(InvalidFullName):
        FullName("", "Ruiz")


def test_requires_a_first_surname() -> None:
    with pytest.raises(InvalidFullName):
        FullName("Ana", "   ")


def test_is_immutable() -> None:
    name = FullName("Ana", "Ruiz")

    with pytest.raises(Exception):
        name.first_name = "Otra"