CREATE TABLE IF NOT EXISTS rooms (
    id           BIGSERIAL PRIMARY KEY,
    workspace_id BIGINT NULL,
    name         VARCHAR(64) NOT NULL,  -- 'general'
    is_public    BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS rooms_name_unique ON rooms(name) WHERE workspace_id IS NULL;

CREATE TABLE IF NOT EXISTS messages (
    id           BIGSERIAL PRIMARY KEY,
    room_id      BIGINT    NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id      BIGINT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content      TEXT      NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    edited_at    TIMESTAMPTZ NULL,
    deleted_at   TIMESTAMPTZ NULL
);
-- keysey pagination
CREATE INDEX IF NOT EXISTS messages_room_created_at ON messages(room_id, created_at, id);

INSERT INTO rooms (workspace_id, name, is_public)
VALUES (NULL, 'general', TRUE)
ON CONFLICT ON CONSTRAINT rooms_name_unique DO NOTHING;