package controller

import (
	dexservice "github.com/alexkalak/whatever-system/src/modules/exchanges/dex/service"
	"github.com/alexkalak/whatever-system/src/modules/trades/dexactions/service"
	sharedhttp "github.com/alexkalak/whatever-system/src/shared/http"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type DexActionController struct {
	dexService dexservice.DexService
	v2Service  service.UniswapV2Service
	v3Service  service.UniswapV3Service
}

func NewDexActionController(
	v2 service.UniswapV2Service,
	v3 service.UniswapV3Service,
	dexSvc ...dexservice.DexService,
) *DexActionController {
	controller := &DexActionController{v2Service: v2, v3Service: v3}
	if len(dexSvc) > 0 {
		controller.dexService = dexSvc[0]
	}
	return controller
}

func (c *DexActionController) RegisterRoutes(router fiber.Router) {
	router.Get("/", c.GetAll)
	router.Get("/tx/:txHash", c.GetByTxHash)
	router.Get("/chain/:chainId/dex/:dexAddress", c.GetByChainIDAndDexAddress)

	router.Get("/uniswap-v2", c.GetUniswapV2All)
	router.Get("/uniswap-v2/tx/:txHash", c.GetUniswapV2ByTxHash)
	router.Get("/uniswap-v2/chain/:chainId/dex/:dexAddress", c.GetUniswapV2ByChainIDAndDexAddress)
	router.Get("/uniswap-v2/:id", c.GetUniswapV2ByID)

	router.Get("/uniswap-v3", c.GetUniswapV3All)
	router.Get("/uniswap-v3/tx/:txHash", c.GetUniswapV3ByTxHash)
	router.Get("/uniswap-v3/chain/:chainId/dex/:dexAddress", c.GetUniswapV3ByChainIDAndDexAddress)
	router.Get("/uniswap-v3/:id", c.GetUniswapV3ByID)
}

func (c *DexActionController) GetAll(ctx *fiber.Ctx) error {
	page, limit := sharedhttp.ParsePagination(ctx)
	orderBy, direction := parseOrdering(ctx)

	v2Actions, v2Total, err := c.v2Service.GetPaginated(page, limit, orderBy, direction)
	if err != nil {
		return sharedhttp.InternalError(ctx, err)
	}

	v3Actions, v3Total, err := c.v3Service.GetPaginated(page, limit, orderBy, direction)
	if err != nil {
		return sharedhttp.InternalError(ctx, err)
	}

	total := v2Total + v3Total
	return ctx.JSON(combinedPaginatedResponse(v2Actions, v3Actions, page, limit, total, orderBy, direction))
}

func (c *DexActionController) GetByTxHash(ctx *fiber.Ctx) error {
	page, limit := sharedhttp.ParsePagination(ctx)
	orderBy, direction := parseOrdering(ctx)
	txHash := ctx.Params("txHash")

	v2Actions, v2Total, err := c.v2Service.GetByTxHash(txHash, page, limit, orderBy, direction)
	if err != nil {
		return sharedhttp.InternalError(ctx, err)
	}

	v3Actions, v3Total, err := c.v3Service.GetByTxHash(txHash, page, limit, orderBy, direction)
	if err != nil {
		return sharedhttp.InternalError(ctx, err)
	}

	total := v2Total + v3Total
	return ctx.JSON(combinedPaginatedResponse(v2Actions, v3Actions, page, limit, total, orderBy, direction))
}

func (c *DexActionController) GetByChainIDAndDexAddress(ctx *fiber.Ctx) error {
	chainID, ok := sharedhttp.ParseUintParam(ctx, "chainId")
	if !ok {
		return sharedhttp.BadRequest(ctx, "invalid chainId")
	}
	if c.dexService == nil {
		return sharedhttp.InternalMessage(ctx, "dex service is not configured")
	}

	dex, err := c.dexService.GetByChainIDAndAddress(chainID, ctx.Params("dexAddress"))
	if err != nil {
		if dexservice.IsNotFound(err) {
			return sharedhttp.NotFound(ctx, "dex not found")
		}
		return sharedhttp.InternalError(ctx, err)
	}

	switch dex.DexType {
	case "uniswapv2":
		return c.GetUniswapV2ByChainIDAndDexAddress(ctx)
	case "uniswapv3":
		return c.GetUniswapV3ByChainIDAndDexAddress(ctx)
	default:
		return sharedhttp.BadRequest(ctx, "unsupported dex type")
	}
}

func (c *DexActionController) GetUniswapV2All(ctx *fiber.Ctx) error {
	page, limit := sharedhttp.ParsePagination(ctx)
	orderBy, direction := parseOrdering(ctx)

	items, total, err := c.v2Service.GetPaginated(page, limit, orderBy, direction)
	if err != nil {
		return sharedhttp.InternalError(ctx, err)
	}

	return ctx.JSON(sharedhttp.OrderedPaginatedResponse(items, page, limit, total, orderBy, direction))
}

func (c *DexActionController) GetUniswapV3All(ctx *fiber.Ctx) error {
	page, limit := sharedhttp.ParsePagination(ctx)
	orderBy, direction := parseOrdering(ctx)

	items, total, err := c.v3Service.GetPaginated(page, limit, orderBy, direction)
	if err != nil {
		return sharedhttp.InternalError(ctx, err)
	}

	return ctx.JSON(sharedhttp.OrderedPaginatedResponse(items, page, limit, total, orderBy, direction))
}

func (c *DexActionController) GetUniswapV2ByTxHash(ctx *fiber.Ctx) error {
	page, limit := sharedhttp.ParsePagination(ctx)
	orderBy, direction := parseOrdering(ctx)

	items, total, err := c.v2Service.GetByTxHash(ctx.Params("txHash"), page, limit, orderBy, direction)
	if err != nil {
		return sharedhttp.InternalError(ctx, err)
	}

	return ctx.JSON(sharedhttp.OrderedPaginatedResponse(items, page, limit, total, orderBy, direction))
}

func (c *DexActionController) GetUniswapV3ByTxHash(ctx *fiber.Ctx) error {
	page, limit := sharedhttp.ParsePagination(ctx)
	orderBy, direction := parseOrdering(ctx)

	items, total, err := c.v3Service.GetByTxHash(ctx.Params("txHash"), page, limit, orderBy, direction)
	if err != nil {
		return sharedhttp.InternalError(ctx, err)
	}

	return ctx.JSON(sharedhttp.OrderedPaginatedResponse(items, page, limit, total, orderBy, direction))
}

func (c *DexActionController) GetUniswapV2ByChainIDAndDexAddress(ctx *fiber.Ctx) error {
	page, limit := sharedhttp.ParsePagination(ctx)
	orderBy, direction := parseOrdering(ctx)
	chainID, ok := sharedhttp.ParseUintParam(ctx, "chainId")
	if !ok {
		return sharedhttp.BadRequest(ctx, "invalid chainId")
	}

	if isCursorPagination(ctx) {
		items, total, nextCursor, err := c.v2Service.GetByChainIDAndDexAddressCursor(
			chainID,
			ctx.Params("dexAddress"),
			ctx.Query("cursor"),
			limit,
			orderBy,
			direction,
		)
		if err != nil {
			return sharedhttp.InternalError(ctx, err)
		}

		return ctx.JSON(cursorResponse(items, limit, total, nextCursor, orderBy, direction))
	}

	items, total, err := c.v2Service.GetByChainIDAndDexAddress(
		chainID,
		ctx.Params("dexAddress"),
		page,
		limit,
		orderBy,
		direction,
	)
	if err != nil {
		return sharedhttp.InternalError(ctx, err)
	}

	return ctx.JSON(sharedhttp.OrderedPaginatedResponse(items, page, limit, total, orderBy, direction))
}

func (c *DexActionController) GetUniswapV3ByChainIDAndDexAddress(ctx *fiber.Ctx) error {
	page, limit := sharedhttp.ParsePagination(ctx)
	orderBy, direction := parseOrdering(ctx)
	chainID, ok := sharedhttp.ParseUintParam(ctx, "chainId")
	if !ok {
		return sharedhttp.BadRequest(ctx, "invalid chainId")
	}

	if isCursorPagination(ctx) {
		items, total, nextCursor, err := c.v3Service.GetByChainIDAndDexAddressCursor(
			chainID,
			ctx.Params("dexAddress"),
			ctx.Query("cursor"),
			limit,
			orderBy,
			direction,
		)
		if err != nil {
			return sharedhttp.InternalError(ctx, err)
		}

		return ctx.JSON(cursorResponse(items, limit, total, nextCursor, orderBy, direction))
	}

	items, total, err := c.v3Service.GetByChainIDAndDexAddress(
		chainID,
		ctx.Params("dexAddress"),
		page,
		limit,
		orderBy,
		direction,
	)
	if err != nil {
		return sharedhttp.InternalError(ctx, err)
	}

	return ctx.JSON(sharedhttp.OrderedPaginatedResponse(items, page, limit, total, orderBy, direction))
}

func (c *DexActionController) GetUniswapV2ByID(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return sharedhttp.BadRequest(ctx, "invalid id")
	}

	item, err := c.v2Service.GetByID(id)
	if err != nil {
		return actionLookupError(ctx, err)
	}

	return ctx.JSON(item)
}

