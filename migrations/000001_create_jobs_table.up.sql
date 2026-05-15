CREATE TABLE jobs (
    id UUID PRIMARY KEY,

    type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,

    status VARCHAR(30) NOT NULL DEFAULT 'queued',
    priority INTEGER NOT NULL DEFAULT 0,

    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,

    run_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locked_at TIMESTAMPTZ NULL,
    worker_id VARCHAR(100) NULL,
    last_error TEXT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT jobs_status_check CHECK (
        status IN ('queued', 'running', 'completed', 'failed', 'cancelled')
    ),

    CONSTRAINT jobs_attempts_check CHECK (attempts >= 0),
    CONSTRAINT jobs_max_attempts_check CHECK (max_attempts > 0),
    CONSTRAINT jobs_priority_check CHECK (priority >= 0)
);

CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_type ON jobs(type);
CREATE INDEX idx_jobs_run_at ON jobs(run_at);
CREATE INDEX idx_jobs_locked_at ON jobs(locked_at);
CREATE INDEX idx_jobs_worker_id ON jobs(worker_id);

CREATE INDEX idx_jobs_claim
ON jobs(priority DESC, run_at ASC, created_at ASC)
WHERE status = 'queued';

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_jobs_updated_at
BEFORE UPDATE ON jobs
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();