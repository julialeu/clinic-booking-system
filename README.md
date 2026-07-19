# Clinic Booking System

Appointment management system for a physiotherapy clinic, built as a
microservices architecture. Replaces a fully manual WhatsApp-based
workflow with automated booking, confirmations and reminders.

> 🚧 **Work in progress** — actively under development.

## Architecture

The system is composed of three independent services:

| Service | Language | Responsibility |
|---|---|---|
| **Clinic Core** | Go | Appointment & schedule management. gRPC API, event publishing to Kafka via the Outbox pattern. PostgreSQL. |
| **Notification Service** | Go | Kafka consumer. Sends WhatsApp notifications (confirmations, reminders) via Twilio. |
| **Patient Management** | PHP / Laravel | Clinical history management. Built with DDD, Hexagonal Architecture and CQRS. |

## Tech Stack

- **Languages:** Go, PHP 8
- **Architecture:** Hexagonal, Domain-Driven Design, CQRS, Event-Driven
- **Communication:** gRPC, Apache Kafka
- **Patterns:** Transactional Outbox, Repository, Domain Events
- **Databases:** PostgreSQL
- **Infrastructure:** Docker, Docker Compose
- **Observability:** OpenTelemetry, Prometheus, Grafana *(planned)*

## Status

| Component | Status |
|---|---|
| Clinic Core — Domain layer | 🚧 In progress |
| Clinic Core — gRPC API | ⏳ Planned |
| Outbox + Kafka | ⏳ Planned |
| Notification Service | ⏳ Planned |
| Patient Management | ⏳ Planned |

---

Built by [Julia Leuenberger](https://github.com/julialeu)