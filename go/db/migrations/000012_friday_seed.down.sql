-- 000012 rollback: hapus seed jadwal petugas Jumat.
DELETE FROM friday_schedules WHERE tanggal IN ('2026-09-25','2026-10-02','2026-10-09','2026-10-16','2026-10-23','2026-10-30','2026-11-06','2026-11-13','2026-11-20','2026-11-27','2026-12-04','2026-12-11');
