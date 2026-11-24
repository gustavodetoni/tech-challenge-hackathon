CREATE TABLE IF NOT EXISTS video_jobs (
  id UUID PRIMARY KEY, user_id TEXT NOT NULL, original_filename TEXT NOT NULL,
  content_type TEXT NOT NULL, size_bytes BIGINT NOT NULL CHECK (size_bytes > 0),
  input_object_key TEXT NOT NULL UNIQUE, output_object_key TEXT NULL,
  status TEXT NOT NULL CHECK (status IN ('AWAITING_UPLOAD','PENDING','PROCESSING','COMPLETED','FAILED')),
  error_code TEXT NULL, error_message TEXT NULL, attempt_count INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  started_at TIMESTAMPTZ NULL, finished_at TIMESTAMPTZ NULL
);
CREATE INDEX IF NOT EXISTS idx_video_jobs_user_created_at ON video_jobs (user_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_video_jobs_status ON video_jobs (status);
