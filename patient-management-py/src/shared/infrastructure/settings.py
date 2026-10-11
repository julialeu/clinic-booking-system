from __future__ import annotations

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", extra="ignore")

    database_url: str = (
        "postgresql+psycopg://clinic:clinic_dev_password@localhost:5432/patients_py"
    )
    kafka_brokers: str = "localhost:9092"
    patients_topic: str = "clinic.patients"


settings = Settings()