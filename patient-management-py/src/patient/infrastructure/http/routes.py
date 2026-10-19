from __future__ import annotations

from fastapi import APIRouter, Depends, HTTPException, Query, status

from src.patient.application.command.register_patient import RegisterPatient
from src.patient.application.command.register_patient_handler import (
    PhoneAlreadyRegistered,
    RegisterPatientHandler,
)
from src.patient.domain.value_object import (
    InvalidFullName,
    InvalidPatientId,
    InvalidPhoneNumber,
    PatientId,
)
from src.patient.infrastructure.http.dependencies import (
    get_register_handler,
    get_repository,
)
from src.patient.infrastructure.http.schemas import (
    PatientListResponse,
    PatientResponse,
    PatientSummary,
    RegisterPatientRequest,
    RegisterPatientResponse,
)
from src.patient.infrastructure.persistence.sqlalchemy_patient_repository import (
    SqlAlchemyPatientRepository,
)

router = APIRouter(prefix="/api/v1", tags=["patients"])


@router.post(
    "/patients",
    status_code=status.HTTP_201_CREATED,
    response_model=RegisterPatientResponse,
)
def register_patient(
    request: RegisterPatientRequest,
    handler: RegisterPatientHandler = Depends(get_register_handler),
) -> RegisterPatientResponse:
    try:
        patient_id = handler(
            RegisterPatient(
                first_name=request.first_name,
                first_surname=request.first_surname,
                second_surname=request.second_surname,
                phone=request.phone,
                email=request.email,
            )
        )
    except PhoneAlreadyRegistered as error:
        
        raise HTTPException(status.HTTP_409_CONFLICT, str(error)) from error
    except (InvalidPhoneNumber, InvalidFullName) as error:
        
        raise HTTPException(
            status.HTTP_422_UNPROCESSABLE_ENTITY, str(error)
        ) from error

    return RegisterPatientResponse(id=patient_id.value)


@router.get("/patients/{patient_id}", response_model=PatientResponse)
def find_patient(
    patient_id: str,
    patients: SqlAlchemyPatientRepository = Depends(get_repository),
) -> PatientResponse:
    try:
        identifier = PatientId(patient_id)
    except InvalidPatientId as error:
        raise HTTPException(
            status.HTTP_422_UNPROCESSABLE_ENTITY, "Invalid patient id"
        ) from error

    patient = patients.find_by_id(identifier)

    if patient is None:
        raise HTTPException(status.HTTP_404_NOT_FOUND, "Patient not found")

    return PatientResponse(
        id=patient.id.value,
        first_name=patient.name.first_name,
        first_surname=patient.name.first_surname,
        second_surname=patient.name.second_surname,
        full_name=patient.name.full,
        phone=patient.phone.value,
        email=patient.email,
        active=patient.is_active,
        registered_at=patient.registered_at,
    )


@router.get("/patients", response_model=PatientListResponse)
def search_patients(
    q: str = Query(min_length=1, description="Name or surname to search for"),
    patients: SqlAlchemyPatientRepository = Depends(get_repository),
) -> PatientListResponse:
    found = patients.search_by_name(q)

    return PatientListResponse(
        data=[
            PatientSummary(
                id=patient.id.value,
                full_name=patient.name.full,
                phone=patient.phone.value,
                active=patient.is_active,
            )
            for patient in found
        ]
    )