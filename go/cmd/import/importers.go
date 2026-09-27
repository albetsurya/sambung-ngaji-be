package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func importSettings(ctx context.Context, pool *pgxpool.Pool, rows []map[string]string) error {
	sql := `INSERT INTO settings (key, value, updated_at) VALUES ($1,$2,$3)
	        ON CONFLICT (key) DO UPDATE SET value=EXCLUDED.value, updated_at=EXCLUDED.updated_at`
	for i, r := range rows {
		key := str(r["key"])
		if key == "" {
			continue
		}
		upd := parseTime(r["updated_at"])
		if upd == nil {
			now := time.Now()
			upd = &now
		}
		if _, err := pool.Exec(ctx, sql, key, str(r["value"]), upd); err != nil {
			return fmt.Errorf("baris %d (%s): %w", i+2, key, err)
		}
	}
	return nil
}

func importAnnouncementTemplates(ctx context.Context, pool *pgxpool.Pool, rows []map[string]string) error {
	sql := `INSERT INTO announcement_templates
	        (template_id, nama_template, kode, isi_template, status_aktif, created_at, updated_at)
	        VALUES ($1,$2,$3,$4,$5,$6,$7)
	        ON CONFLICT (template_id) DO NOTHING`
	for i, r := range rows {
		id := str(r["template_id"])
		if id == "" {
			continue
		}
		ca := parseTime(r["created_at"])
		ua := parseTime(r["updated_at"])
		if ca == nil {
			now := time.Now()
			ca = &now
		}
		if ua == nil {
			ua = ca
		}
		if _, err := pool.Exec(ctx, sql,
			id, str(r["nama_template"]), str(r["kode"]), str(r["isi_template"]),
			parseBool(r["status_aktif"]), ca, ua,
		); err != nil {
			return fmt.Errorf("baris %d (%s): %w", i+2, id, err)
		}
	}
	return nil
}

func importMembers(ctx context.Context, pool *pgxpool.Pool, rows []map[string]string) error {
	sql := `INSERT INTO members (
	  member_id, nama_lengkap, nama_panggilan, jenis_kelamin,
	  tempat_lahir, tanggal_lahir, foto_url, no_wa,
	  alamat_rumah, desa, daerah, kelompok,
	  is_muballigh, is_kerja, is_nikah, tinggi_badan, berat_badan,
	  hobi, pekerjaan, status_pembinaan, status_aktif,
	  tanggal_masuk, tanggal_keluar,
	  jenjang_pendidikan, sekolah, jurusan,
	  tahun_mulai_pendidikan, tahun_selesai_pendidikan,
	  created_at, updated_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30)
	ON CONFLICT (member_id) DO NOTHING`

	for i, r := range rows {
		id := str(r["member_id"])
		if id == "" {
			continue
		}
		ca := parseTime(r["created_at"])
		ua := parseTime(r["updated_at"])
		if ca == nil {
			now := time.Now()
			ca = &now
		}
		if ua == nil {
			ua = ca
		}
		if _, err := pool.Exec(ctx, sql,
			id,
			str(r["nama_lengkap"]),
			str(r["nama_panggilan"]),
			sPtr(parseGender(r["jenis_kelamin"])),
			str(r["tempat_lahir"]),
			tPtr(parseDate(r["tanggal_lahir"])),
			str(r["foto_url"]),
			str(r["no_wa"]),
			str(r["alamat_rumah"]),
			str(r["desa"]),
			str(r["daerah"]),
			str(r["kelompok"]),
			parseBool(r["is_muballigh"]),
			parseBool(r["is_kerja"]),
			parseBool(r["is_nikah"]),
			str(r["tinggi_badan"]),
			str(r["berat_badan"]),
			str(r["hobi"]),
			str(r["pekerjaan"]),
			strDef(r["status_pembinaan"], "AKTIF"),
			parseBool(r["status_aktif"]),
			tPtr(parseDate(r["tanggal_masuk"])),
			tPtr(parseDate(r["tanggal_keluar"])),
			str(r["jenjang_pendidikan"]),
			str(r["sekolah"]),
			str(r["jurusan"]),
			str(r["tahun_mulai_pendidikan"]),
			str(r["tahun_selesai_pendidikan"]),
			ca, ua,
		); err != nil {
			return fmt.Errorf("baris %d (%s): %w", i+2, id, err)
		}
	}
	return nil
}

