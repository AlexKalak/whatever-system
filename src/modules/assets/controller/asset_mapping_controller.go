package controller

import (
	"github.com/alexkalak/whatever-system/src/modules/assets/entities"
	"github.com/alexkalak/whatever-system/src/modules/assets/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AssetMappingController struct {
	service service.AssetMappingService
}

func NewAssetMappingController(service service.AssetMappingService) *AssetMappingController {
	return &AssetMappingController{service: service}
}

func (c *AssetMappingController) RegisterRoutes(router fiber.Router) {
	router.Post("/", c.Create)
	router.Get("/", c.GetAll)
	router.Get("/:id", c.GetByID)
	router.Put("/:id", c.Update)
	router.Delete("/:id", c.Delete)
}

func (c *AssetMappingController) Create(ctx *fiber.Ctx) error {
	var payload entities.AssetMapping
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid payload"})
	}

	if err := c.service.Create(&payload); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(payload)
}

func (c *AssetMappingController) GetAll(ctx *fiber.Ctx) error {
	assetMappings, err := c.service.GetAll()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return ctx.JSON(assetMappings)
}

func (c *AssetMappingController) GetByID(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid id"})
	}

	assetMapping, err := c.service.GetByID(id)
	if err != nil {
		if service.IsNotFound(err) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "asset mapping not found"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return ctx.JSON(assetMapping)
}

func (c *AssetMappingController) Update(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid id"})
	}

	var payload entities.AssetMapping
	if err := ctx.BodyParser(&payload); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid payload"})
	}

	assetMapping, err := c.service.Update(id, &payload)
	if err != nil {
		if service.IsNotFound(err) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "asset mapping not found"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return ctx.JSON(assetMapping)
}

func (c *AssetMappingController) Delete(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid id"})
	}

	if err := c.service.Delete(id); err != nil {
		if service.IsNotFound(err) {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "asset mapping not found"})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
