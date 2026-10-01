package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"pengajian-backend/internal/ai"
	"pengajian-backend/internal/auth"
	"pengajian-backend/internal/model"
	"pengajian-backend/internal/repository"
	"pengajian-backend/internal/util"
)

const (
	maxToolRounds = 6
	maxHistoryMsg = 10
)

type AIService struct {
	providers map[string]ai.Provider
	order     []string
	repo      *repository.AIRepo
	executor  *AIToolExecutor
	settings  *repository.SettingsRepo
}

func NewAIService(
	providers map[string]ai.Provider,
	order []string,
	repo *repository.AIRepo,
	executor *AIToolExecutor,
	settings *repository.SettingsRepo,
) *AIService {
	return &AIService{
		providers: providers,
		order:     order,
		repo:      repo,
		executor:  executor,
		settings:  settings,
	}
}

var aiDailyLimit = map[string]int{
	"SUPER_ADMIN": 100,
	"ADMIN":       50,
	"TIM_PNKB":    30,
	"TIM_ABSENSI": 30,
	"MEMBER":      10,
	"PENGAWAS":    30,
}

func (s *AIService) getStoredProvider(ctx context.Context) string {
	if s.settings == nil {
		return "auto"
	}
	all, err := s.settings.GetAll(ctx)
	if err != nil {
		return "auto"
	}
	p := strings.ToLower(strings.TrimSpace(all["AI_PROVIDER"]))
	if p == "" {
		return "auto"
	}
	return p
}

func (s *AIService) SetProvider(ctx context.Context, provider string) error {
	provider = strings.ToLower(strings.TrimSpace(provider))
	allowed := map[string]bool{"gemini": true, "groq": true, "nvidia": true, "auto": true}
	if !allowed[provider] {
		return errors.New("provider tidak valid")
	}
	return s.settings.Upsert(ctx, "AI_PROVIDER", provider)
}

func (s *AIService) GetProviderInfo(ctx context.Context) map[string]string {
	stored := s.getStoredProvider(ctx)
	active := stored
	if stored == "auto" {
		active = s.firstAvailable()
	}
	return map[string]string{
		"provider": stored,
		"active":   active,
	}
}

func (s *AIService) firstAvailable() string {
	for _, name := range s.order {
		if _, ok := s.providers[name]; ok {
			return name
		}
	}
	return "gemini"
}

func (s *AIService) buildChain(requested string) []string {
	requested = strings.ToLower(strings.TrimSpace(requested))
	chain := []string{}
	seen := map[string]bool{}

	addIfAvail := func(name string) {
		if seen[name] {
			return
		}
		if _, ok := s.providers[name]; ok {
			chain = append(chain, name)
			seen[name] = true
		}
	}

	if requested != "" && requested != "auto" {
		addIfAvail(requested)
	}
	for _, name := range s.order {
		addIfAvail(name)
	}
	return chain
}

