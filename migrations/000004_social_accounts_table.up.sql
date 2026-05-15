CREATE TABLE social_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    platform VARCHAR(30) NOT NULL,
    provider_account_id VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NULL,
    username VARCHAR(255) NULL,
    avatar_url TEXT NULL,

    access_token TEXT NOT NULL,
    refresh_token TEXT NULL,
    token_expires_at TIMESTAMPTZ NULL,

    status VARCHAR(30) NOT NULL DEFAULT 'connected',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT social_accounts_platform_check CHECK (
        platform IN ('youtube', 'tiktok')
    ),

    CONSTRAINT social_accounts_status_check CHECK (
        status IN ('connected', 'expired', 'revoked', 'disabled')
    ),

    CONSTRAINT social_accounts_unique_provider_account UNIQUE (
        platform,
        provider_account_id
    )
);

CREATE INDEX idx_social_accounts_user_id ON social_accounts(user_id);
CREATE INDEX idx_social_accounts_platform ON social_accounts(platform);
CREATE INDEX idx_social_accounts_status ON social_accounts(status);
CREATE INDEX idx_social_accounts_token_expires_at ON social_accounts(token_expires_at);

CREATE TRIGGER trg_social_accounts_updated_at
BEFORE UPDATE ON social_accounts
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();