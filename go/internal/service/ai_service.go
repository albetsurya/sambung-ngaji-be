package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"pengajian-backend/internal/ai"
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
	order     []string // urutan fallback
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

// getStoredProvider: baca preferensi dari settings. Default "auto".
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

// SetProvider: simpan preferensi provider.
func (s *AIService) SetProvider(ctx context.Context, provider string) error {
	provider = strings.ToLower(strings.TrimSpace(provider))
	allowed := map[string]bool{"omniroute": true, "gemini": true, "groq": true, "auto": true}
	if !allowed[provider] {
		return errors.New("provider tidak valid")
	}
	return s.settings.Upsert(ctx, "AI_PROVIDER", provider)
}

// GetProviderInfo: return stored + active.
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

// firstAvailable: pilih provider pertama yang tersedia.
func (s *AIService) firstAvailable() string {
	for _, name := range s.order {
		if _, ok := s.providers[name]; ok {
			return name
		}
	}
	return "omniroute"
}

// buildChain: tentukan urutan provider yang dicoba.
// Kalau requested != "auto", provider itu dulu, lalu fallback ke sisanya.
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

// Chat: entry point utama.
func (s *AIService) Chat(ctx context.Context, user *model.User, req model.ChatRequest) (*model.ChatResponse, error) {
	// 1. Cek quota
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

	// 2. Pilih tools + system prompt
	var tools []model.LLMToolDef
	var sysPrompt string
	if isMember {
		tools = ai.MemberTools()
		sysPrompt = buildSystemPromptMember(user.Nama)
	} else {
		tools = ai.AdminTools()
		sysPrompt = buildSystemPrompt(user.Nama)
	}

	// 3. Bangun messages
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

	// 4. Tentukan provider chain
	requested := req.Provider
	if requested == "" {
		requested = s.getStoredProvider(ctx)
	}
	chain := s.buildChain(requested)
	if len(chain) == 0 {
		return nil, errors.New("tidak ada provider AI yang tersedia")
	}

	// 5. Coba setiap provider sampai sukses
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
		// Kalau bukan quota error, tidak usah fallback
		if !ai.IsQuotaError(err) {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
	}
	return nil, fmt.Errorf("semua provider gagal: %w", lastErr)
}

// runProvider: jalankan loop tool calling untuk satu provider.
func (s *AIService) runProvider(
	ctx context.Context,
	provider ai.Provider,
	messages []model.LLMMessage,
	tools []model.LLMToolDef,
	user *model.User,
	memberID string,
	isMember bool,
) (*model.ChatResponse, error) {
	// copy messages biar tidak mengganggu fallback berikutnya
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
			toolResult, terr := s.executor.Execute(ctx, tc.Function.Name, args, user.UserID, memberID, isMember)
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

	// Log usage (fire-and-forget)
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

// GetUsageStats: untuk admin monitoring.
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

/* ===== Helpers ===== */

func buildSystemPrompt(user string) string {
	today := time.Now().Format("Monday, 2 January 2006")
	return "Kamu adalah asisten AI untuk aplikasi Manajemen Pengajian yang mencakup data jamaah, kelompok, absensi, monitoring, dan pengumuman.\n" +
		"Hari ini: " + today + ".\n" +
		"Kamu sedang berbicara dengan: " + user + ".\n\n" +
		"FORMAT JAWABAN (WAJIB DIIKUTI):\n" +
		"1. Selalu pakai bullet list dengan tanda '-' untuk daftar.\n" +
		"2. Setiap item di baris terpisah.\n" +
		"3. Pakai **bold** untuk nama orang dan angka penting.\n" +
		"4. Beri jarak kosong antar bagian.\n" +
		"5. Akhiri dengan ringkasan singkat atau catatan penting.\n\n" +
		"PENTING:\n" +
		"- Gunakan tools untuk mengambil data ASLI sebelum menjawab pertanyaan yang menyebut angka, nama, atau statistik.\n" +
		"- JANGAN pernah mengarang atau menebak data.\n" +
		"- Jika data tidak ditemukan, katakan dengan jujur.\n" +
		"- Jawab dalam Bahasa Indonesia yang ramah dan profesional."
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
