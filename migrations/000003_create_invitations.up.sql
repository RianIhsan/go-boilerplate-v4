CREATE TABLE IF NOT EXISTS invitations (
    id            VARCHAR(36)  PRIMARY KEY,
    user_id       VARCHAR(36)  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title         VARCHAR(255) NOT NULL,
    slug          VARCHAR(100) NOT NULL UNIQUE,
    event_type    VARCHAR(50)  NOT NULL,
    event_date    TIMESTAMPTZ  NOT NULL,
    venue_name    VARCHAR(255) NOT NULL,
    venue_address TEXT         NOT NULL,
    venue_lat     DOUBLE PRECISION,
    venue_lng     DOUBLE PRECISION,
    status        VARCHAR(20)  NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'expired')),
    is_published  BOOLEAN      NOT NULL DEFAULT FALSE,
    published_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invitations_user_id ON invitations(user_id);
