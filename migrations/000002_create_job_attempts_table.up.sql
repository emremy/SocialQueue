CREATE TABLE job_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,

    attempt_number INTEGER NOT NULL,
    status VARCHAR(30) NOT NULL,
    error_message TEXT NULL,

    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT job_attempts_status_check CHECK (
        status IN ('running', 'completed', 'failed')
    ),

    CONSTRAINT job_attempts_attempt_number_check CHECK (attempt_number > 0)
);

CREATE INDEX idx_job_attempts_job_id ON job_attempts(job_id);
CREATE INDEX idx_job_attempts_status ON job_attempts(status);
CREATE INDEX idx_job_attempts_created_at ON job_attempts(created_at);