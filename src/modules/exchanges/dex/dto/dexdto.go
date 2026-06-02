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
}

type DexUpdateRequest struct {
	DexType string `json:"dexType"`
	ChainID uint64 `json:"chainId"`
	Address string `json:"address"`
}

type DexResponse struct {
	ID            string    `json:"id"`
	DexType       string    `json:"dexType"`
	ChainID       uint64    `json:"chainId"`
	Address       string    `json:"address"`
	Token0Address string    `json:"token0Address,omitempty"`
	Token1Address string    `json:"token1Address,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type DexPaginatedResponse struct {
	Items      []DexResponse `json:"items"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	Total      int64         `json:"total"`
	TotalPages int           `json:"totalPages"`
}

type UniswapV2DexResponse struct {
	DexResponse
	Token0Amount string `json:"token0Amount"`
	Token1Amount string `json:"token1Amount"`
	FeeTier      uint   `json:"feeTier"`
	ExchangeName string `json:"exchangeName"`
}

type UniswapV3DexResponse struct {
	DexResponse
	FeeTier      uint   `json:"feeTier"`
	SqrtPriceX96 string `json:"sqrtPriceX96"`
	Liquidity    string `json:"liquidity"`
	Tick         int64  `json:"tick"`
	TickSpacing  int64  `json:"tickSpacing"`
}

type UniswapV2DexPaginatedResponse struct {
	Items      []UniswapV2DexResponse `json:"items"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	Total      int64                  `json:"total"`
	TotalPages int                    `json:"totalPages"`
}

type UniswapV3DexPaginatedResponse struct {
	Items      []UniswapV3DexResponse `json:"items"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	Total      int64                  `json:"total"`
	TotalPages int                    `json:"totalPages"`
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
	return &entities.Dex{DexType: strings.ToLower(strings.TrimSpace(r.DexType)), ChainID: r.ChainID, Address: strings.ToLower(strings.TrimSpace(r.Address))}
}

func (r DexUpdateRequest) ToEntity() *entities.Dex {
	return &entities.Dex{DexType: strings.ToLower(strings.TrimSpace(r.DexType)), ChainID: r.ChainID, Address: strings.ToLower(strings.TrimSpace(r.Address))}
}

func ToDexResponse(dex *entities.Dex) DexResponse {
	return DexResponse{ID: dex.ID.String(), DexType: dex.DexType, ChainID: dex.ChainID, Address: dex.Address, CreatedAt: dex.CreatedAt, UpdatedAt: dex.UpdatedAt}
}

func ToDexResponses(dexes []entities.Dex) []DexResponse {
	result := make([]DexResponse, 0, len(dexes))
	for i := range dexes {
		result = append(result, ToDexResponse(&dexes[i]))
	}
	return result
}

func ToDexPaginatedResponse(dexes []entities.Dex, page, limit int, total int64) DexPaginatedResponse {
	return DexPaginatedResponse{Items: ToDexResponses(dexes), Page: page, Limit: limit, Total: total, TotalPages: totalPages(total, limit)}
}

func ToUniswapV2DexResponse(dex *entities.Dex, details *entities.UniswapV2Dex) UniswapV2DexResponse {
	response := UniswapV2DexResponse{DexResponse: ToDexResponse(dex)}
	if details == nil {
		return response
	}
	response.Token0Address = details.Token0Address
	response.Token1Address = details.Token1Address
	response.Token0Amount = details.Token0Amount
	response.Token1Amount = details.Token1Amount
	response.FeeTier = details.FeeTier
	response.ExchangeName = details.ExchangeName
	return response
}

func ToUniswapV3DexResponse(dex *entities.Dex, details *entities.UniswapV3Dex) UniswapV3DexResponse {
	response := UniswapV3DexResponse{DexResponse: ToDexResponse(dex)}
	if details == nil {
		return response
	}
	response.Token0Address = details.Token0Address
	response.Token1Address = details.Token1Address
	response.FeeTier = details.FeeTier
	if details.SqrtPriceX96.Int != nil {
		response.SqrtPriceX96 = details.SqrtPriceX96.String()
	}
	if details.Liquidity.Int != nil {
		response.Liquidity = details.Liquidity.String()
	}
	response.Tick = details.Tick
	response.TickSpacing = details.TickSpacing
	return response
}

func ToUniswapV2DexPaginatedResponse(items []UniswapV2DexResponse, page, limit int, total int64) UniswapV2DexPaginatedResponse {
	return UniswapV2DexPaginatedResponse{Items: items, Page: page, Limit: limit, Total: total, TotalPages: totalPages(total, limit)}
}

func ToUniswapV3DexPaginatedResponse(items []UniswapV3DexResponse, page, limit int, total int64) UniswapV3DexPaginatedResponse {
	return UniswapV3DexPaginatedResponse{Items: items, Page: page, Limit: limit, Total: total, TotalPages: totalPages(total, limit)}
}

func totalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	return int((total + int64(limit) - 1) / int64(limit))
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
