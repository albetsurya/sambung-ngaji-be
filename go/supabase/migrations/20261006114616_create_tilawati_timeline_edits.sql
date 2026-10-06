
CREATE TABLE tilawati_timeline_edits (
    jilid INT NOT NULL,
    page INT NOT NULL,
    published_timeline_json JSONB NOT NULL,
    revision TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (jilid, page)
);

CREATE TRIGGER set_timestamp
BEFORE UPDATE ON tilawati_timeline_edits
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

-- Add a comment to the table and columns
COMMENT ON TABLE tilawati_timeline_edits IS 'Stores published timeline edits for Tilawati audio slicing, managed by admin.';
COMMENT ON COLUMN tilawati_timeline_edits.jilid IS 'Tilawati Jilid number.';
COMMENT ON COLUMN tilawati_timeline_edits.page IS 'Tilawati Page number within the Jilid.';
COMMENT ON COLUMN tilawati_timeline_edits.published_timeline_json IS 'JSON representation of the published timeline data (clips, texts, etc.).';
COMMENT ON COLUMN tilawati_timeline_edits.revision IS 'Unique revision identifier (e.g., timestamp) for this published timeline.';
COMMENT ON COLUMN tilawati_timeline_edits.created_at IS 'Timestamp of when the timeline was first published.';
COMMENT ON COLUMN tilawati_timeline_edits.created_at IS 'Timestamp of when the timeline was first published.';
COMMENT ON COLUMN tilawati_timeline_edits.updated_at IS 'Timestamp of the last update to the published timeline.';

