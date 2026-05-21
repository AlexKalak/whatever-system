package dto

import (
	"strings"
	"time"

	"github.com/alexkalak/whatever-system/src/modules/exchanges/dex/entities"
)

type DexCreateRequest struct {
	DexType string `json:"dexType"`
	ChainID uint64 `json:"chainId"`
	Address string `json:"address"`
	FeeTier uint   `json:"feeTier"`
}

type DexUpdateRequest struct {
	DexType string `json:"dexType"`
	ChainID uint64 `json:"chainId"`
	Address string `json:"address"`
	FeeTier uint   `json:"feeTier"`
}

type DexResponse struct {
	ID        string    `json:"id"`
	DexType   string    `json:"dexType"`
	ChainID   uint64    `json:"chainId"`
	Address   string    `json:"address"`
	FeeTier   uint      `json:"feeTier"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

var allowedDexTypes = map[string]bool{
	"uniswapv2":     true,
	"uniswapv3":     true,
	"pancakeswapv3": true,
	"sushiswapv3":   true,
}

func (r DexCreateRequest) Validate() map[string]string {
	return validateDexPayload(r.DexType, r.Address)
}

func (r DexUpdateRequest) Validate() map[string]string {
	return validateDexPayload(r.DexType, r.Address)
}

func (r DexCreateRequest) ToEntity() *entities.Dex {
	return &entities.Dex{DexType: strings.ToLower(strings.TrimSpace(r.DexType)), ChainID: r.ChainID, Address: strings.ToLower(strings.TrimSpace(r.Address)), FeeTier: r.FeeTier}
}

func (r DexUpdateRequest) ToEntity() *entities.Dex {
	return &entities.Dex{DexType: strings.ToLower(strings.TrimSpace(r.DexType)), ChainID: r.ChainID, Address: strings.ToLower(strings.TrimSpace(r.Address)), FeeTier: r.FeeTier}
}

func ToDexResponse(dex *entities.Dex) DexResponse {
	return DexResponse{ID: dex.ID.String(), DexType: dex.DexType, ChainID: dex.ChainID, Address: dex.Address, FeeTier: dex.FeeTier, CreatedAt: dex.CreatedAt, UpdatedAt: dex.UpdatedAt}
}

func ToDexResponses(dexes []entities.Dex) []DexResponse {
	result := make([]DexResponse, 0, len(dexes))
	for i := range dexes {
		result = append(result, ToDexResponse(&dexes[i]))
	}
	return result
}

func validateDexPayload(dexType, address string) map[string]string {
	errors := map[string]string{}

	dexType = strings.ToLower(strings.TrimSpace(dexType))
	if dexType == "" {
		errors["dexType"] = "dexType is required"
	} else if !allowedDexTypes[dexType] {
		errors["dexType"] = "dexType must be one of: uniswapv2, uniswapv3, pancakeswapv3, sushiswapv3"
	}

	address = strings.TrimSpace(address)
	if address == "" {
		errors["address"] = "address is required"
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}
