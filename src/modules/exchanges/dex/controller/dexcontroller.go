package controller

import (
	"errors"
	"strconv"
	"strings"

	"github.com/alexkalak/whatever-system/src/modules/exchanges/dex/dto"
	"github.com/alexkalak/whatever-system/src/modules/exchanges/dex/entities"
	"github.com/alexkalak/whatever-system/src/modules/exchanges/dex/repository"
	"github.com/alexkalak/whatever-system/src/modules/exchanges/dex/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DexController struct {
	service service.DexService
	v2Repo  repository.UniswapV2Repository
	v3Repo  repository.UniswapV3Repository
}

type DexControllerDeps struct {
	UniswapV2Repo repository.UniswapV2Repository
	UniswapV3Repo repository.UniswapV3Repository
}

func NewDexController(service service.DexService, deps ...DexControllerDeps) *DexController {
	controller := &DexController{service: service}
	if len(deps) > 0 {
		controller.v2Repo = deps[0].UniswapV2Repo
		controller.v3Repo = deps[0].UniswapV3Repo
	}
	return controller
}

func (c *DexController) RegisterRoutes(router fiber.Router) {
	router.Post("/", c.Create)
	router.Get("/", c.GetAll)
	router.Get("/chain/:chainId/address/:address", c.GetByChainIDAndAddress)
	router.Get("/uniswap-v2", c.GetUniswapV2All)
	router.Get("/uniswap-v2/chain/:chainId/address/:address", c.GetUniswapV2ByChainIDAndAddress)
	router.Get("/uniswap-v3", c.GetUniswapV3All)
	router.Get("/uniswap-v3/chain/:chainId/address/:address", c.GetUniswapV3ByChainIDAndAddress)
	router.Get("/:id", c.GetByID)
	router.Put("/:id", c.Update)
	router.Delete("/:id", c.Delete)
}

func (c *DexController) Create(ctx *fiber.Ctx) error {
	var payload dto.DexCreateRequest
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid payload"})
	}

	if errs := payload.Validate(); errs != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "validation error", "errors": errs})
	}

	dex := payload.ToEntity()
	created, err := c.service.Create(dex)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(dto.ToDexResponse(created))
}

