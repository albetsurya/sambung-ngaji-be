-- Data migration: Title Case nama & tempat di members + pending_members
-- Skema tidak berubah, hanya normalisasi data existing.

UPDATE members SET
  nama_lengkap   = initcap(nama_lengkap),
  nama_panggilan = initcap(nama_panggilan),
  tempat_lahir   = initcap(tempat_lahir),
  desa           = initcap(desa),
  daerah         = initcap(daerah)
WHERE
  nama_lengkap   IS DISTINCT FROM initcap(nama_lengkap)
  OR nama_panggilan IS DISTINCT FROM initcap(nama_panggilan)
  OR tempat_lahir   IS DISTINCT FROM initcap(tempat_lahir)
  OR desa           IS DISTINCT FROM initcap(desa)
  OR daerah         IS DISTINCT FROM initcap(daerah);

UPDATE pending_members SET
  nama_lengkap   = initcap(nama_lengkap),
  nama_panggilan = initcap(nama_panggilan),
  desa           = initcap(desa),
  daerah         = initcap(daerah)
WHERE
  nama_lengkap   IS DISTINCT FROM initcap(nama_lengkap)
  OR nama_panggilan IS DISTINCT FROM initcap(nama_panggilan)
  OR desa           IS DISTINCT FROM initcap(desa)
  OR daerah         IS DISTINCT FROM initcap(daerah);
