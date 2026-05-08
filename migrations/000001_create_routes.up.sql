CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE routes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    method      VARCHAR(10)  NOT NULL,
    path        VARCHAR(255) NOT NULL,
    description TEXT,
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW(),
    UNIQUE(method, path)
);
