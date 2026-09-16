package ai

import "pengajian-backend/internal/model"

// AdminTools: 9 tool dari AI_TOOL_DEFS_ di Apps Script.
func AdminTools() []model.LLMToolDef {
	return []model.LLMToolDef{
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
				Description: "Ambil daftar jamaah. Bisa difilter berdasarkan kategori, jenis kelamin, kelompok, atau pencarian nama.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"kategori":      map[string]interface{}{"type": "string", "description": "BALITA|CABERAWIT|PRA_REMAJA|REMAJA|PRA_NIKAH|DEWASA|ISTIMEWA"},
						"jenis_kelamin": map[string]interface{}{"type": "string", "description": "L atau P"},
						"kelompok":      map[string]interface{}{"type": "string", "description": "group_id"},
						"search":        map[string]interface{}{"type": "string", "description": "Kata kunci nama"},
						"limit":         map[string]interface{}{"type": "number", "description": "Maksimal hasil, default 50"},
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
				Description: "Ambil daftar kelompok pengajian beserta pembina dan jadwal.",
				Parameters:  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}, "required": []string{}},
			},
		},
		{
			Type: "function",
			Function: model.LLMToolDefFunc{
				Name:        "get_attendance_summary",
				Description: "Ambil ringkasan absensi: total hadir/ijin/sakit/alpa dalam periode tertentu.",
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
				Description: "Ambil daftar jadwal pengajian yang akan datang.",
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
}

// MemberTools: 5 tool dari AI_TOOL_DEFS_MEMBER_ di Apps Script.
func MemberTools() []model.LLMToolDef {
	return []model.LLMToolDef{
		{
			Type: "function",
			Function: model.LLMToolDefFunc{
				Name:        "get_my_profile",
				Description: "Ambil biodata diri sendiri (nama, usia, kelompok, alamat, pendidikan, dll).",
				Parameters:  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}, "required": []string{}},
			},
		},
		{
			Type: "function",
			Function: model.LLMToolDefFunc{
				Name:        "get_my_attendance",
				Description: "Ambil riwayat absensi pengajian diri sendiri (tanggal, acara, status).",
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
				Description: "Ambil jadwal pengajian yang akan datang untuk informasi pribadi.",
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
