CREATE TABLE processed_events (
    topic        VARCHAR(120) NOT NULL,
    partition_id INTEGER      NOT NULL,
    offset_id    BIGINT       NOT NULL,
    event_type   VARCHAR(120) NOT NULL,
    processed_at TIMESTAMPTZ  NOT NULL DEFAULT now(),

    PRIMARY KEY (topic, partition_id, offset_id)
);

CREATE INDEX processed_events_processed_at_idx
    ON processed_events (processed_at);