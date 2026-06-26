CREATE TABLE IF NOT EXISTS rsvps (
    id             VARCHAR(36) PRIMARY KEY,
    invitation_id  VARCHAR(36) NOT NULL REFERENCES invitations(id) ON DELETE CASCADE,
    guest_id       VARCHAR(36) REFERENCES guests(id) ON DELETE SET NULL,
    name           VARCHAR(150) NOT NULL,
    status         VARCHAR(20) NOT NULL CHECK (status IN ('attending', 'not_attending', 'maybe')),
    attendee_count INT NOT NULL DEFAULT 1,
    message        TEXT,
    responded_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rsvps_invitation_id ON rsvps(invitation_id);
CREATE INDEX idx_rsvps_guest_id ON rsvps(guest_id);
