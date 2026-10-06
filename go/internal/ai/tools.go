package ai

import "pengajian-backend/internal/model"

func FinanceTools() []model.LLMToolDef {
	return []model.LLMToolDef{
		{
			Type: "function",
			Function: model.LLMToolDefFunc{
				Name:        "get_finance_summary",
				Description: "Ambil ringkasan kas (kas utama & kas kedua/event), total pemasukan, pengeluaran, saldo, dan daftar transaksi terbaru.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"cash_type": map[string]interface{}{"type": "string", "description": "main atau secondary"},
						"group_id":  map[string]interface{}{"type": "string", "description": "group_id group_label (opsional)"},
					},
					"required": []string{},
				},
			},
		},
		{
			Type: "function",
			Function: model.LLMToolDefFunc{
				Name:        "get_shodaqoh_summary",
				Description: "Ambil rekap iuran shodaqoh bulanan (total penerimaan, target, daftar anggota lunas & tunggakan).",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"group_id": map[string]interface{}{"type": "string", "description": "group_id group_label (opsional)"},
						"month":    map[string]interface{}{"type": "string", "description": "Format YYYY-MM (contoh: 2026-09)"},
					},
					"required": []string{},
				},
			},
		},
		{
			Type: "function",
			Function: model.LLMToolDefFunc{
				Name:        "get_zakat_summary",
				Description: "Ambil rekap data zakat fitrah & zakat mal (total jiwa, total beras, total uang, alokasi per kategori, daftar muzaki & mustahik).",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"group_id": map[string]interface{}{"type": "string", "description": "group_id group_label (opsional)"},
						"zakat_id": map[string]interface{}{"type": "string", "description": "zakat_id (opsional jika ingin detail zakat spesifik)"},
					},
					"required": []string{},
				},
			},
		},
	}
}

func AdminTools(canFinance bool) []model.LLMToolDef {
	tools := []model.LLMToolDef{
		{
			Type: "function",
			Function: model.LLMToolDefFunc{
				Name:        "get_dashboard_summary",
				Description: "Ambil ringkasan dashboard: total jamaah aktif, rata-rata kehadiran, jumlah yang perlu perhatian, dan data belum lengkap.",
				Parameters:  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}, "required": []string{}},
			},
		},
		{
			Type: "function",
			Function: model.LLMToolDefFunc{
				Name:        "get_members_list",
				Description: "Ambil daftar jamaah. Bisa difilter berdasarkan kategori, type kelamin, group_label, atau pencarian name.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"kategori":    map[string]interface{}{"type": "string", "description": "BALITA|CABERAWIT|PRA_REMAJA|REMAJA|PRA_NIKAH|DEWASA|ISTIMEWA"},
						"gender":      map[string]interface{}{"type": "string", "description": "L atau P"},
						"group_label": map[string]interface{}{"type": "string", "description": "group_id"},
						"search":      map[string]interface{}{"type": "string", "description": "Kata kunci name"},
						"limit":       map[string]interface{}{"type": "number", "description": "Maksimal hasil, default 50"},
					},
					"required": []string{},
				},
			},
		},
		{
			Type: "function",
			Function: model.LLMToolDefFunc{
				Name:        "get_member_detail",
				Description: "Ambil detail lengkap satu jamaah berdasarkan member_id.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"member_id": map[string]interface{}{"type": "string"},
					},
					"required": []string{"member_id"},
				},
			},
		},
		{
			Type: "function",
			Function: model.LLMToolDefFunc{
				Name:        "get_groups_list",
				Description: "Ambil daftar group_label pengajian beserta mentor dan schedule.",
				Parameters:  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}, "required": []string{}},
			},
		},
		{
			Type: "function",
			Function: model.LLMToolDefFunc{
				Name:        "get_attendance_summary",
				Description: "Ambil ringkasan absensi: total hadir/izin/sakit/alpa/dispensasi dalam periode tertentu.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"from":     map[string]interface{}{"type": "string", "description": "YYYY-MM-DD"},
						"to":       map[string]interface{}{"type": "string", "description": "YYYY-MM-DD"},
						"group_id": map[string]interface{}{"type": "string"},
					},
					"required": []string{},
				},
			},
		},
		{
			Type: "function",
			Function: model.LLMToolDefFunc{
				Name:        "get_attendance_by_meeting",
				Description: "Ambil detail absensi satu meeting.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"meeting_id": map[string]interface{}{"type": "string"},
					},
					"required": []string{"meeting_id"},
				},
			},
		},
		{
			Type: "function",
			Function: model.LLMToolDefFunc{
				Name:        "get_upcoming_meetings",
				Description: "Ambil daftar schedule pengajian yang akan datang.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"limit": map[string]interface{}{"type": "number"},
					},
					"required": []string{},
				},
			},
		},
		{
			Type: "function",
			Function: model.LLMToolDefFunc{
				Name:        "get_monitoring_list",
				Description: "Ambil daftar monitoring jamaah. Bisa filter status.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"status":    map[string]interface{}{"type": "string", "description": "AKTIF|PERLU_PERHATIAN|KURANG_AKTIF|TIDAK_AKTIF"},
						"member_id": map[string]interface{}{"type": "string"},
					},
					"required": []string{},
				},
			},
		},
		{
			Type: "function",
			Function: model.LLMToolDefFunc{
				Name:        "get_announcements_list",
				Description: "Ambil daftar pengumuman yang pernah dibuat.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"status":   map[string]interface{}{"type": "string", "description": "DRAFT|READY|SHARED|CANCELLED"},
						"group_id": map[string]interface{}{"type": "string"},
					},
					"required": []string{},
				},
			},
		},
	}
	if canFinance {
		tools = append(tools, FinanceTools()...)
	}
	return tools
}

func MemberTools() []model.LLMToolDef {
	return []model.LLMToolDef{
		{
			Type: "function",
			Function: model.LLMToolDefFunc{
				Name:        "get_my_profile",
				Description: "Ambil biodata diri sendiri (name, usia, group_label, alamat, pendidikan, dll).",
				Parameters:  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}, "required": []string{}},
			},
		},
		{
			Type: "function",
			Function: model.LLMToolDefFunc{
				Name:        "get_my_attendance",
				Description: "Ambil riwayat absensi pengajian diri sendiri (date, event, status).",
				Parameters:  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}, "required": []string{}},
			},
		},
		{
			Type: "function",
			Function: model.LLMToolDefFunc{
				Name:        "get_my_attendance_stats",
				Description: "Ambil statistik kehadiran pribadi (total hadir, ijin, sakit, alpa, persentase).",
				Parameters:  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}, "required": []string{}},
			},
		},
		{
			Type: "function",
			Function: model.LLMToolDefFunc{
				Name:        "get_my_monitoring",
				Description: "Ambil riwayat pembinaan/monitoring diri sendiri.",
				Parameters:  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}, "required": []string{}},
			},
		},
		{
			Type: "function",
			Function: model.LLMToolDefFunc{
				Name:        "get_my_upcoming_meetings",
				Description: "Ambil schedule pengajian yang akan datang untuk informasi pribadi.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"limit": map[string]interface{}{"type": "number"},
					},
					"required": []string{},
				},
			},
		},
	}
}