func importGroups(ctx context.Context, pool *pgxpool.Pool, rows []map[string]string) error {
	sql := `INSERT INTO groups
	        (group_id, group_code, group_name, pembina, penandatangan, jadwal, status_aktif, created_at, updated_at)
	        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	        ON CONFLICT (group_id) DO NOTHING`
	for i, r := range rows {
		id := str(r["group_id"])
		if id == "" {
			continue
		}
		ca := parseTime(r["created_at"])
		ua := parseTime(r["updated_at"])
		if ca == nil {
			now := time.Now()
			ca = &now
		}
		if ua == nil {
			ua = ca
		}
		if _, err := pool.Exec(ctx, sql,
			id, str(r["group_code"]), str(r["group_name"]),
			str(r["pembina"]), str(r["penandatangan"]), str(r["jadwal"]),
			parseBool(r["status_aktif"]), ca, ua,
		); err != nil {
			return fmt.Errorf("baris %d (%s): %w", i+2, id, err)
		}
	}
	return nil
}

func importUsers(ctx context.Context, pool *pgxpool.Pool, rows []map[string]string) error {
	existing := map[string]bool{}
	rs, err := pool.Query(ctx, "SELECT member_id FROM members")
	if err != nil {
		return fmt.Errorf("load members: %w", err)
	}
	for rs.Next() {
		var id string
		_ = rs.Scan(&id)
		existing[id] = true
	}
	rs.Close()

	sql := `INSERT INTO users
	        (user_id, username, password_hash, nama, role, member_id, status_aktif, created_at, updated_at, last_login_at)
	        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	        ON CONFLICT (user_id) DO NOTHING`

	for i, r := range rows {
		id := str(r["user_id"])
		username := str(r["username"])
		hash := str(r["password_hash"])
		if id == "" || username == "" || hash == "" {
			continue
		}

		memberID := str(r["member_id"])
		var memberPtr *string
		if memberID != "" {
			if existing[memberID] {
				memberPtr = &memberID
			} else {
				fmt.Printf("  [users] warn: %s member_id %s tidak ada di members → NULL\n", id, memberID)
			}
		}

		ca := parseTime(r["created_at"])
		ua := parseTime(r["updated_at"])
		if ca == nil {
			now := time.Now()
			ca = &now
		}
		if ua == nil {
			ua = ca
		}
		if _, err := pool.Exec(ctx, sql,
			id, username, hash, str(r["nama"]), str(r["role"]),
			sPtr(memberPtr), parseBool(r["status_aktif"]),
			ca, ua, tPtr(parseTime(r["last_login_at"])),
		); err != nil {
			return fmt.Errorf("baris %d (%s): %w", i+2, id, err)
		}
	}
	return nil
}

func importMeetings(ctx context.Context, pool *pgxpool.Pool, rows []map[string]string) error {
	sql := `INSERT INTO meetings
	        (meeting_id, tanggal, hari, jam, jam_start, group_id, acara, materi, status, catatan, kategori_target, created_by, created_at, updated_at)
	        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::jsonb,$12,$13,$14)
	        ON CONFLICT (meeting_id) DO NOTHING`
	for i, r := range rows {
		id := str(r["meeting_id"])
		if id == "" {
			continue
		}
		kat := strDef(r["kategori_target"], "[]")
		if kat == "" {
			kat = "[]"
		}
		ca := parseTime(r["created_at"])
		ua := parseTime(r["updated_at"])
		if ca == nil {
			now := time.Now()
			ca = &now
		}
		if ua == nil {
			ua = ca
		}
		groupID := str(r["group_id"])
		var groupPtr *string
		if groupID != "" {
			groupPtr = &groupID
		}
		if _, err := pool.Exec(ctx, sql,
			id,
			tPtr(parseDate(r["tanggal"])),
			str(r["hari"]),
			str(r["jam"]),
			str(r["jam_start"]),
			sPtr(groupPtr),
			str(r["acara"]),
			str(r["materi"]),
			strDef(r["status"], "SCHEDULED"),
			str(r["catatan"]),
			kat,
			str(r["created_by"]),
			ca, ua,
		); err != nil {
			return fmt.Errorf("baris %d (%s): %w", i+2, id, err)
		}
	}
	return nil
}

