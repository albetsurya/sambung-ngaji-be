package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// FonnteService — unofficial WhatsApp API (fonnte.com).
// Kirim teks ke nomor atau grup via REST API.
type FonnteService struct {
	enabled    bool
	token      string
	apiURL     string
	httpClient *http.Client
}

func NewFonnteService() *FonnteService {
	return &FonnteService{
		enabled:    os.Getenv("FONNTE_ENABLED") == "true",
		token:      os.Getenv("FONNTE_TOKEN"),
		apiURL:     getEnvOr("FONNTE_API_URL", "https://api.fonnte.com/send"),
		httpClient: &http.Client{Timeout: 20 * time.Second},
	}
}

func (s *FonnteService) IsEnabled() bool {
	return s.enabled && s.token != ""
}

// SendText — kirim pesan teks ke target (nomor atau group ID).
func (s *FonnteService) SendText(ctx context.Context, target, message string) error {
	if !s.IsEnabled() {
		return errors.New("fonnte service disabled")
	}
	if strings.TrimSpace(target) == "" {
		return errors.New("target wajib diisi")
	}
	if strings.TrimSpace(message) == "" {
		return errors.New("message wajib diisi")
	}

	form := url.Values{}
	form.Set("target", target)
	form.Set("message", message)
	form.Set("countryCode", "62")
	form.Set("delay", "2")

	req, err := http.NewRequestWithContext(
		ctx, "POST", s.apiURL, bytes.NewBufferString(form.Encode()),
	)
	if err != nil {
		return fmt.Errorf("buat request: %w", err)
	}
	req.Header.Set("Authorization", s.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("kirim request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("fonnte error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Status bool   `json:"status"`
		Detail string `json:"detail"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parse response: %w (body: %s)", err, string(body))
	}
	if !result.Status {
		return fmt.Errorf("fonnte gagal: %s", result.Detail)
	}

	return nil
}

func getEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