func (s *AIService) Chat(ctx context.Context, user *model.User, req model.ChatRequest) (*model.ChatResponse, error) {
	limit := aiDailyLimit[user.Role]
	if limit > 0 {
		used, err := s.repo.GetQuota(ctx, user.UserID, time.Now())
		if err == nil && used >= limit {
			return nil, fmt.Errorf("kuota AI chat harian sudah habis (%d/%d). Coba lagi besok.", used, limit)
		}
	}

	isMember := user.Role == "MEMBER"
	memberID := ""
	if user.MemberID != nil {
		memberID = *user.MemberID
	}

	var tools []model.LLMToolDef
	var sysPrompt string
	if isMember {
		tools = ai.MemberTools()
		sysPrompt = buildSystemPromptMember(user.Nama)
	} else {
		canFinance := auth.CanAccess(user.Role, "getCashLedger")
		tools = ai.AdminTools(canFinance)
		sysPrompt = buildSystemPrompt(user)
	}

	messages := []model.LLMMessage{
		{Role: "system", Content: sysPrompt},
	}
	for _, h := range trimHistory(req.History, maxHistoryMsg) {
		role := "user"
		if h.Role == "assistant" {
			role = "assistant"
		}
		messages = append(messages, model.LLMMessage{
			Role:    role,
			Content: shortStr(h.Text, 1000),
		})
	}
	messages = append(messages, model.LLMMessage{
		Role:    "user",
		Content: req.Message,
	})

	requested := req.Provider
	if requested == "" {
		requested = s.getStoredProvider(ctx)
	}
	chain := s.buildChain(requested)
	if len(chain) == 0 {
		return nil, errors.New("tidak ada provider AI yang tersedia")
	}

	var lastErr error
	for _, name := range chain {
		provider := s.providers[name]
		resp, err := s.runProvider(ctx, provider, messages, tools, user, memberID, isMember)
		if err == nil {
			resp.Provider = name
			if name != requested && requested != "auto" {
				resp.RequestedProvider = requested
			}
			return resp, nil
		}
		lastErr = err
		if !ai.IsFallbackable(err) {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
	}
	return nil, fmt.Errorf("semua provider gagal: %w", lastErr)
}

func (s *AIService) ActiveProvider() ai.Provider {
	chain := s.buildChain(s.getStoredProvider(context.Background()))
	if len(chain) == 0 {
		return nil
	}
	return s.providers[chain[0]]
}

func (s *AIService) runProvider(
	ctx context.Context,
	provider ai.Provider,
	messages []model.LLMMessage,
	tools []model.LLMToolDef,
	user *model.User,
	memberID string,
	isMember bool,
) (*model.ChatResponse, error) {
	msgs := make([]model.LLMMessage, len(messages))
	copy(msgs, messages)

	var lastResult *ai.LLMResult
	rounds := 0

	for rounds < maxToolRounds {
		rounds++
		result, err := provider.Chat(ctx, msgs, tools)
		if err != nil {
			return nil, err
		}
		lastResult = result

		if len(result.ToolCalls) == 0 {
			break
		}

		msgs = append(msgs, model.LLMMessage{
			Role:      "assistant",
			Content:   result.Content,
			ToolCalls: result.ToolCalls,
		})

		for _, tc := range result.ToolCalls {
			args := map[string]interface{}{}
			if tc.Function.Arguments != "" {
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
			}
			toolResult, terr := s.executor.Execute(ctx, tc.Function.Name, args, user, memberID, isMember)
			var payload string
			if terr != nil {
				payload = `{"success":false,"message":` + jsonQuote(terr.Error()) + `}`
			} else {
				payload = toJSON(toolResult)
			}
			msgs = append(msgs, model.LLMMessage{
				Role:       "tool",
				ToolCallID: tc.ID,
				Content:    payload,
			})
		}
	}

	if lastResult == nil {
		return nil, errors.New("tidak ada respons dari provider")
	}

	_ = s.repo.InsertUsage(ctx, &model.AIUsageLog{
		UsageID:      util.NewID("USE"),
		UserID:       &user.UserID,
		UserNama:     user.Nama,
		Role:         user.Role,
		Provider:     provider.Name(),
		InputTokens:  lastResult.InputTokens,
		OutputTokens: lastResult.OutputTokens,
		TotalTokens:  lastResult.TotalTokens,
	})
	_ = s.repo.IncrementQuota(ctx, user.UserID, time.Now())

	reply := strings.TrimSpace(lastResult.Content)
	if reply == "" {
		reply = "Maaf, saya tidak mendapatkan jawaban."
	}
	return &model.ChatResponse{
		Reply: reply,
		Model: lastResult.Model,
	}, nil
}

func (s *AIService) GetUsageStats(ctx context.Context, user *model.User) (*model.AIUsageStats, error) {
	if user.Role != "SUPER_ADMIN" && user.Role != "ADMIN" {
		return nil, errors.New("hanya admin yang bisa akses monitoring AI")
	}
	rows, err := s.repo.GetUsageStats(ctx)
	if err != nil {
		return nil, err
	}

	today := time.Now().Format("2006-01-02")
	monthStart := time.Now().AddDate(0, 0, -30).Format("2006-01-02")

	var todaySummary, monthSummary model.AIUsageSummary
	byProvider := map[string]*model.AIUsageByField{}
	byRole := map[string]*model.AIUsageByField{}
	byUser := map[string]*model.AIUsageByUser{}

	for _, u := range rows {
		ts := u.Timestamp.Format("2006-01-02")
		if ts == today {
			todaySummary.ChatCount++
			todaySummary.TotalTokens += u.TotalTokens
		}
		if ts >= monthStart {
			monthSummary.ChatCount++
			monthSummary.TotalTokens += u.TotalTokens

			p := u.Provider
			if p == "" {
				p = "unknown"
			}
			if byProvider[p] == nil {
				byProvider[p] = &model.AIUsageByField{Field: p}
			}
			byProvider[p].ChatCount++
			byProvider[p].TotalTokens += u.TotalTokens

			r := u.Role
			if r == "" {
				r = "unknown"
			}
			if byRole[r] == nil {
				byRole[r] = &model.AIUsageByField{Role: r}
			}
			byRole[r].ChatCount++
			byRole[r].TotalTokens += u.TotalTokens

			uid := ""
			if u.UserID != nil {
				uid = *u.UserID
			}
			if uid != "" {
				if byUser[uid] == nil {
					byUser[uid] = &model.AIUsageByUser{
						UserID:   uid,
						UserNama: u.UserNama,
						Role:     u.Role,
					}
				}
				byUser[uid].ChatCount++
				byUser[uid].TotalTokens += u.TotalTokens
			}
		}
	}

	providerList := []model.AIUsageByField{}
	for _, v := range byProvider {
		providerList = append(providerList, *v)
	}
	roleList := []model.AIUsageByField{}
	for _, v := range byRole {
		roleList = append(roleList, *v)
	}
	userList := []model.AIUsageByUser{}
	for _, v := range byUser {
		userList = append(userList, *v)
	}
	sortUsageByCount(userList)

	return &model.AIUsageStats{
		Today:      todaySummary,
		Month:      monthSummary,
		ByProvider: providerList,
		ByRole:     roleList,
		TopUsers:   topN(userList, 10),
	}, nil
}

func buildSystemPrompt(user *model.User) string {
	today := time.Now().Format("Monday, 2 January 2006")
	canFinance := auth.CanAccess(user.Role, "getCashLedger")

	prompt := "Kamu adalah Asisten AI Profesional untuk aplikasi Sambung Ngaji.\n"
	prompt += "Domain utama: manajemen jamaah, kelompok pengajian, absensi, monitoring kehadiran, pengumuman"
	if canFinance {
		prompt += ", serta manajemen keuangan (kas ledger, iuran shodaqoh, zakat fitrah & mal)"
	}
	prompt += ".\n"
	prompt += "Hari ini: " + today + ".\n"
	prompt += "Kamu sedang berbicara dengan: " + user.Nama + " (Role: " + user.Role + ").\n\n"

	if canFinance {
		prompt += "HAK AKSES KEUANGAN:\n" +
			"- User ini (" + user.Role + ") MEMILIKI HAK AKSES data keuangan.\n" +
			"- Tools keuangan tersedia: get_finance_summary (kas utama/secunder), get_shodaqoh_summary (iuran bulanan), get_zakat_summary (zakat fitrah/mal).\n" +
			"- SELALU gunakan tools untuk data angka/nominal keuangan, JANGAN mengarang.\n" +
			"- Format nominal: **Rp X.XXX.XXX** dengan bold.\n" +
			"- Untuk ringkasan, tampilkan total pemasukan, pengeluaran, saldo, dan transaksi terbaru.\n" +
			"- Jika data kosong/tidak ditemukan, katakan 'Belum ada data untuk periode tersebut'.\n\n"
	} else {
		prompt += "HAK AKSES KEUANGAN:\n" +
			"- User ini (" + user.Role + ") TIDAK MEMILIKI HAK AKSES ke data keuangan.\n" +
			"- Tools keuangan DITUTUP untuk role ini.\n" +
			"- Jika user menanyakan saldo kas, iuran shodaqoh, atau zakat, JAWAB TEGAS: \"Maaf, role Anda (" + user.Role + ") tidak memiliki hak akses untuk melihat data keuangan.\"\n" +
			"- DILARANG KERAS memanggil tool keuangan (get_finance_summary, get_shodaqoh_summary, get_zakat_summary).\n\n"
	}

	prompt += "STANDAR FORMAT JAWABAN PROFESIONAL:\n" +
		"\n1. STRUKTUR DASAR (WAJIB):\n" +
		"   - Mulai dengan ringkasan 1-2 kalimat\n" +
		"   - Gunakan **JUDUL BAGIAN** dalam bold caps\n" +
		"   - Bullet list dengan '-' untuk item\n" +
		"   - Setiap item di baris terpisah\n" +
		"   - Beri jarak kosong antar bagian\n" +
		"   - Akhiri dengan KESIMPULAN/RINGKASAN\n" +
		"\n2. FORMAT TEKS UNIVERSAL:\n" +
		"   - Nominal uang: **Rp X.XXX.XXX** (selalu bold)\n" +
		"   - Nama orang: **Nama Lengkap** (bold)\n" +
		"   - Angka/Statistik: **123** (bold)\n" +
		"   - Tanggal: DD-MM-YYYY\n" +
		"   - Persentase: **95%** (bold)\n" +
		"   - Kode/ID: `ID123` (code format)\n" +
		"\n3. FORMAT KHUSUS KEUANGAN:\n" +
		"   - Kas: Saldo Awal **Rp X** → Pemasukan **Rp Y** → Pengeluaran **Rp Z** → Saldo Akhir **Rp W**\n" +
		"   - Shodaqoh: Target **Rp X** | Realisasi **Rp Y** | Pencapaian **Z%** | Sisa **Rp W**\n" +
		"   - Zakat: Jiwa **N** | Beras **N kg** | Uang **Rp X** | Alokasi: [kategori]\n" +
		"\n4. TATA BAHASA:\n" +
		"   - Bahasa Indonesia formal & profesional\n" +
		"   - Kalimat aktif, jelas, tidak bertele-tele\n" +
		"   - Hindari singkatan: tulis lengkap\n" +
		"   - Gunakan kata transisi untuk alur\n" +
		"\n5. TIPE JAWABAN:\n" +
		"   - Ringkasan Eksekutif: 3-5 poin utama\n" +
		"   - Detail Laporan: Maks 10 item/bagian\n" +
		"   - Daftar/Tabular: Urut relevansi/tanggal\n" +
		"   - Analisis: Temuan + Rekomendasi\n" +
		"\n6. VALIDASI DATA:\n" +
		"   - SELALU panggil tools untuk angka/nama/statistik\n" +
		"   - JANGAN mengarang atau menebak\n" +
		"   - Jika data kosong: 'Belum ada data untuk periode tersebut'\n" +
		"   - Jika data error: 'Sistem sedang maintenance, silakan coba lagi'\n" +
		"   - Validasi konsistensi angka\n" +
		"\n7. KEAMANAN & PRIVASI:\n" +
		"   - JANGAN bocorkan data sensitif non-keuangan\n" +
		"   - Hormati batas akses role user\n" +
		"   - Tidak menyebut user lain tanpa izin\n" +
		"   - Gunakan bahasa sopan & menghargai"

	return prompt
}

func buildSystemPromptMember(user string) string {
	today := time.Now().Format("Monday, 2 January 2006")
	return "Kamu adalah asisten AI pribadi untuk jamaah pengajian.\n" +
		"Hari ini: " + today + ".\n" +
		"Kamu sedang berbicara dengan: " + user + ".\n\n" +
		"ATURAN KETAT:\n" +
		"1. Kamu HANYA boleh menjawab pertanyaan tentang DATA DIRI user ini:\n" +
		"   - Biodata pribadi\n   - Riwayat absensi pribadi\n   - Statistik kehadiran pribadi\n   - Riwayat pembinaan pribadi\n   - Jadwal pengajian mendatang\n\n" +
		"2. Kamu DILARANG KERAS:\n" +
		"   - Menyebut nama jamaah lain\n   - Memberi data statistik global\n   - Menjawab tentang kelompok lain\n   - Membahas dashboard atau struktur organisasi\n\n" +
		"3. Kalau user bertanya tentang orang lain atau data global, jawab:\n" +
		"   Maaf, saya hanya bisa membantu dengan data pribadi Anda.\n\n" +
		"4. Jangan pernah mengarang data.\n\n" +
		"FORMAT JAWABAN:\n- Pakai bullet list dengan tanda '-'\n- Pakai **bold** untuk angka penting\n- Jawab dengan ramah dan hangat\n- Bahasa Indonesia"
}

func trimHistory(h []model.ChatHistory, max int) []model.ChatHistory {
	if len(h) <= max {
		return h
	}
	return h[len(h)-max:]
}

func shortStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func toJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return `{"error":"marshal"}`
	}
	return string(b)
}

func jsonQuote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func sortUsageByCount(list []model.AIUsageByUser) {
	for i := 1; i < len(list); i++ {
		for j := i; j > 0 && list[j].ChatCount > list[j-1].ChatCount; j-- {
			list[j], list[j-1] = list[j-1], list[j]
		}
	}
}

func topN(list []model.AIUsageByUser, n int) []model.AIUsageByUser {
	if len(list) <= n {
		return list
	}
	return list[:n]
}
