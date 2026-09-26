package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type StorageService struct {
	url        string
	serviceKey string
	bucket     string
	client     *http.Client
}

func NewStorageService(url, serviceKey, bucket string) *StorageService {
	return &StorageService{
		url:        strings.TrimRight(url, "/"),
		serviceKey: serviceKey,
		bucket:     bucket,
		client:     &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *StorageService) UploadPhoto(ctx context.Context, memberID, base64Data, mimeType string) (string, error) {
	if s.url == "" || s.serviceKey == "" || s.bucket == "" {
		return "", errors.New("Supabase Storage belum dikonfigurasi")
	}
	if memberID == "" {
		return "", errors.New("member_id wajib diisi")
	}
	if base64Data == "" || mimeType == "" {
		return "", errors.New("file foto wajib diisi")
	}

	allowed := map[string]string{
		"image/jpeg": "jpg",
		"image/png":  "png",
		"image/webp": "webp",
	}
	ext, ok := allowed[mimeType]
	if !ok {
		return "", errors.New("tipe file harus JPEG/PNG/WEBP")
	}

	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		data, err = base64.RawStdEncoding.DecodeString(base64Data)
		if err != nil {
			return "", errors.New("data foto tidak valid")
		}
	}
	if len(data) > 5*1024*1024 {
		return "", errors.New("ukuran foto maksimum 5MB")
	}

	filename := fmt.Sprintf("%s_%d.%s", memberID, time.Now().Unix(), ext)
	path := filename

	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.url, s.bucket, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Content-Type", mimeType)
	req.Header.Set("x-upsert", "true")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload storage: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("storage HTTP %d: %s", resp.StatusCode, string(body))
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.url, s.bucket, path)
	return publicURL, nil
}

func (s *StorageService) DeletePhoto(ctx context.Context, publicURL string) error {
	if publicURL == "" {
		return nil
	}
	prefix := fmt.Sprintf("%s/storage/v1/object/public/%s/", s.url, s.bucket)
	if !strings.HasPrefix(publicURL, prefix) {
		return nil
	}
	path := strings.TrimPrefix(publicURL, prefix)
	if path == "" {
		return nil
	}

	deleteURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.url, s.bucket, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, deleteURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("hapus storage: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 && resp.StatusCode != 404 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("hapus HTTP %d: %s", resp.StatusCode, body)
	}
	return nil
}
