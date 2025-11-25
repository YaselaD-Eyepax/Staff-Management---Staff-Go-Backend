CREATE TABLE publish_audit (
    id SERIAL PRIMARY KEY,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    channel TEXT NOT NULL,
    status TEXT NOT NULL,
    details JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_publish_audit_event_id ON publish_audit(event_id);
CREATE INDEX idx_publish_audit_channel ON publish_audit(channel);
