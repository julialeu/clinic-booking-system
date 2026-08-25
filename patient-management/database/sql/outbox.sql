CREATE TABLE outbox_events (
    id           BIGSERIAL PRIMARY KEY,
    aggregate_id VARCHAR(36)  NOT NULL,
    event_name   VARCHAR(120) NOT NULL,
    payload      JSONB        NOT NULL,
    occurred_on  TIMESTAMPTZ  NOT NULL,
    published_at TIMESTAMPTZ,
    attempts     INTEGER      NOT NULL DEFAULT 0,
    last_error   TEXT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX outbox_events_pending_idx
    ON outbox_events (id)
    WHERE published_at IS NULL;