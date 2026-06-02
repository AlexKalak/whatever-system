package controller

import (
	"strings"

	"github.com/alexkalak/whatever-system/src/modules/tokens/entities"
	"github.com/alexkalak/whatever-system/src/modules/tokens/service"
	sharedhttp "github.com/alexkalak/whatever-system/src/shared/http"
	"github.com/gofiber/fiber/v2"
)

type TokenController struct {
	service service.TokenService
}

func NewTokenController(service service.TokenService) *TokenController {
	return &TokenController{service: service}
}

func (c *TokenController) RegisterRoutes(router fiber.Router) {
	router.Get("/", c.GetAll)
	router.Get("/:chainId/:address", c.GetByChainIDAndAddress)
}

func (c *TokenController) GetAll(ctx *fiber.Ctx) error {
	tokens, err := c.service.GetAll()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}
	return ctx.JSON(tokens)
}

func (c *TokenController) GetByChainIDAndAddress(ctx *fiber.Ctx) error {
	chainID, ok := sharedhttp.ParseUintParam(ctx, "chainId")
	if !ok {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid chainId"})
	}

	address := strings.ToLower(strings.TrimSpace(ctx.Params("address")))
	if address == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "address is required"})
	}

	token, err := c.service.GetByChainIDAndAddress(uint(chainID), address)
	if err != nil {
		if service.IsNotFound(err) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "token not found"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return ctx.JSON(toTokenResponse(token))
}

type TokenResponse struct {
	ChainID  uint   `json:"chainId"`
	Address  string `json:"address"`
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Decimals uint8  `json:"decimals"`
}

func toTokenResponse(token *entities.Token) TokenResponse {
	return TokenResponse{
		ChainID:  token.ChainID,
		Address:  token.Address,
		Symbol:   token.Symbol,
		Name:     token.Name,
		Decimals: token.Decimals,
	}
}
