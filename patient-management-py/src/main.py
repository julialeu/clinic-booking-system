from __future__ import annotations

from contextlib import asynccontextmanager
from collections.abc import AsyncIterator

from fastapi import FastAPI

from src.patient.infrastructure.http.routes import router
from src.patient.infrastructure.persistence.mapping import configure_mappings


@asynccontextmanager
async def lifespan(_app: FastAPI) -> AsyncIterator[None]:
    configure_mappings()
    yield


app = FastAPI(
    title="Patient Management",
    description="Clinical records service for the clinic booking system",
    version="1.0.0",
    lifespan=lifespan,
)

app.include_router(router)


@app.get("/health", tags=["ops"])
def health() -> dict[str, str]:
    return {"status": "ok"}