-- 003_create_message_recipients (UP)

CREATE TABLE message_recipients (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_id    UUID         NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    recipient     VARCHAR(255) NOT NULL,
    status        SMALLINT     NOT NULL DEFAULT 0,
    provider_id   VARCHAR(255),
    error_message TEXT,
    retry_count   SMALLINT     NOT NULL DEFAULT 0,
    sent_at       TIMESTAMPTZ,
    delivered_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_recipients_message_id ON message_recipients (message_id);
CREATE INDEX idx_recipients_status ON message_recipients (status);
CREATE INDEX idx_recipients_recipient ON message_recipients (recipient);
