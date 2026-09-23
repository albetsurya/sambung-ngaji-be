-- 000010: Jadwal petugas shalat Jumat (khatib, muadzin, dll).
CREATE TABLE IF NOT EXISTS friday_schedules (
    friday_id      text PRIMARY KEY,
    tanggal        date NOT NULL UNIQUE,
    khatib_imam    text NOT NULL DEFAULT '',
    muadzin        text NOT NULL DEFAULT '',
    penasihat      text NOT NULL DEFAULT '',
    petugas_parkir text NOT NULL DEFAULT '',
    penata_sandal  text NOT NULL DEFAULT '',
    catatan        text NOT NULL DEFAULT '',
    created_by     text,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_friday_schedules_tanggal ON friday_schedules (tanggal);
