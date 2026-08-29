CREATE TABLE patients (
    id         UUID PRIMARY KEY,
    first_name VARCHAR(120) NOT NULL,
    full_name  VARCHAR(250) NOT NULL,
    phone      VARCHAR(20)  NOT NULL,
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);