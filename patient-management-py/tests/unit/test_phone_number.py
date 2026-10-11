from __future__ import annotations

import pytest

from src.patient.domain.value_object import InvalidPhoneNumber, PhoneNumber


def test_accepts_a_valid_international_number() -> None:
    assert PhoneNumber("+34600111222").value == "+34600111222"


@pytest.mark.parametrize(
    "raw",
    [
        "+34 600 111 222",
        "+34-600-111-222",
        "  +34600111222  ",
        "+34 (600) 111222",
    ],
)
def test_normalises_spaces_and_separators(raw: str) -> None:
    assert PhoneNumber(raw).value == "+34600111222"


def test_rejects_numbers_without_an_international_prefix() -> None:
    with pytest.raises(InvalidPhoneNumber):
        PhoneNumber("600111222")


@pytest.mark.parametrize(
    "raw",
    ["", "+", "+0600111222", "+3460011122233445566", "not a phone"],
)
def test_rejects_malformed_numbers(raw: str) -> None:
    with pytest.raises(InvalidPhoneNumber):
        PhoneNumber(raw)


def test_compares_by_value() -> None:
    assert PhoneNumber("+34600111222") == PhoneNumber("+34 600 111 222")