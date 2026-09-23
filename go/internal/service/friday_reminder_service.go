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

// FridayReminderService — kirim info petugas Jumat besok ke grup WA pengurus
// via Fonnte. Dipanggil dari cron tiap 1 jam, tapi hanya benar-benar mengirim
// pada hari Kamis jam 12 siang WIB untuk Jumat keesokan harinya, sekali per
// jadwal (kolom reminder_sent_at sebagai anti double-kirim).
//
// Override untuk pengujian:
//   - FRIDAY_REMINDER_FORCE=true → abaikan jendela hari/jam.
//   - FRIDAY_REMINDER_DATE=YYYY-MM-DD → tanggal target (default: besok).
type FridayReminderService struct {
	repo    *repository.FridayRepo
	fonnte  *FonnteService
	groupID string
}

func NewFridayReminderService(
	repo *repository.FridayRepo,
	fonnte *FonnteService,
) *FridayReminderService {
	return &FridayReminderService{
		repo:    repo,
		fonnte:  fonnte,
		groupID: os.Getenv("FONNTE_REMINDER_GROUP"),
	}
}

// RunOnce — kirim sekali kalau waktunya tepat dan ada jadwal belum terkirim.
func (s *FridayReminderService) RunOnce(ctx context.Context) error {
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
		// Hanya Kamis jam 12 siang (12:00–12:59 waktu lokal,
		// container TZ=Asia/Jakarta), dan target harus hari Jumat (besok).
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
