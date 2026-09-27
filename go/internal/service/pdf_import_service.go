package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"pengajian-backend/internal/model"
	"pengajian-backend/internal/util"
)

type ParsedPdfMeeting struct {
	Tanggal string          `json:"tanggal"`
	Hari    string          `json:"hari"`
	Acara   string          `json:"acara"`
	Tempat  string          `json:"tempat"`
	Peserta []ParsedPeserta `json:"peserta"`
	Agenda  []string        `json:"agenda"`
	Warning string          `json:"warning,omitempty"`
}

type ParsedPeserta struct {
	Nama    string `json:"nama"`
	Jabatan string `json:"jabatan"`
	Alamat  string `json:"alamat"`
	Hadir   bool   `json:"hadir"`
}

type PDFImportService struct {
	aiSvc *AIService
}

func NewPDFImportService(aiSvc *AIService) *PDFImportService {
	return &PDFImportService{aiSvc: aiSvc}
}

const pdfParsePrompt = `Kamu adalah parser PDF untuk aplikasi Manajemen Pengajian.

TUGAS: ekstrak data dari teks PDF (biasanya daftar hadir rapat atau undangan pengajian) menjadi JSON.

OUTPUT WAJIB: JSON murni tanpa markdown code block, tanpa penjelasan, tanpa preamble.

SCHEMA:
{
  "tanggal": "YYYY-MM-DD",
  "hari": "Minggu|Senin|Selasa|Rabu|Kamis|Jumat|Sabtu",
  "acara": "judul rapat/pengajian",
  "tempat": "lokasi kalau ada",
  "peserta": [
    { "nama": "...", "jabatan": "...", "alamat": "...", "hadir": true }
  ],
  "agenda": ["poin 1", "poin 2"]
}

ATURAN:
- Kalau tanggal tidak ada, pakai "".
- Kalau tahun tidak tertulis, asumsikan tahun sekarang.
- Format tanggal di PDF bisa "14-09-2026", "14 September 2026", atau "Senin, 14/09/2026" — konversi ke YYYY-MM-DD.
- Untuk peserta, "HADIR/TIDAK HADIR" atau status apapun → boolean hadir (true kalau hadir).
- Kalau ada kolom "IZIN" → hadir: false.
- Agenda: ambil dari bagian bernomor/bulleted list kalau ada.
- Kalau field tidak ada, isi dengan "" atau [] (jangan null).

TEKS PDF:
`

func (s *PDFImportService) ParsePDF(ctx context.Context, rawText string) (*ParsedPdfMeeting, error) {
	if strings.TrimSpace(rawText) == "" {
		return nil, errors.New("teks PDF kosong")
	}
	if len(rawText) > 50_000 {
		rawText = rawText[:50_000]
	}

	provider := s.aiSvc.ActiveProvider()
	if provider == nil {
		return nil, errors.New("tidak ada provider AI yang tersedia")
	}

	messages := []model.LLMMessage{
		{Role: "system", Content: "Kamu adalah parser PDF. Output WAJIB JSON murni tanpa markdown."},
		{Role: "user", Content: pdfParsePrompt + "\n\n" + rawText},
	}

	result, err := provider.Chat(ctx, messages, nil)
	if err != nil {
		return nil, fmt.Errorf("AI gagal parse: %w", err)
	}

	content := strings.TrimSpace(result.Content)
	content = stripCodeBlock(content)

	var parsed ParsedPdfMeeting
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, fmt.Errorf("AI tidak mengembalikan JSON valid: %w\nraw: %s", err, truncateStr(content, 300))
	}

	if parsed.Tanggal != "" {
		if t, err := util.ParseFlexibleDate(parsed.Tanggal); err == nil {
			parsed.Tanggal = t.Format("2006-01-02")
			if parsed.Hari == "" {
				parsed.Hari = util.GetHariFromDate(t)
			}
		}
	}

	if parsed.Tanggal == "" {
		parsed.Warning = appendWarning(parsed.Warning, "Tanggal tidak terdeteksi — perlu diisi manual")
	}
	if len(parsed.Peserta) == 0 {
		parsed.Warning = appendWarning(parsed.Warning, "Tidak ada peserta terdeteksi")
	}
	if parsed.Acara == "" {
		parsed.Warning = appendWarning(parsed.Warning, "Acara kosong — perlu diisi manual")
	}

	return &parsed, nil
}

func stripCodeBlock(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		if i := strings.Index(s, "\n"); i != -1 {
			s = s[i+1:]
		}
		s = strings.TrimSuffix(strings.TrimSpace(s), "```")
	}
	return strings.TrimSpace(s)
}

func appendWarning(existing, extra string) string {
	if existing == "" {
		return extra
	}
	return existing + "; " + extra
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