func (c *DexController) GetAll(ctx *fiber.Ctx) error {
	page, limit := parseDexPagination(ctx)

	dexes, total, err := c.service.GetPaginated(page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return ctx.JSON(dto.ToDexPaginatedResponse(dexes, page, limit, total))
}

func (c *DexController) GetUniswapV2All(ctx *fiber.Ctx) error {
	page, limit := parseDexPagination(ctx)

	dexes, total, err := c.service.GetPaginatedByType("uniswapv2", page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	items := make([]dto.UniswapV2DexResponse, 0, len(dexes))
	for i := range dexes {
		items = append(items, c.toUniswapV2DexResponse(&dexes[i]))
	}
	return ctx.JSON(dto.ToUniswapV2DexPaginatedResponse(items, page, limit, total))
}

func (c *DexController) GetUniswapV3All(ctx *fiber.Ctx) error {
	page, limit := parseDexPagination(ctx)

	dexes, total, err := c.service.GetPaginatedByType("uniswapv3", page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	items := make([]dto.UniswapV3DexResponse, 0, len(dexes))
	for i := range dexes {
		items = append(items, c.toUniswapV3DexResponse(&dexes[i]))
	}
	return ctx.JSON(dto.ToUniswapV3DexPaginatedResponse(items, page, limit, total))
}

func (c *DexController) GetByChainIDAndAddress(ctx *fiber.Ctx) error {
	return c.getByChainIDAndAddress(ctx, "")
}

func (c *DexController) GetUniswapV2ByChainIDAndAddress(ctx *fiber.Ctx) error {
	return c.getByChainIDAndAddress(ctx, "uniswapv2")
}

func (c *DexController) GetUniswapV3ByChainIDAndAddress(ctx *fiber.Ctx) error {
	return c.getByChainIDAndAddress(ctx, "uniswapv3")
}

func (c *DexController) getByChainIDAndAddress(ctx *fiber.Ctx, dexType string) error {
	chainID, err := strconv.ParseUint(ctx.Params("chainId"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid chainId"})
	}

	address := strings.ToLower(strings.TrimSpace(ctx.Params("address")))
	if address == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "address is required"})
	}

	var dex *entities.Dex
	if dexType == "" {
		dex, err = c.service.GetByChainIDAndAddress(chainID, address)
	} else {
		dex, err = c.service.GetByTypeChainIDAndAddress(dexType, chainID, address)
	}
	if err != nil {
		if service.IsNotFound(err) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "dex not found"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	dexType = dex.DexType

	if dexType == "uniswapv2" {
		return ctx.JSON(c.toUniswapV2DexResponse(dex))
	}
	if dexType == "uniswapv3" {
		return ctx.JSON(c.toUniswapV3DexResponse(dex))
	}

	return ctx.JSON(c.toDexResponse(dex))
}

func (c *DexController) GetByID(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid id"})
	}

	dex, err := c.service.GetByID(id)
	if err != nil {
		if service.IsNotFound(err) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "dex not found"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return ctx.JSON(c.toDexResponse(dex))
}

func (c *DexController) toDexResponse(dex *entities.Dex) dto.DexResponse {
	response := dto.ToDexResponse(dex)
	if c.v2Repo == nil || c.v3Repo == nil {
		return response
	}

	token0Address, token1Address, err := c.dexTokenAddresses(dex.ID)
	if err != nil {
		return response
	}
	response.Token0Address = token0Address
	response.Token1Address = token1Address
	return response
}

func (c *DexController) toUniswapV2DexResponse(dex *entities.Dex) dto.UniswapV2DexResponse {
	if c.v2Repo == nil {
		return dto.ToUniswapV2DexResponse(dex, nil)
	}
	details, err := c.v2Repo.GetByDexID(dex.ID)
	if err != nil {
		return dto.ToUniswapV2DexResponse(dex, nil)
	}
	return dto.ToUniswapV2DexResponse(dex, details)
}

func (c *DexController) toUniswapV3DexResponse(dex *entities.Dex) dto.UniswapV3DexResponse {
	if c.v3Repo == nil {
		return dto.ToUniswapV3DexResponse(dex, nil)
	}
	details, err := c.v3Repo.GetByDexID(dex.ID)
	if err != nil {
		return dto.ToUniswapV3DexResponse(dex, nil)
	}
	return dto.ToUniswapV3DexResponse(dex, details)
}

func (c *DexController) dexTokenAddresses(dexID uuid.UUID) (string, string, error) {
	v2, err := c.v2Repo.GetByDexID(dexID)
	if err == nil {
		return v2.Token0Address, v2.Token1Address, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", "", err
	}

	v3, err := c.v3Repo.GetByDexID(dexID)
	if err != nil {
		return "", "", err
	}
	return v3.Token0Address, v3.Token1Address, nil
}

func parseDexPagination(ctx *fiber.Ctx) (int, int) {
	page := ctx.QueryInt("page", 1)
	limit := ctx.QueryInt("limit", 10)

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return page, limit
}

func (c *DexController) Update(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid id"})
	}

	var payload dto.DexUpdateRequest
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid payload"})
	}

	if errs := payload.Validate(); errs != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "validation error", "errors": errs})
	}

	dex, err := c.service.Update(id, payload.ToEntity())
	if err != nil {
		if service.IsNotFound(err) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "dex not found"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return ctx.JSON(dto.ToDexResponse(dex))
}

func (c *DexController) Delete(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid id"})
	}

	if err := c.service.Delete(id); err != nil {
		if service.IsNotFound(err) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "dex not found"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
