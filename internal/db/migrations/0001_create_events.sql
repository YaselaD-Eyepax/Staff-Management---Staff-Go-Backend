CREATE TABLE events (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    summary TEXT,
    created_by UUID NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('draft','pending','approved','rejected')),
    scheduled_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_events_status ON events(status);
CREATE INDEX idx_events_scheduled_at ON events(scheduled_at);
