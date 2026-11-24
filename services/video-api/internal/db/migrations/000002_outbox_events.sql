CREATE TABLE IF NOT EXISTS outbox_events (
  id UUID PRIMARY KEY,
  aggregate_id UUID NOT NULL,
  type TEXT NOT NULL,
  payload JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  published_at TIMESTAMPTZ NULL,
  attempts INTEGER NOT NULL DEFAULT 0,
  publish_error TEXT NULL
);
CREATE INDEX IF NOT EXISTS idx_outbox_events_unpublished ON outbox_events (created_at) WHERE published_at IS NULL;
