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

ALTER TABLE activities ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ;

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
    plan_id       UUID,
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

CREATE TABLE IF NOT EXISTS week_summaries (
    week           DATE        NOT NULL,
    prompt_version INT         NOT NULL,
    text           TEXT        NOT NULL,
    input_hash     TEXT        NOT NULL,
    generated_at   TIMESTAMPTZ NOT NULL,
    expires_at     TIMESTAMPTZ,
    PRIMARY KEY (week, prompt_version)
);

ALTER TABLE entries DROP CONSTRAINT IF EXISTS entries_plan_id_fkey;

CREATE OR REPLACE FUNCTION pg_temp.add_constraint(tbl TEXT, name TEXT, definition TEXT) RETURNS VOID AS $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = name) THEN
        EXECUTE format('ALTER TABLE %I ADD CONSTRAINT %I %s', tbl, name, definition);
    END IF;
END
$$ LANGUAGE plpgsql;

SELECT pg_temp.add_constraint('plans', 'plans_has_day',
    'CHECK (monday OR tuesday OR wednesday OR thursday OR friday OR saturday OR sunday)');
SELECT pg_temp.add_constraint('plans', 'plans_dates',
    'CHECK (ends_on IS NULL OR ends_on >= starts_on)');
SELECT pg_temp.add_constraint('plans', 'plans_title',
    $c$CHECK (btrim(title) <> '')$c$);
SELECT pg_temp.add_constraint('plans', 'plans_title_length',
    'CHECK (char_length(title) <= 100) NOT VALID');
SELECT pg_temp.add_constraint('plans', 'plans_identity',
    'UNIQUE (id, activity_id, user_id)');

SELECT pg_temp.add_constraint('entries', 'entries_plan',
    'FOREIGN KEY (plan_id, activity_id, user_id) REFERENCES plans (id, activity_id, user_id)');
SELECT pg_temp.add_constraint('entries', 'entries_planned_day',
    'CHECK ((plan_id IS NULL) = (scheduled_for IS NULL))');
SELECT pg_temp.add_constraint('entries', 'entries_excuse_planned',
    'CHECK (NOT excused OR plan_id IS NOT NULL)');
SELECT pg_temp.add_constraint('entries', 'entries_excuse_reason',
    $c$CHECK (NOT excused OR btrim(description) <> '')$c$);
SELECT pg_temp.add_constraint('entries', 'entries_description_length',
    'CHECK (char_length(description) <= 2000) NOT VALID');
SELECT pg_temp.add_constraint('activities', 'activities_name_length',
    'CHECK (char_length(name) <= 50) NOT VALID');
SELECT pg_temp.add_constraint('activities', 'activities_description_length',
    'CHECK (char_length(description) <= 500) NOT VALID');
SELECT pg_temp.add_constraint('media', 'media_content_type',
    $c$CHECK (content_type IN ('image/jpeg', 'image/png', 'image/gif', 'image/webp', 'video/mp4', 'video/webm', 'video/quicktime')) NOT VALID$c$);

CREATE INDEX IF NOT EXISTS entries_occurred_at_idx ON entries (occurred_at);
CREATE INDEX IF NOT EXISTS entries_user_occurred_at_idx ON entries (user_id, occurred_at);
CREATE INDEX IF NOT EXISTS entries_activity_occurred_at_idx ON entries (activity_id, occurred_at);
CREATE INDEX IF NOT EXISTS entries_scheduled_for_idx ON entries (scheduled_for) WHERE plan_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS plans_activity_idx ON plans (activity_id);
CREATE INDEX IF NOT EXISTS media_entry_idx ON media (entry_id);
