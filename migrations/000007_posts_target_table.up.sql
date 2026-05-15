CREATE TABLE post_targets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    social_account_id UUID NOT NULL REFERENCES social_accounts(id) ON DELETE RESTRICT,

    platform VARCHAR(30) NOT NULL,

    status VARCHAR(30) NOT NULL DEFAULT 'pending',

    platform_post_id VARCHAR(255) NULL,
    platform_url TEXT NULL,

    error_message TEXT NULL,

    published_at TIMESTAMPTZ NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT post_targets_platform_check CHECK (
        platform IN ('youtube', 'tiktok')
    ),

    CONSTRAINT post_targets_status_check CHECK (
        status IN (
            'pending',
            'processing',
            'published',
            'failed',
            'cancelled'
        )
    ),

    CONSTRAINT post_targets_unique_account_per_post UNIQUE (
        post_id,
        social_account_id
    )
);

CREATE INDEX idx_post_targets_post_id
ON post_targets(post_id);

CREATE INDEX idx_post_targets_social_account_id
ON post_targets(social_account_id);

CREATE INDEX idx_post_targets_platform
ON post_targets(platform);

CREATE INDEX idx_post_targets_status
ON post_targets(status);

CREATE INDEX idx_post_targets_published_at
ON post_targets(published_at);

CREATE TRIGGER trg_post_targets_updated_at
BEFORE UPDATE ON post_targets
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();