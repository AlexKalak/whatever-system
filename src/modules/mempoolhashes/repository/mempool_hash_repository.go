package repository

import (
	"context"
	"time"

	"github.com/alexkalak/whatever-system/src/modules/mempoolhashes/entities"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MempoolHashRepository struct {
	db *gorm.DB
}

func NewMempoolHashRepository(db *gorm.DB) *MempoolHashRepository {
	return &MempoolHashRepository{db: db}
}

func (r *MempoolHashRepository) Save(ctx context.Context, hash string, chainID uint, timestamp time.Time) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&entities.MempoolHash{
		Hash:      hash,
		ChainID:   chainID,
		Timestamp: timestamp,
	}).Error
}
