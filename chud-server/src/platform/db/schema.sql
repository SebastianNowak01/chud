CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY,
    username      TEXT UNIQUE NOT NULL CHECK (LENGTH(username) BETWEEN 3 AND 32),
    password_hash TEXT        NOT NULL,
    is_admin      BOOLEAN     NOT NULL DEFAULT FALSE,

    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE users ADD COLUMN IF NOT EXISTS color TEXT NOT NULL DEFAULT '#4363d8';

CREATE TABLE IF NOT EXISTS activities (
    id          UUID PRIMARY KEY,
    name        TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_by  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS plans (
    id          UUID PRIMARY KEY,
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    monday      BOOLEAN NOT NULL DEFAULT FALSE,
    tuesday     BOOLEAN NOT NULL DEFAULT FALSE,
    wednesday   BOOLEAN NOT NULL DEFAULT FALSE,
    thursday    BOOLEAN NOT NULL DEFAULT FALSE,
    friday      BOOLEAN NOT NULL DEFAULT FALSE,
    saturday    BOOLEAN NOT NULL DEFAULT FALSE,
    sunday      BOOLEAN NOT NULL DEFAULT FALSE,
    starts_on   DATE NOT NULL,
    ends_on     DATE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS entries (
    id            UUID PRIMARY KEY,
    activity_id   UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id       UUID REFERENCES plans(id) ON DELETE SET NULL,
    scheduled_for DATE,
    excused       BOOLEAN NOT NULL DEFAULT FALSE,
    description   TEXT NOT NULL DEFAULT '',
    occurred_at   TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (plan_id, scheduled_for)
);

CREATE TABLE IF NOT EXISTS media (
    id           UUID PRIMARY KEY,
    entry_id     UUID NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    content_type TEXT NOT NULL,
    data         BYTEA NOT NULL
);
