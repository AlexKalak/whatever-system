package http

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func ParsePagination(ctx *fiber.Ctx) (int, int) {
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

func ParseOrdering(ctx *fiber.Ctx, defaultOrderBy string, allowedOrderBy ...string) (string, string) {
	defaultOrderBy = strings.ToLower(strings.TrimSpace(defaultOrderBy))
	if defaultOrderBy == "" {
		defaultOrderBy = "time"
	}

	orderBy := strings.ToLower(strings.TrimSpace(ctx.Query("orderBy", defaultOrderBy)))
	if len(allowedOrderBy) > 0 {
		allowed := false
		for _, value := range allowedOrderBy {
			if orderBy == strings.ToLower(strings.TrimSpace(value)) {
				allowed = true
				break
			}
		}
		if !allowed {
			orderBy = defaultOrderBy
		}
	}

	direction := strings.ToLower(strings.TrimSpace(ctx.Query("direction", "desc")))
	if direction != "asc" {
		direction = "desc"
	}

	return orderBy, direction
}

func PaginatedResponse(items any, page, limit int, total int64) fiber.Map {
	return fiber.Map{
		"items":      items,
		"page":       page,
		"limit":      limit,
		"total":      total,
		"totalPages": TotalPages(total, limit),
	}
}

func OrderedPaginatedResponse(items any, page, limit int, total int64, orderBy, direction string) fiber.Map {
	response := PaginatedResponse(items, page, limit, total)
	response["orderBy"] = orderBy
	response["direction"] = direction
	return response
}

func TotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	return int((total + int64(limit) - 1) / int64(limit))
}

func ParseUintParam(ctx *fiber.Ctx, name string) (uint64, bool) {
	value, err := strconv.ParseUint(ctx.Params(name), 10, 64)
	return value, err == nil
}

func BadRequest(ctx *fiber.Ctx, message string) error {
	return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": message})
}

func NotFound(ctx *fiber.Ctx, message string) error {
	return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": message})
}

func InternalError(ctx *fiber.Ctx, err error) error {
	return InternalMessage(ctx, err.Error())
}

func InternalMessage(ctx *fiber.Ctx, message string) error {
	return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": message})
}