func (c *DexActionController) GetUniswapV3ByID(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return sharedhttp.BadRequest(ctx, "invalid id")
	}

	item, err := c.v3Service.GetByID(id)
	if err != nil {
		return actionLookupError(ctx, err)
	}

	return ctx.JSON(item)
}

func combinedPaginatedResponse(v2Items, v3Items any, page int, limit int, total int64, orderBy string, direction string) fiber.Map {
	return fiber.Map{
		"items": fiber.Map{
			"uniswapV2": v2Items,
			"uniswapV3": v3Items,
		},
		"page":       page,
		"limit":      limit,
		"total":      total,
		"totalPages": sharedhttp.TotalPages(total, limit),
		"orderBy":    orderBy,
		"direction":  direction,
	}
}

func cursorResponse(items any, limit int, total int64, nextCursor string, orderBy string, direction string) fiber.Map {
	return fiber.Map{
		"items":      items,
		"limit":      limit,
		"total":      total,
		"nextCursor": nextCursor,
		"hasMore":    nextCursor != "",
		"orderBy":    orderBy,
		"direction":  direction,
	}
}

func isCursorPagination(ctx *fiber.Ctx) bool {
	return ctx.Query("pagination") == "cursor"
}

func parseOrdering(ctx *fiber.Ctx) (string, string) {
	return sharedhttp.ParseOrdering(ctx, "time", "time", "amount0", "amount1")
}

func actionLookupError(ctx *fiber.Ctx, err error) error {
	if service.IsNotFound(err) {
		return sharedhttp.NotFound(ctx, "dex action not found")
	}

	return sharedhttp.InternalError(ctx, err)
}
