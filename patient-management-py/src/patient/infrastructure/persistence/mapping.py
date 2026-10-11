from __future__ import annotations

from sqlalchemy import (
    Column,
    DateTime,
    ForeignKey,
    MetaData,
    String,
    Table,
    Text,
)
from sqlalchemy.orm import composite, registry, relationship

from src.patient.domain.model.clinical_record import ClinicalRecord
from src.patient.domain.model.patient import Patient
from src.patient.domain.value_object import FullName

metadata = MetaData()
mapper_registry = registry(metadata=metadata)

patients_table = Table(
    "patients",
    metadata,
    Column("id", String(36), primary_key=True),
    Column("first_name", String(80), nullable=False),
    Column("first_surname", String(80), nullable=False),
    Column("second_surname", String(80), nullable=True),
    Column("phone", String(20), nullable=False, index=True),
    Column("email", String(180), nullable=True),
    # Con zona horaria: sin ella, el historial se desplaza una hora
    # dos veces al año.
    Column("registered_at", DateTime(timezone=True), nullable=False),
    Column("discharged_at", DateTime(timezone=True), nullable=True),
)

clinical_records_table = Table(
    "clinical_records",
    metadata,
    Column("id", String(36), primary_key=True),
    Column(
        "patient_id",
        String(36),
        ForeignKey("patients.id"),
        nullable=False,
        index=True,
    ),
    Column("appointment_id", String(36), nullable=True),
    Column("consultation_reason", Text, nullable=False),
    Column("treatment", Text, nullable=False),
    Column("evolution", Text, nullable=True),
    Column("recorded_at", DateTime(timezone=True), nullable=False),
)


def configure_mappings() -> None:
    mapper_registry.map_imperatively(
        ClinicalRecord,
        clinical_records_table,
        properties={
            "_id": clinical_records_table.c.id,
            "_patient_id": clinical_records_table.c.patient_id,
            "_appointment_id": clinical_records_table.c.appointment_id,
            "_consultation_reason": clinical_records_table.c.consultation_reason,
            "_treatment": clinical_records_table.c.treatment,
            "_evolution": clinical_records_table.c.evolution,
            "_recorded_at": clinical_records_table.c.recorded_at,
        },
    )

    mapper_registry.map_imperatively(
        Patient,
        patients_table,
        properties={
            "_id": patients_table.c.id,
            # El nombre se aplana en tres columnas de la misma tabla,
            # igual que un Embeddable: no tiene existencia propia.
            "_name": composite(
                FullName,
                patients_table.c.first_name,
                patients_table.c.first_surname,
                patients_table.c.second_surname,
            ),
            "_phone": patients_table.c.phone,
            "_email": patients_table.c.email,
            "_registered_at": patients_table.c.registered_at,
            "_discharged_at": patients_table.c.discharged_at,
            "_clinical_records": relationship(
                ClinicalRecord,
                lazy="select",
                cascade="all, delete-orphan",
                order_by=clinical_records_table.c.recorded_at.desc(),
            ),
        },
    )