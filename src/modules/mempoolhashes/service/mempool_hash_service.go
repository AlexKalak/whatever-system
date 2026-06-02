package service

import (
	"context"
	"strings"
	"time"

	"github.com/alexkalak/whatever-system/src/modules/mempoolhashes/repository"
)

type MempoolHashService interface {
	Save(ctx context.Context, hash string, chainID uint, timestamp time.Time) error
}

type mempoolHashService struct {
	repo *repository.MempoolHashRepository
}

func NewMempoolHashService(repo *repository.MempoolHashRepository) MempoolHashService {
	return &mempoolHashService{repo: repo}
}

func (s *mempoolHashService) Save(ctx context.Context, hash string, chainID uint, timestamp time.Time) error {
	hash = strings.TrimSpace(hash)
	if hash == "" {
		return nil
	}
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	return s.repo.Save(ctx, hash, chainID, timestamp)
}
