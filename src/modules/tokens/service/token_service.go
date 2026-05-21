package service

import (
	"context"
	"errors"
	"strings"

	chainservice "github.com/alexkalak/whatever-system/src/modules/chain/service"
	"github.com/alexkalak/whatever-system/src/modules/tokens/entities"
	"github.com/alexkalak/whatever-system/src/modules/tokens/repository"
	"gorm.io/gorm"
)

type TokenService interface {
	Create(token *entities.Token) error
	GetAll() ([]entities.Token, error)
	GetByChainIDAndAddress(chainID uint, address string) (*entities.Token, error)
	Update(chainID uint, address string, payload *entities.Token) (*entities.Token, error)
	Delete(chainID uint, address string) error
	EnsureTokenExists(ctx context.Context, chainID uint, address string) error
}

type tokenService struct {
	repo      repository.TokenRepository
	chainData chainservice.ChainDataService
}

func NewTokenService(repo repository.TokenRepository, chainData chainservice.ChainDataService) TokenService {
	return &tokenService{repo: repo, chainData: chainData}
}

func (s *tokenService) Create(token *entities.Token) error {
	return s.repo.Create(token)
}

func (s *tokenService) GetAll() ([]entities.Token, error) {
	return s.repo.GetAll()
}

func (s *tokenService) GetByChainIDAndAddress(chainID uint, address string) (*entities.Token, error) {
	return s.repo.GetByChainIDAndAddress(chainID, address)
}

func (s *tokenService) Update(chainID uint, address string, payload *entities.Token) (*entities.Token, error) {
	token, err := s.repo.GetByChainIDAndAddress(chainID, address)
	if err != nil {
		return nil, err
	}

	token.Symbol = payload.Symbol
	token.Name = payload.Name
	token.Decimals = payload.Decimals

	if err := s.repo.Update(token); err != nil {
		return nil, err
	}

	return token, nil
}

func (s *tokenService) Delete(chainID uint, address string) error {
	_, err := s.repo.GetByChainIDAndAddress(chainID, address)
	if err != nil {
		return err
	}
	return s.repo.Delete(chainID, address)
}

func (s *tokenService) EnsureTokenExists(ctx context.Context, chainID uint, address string) error {
	address = strings.ToLower(strings.TrimSpace(address))
	_, err := s.repo.GetByChainIDAndAddress(chainID, address)
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	tokenInfo, err := s.chainData.GetTokenInfo(ctx, chainID, address)
	if err != nil {
		return err
	}

	return s.repo.Create(&entities.Token{
		ChainID:  tokenInfo.ChainID,
		Address:  tokenInfo.Address,
		Symbol:   tokenInfo.Symbol,
		Name:     tokenInfo.Name,
		Decimals: tokenInfo.Decimals,
	})
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
