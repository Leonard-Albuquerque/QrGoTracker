package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"qr-tracker/internal/config"
	"qr-tracker/internal/models"
	"qr-tracker/internal/repository"
	"qr-tracker/internal/utils"
)

type LinkService struct {
	repo repository.LinkRepository
	cfg  *config.Config
}

var (
	ErrInvalidURL           = errors.New("invalid_url")
	ErrInvalidCode          = errors.New("invalid_code")
	ErrCodeUnavailable      = errors.New("code_unavailable")
	ErrCodeGenerationFailed = errors.New("code_generation_failed")
)

func NewLinkService(r repository.LinkRepository, cfg *config.Config) *LinkService {
	return &LinkService{repo: r, cfg: cfg}
}

func (s *LinkService) Create(ctx context.Context, target, requestedCode string) (*models.Link, error) {
	if !utils.ValidURL(target) {
		return nil, ErrInvalidURL
	}
	idb := make([]byte, 12)
	_, _ = rand.Read(idb)
	id := hex.EncodeToString(idb)

	var code string
	requestedCode = strings.TrimSpace(requestedCode)

	if requestedCode != "" {
		if !utils.ValidShortCode(requestedCode) {
			return nil, ErrInvalidCode
		}
		existing, err := s.repo.GetByCode(requestedCode)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		if existing != nil {
			return nil, ErrCodeUnavailable
		}
		code = requestedCode
	} else {
		for i := 0; i < 10; i++ {
			code = utils.GenerateCode(7)
			existing, err := s.repo.GetByCode(code)
			if err != nil && err != sql.ErrNoRows {
				return nil, err
			}
			if existing == nil {
				break
			}
		}
		if code == "" {
			return nil, ErrCodeGenerationFailed
		}
	}

	link := &models.Link{
		ID:        id,
		Code:      code,
		TargetURL: target,
		Clicks:    0,
		CreatedAt: time.Now().UTC(),
		IsActive:  true,
	}
	if err := s.repo.Create(link); err != nil {
		return nil, err
	}
	return link, nil
}

func (s *LinkService) GetByCode(ctx context.Context, code string) (*models.Link, error) {
	return s.repo.GetByCode(code)
}

func (s *LinkService) TrackClick(ctx context.Context, code string) error {
	return s.repo.IncrementClick(code)
}

func (s *LinkService) BuildShortURL(code string) string {
	return fmt.Sprintf("%s/r/%s", s.cfg.BaseURL, code)
}

func (s *LinkService) BuildQRURL(code string) string {
	return fmt.Sprintf("%s/qr/%s.png", s.cfg.BaseURL, code)
}
