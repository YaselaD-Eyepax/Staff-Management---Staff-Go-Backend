CREATE TABLE IF announcement_bodies (
    id UUID PRIMARY KEY,
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    attachments JSONB DEFAULT '[]'
);

CREATE INDEX idx_bodies_event_id ON announcement_bodies(event_id);
