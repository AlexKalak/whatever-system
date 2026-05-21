package repository

import (
	"github.com/alexkalak/whatever-system/src/modules/tokens/entities"
	"gorm.io/gorm"
)

type TokenRepository interface {
	Create(token *entities.Token) error
	GetAll() ([]entities.Token, error)
	GetByChainIDAndAddress(chainID uint, address string) (*entities.Token, error)
	Update(token *entities.Token) error
	Delete(chainID uint, address string) error
}

type tokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) Create(token *entities.Token) error {
	return r.db.Create(token).Error
}

func (r *tokenRepository) GetAll() ([]entities.Token, error) {
	var tokens []entities.Token
	err := r.db.Find(&tokens).Error
	return tokens, err
}

func (r *tokenRepository) GetByChainIDAndAddress(chainID uint, address string) (*entities.Token, error) {
	var token entities.Token
	err := r.db.First(&token, "chain_id = ? AND address = ?", chainID, address).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *tokenRepository) Update(token *entities.Token) error {
	return r.db.Save(token).Error
}

func (r *tokenRepository) Delete(chainID uint, address string) error {
	return r.db.Delete(&entities.Token{}, "chain_id = ? AND address = ?", chainID, address).Error
}
