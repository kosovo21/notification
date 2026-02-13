-- 002_create_messages (UP)

CREATE TABLE messages (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id       UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject       VARCHAR(200) NOT NULL,
    body          TEXT         NOT NULL,
    sender        VARCHAR(100) NOT NULL,
    platform      VARCHAR(20)  NOT NULL CHECK (platform IN ('sms', 'whatsapp', 'telegram', 'email')),
    priority      SMALLINT     NOT NULL DEFAULT 1 CHECK (priority IN (0, 1, 2)),
    status        SMALLINT     NOT NULL DEFAULT 0,
    scheduled_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Status values:
-- 0 = Queued
-- 1 = Processing
-- 2 = Sent
-- 3 = Delivered
-- 4 = Failed
-- 5 = Pending (retry)
-- 6 = Cancelled

CREATE INDEX idx_messages_user_id ON messages (user_id);
CREATE INDEX idx_messages_status ON messages (status);
CREATE INDEX idx_messages_platform ON messages (platform);
CREATE INDEX idx_messages_created_at ON messages (created_at);
CREATE INDEX idx_messages_scheduled_at ON messages (scheduled_at) WHERE scheduled_at IS NOT NULL AND status = 0;
