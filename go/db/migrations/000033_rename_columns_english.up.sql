-- 033: Rename kolom Bahasa Indonesia -> English (snake_case) di semua tabel.
-- Idempotent: setiap rename dicek dulu via information_schema, aman di-rerun.
-- Nilai enum/data (HADIR, AKTIF, PRA_NIKAH, MUZAKKI, dsb) TIDAK diubah.
-- Nama tabel & istilah domain baku (shodaqoh, zakat, muadzin) TIDAK diubah.

DO $$
DECLARE
  m TEXT[][];
  r TEXT[];
BEGIN
  m := ARRAY[
    -- members
    ['members','nama_lengkap','full_name'],
    ['members','nama_panggilan','nickname'],
    ['members','jenis_kelamin','gender'],
    ['members','tempat_lahir','birth_place'],
    ['members','tanggal_lahir','birth_date'],
    ['members','foto_url','photo_url'],
    ['members','no_wa','whatsapp_number'],
    ['members','alamat_rumah','home_address'],
    ['members','desa','village'],
    ['members','daerah','region'],
    ['members','kelompok','group_label'],
    ['members','is_muballigh','is_preacher'],
    ['members','is_kerja','is_employed'],
    ['members','is_nikah','is_married'],
    ['members','tinggi_badan','height'],
    ['members','berat_badan','weight'],
    ['members','hobi','hobby'],
    ['members','pekerjaan','occupation'],
    ['members','status_pembinaan','mentoring_status'],
    ['members','status_aktif','is_active'],
    ['members','tanggal_masuk','joined_date'],
    ['members','tanggal_keluar','left_date'],
    ['members','jenjang_pendidikan','education_level'],
    ['members','sekolah','school'],
    ['members','jurusan','major'],
    ['members','tahun_mulai_pendidikan','education_start_year'],
    ['members','tahun_selesai_pendidikan','education_end_year'],
    -- pending_members
    ['pending_members','nama_lengkap','full_name'],
    ['pending_members','nama_panggilan','nickname'],
    ['pending_members','jenis_kelamin','gender'],
    ['pending_members','tempat_lahir','birth_place'],
    ['pending_members','tanggal_lahir','birth_date'],
    ['pending_members','foto_url','photo_url'],
    ['pending_members','no_wa','whatsapp_number'],
    ['pending_members','alamat_rumah','home_address'],
    ['pending_members','desa','village'],
    ['pending_members','daerah','region'],
    ['pending_members','is_nikah','is_married'],
    ['pending_members','pekerjaan','occupation'],
    ['pending_members','hobi','hobby'],
    ['pending_members','jenjang_pendidikan','education_level'],
    ['pending_members','sekolah','school'],
    ['pending_members','jurusan','major'],
    ['pending_members','tahun_mulai_pendidikan','education_start_year'],
    ['pending_members','tahun_selesai_pendidikan','education_end_year'],
    -- users
    ['users','nama','name'],
    ['users','status_aktif','is_active'],
    -- groups
    ['groups','pembina','mentor'],
    ['groups','penandatangan','signatory'],
    ['groups','jadwal','schedule'],
    ['groups','status_aktif','is_active'],
    -- meetings
    ['meetings','tanggal','date'],
    ['meetings','hari','day'],
    ['meetings','jam_start','start_time'],
    ['meetings','jam','time'],
    ['meetings','acara','event'],
    ['meetings','materi','topic'],
    ['meetings','catatan','notes'],
    ['meetings','kategori_target','target_categories'],
    -- announcements
    ['announcements','tanggal','date'],
    ['announcements','hari','day'],
    ['announcements','jam','time'],
    ['announcements','acara','event'],
    ['announcements','materi','topic'],
    ['announcements','catatan','notes'],
    -- announcement_templates
    ['announcement_templates','nama_template','template_name'],
    ['announcement_templates','isi_template','template_body'],
    ['announcement_templates','status_aktif','is_active'],
    -- attendance
    ['attendance','catatan','notes'],
    -- monitoring
    ['monitoring','tanggal','date'],
    ['monitoring','jenis','type'],
    ['monitoring','catatan','notes'],
    ['monitoring','tindak_lanjut','follow_up'],
    -- friday_schedules
    ['friday_schedules','tanggal','date'],
    ['friday_schedules','khatib_imam','sermon_leader'],
    ['friday_schedules','penasihat','advisor'],
    ['friday_schedules','petugas_parkir','parking_attendant'],
    ['friday_schedules','penata_sandal','footwear_attendant'],
    ['friday_schedules','catatan','notes'],
    -- kas_transactions
    ['kas_transactions','tanggal','date'],
    -- audit_logs
    ['audit_logs','user_nama','user_name'],
    -- ai_usage
    ['ai_usage','user_nama','user_name'],
    -- wa_queue
    ['wa_queue','jam_start','start_time']
  ];
  FOREACH r SLICE 1 IN ARRAY m LOOP
    IF EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_name = r[1] AND column_name = r[2]
    ) THEN
      EXECUTE format('ALTER TABLE %I RENAME COLUMN %I TO %I', r[1], r[2], r[3]);
    END IF;
  END LOOP;
END $$;
