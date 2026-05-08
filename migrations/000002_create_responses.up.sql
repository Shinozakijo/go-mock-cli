CREATE TABLE responses (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    route_id    UUID REFERENCES routes(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    status_code INT  NOT NULL DEFAULT 200,
    body        JSONB NOT NULL DEFAULT '{}',
    headers     JSONB NOT NULL DEFAULT '{}',
    delay_ms    INT  NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW()
);