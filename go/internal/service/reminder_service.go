package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"pengajian-backend/internal/repository"
)

// ReminderService — kirim reminder pengajian ke grup WA via Fonnte.
// Dipanggil dari cron goroutine di main.go setiap 1 jam.
type ReminderService struct {
	meetingRepo *repository.MeetingRepo
	fonnte      *FonnteService
	groupID     string
}

func NewReminderService(
	meetingRepo *repository.MeetingRepo,
	fonnte *FonnteService,
) *ReminderService {
	return &ReminderService{
		meetingRepo: meetingRepo,
		fonnte:      fonnte,
		groupID:     os.Getenv("FONNTE_REMINDER_GROUP"),
	}
}

// RunOnce — cari meeting dalam window H-7 s/d H-9 jam, kirim reminder kalau belum.
// Sengaja MATI secara default; aktifkan eksplisit via MEETING_REMINDER_ENABLED=true
// agar mengaktifkan Fonnte untuk keperluan lain (misal reminder Jumat) tidak
// ikut menghidupkan reminder meeting.
func (s *ReminderService) RunOnce(ctx context.Context) error {
	if os.Getenv("MEETING_REMINDER_ENABLED") != "true" {
		return nil
	}
	if !s.fonnte.IsEnabled() || s.groupID == "" {
		return nil
	}

	from := time.Now().Add(7 * time.Hour)
	to := time.Now().Add(9 * time.Hour)

	meetings, err := s.meetingRepo.FindPendingReminder(ctx, from, to)
	if err != nil {
		return fmt.Errorf("query meetings: %w", err)
	}

	if len(meetings) == 0 {
		return nil
	}

	for _, m := range meetings {
		msg := buildReminderMessage(m)
		if err := s.fonnte.SendText(ctx, s.groupID, msg); err != nil {
			log.Error().Err(err).Str("meeting_id", m.MeetingID).Msg("gagal kirim reminder")
			continue
		}
		if err := s.meetingRepo.MarkReminderSent(ctx, m.MeetingID); err != nil {
			log.Error().Err(err).Str("meeting_id", m.MeetingID).Msg("gagal update reminder_sent_at")
		}
		log.Info().Str("meeting_id", m.MeetingID).Str("acara", m.Acara).Msg("reminder terkirim")
	}

	return nil
}

func buildReminderMessage(m repository.ReminderMeetingRow) string {
	return fmt.Sprintf(
		"Assalamu'alaikum,\n\n"+
			"Reminder pengajian *%s* akan dilaksanakan:\n"+
			"📅 %s\n"+
			"🕐 %s\n\n"+
			"Mohon hadir tepat waktu. Barakallahu fiik.",
		m.Acara, m.Tanggal, m.Jam,
	)
}
