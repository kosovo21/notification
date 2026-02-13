-- 001_create_users (UP)

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email           VARCHAR(255) NOT NULL UNIQUE,
    api_key_hash    VARCHAR(255) NOT NULL UNIQUE,
    role            VARCHAR(50)  NOT NULL DEFAULT 'user',
    rate_limit_tier VARCHAR(50)  NOT NULL DEFAULT 'free',
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_api_key_hash ON users (api_key_hash);
CREATE INDEX idx_users_email ON users (email);
