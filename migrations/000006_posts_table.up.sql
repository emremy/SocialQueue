CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    media_asset_id UUID NOT NULL REFERENCES media_assets(id) ON DELETE RESTRICT,

    title VARCHAR(255) NOT NULL,
    description TEXT NULL,

    status VARCHAR(30) NOT NULL DEFAULT 'draft',

    scheduled_at TIMESTAMPTZ NULL,
    published_at TIMESTAMPTZ NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT posts_status_check CHECK (
        status IN (
            'draft',
            'scheduled',
            'processing',
            'published',
            'failed',
            'cancelled'
        )
    ),

    CONSTRAINT posts_title_not_empty_check CHECK (
        length(trim(title)) > 0
    )
);

CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_media_asset_id ON posts(media_asset_id);
CREATE INDEX idx_posts_status ON posts(status);
CREATE INDEX idx_posts_scheduled_at ON posts(scheduled_at);
CREATE INDEX idx_posts_created_at ON posts(created_at);

CREATE TRIGGER trg_posts_updated_at
BEFORE UPDATE ON posts
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();