package service

import (
	"context"
	"encoding/json"
	"errors"

	"pengajian-backend/internal/repository"
)

type SettingsService struct {
	repo *repository.SettingsRepo
}

func NewSettingsService(repo *repository.SettingsRepo) *SettingsService {
	return &SettingsService{repo: repo}
}

func (s *SettingsService) GetSettings(ctx context.Context) (map[string]interface{}, error) {
	raw, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string]interface{}, len(raw))
	for k, v := range raw {
		var parsed interface{}
		if err := json.Unmarshal([]byte(v), &parsed); err == nil {
			out[k] = parsed
		} else {
			out[k] = v
		}
	}
	return out, nil
}

func (s *SettingsService) UpdateSettings(ctx context.Context, key string, value interface{}) error {
	if key == "" {
		return errors.New("key wajib diisi")
	}

	var encoded string
	switch v := value.(type) {
	case string:
		encoded = v
	default:
		b, err := json.Marshal(value)
		if err != nil {
			return err
		}
		encoded = string(b)
	}

	return s.repo.Upsert(ctx, key, encoded)
}
