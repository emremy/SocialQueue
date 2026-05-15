CREATE TABLE media_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    storage_provider VARCHAR(30) NOT NULL,
    storage_key TEXT NOT NULL,

    original_filename TEXT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,

    size_bytes BIGINT NOT NULL,

    duration_seconds INTEGER NULL,

    width INTEGER NULL,
    height INTEGER NULL,

    status VARCHAR(30) NOT NULL DEFAULT 'uploaded',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT media_assets_storage_provider_check CHECK (
        storage_provider IN ('local', 's3', 'r2', 'gcs')
    ),

    CONSTRAINT media_assets_status_check CHECK (
        status IN (
            'uploaded',
            'processing',
            'ready',
            'failed',
            'deleted'
        )
    ),

    CONSTRAINT media_assets_size_bytes_check CHECK (
        size_bytes >= 0
    )
);

CREATE INDEX idx_media_assets_user_id
ON media_assets(user_id);

CREATE INDEX idx_media_assets_status
ON media_assets(status);

CREATE INDEX idx_media_assets_created_at
ON media_assets(created_at);

CREATE TRIGGER trg_media_assets_updated_at
BEFORE UPDATE ON media_assets
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();