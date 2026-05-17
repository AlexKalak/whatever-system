package controller

import (
	"github.com/alexkalak/whatever-system/src/modules/assets/dto"
	"github.com/alexkalak/whatever-system/src/modules/assets/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AssetController struct {
	service service.AssetService
}

func NewAssetController(service service.AssetService) *AssetController {
	return &AssetController{service: service}
}

func (c *AssetController) RegisterRoutes(router fiber.Router) {
	router.Post("/", c.Create)
	router.Get("/", c.GetAll)
	router.Get("/:id", c.GetByID)
	router.Put("/:id", c.Update)
	router.Delete("/:id", c.Delete)
}

func (c *AssetController) Create(ctx *fiber.Ctx) error {
	var payload dto.AssetCreateRequest
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid payload"})
	}

	if errs := payload.Validate(); errs != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "validation error", "errors": errs})
	}

	asset := payload.ToEntity()
	if err := c.service.Create(asset); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(dto.ToAssetResponse(asset))
}

func (c *AssetController) GetAll(ctx *fiber.Ctx) error {
	assets, err := c.service.GetAll()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return ctx.JSON(dto.ToAssetResponses(assets))
}

func (c *AssetController) GetByID(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid id"})
	}

	asset, err := c.service.GetByID(id)
	if err != nil {
		if service.IsNotFound(err) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "asset not found"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return ctx.JSON(dto.ToAssetResponse(asset))
}

func (c *AssetController) Update(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid id"})
	}

	var payload dto.AssetUpdateRequest
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid payload"})
	}

	if errs := payload.Validate(); errs != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "validation error", "errors": errs})
	}

	asset, err := c.service.Update(id, payload.ToEntity())
	if err != nil {
		if service.IsNotFound(err) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "asset not found"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return ctx.JSON(dto.ToAssetResponse(asset))
}

func (c *AssetController) Delete(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid id"})
	}

	if err := c.service.Delete(id); err != nil {
		if service.IsNotFound(err) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "asset not found"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
