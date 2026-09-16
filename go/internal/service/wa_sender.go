package service

import (
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

	"pengajian-backend/internal/util"
)

type WASender struct {
	enabled bool
	url     string
	token   string
	client  *http.Client
}

func NewWASender() *WASender {
	return &WASender{
		enabled: strings.EqualFold(os.Getenv("WA_GATEWAY_ENABLED"), "true"),
		url:     os.Getenv("WA_GATEWAY_URL"),
		token:   os.Getenv("WA_GATEWAY_TOKEN"),
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

// Send: kirim pesan WA. Return nil kalau gateway disabled (soft-skip).
func (w *WASender) Send(ctx context.Context, noWA, message string) error {
	if !w.enabled {
		return nil
	}
	if w.url == "" || w.token == "" {
		return nil
	}
	target := util.NormalizePhone(noWA)
	if target == "" {
		return errors.New("nomor WA tidak valid")
	}

	form := url.Values{}
	form.Set("target", target)
	form.Set("message", message)
	form.Set("countryCode", "62")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", w.token)

	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("WA gateway HTTP %d: %s", resp.StatusCode, string(body))
	}

	var parsed struct {
		Status bool   `json:"status"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil && !parsed.Status {
		return fmt.Errorf("WA gagal: %s", parsed.Reason)
	}
	return nil
}
