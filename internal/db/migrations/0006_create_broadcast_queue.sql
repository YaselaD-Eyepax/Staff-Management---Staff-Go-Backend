CREATE TABLE IF NOT EXISTS broadcast_queue (
    id SERIAL PRIMARY KEY,
    event_id UUID NOT NULL,
    channel VARCHAR(50) NOT NULL,
    payload JSONB,
    attempts INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    last_error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
