-- 000012: Seed jadwal petugas Jumat (12 pekan, 25 Sep-11 Des 2026).
-- Sumber: data docker lokal. Idempotent via ON CONFLICT(tanggal).
INSERT INTO friday_schedules
    (friday_id, tanggal, khatib_imam, muadzin, penasihat, petugas_parkir, penata_sandal, catatan, created_by)
VALUES
('JMT26A5ABE4','2026-09-25','Bp H. Budi Mulyono','Bp Usman','Bp. H. Choirul Umam','Senkom','Kiki & King','','albetsurya'),
('JMT31882D3A','2026-10-02','Ust Albet','Bp Usman','Bp H. Didik Supriyono','Senkom','Midkal & Sabiq','','albetsurya'),
('JMTFF31A6AA','2026-10-09','Bp H. Didik Supriyono','Bp Usman','Bp H. Budi Mulyono','Senkom','Kiki & King','','albetsurya'),
('JMT40CC3E01','2026-10-16','Bp H. Budi Mulyono','Bp Usman','Ust Albet','Senkom','Midkal & Sabiq','','albetsurya'),
('JMTDDB2C31E','2026-10-23','Ust Albet','Bp Usman','Ust Umar Buang','Senkom','Kiki & King','','albetsurya'),
('JMTF4C20443','2026-10-30','Bp H. Didik Supriyono','Bp Usman','Ust Galih','Senkom','Midkal & Sabiq','','albetsurya'),
('JMTBAE9594A','2026-11-06','Bp H. Budi Mulyono','Bp Usman','Bp. H. Olron Lanbadat','Senkom','Kiki & King','','albetsurya'),
('JMTBD9F69D6','2026-11-13','Ust Albet','Bp Usman','Bp. H. Choirul Umam','Senkom','Midkal & Sabiq','','albetsurya'),
('JMTFC7F7376','2026-11-20','Bp H. Didik Supriyono','Bp Usman','Bp H. Didik Supriyono','Senkom','Kiki & King','','albetsurya'),
('JMT73D54345','2026-11-27','Bp H. Budi Mulyono','Bp Usman','Bp H. Budi Mulyono','Senkom','Midkal & Sabiq','','albetsurya'),
('JMT2F3A245A','2026-12-04','Ust Albet','Bp Usman','Ust Albet','Senkom','Kiki & King','','albetsurya'),
('JMTED7E0629','2026-12-11','Bp H. Didik Supriyono','Bp Usman','Ust Umar Buang','Senkom','Midkal & Sabiq','','albetsurya')
ON CONFLICT (tanggal) DO NOTHING;
