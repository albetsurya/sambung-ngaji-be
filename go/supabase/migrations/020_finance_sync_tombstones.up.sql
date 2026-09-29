-- ============================================================
-- 000020: Finance sync delete tombstones
-- ============================================================

CREATE TABLE IF NOT EXISTS finance_sync_deleted (
  group_id    TEXT REFERENCES groups(group_id),
  entity      TEXT NOT NULL DEFAULT '',
  entity_id   TEXT NOT NULL DEFAULT '',
  deleted_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (group_id, entity, entity_id)
);
