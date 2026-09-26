package service

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
)

type FridayReminderService struct {
	repo     *repository.FridayRepo
	settings *repository.SettingsRepo
	fonnte   *FonnteService
	groupID  string
}

func NewFridayReminderService(
	repo *repository.FridayRepo,
	fonnte *FonnteService,
	settings *repository.SettingsRepo,
) *FridayReminderService {
	return &FridayReminderService{
		repo:     repo,
		settings: settings,
		fonnte:   fonnte,
		groupID:  os.Getenv("FONNTE_REMINDER_GROUP"),
	}
}

type ReminderStatus struct {
	ServerNow     string                    `json:"server_now"`
	ServerWeekday string                    `json:"server_weekday"`
	FonnteEnabled bool                      `json:"fonnte_enabled"`
	GroupSet      bool                      `json:"group_set"`
	LastCronHit   string                    `json:"last_cron_hit"`
	LastSent      string                    `json:"last_sent"`
	NextWindow    string                    `json:"next_window"`
	Upcoming      []model.FridayScheduleDTO `json:"upcoming"`
}

func (s *FridayReminderService) GetReminderStatus(ctx context.Context) (*ReminderStatus, error) {
	now := time.Now()
	lastSent, _ := s.repo.MaxReminderSent(ctx)
	lastSentStr := ""
	if lastSent != nil {
		lastSentStr = *lastSent
	}
	lastCron := ""
	if s.settings != nil {
		if all, err := s.settings.GetAll(ctx); err == nil {
			lastCron = all["last_cron_hit"]
		}
	}
	upcomingDTO := []model.FridayScheduleDTO{}
	if rows, err := s.repo.FindUpcoming(ctx, 4); err == nil {
		for _, r := range rows {
			upcomingDTO = append(upcomingDTO, toFridayDTO(r))
		}
	}
	return &ReminderStatus{
		ServerNow:     now.Format("2006-01-02T15:04:05.000Z07:00"),
		ServerWeekday: now.Weekday().String(),
		FonnteEnabled: s.fonnte.IsEnabled(),
		GroupSet:      s.groupID != "",
		LastCronHit:   lastCron,
		LastSent:      lastSentStr,
		NextWindow:    nextThursdayNoon(now).Format("2006-01-02T15:04:05.000Z07:00"),
		Upcoming:      upcomingDTO,
	}, nil
}

func nextThursdayNoon(now time.Time) time.Time {
	d := now
	for {
		if d.Weekday() == time.Thursday {
			noon := time.Date(d.Year(), d.Month(), d.Day(), 12, 0, 0, 0, d.Location())
			if now.Before(noon) {
				return noon
			}
		}
		d = d.AddDate(0, 0, 1)
		if d.Weekday() == time.Thursday {
			return time.Date(d.Year(), d.Month(), d.Day(), 12, 0, 0, 0, d.Location())
		}
	}
}

func (s *FridayReminderService) RunOnce(ctx context.Context) error {
	if os.Getenv("FRIDAY_REMINDER_ENABLED") != "true" {
		return nil
	}
	if !s.fonnte.IsEnabled() || s.groupID == "" {
		return nil
	}

	force := os.Getenv("FRIDAY_REMINDER_FORCE") == "true"
	now := time.Now()

	target := os.Getenv("FRIDAY_REMINDER_DATE")
	if target == "" {
		target = now.AddDate(0, 0, 1).Format("2006-01-02")
	}
	tgl, err := time.Parse("2006-01-02", target)
	if err != nil {
		return errors.New("FRIDAY_REMINDER_DATE tidak valid (YYYY-MM-DD)")
	}

	if !force {
		if now.Weekday() != time.Thursday {
			return nil
		}
		if h := now.Hour(); h < 12 || h >= 13 {
			return nil
		}
		if tgl.Weekday() != time.Friday {
			return nil
		}
	}

	f, err := s.repo.FindUnsentByDate(ctx, target)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}

	msg := buildFridayReminderMessage(f)
	if err := s.fonnte.SendText(ctx, s.groupID, msg); err != nil {
		log.Error().Err(err).Str("tanggal", target).Msg("gagal kirim reminder jumat")
		return err
	}
	if err := s.repo.MarkReminderSent(ctx, target); err != nil {
		log.Error().Err(err).Str("tanggal", target).Msg("gagal tandai reminder jumat")
		return err
	}
	log.Info().Str("tanggal", target).Msg("reminder jumat terkirim")
	return nil
}

func orBelumDiisi(v string) string {
	if strings.TrimSpace(v) == "" {
		return "(belum diisi)"
	}
	return strings.TrimSpace(v)
}

func formatTanggalPanjang(t time.Time) string {
	bulan := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	hari := []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
	return hari[int(t.Weekday())] + ", " + strconv.Itoa(t.Day()) + " " +
		bulan[int(t.Month())-1] + " " + strconv.Itoa(t.Year())
}

func buildFridayReminderMessage(f *model.FridaySchedule) string {
	var b strings.Builder
	b.WriteString("╔════════════════════╗\n")
	b.WriteString("🕌 JADWAL PETUGAS SHALAT JUMAT\n")
	b.WriteString("🗓️ " + formatTanggalPanjang(f.Tanggal) + "\n")
	b.WriteString("╚════════════════════╝\n")
	b.WriteString("\n")
	b.WriteString("👤 Khatib & Imam : " + orBelumDiisi(f.KhatibImam) + "\n")
	b.WriteString("🎙️ Muadzin          : " + orBelumDiisi(f.Muadzin) + "\n")
	b.WriteString("📖 Penasihat        : " + orBelumDiisi(f.Penasihat) + "\n")
	b.WriteString("🚗 Petugas Parkir : " + orBelumDiisi(f.PetugasParkir) + "\n")
	b.WriteString("👞 Penata Sandal : " + orBelumDiisi(f.PenataSandal) + "\n")
	b.WriteString("\n")
	b.WriteString("Semoga Allah ﷻ memberikan pahala dan kebarokahan.\n")
	b.WriteString("\n")
	b.WriteString("جزاكم الله خيرًا")
	return b.String()
}

func (s *FridayReminderService) MarkSent(ctx context.Context, tanggal string) error {
	if tanggal == "" {
		return errors.New("tanggal wajib diisi (YYYY-MM-DD)")
	}
	if _, err := time.Parse("2006-01-02", tanggal); err != nil {
		return errors.New("tanggal tidak valid (YYYY-MM-DD)")
	}
	if _, err := s.repo.FindByDate(ctx, "", tanggal); err != nil {
		return errors.New("jadwal tidak ditemukan")
	}
	return s.repo.MarkReminderSent(ctx, tanggal)
}
