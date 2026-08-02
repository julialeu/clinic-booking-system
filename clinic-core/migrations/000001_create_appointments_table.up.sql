CREATE TABLE appointments (
    id                UUID PRIMARY KEY,
    patient_id        UUID        NOT NULL,
    starts_at         TIMESTAMPTZ NOT NULL,
    ends_at           TIMESTAMPTZ NOT NULL,
    status            SMALLINT    NOT NULL,
    reserved_until    TIMESTAMPTZ,

    type_name         VARCHAR(120) NOT NULL,
    type_duration_min INTEGER      NOT NULL,
    type_color        VARCHAR(9)   NOT NULL,
    price_cents       INTEGER      NOT NULL,
    price_currency    CHAR(3)      NOT NULL,

    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT appointments_slot_order CHECK (ends_at > starts_at),
    CONSTRAINT appointments_price_non_negative CHECK (price_cents >= 0)
);

CREATE INDEX appointments_active_slot_idx
    ON appointments (starts_at, ends_at)
    WHERE status IN (0, 1);

CREATE INDEX appointments_reserved_until_idx
    ON appointments (reserved_until)
    WHERE status = 0;

CREATE INDEX appointments_patient_idx
    ON appointments (patient_id);