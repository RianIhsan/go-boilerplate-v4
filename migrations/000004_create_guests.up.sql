CREATE TABLE IF NOT EXISTS guests (
    id            VARCHAR(36)  PRIMARY KEY,
    invitation_id VARCHAR(36)  NOT NULL REFERENCES invitations(id) ON DELETE CASCADE,
    name          VARCHAR(150) NOT NULL,
    phone         VARCHAR(20),
    email         VARCHAR(255),
    unique_token  VARCHAR(64)  NOT NULL UNIQUE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_guests_invitation_id ON guests(invitation_id);
