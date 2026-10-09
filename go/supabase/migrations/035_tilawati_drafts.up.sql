CREATE TABLE IF NOT EXISTS tilawati_timeline_edits (
    jilid INTEGER NOT NULL,
    page INTEGER NOT NULL,
    published_timeline_json JSONB NOT NULL,
    revision TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (jilid, page)
);

CREATE TABLE IF NOT EXISTS tilawati_timeline_drafts (
    user_id TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    jilid INTEGER NOT NULL CHECK (jilid BETWEEN 1 AND 6),
    page INTEGER NOT NULL CHECK (page BETWEEN 1 AND 44),
    timeline_json JSONB NOT NULL,
    base_revision TEXT NOT NULL DEFAULT '',
    updated_at_ms BIGINT NOT NULL,
    PRIMARY KEY (user_id, jilid, page)
);
