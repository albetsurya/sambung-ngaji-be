package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"pengajian-backend/internal/model"
)

type AnnouncementRepo struct {
	pool *pgxpool.Pool
}

func NewAnnouncementRepo(pool *pgxpool.Pool) *AnnouncementRepo {
	return &AnnouncementRepo{pool: pool}
}

/* ===== Templates ===== */

const templateSelectCols = `
	template_id, nama_template, kode, isi_template, status_aktif, created_at, updated_at`

func (r *AnnouncementRepo) FindTemplates(ctx context.Context, includeInactive bool) ([]model.AnnouncementTemplate, error) {
	q := `SELECT ` + templateSelectCols + ` FROM announcement_templates`
	if !includeInactive {
		q += ` WHERE status_aktif = true`
	}
	q += ` ORDER BY nama_template`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.AnnouncementTemplate
	for rows.Next() {
		var t model.AnnouncementTemplate
		if err := rows.Scan(
			&t.TemplateID, &t.NamaTemplate, &t.Kode, &t.IsiTemplate,
			&t.StatusAktif, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *AnnouncementRepo) FindTemplateByID(ctx context.Context, id string) (*model.AnnouncementTemplate, error) {
	var t model.AnnouncementTemplate
	err := r.pool.QueryRow(ctx,
		`SELECT `+templateSelectCols+` FROM announcement_templates WHERE template_id = $1`, id,
	).Scan(
		&t.TemplateID, &t.NamaTemplate, &t.Kode, &t.IsiTemplate,
		&t.StatusAktif, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *AnnouncementRepo) FindTemplateByKode(ctx context.Context, kode string) (*model.AnnouncementTemplate, error) {
	var t model.AnnouncementTemplate
	err := r.pool.QueryRow(ctx,
		`SELECT `+templateSelectCols+` FROM announcement_templates WHERE kode = $1`, kode,
	).Scan(
		&t.TemplateID, &t.NamaTemplate, &t.Kode, &t.IsiTemplate,
		&t.StatusAktif, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *AnnouncementRepo) InsertTemplate(ctx context.Context, t *model.AnnouncementTemplate) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO announcement_templates
		(template_id, nama_template, kode, isi_template, status_aktif, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,now(),now())
	`, t.TemplateID, t.NamaTemplate, t.Kode, t.IsiTemplate, t.StatusAktif)
	return err
}

func (r *AnnouncementRepo) UpdateTemplate(ctx context.Context, id string, patch map[string]interface{}) error {
	q := `UPDATE announcement_templates SET updated_at = now()`
	args := []interface{}{}
	n := 1
	for k, v := range patch {
		q += `, ` + k + ` = $` + itoa(n)
		args = append(args, v)
		n++
	}
	q += ` WHERE template_id = $` + itoa(n)
	args = append(args, id)
	_, err := r.pool.Exec(ctx, q, args...)
	return err
}

func (r *AnnouncementRepo) SoftDeleteTemplate(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE announcement_templates SET status_aktif = false, updated_at = now() WHERE template_id = $1`, id)
	return err
}

/* ===== Announcements ===== */

const announcementSelectCols = `
	announcement_id, template_id, meeting_id, group_id, tanggal, hari,
	jam, acara, materi, catatan, generated_text, status, created_by,
	created_at, updated_at`

func (r *AnnouncementRepo) FindAnnouncements(ctx context.Context, groupID, status string) ([]model.Announcement, error) {
	q := `SELECT ` + announcementSelectCols + ` FROM announcements WHERE 1=1`
	args := []interface{}{}
	n := 1
	if groupID != "" {
		q += ` AND group_id = $` + itoa(n)
		args = append(args, groupID)
		n++
	}
	if status != "" {
		q += ` AND status = $` + itoa(n)
		args = append(args, status)
	}
	q += ` ORDER BY tanggal DESC`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAnnouncements(rows)
}

func (r *AnnouncementRepo) FindAnnouncementByID(ctx context.Context, id string) (*model.Announcement, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+announcementSelectCols+` FROM announcements WHERE announcement_id = $1`, id)
	return scanAnnouncement(row)
}

func scanAnnouncement(s rowScanner) (*model.Announcement, error) {
	var a model.Announcement
	err := s.Scan(
		&a.AnnouncementID, &a.TemplateID, &a.MeetingID, &a.GroupID, &a.Tanggal, &a.Hari,
		&a.Jam, &a.Acara, &a.Materi, &a.Catatan, &a.GeneratedText, &a.Status, &a.CreatedBy,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func scanAnnouncements(rows rowsScanner) ([]model.Announcement, error) {
	var out []model.Announcement
	for rows.Next() {
		var a model.Announcement
		err := rows.Scan(
			&a.AnnouncementID, &a.TemplateID, &a.MeetingID, &a.GroupID, &a.Tanggal, &a.Hari,
			&a.Jam, &a.Acara, &a.Materi, &a.Catatan, &a.GeneratedText, &a.Status, &a.CreatedBy,
			&a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AnnouncementRepo) InsertAnnouncement(ctx context.Context, a *model.Announcement) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO announcements
		(announcement_id, template_id, meeting_id, group_id, tanggal, hari,
		 jam, acara, materi, catatan, generated_text, status, created_by,
		 created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,now(),now())
	`, a.AnnouncementID, a.TemplateID, a.MeetingID, a.GroupID, a.Tanggal, a.Hari,
		a.Jam, a.Acara, a.Materi, a.Catatan, a.GeneratedText, a.Status, a.CreatedBy)
	return err
}

func (r *AnnouncementRepo) UpdateAnnouncement(ctx context.Context, id string, patch map[string]interface{}) error {
	q := `UPDATE announcements SET updated_at = now()`
	args := []interface{}{}
	n := 1
	for k, v := range patch {
		q += `, ` + k + ` = $` + itoa(n)
		args = append(args, v)
		n++
	}
	q += ` WHERE announcement_id = $` + itoa(n)
	args = append(args, id)
	_, err := r.pool.Exec(ctx, q, args...)
	return err
}