func importAttendance(ctx context.Context, pool *pgxpool.Pool, rows []map[string]string) error {
	sql := `INSERT INTO attendance
	        (attendance_id, meeting_id, member_id, status, catatan, created_by, created_at, updated_at)
	        VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	        ON CONFLICT (attendance_id) DO NOTHING`
	for i, r := range rows {
		id := str(r["attendance_id"])
		if id == "" {
			continue
		}
		ca := parseTime(r["created_at"])
		ua := parseTime(r["updated_at"])
		if ca == nil {
			now := time.Now()
			ca = &now
		}
		if ua == nil {
			ua = ca
		}
		if _, err := pool.Exec(ctx, sql,
			id, str(r["meeting_id"]), str(r["member_id"]),
			str(r["status"]), str(r["catatan"]), str(r["created_by"]),
			ca, ua,
		); err != nil {
			return fmt.Errorf("baris %d (%s): %w", i+2, id, err)
		}
	}
	return nil
}

func importMonitoring(ctx context.Context, pool *pgxpool.Pool, rows []map[string]string) error {
	sql := `INSERT INTO monitoring
	        (monitoring_id, member_id, tanggal, jenis, status, catatan, tindak_lanjut, created_by, created_at, updated_at)
	        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	        ON CONFLICT (monitoring_id) DO NOTHING`
	for i, r := range rows {
		id := str(r["monitoring_id"])
		if id == "" {
			continue
		}
		ca := parseTime(r["created_at"])
		ua := parseTime(r["updated_at"])
		if ca == nil {
			now := time.Now()
			ca = &now
		}
		if ua == nil {
			ua = ca
		}
		if _, err := pool.Exec(ctx, sql,
			id, str(r["member_id"]),
			tPtr(parseDate(r["tanggal"])),
			strDef(r["jenis"], "UMUM"),
			strDef(r["status"], "AKTIF"),
			str(r["catatan"]), str(r["tindak_lanjut"]),
			str(r["created_by"]),
			ca, ua,
		); err != nil {
			return fmt.Errorf("baris %d (%s): %w", i+2, id, err)
		}
	}
	return nil
}

func importAnnouncements(ctx context.Context, pool *pgxpool.Pool, rows []map[string]string) error {
	sql := `INSERT INTO announcements
	        (announcement_id, template_id, meeting_id, group_id, tanggal, hari, jam, acara, materi, catatan, generated_text, status, created_by, created_at, updated_at)
	        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
	        ON CONFLICT (announcement_id) DO NOTHING`
	for i, r := range rows {
		id := str(r["announcement_id"])
		if id == "" {
			continue
		}
		ca := parseTime(r["created_at"])
		ua := parseTime(r["updated_at"])
		if ca == nil {
			now := time.Now()
			ca = &now
		}
		if ua == nil {
			ua = ca
		}
		if _, err := pool.Exec(ctx, sql,
			id, str(r["template_id"]), str(r["meeting_id"]), str(r["group_id"]),
			tPtr(parseDate(r["tanggal"])),
			str(r["hari"]), str(r["jam"]), str(r["acara"]),
			str(r["materi"]), str(r["catatan"]), str(r["generated_text"]),
			strDef(r["status"], "DRAFT"),
			str(r["created_by"]),
			ca, ua,
		); err != nil {
			return fmt.Errorf("baris %d (%s): %w", i+2, id, err)
		}
	}
	return nil
}

func importPendingMembers(ctx context.Context, pool *pgxpool.Pool, rows []map[string]string) error {
	sql := `INSERT INTO pending_members
	        (submission_id, nama_lengkap, nama_panggilan, jenis_kelamin, tempat_lahir, tanggal_lahir,
	         no_wa, alamat_rumah, desa, daerah, pekerjaan, hobi, is_nikah, jenjang_pendidikan, sekolah, jurusan,
	         tahun_mulai_pendidikan, tahun_selesai_pendidikan, foto_url, username, password_hash, status,
	         submitted_at, submitted_ip, reviewed_by, reviewed_at, rejection_reason, created_member_id)
	        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28)
	        ON CONFLICT (submission_id) DO NOTHING`
	for i, r := range rows {
		id := str(r["submission_id"])
		if id == "" {
			continue
		}
		sa := parseTime(r["submitted_at"])
		if sa == nil {
			now := time.Now()
			sa = &now
		}
		reviewedBy := str(r["reviewed_by"])
		var rvBy *string
		if reviewedBy != "" {
			rvBy = &reviewedBy
		}
		createdMemberID := str(r["created_member_id"])
		var cmPtr *string
		if createdMemberID != "" {
			cmPtr = &createdMemberID
		}
		if _, err := pool.Exec(ctx, sql,
			id,
			str(r["nama_lengkap"]),
			str(r["nama_panggilan"]),
			sPtr(parseGender(r["jenis_kelamin"])),
			str(r["tempat_lahir"]),
			tPtr(parseDate(r["tanggal_lahir"])),
			str(r["no_wa"]),
			str(r["alamat_rumah"]),
			str(r["desa"]),
			str(r["daerah"]),
			str(r["pekerjaan"]),
			str(r["hobi"]),
			parseBool(r["is_nikah"]),
			str(r["jenjang_pendidikan"]),
			str(r["sekolah"]),
			str(r["jurusan"]),
			str(r["tahun_mulai_pendidikan"]),
			str(r["tahun_selesai_pendidikan"]),
			str(r["foto_url"]),
			strDef(r["username"], ""),
			strDef(r["password_hash"], ""),
			strDef(r["status"], "PENDING"),
			sa, str(r["submitted_ip"]),
			sPtr(rvBy), tPtr(parseTime(r["reviewed_at"])),
			str(r["rejection_reason"]),
			sPtr(cmPtr),
		); err != nil {
			return fmt.Errorf("baris %d (%s): %w", i+2, id, err)
		}
	}
	return nil
}

