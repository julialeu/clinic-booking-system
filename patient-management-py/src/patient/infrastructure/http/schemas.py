from __future__ import annotations

from datetime import datetime

from pydantic import BaseModel, Field


class RegisterPatientRequest(BaseModel):
    first_name: str = Field(min_length=1, max_length=80)
    first_surname: str = Field(min_length=1, max_length=80)
    second_surname: str | None = Field(default=None, max_length=80)
    phone: str = Field(min_length=1, max_length=20)
    email: str | None = Field(default=None, max_length=180)


class RegisterPatientResponse(BaseModel):
    id: str


class PatientResponse(BaseModel):
    id: str
    first_name: str
    first_surname: str
    second_surname: str | None
    full_name: str
    phone: str
    email: str | None
    active: bool
    registered_at: datetime


class PatientSummary(BaseModel):
    id: str
    full_name: str
    phone: str
    active: bool


class PatientListResponse(BaseModel):
    data: list[PatientSummary]