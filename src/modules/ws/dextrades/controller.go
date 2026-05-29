package dextrades

import (
	"strconv"
	"strings"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type Controller struct {
	hub *Hub
}

func NewController(hub *Hub) *Controller {
	return &Controller{hub: hub}
}

func (c *Controller) RegisterRoutes(router fiber.Router) {
	router.Use("/dextrades", func(ctx *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(ctx) {
			return fiber.ErrUpgradeRequired
		}

		address := strings.TrimSpace(ctx.Query("address"))
		if address == "" {
			return fiber.NewError(fiber.StatusBadRequest, "address query parameter is required")
		}

		chainID, err := strconv.ParseUint(ctx.Query("chainId"), 10, 64)
		if err != nil || chainID == 0 {
			return fiber.NewError(fiber.StatusBadRequest, "valid chainId query parameter is required")
		}

		return ctx.Next()
	})

	router.Get("/dextrades", websocket.New(func(conn *websocket.Conn) {
		chainID, _ := strconv.ParseUint(conn.Query("chainId"), 10, 64)
		NewClient(conn, c.hub, conn.Query("address"), chainID).Run()
	}))
}
