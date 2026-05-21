package dto

import (
	"strings"
	"time"

	"github.com/alexkalak/whatever-system/src/modules/assets/entities"
	"github.com/google/uuid"
)

type AssetCreateRequest struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Type   string `json:"type"`
}

type AssetUpdateRequest struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Type   string `json:"type"`
}

type AssetResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Symbol    string    `json:"symbol"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (r AssetCreateRequest) Validate() map[string]string {
	return validateAssetPayload(r.Name, r.Symbol, r.Type)
}

func (r AssetUpdateRequest) Validate() map[string]string {
	return validateAssetPayload(r.Name, r.Symbol, r.Type)
}

func (r AssetCreateRequest) ToEntity() *entities.Asset {
	return &entities.Asset{
		Name:   strings.TrimSpace(r.Name),
		Symbol: strings.TrimSpace(r.Symbol),
		Type:   strings.ToLower(strings.TrimSpace(r.Type)),
	}
}

func (r AssetUpdateRequest) ToEntity() *entities.Asset {
	return &entities.Asset{
		Name:   strings.TrimSpace(r.Name),
		Symbol: strings.TrimSpace(r.Symbol),
		Type:   strings.ToLower(strings.TrimSpace(r.Type)),
	}
}

func ToAssetResponse(asset *entities.Asset) AssetResponse {
	return AssetResponse{
		ID:        asset.ID,
		Name:      asset.Name,
		Symbol:    asset.Symbol,
		Type:      asset.Type,
		CreatedAt: asset.CreatedAt,
		UpdatedAt: asset.UpdatedAt,
	}
}

func ToAssetResponses(assets []entities.Asset) []AssetResponse {
	result := make([]AssetResponse, 0, len(assets))
	for i := range assets {
		result = append(result, ToAssetResponse(&assets[i]))
	}
	return result
}

var allowedTypes = map[string]bool{
	"crypto":     true,
	"stock":      true,
	"fx":         true,
	"derivative": true,
}

func validateAssetPayload(name, symbol, assetType string) map[string]string {
	errors := map[string]string{}

	name = strings.TrimSpace(name)
	symbol = strings.TrimSpace(symbol)
	assetType = strings.ToLower(strings.TrimSpace(assetType))

	if name == "" {
		errors["name"] = "name is required"
	} else if len(name) > 50 {
		errors["name"] = "name max length is 50"
	}

	if symbol == "" {
		errors["symbol"] = "symbol is required"
	} else if len(symbol) > 50 {
		errors["symbol"] = "symbol max length is 50"
	}

	if assetType == "" {
		errors["type"] = "type is required"
	} else {
		if !allowedTypes[assetType] {
			errors["type"] = "type must be one of: crypto, stock, fx, derivative"
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}
