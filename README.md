# Clinic Booking System

Appointment management for a physiotherapy clinic. Built to replace a
WhatsApp-only workflow where the physio was manually tracking every booking
by hand.

Three services, event-driven, running on Go and PHP.

## Why this exists

A friend runs a small physio practice on his own. Patients message him on
WhatsApp, he checks his memory or a paper diary, replies, and hopes he
didn't double-book anyone. No availability view, no reminders, no history.

So the requirements here aren't invented. Slot locking, cancellation
windows and reminders come from an actual problem, which is also why the
concurrency handling gets more attention than a demo project would
normally justify.

## Architecture

```mermaid
flowchart TB
    client([Client])

    subgraph core["Clinic Core · Go"]
        api["gRPC API<br/><small>:50051</small>"]
        domain["Appointment aggregate"]
        relay["outbox-relay"]
    end

    db[("PostgreSQL<br/><small>appointments + outbox</small>")]
    kafka{{"Kafka<br/><small>clinic.appointments</small>"}}

    subgraph notif["Notification Service · Go"]
        consumer["Kafka consumer"]
        sender["Sender"]
    end

    notifdb[("PostgreSQL<br/><small>processed events</small>")]
    whatsapp([WhatsApp])

    client -->|gRPC| api
    api --> domain
    domain -->|"appointment + events<br/>same transaction"| db
    db -.->|polls pending| relay
    relay -->|publishes| kafka
    kafka --> consumer
    consumer -->|dedup check| notifdb
    consumer --> sender
    sender --> whatsapp

    classDef service fill:#f0f7ff,stroke:#4a7fb5
    classDef storage fill:#fff8ec,stroke:#c99a3f
    class core,notif service
    class db,notifdb storage
```

The dotted line is the important one. Events are written to the outbox in
the same transaction as the appointment, then picked up separately. Nothing
publishes to Kafka during the request.

| Service | Stack | Does |
|---|---|---|
| **Clinic Core** | Go, PostgreSQL, gRPC | Scheduling, availability, booking lifecycle |
| **Notification Service** | Go, PostgreSQL, Kafka | Consumes appointment events, sends messages |
| **Patient Management** | PHP 8 / Laravel | Clinical history (in progress) |

Each service owns its database. They talk through events, never through
shared tables.

## Running it

You need Docker and Go 1.23+.

```bash
docker compose up -d
```

That brings up PostgreSQL, Redpanda (Kafka-compatible, lighter for local
work) and a web console at `localhost:8080`. Topics get created
automatically.

Apply migrations:

```bash
cd clinic-core
migrate -path migrations -database "postgres://clinic:clinic_dev_password@localhost:5432/clinic_core?sslmode=disable" up
```

Then run the three processes, one per terminal:

```bash
go run ./cmd/api            # gRPC server on :50051
go run ./cmd/outbox-relay   # publishes events to Kafka
```

```bash
cd notification-service
go run ./cmd/consumer
```

Book something:

```bash
grpcurl -plaintext -d @ localhost:50051 \
  appointment.v1.AppointmentService/ReserveAppointment < reserve.json
```

Confirm it within five minutes and the notification service will log the
message it would have sent.

## Design notes

### Preventing double bookings

Two patients tapping the same slot at the same time is the failure mode
that matters most here. The check happens inside a transaction:
`SELECT ... FOR UPDATE` locks any overlapping rows, and they stay locked
until the write commits. The second request blocks, wakes up, sees the
booking and gets rejected.

The lock only means something if the read and the write share a
transaction, so the application layer owns it. The transaction travels
through `context.Context`, which keeps the repository interface free of
persistence details and lets the same code work inside or outside a
transaction.

`TestConcurrentReservationsOnSameSlot` fires eight goroutines released
simultaneously. Exactly one wins, and the table ends with one row.

### The outbox

Writing to Postgres and publishing to Kafka are separate systems, so they
can't be made atomic without distributed transactions. If the publish fails
after the commit, the appointment exists and nobody gets told about it.

Instead, the event goes into a table in the same database, in the same
transaction. Either both land or neither does. A separate process polls
that table and publishes at its own pace, retrying when the broker is
unhappy.

That gives at-least-once delivery, which means duplicates. The consumer
handles them with a table keyed on topic + partition + offset (Kafka
guarantees that combination is unique). Using the appointment id wouldn't
work, since one appointment legitimately produces several events.

The relay runs `FOR UPDATE SKIP LOCKED`, turning the outbox into a work
queue. Multiple instances claim different rows instead of blocking on each
other.

### Time is injected, never read

Nothing in the domain calls `time.Now()`. It arrives as a parameter.

Sounds fussy until you need to test what happens when a five-minute hold
expires, and the alternative is waiting five minutes. Same reason the
handlers take a `Clock` interface.

### Where DDD applies and where it doesn't

Clinic Core has real invariants (you can't confirm a lapsed hold, can't
cancel a session that already started, slot duration has to match the
appointment type), so it gets a proper aggregate protecting them.

Notification Service doesn't. A message gets composed and sent. Modelling
that with a state machine would be ceremony, so it's a value type and two
functions.

Same for the read side. The weekly agenda query goes straight from table to
DTO without loading aggregates. Building rich objects to immediately throw
them away would be wasteful, and there's nothing to protect on a read.

### Choices worth explaining

**pgx over an ORM.** In hexagonal architecture the adapter already does the
mapping explicitly, so an ORM adds a layer that hides SQL exactly where
transactional locking needs the most control.

**Redpanda locally.** Same protocol as Kafka, one binary, no JVM, no
ZooKeeper. Starts in seconds instead of a minute and uses about 200MB.
Deploying against managed Kafka wouldn't change a line of code.

**Contract duplication between services.** Notification Service defines its
own copy of the event structs rather than importing them. Compile-time
coupling would mean any internal refactor in Clinic Core breaks the other
service.

**Partial indexes.** Cancelled and completed appointments can never block a
slot, so they're excluded from the overlap index (`WHERE status IN (0,1)`).
The table can grow to millions of rows while the index stays small.

## Testing

```bash
cd clinic-core && go test ./...
```

Around 60 tests. Domain logic runs as pure unit tests with no dependencies.
Everything touching the database runs against real PostgreSQL through
testcontainers, spun up and torn down per package.

Mocked repositories confirm a method was called. They don't catch a typo in
your SQL, and they can't verify that `FOR UPDATE` does anything.

## Not built yet

Waiting list, admin panel, statistics, GDPR export, JWT auth, multi-clinic
support. The product spec covers all of it. The architecture leaves room
for it.

Shipping something the physio can actually use came first.