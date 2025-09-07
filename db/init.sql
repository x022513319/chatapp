CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS users (
    id                  BIGSERIAL PRIMARY KEY,
    username            VARCHAR(32) NOT NULL UNIQUE,
    nickname            VARCHAR(64) NULL,
    email               CITEXT      NOT NULL UNIQUE, -- CITEXT: case-insensitive character string type
    password_hash       TEXT        NOT NULL, -- bcrypt hashed
    status              SMALLINT    NOT NULL DEFAULT 1, -- 1=active, 0=locked ... (forbidden/block...)
    last_login_at       TIMESTAMPTZ NULL,
    email_verified_at   TIMESTAMPTZ NULL, -- use for email verifying
    password_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_users_updated_at ON users;
CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION set_updated_at();