func importAuditLogs(ctx context.Context, pool *pgxpool.Pool, rows []map[string]string) error {
	sql := `INSERT INTO audit_logs
	        (log_id, user_id, user_nama, action, target_type, target_id, timestamp)
	        VALUES ($1,$2,$3,$4,$5,$6,$7)
	        ON CONFLICT (log_id) DO NOTHING`
	for i, r := range rows {
		id := str(r["log_id"])
		if id == "" {
			continue
		}
		ts := parseTime(r["timestamp"])
		if ts == nil {
			now := time.Now()
			ts = &now
		}
		userID := str(r["user_id"])
		var uidPtr *string
		if userID != "" {
			uidPtr = &userID
		}
		if _, err := pool.Exec(ctx, sql,
			id, sPtr(uidPtr), str(r["user_nama"]), str(r["action"]),
			str(r["target_type"]), str(r["target_id"]), ts,
		); err != nil {
			return fmt.Errorf("baris %d (%s): %w", i+2, id, err)
		}
	}
	return nil
}

func importAiUsage(ctx context.Context, pool *pgxpool.Pool, rows []map[string]string) error {
	sql := `INSERT INTO ai_usage
	        (usage_id, user_id, user_nama, role, provider, input_tokens, output_tokens, total_tokens, timestamp)
	        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	        ON CONFLICT (usage_id) DO NOTHING`
	for i, r := range rows {
		id := str(r["usage_id"])
		if id == "" {
			continue
		}
		ts := parseTime(r["timestamp"])
		if ts == nil {
			now := time.Now()
			ts = &now
		}
		userID := str(r["user_id"])
		var uidPtr *string
		if userID != "" {
			uidPtr = &userID
		}
		if _, err := pool.Exec(ctx, sql,
			id, sPtr(uidPtr), str(r["user_nama"]), str(r["role"]), str(r["provider"]),
			atoi(r["input_tokens"]), atoi(r["output_tokens"]), atoi(r["total_tokens"]),
			ts,
		); err != nil {
			return fmt.Errorf("baris %d (%s): %w", i+2, id, err)
		}
	}
	return nil
}

func atoi(s string) int {
	s = str(s)
	if s == "" {
		return 0
	}
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

func importWaQueue(ctx context.Context, pool *pgxpool.Pool, rows []map[string]string) error {
	sql := `INSERT INTO wa_queue
	        (queue_id, meeting_id, meeting_date, jam_start, send_at, status, template_id, member_count, sent_count, failed_count, error_log, sent_at, created_at, updated_at)
	        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	        ON CONFLICT (queue_id) DO NOTHING`
	for i, r := range rows {
		id := str(r["queue_id"])
		if id == "" {
			continue
		}
		ca := parseTime(r["created_at"])
		ua := parseTime(r["updated_at"])
		if ca == nil {
			now := time.Now()
			ca = &now
		}
		if ua == nil {
			ua = ca
		}
		meetingID := str(r["meeting_id"])
		var mPtr *string
		if meetingID != "" {
			mPtr = &meetingID
		}
		if _, err := pool.Exec(ctx, sql,
			id, sPtr(mPtr), tPtr(parseDate(r["meeting_date"])),
			str(r["jam_start"]), tPtr(parseTime(r["send_at"])),
			strDef(r["status"], "PENDING"),
			str(r["template_id"]),
			atoi(r["member_count"]), atoi(r["sent_count"]), atoi(r["failed_count"]),
			str(r["error_log"]), tPtr(parseTime(r["sent_at"])),
			ca, ua,
		); err != nil {
			return fmt.Errorf("baris %d (%s): %w", i+2, id, err)
		}
	}
	return nil
}
