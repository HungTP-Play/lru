package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/HungTP-Play/lru/gateway/dto"
	"github.com/HungTP-Play/lru/shared"
	"github.com/gofiber/fiber/v2"
)

func shortenHandler(c *fiber.Ctx) error {
	body := c.Body()
	var shortenDto dto.ShortenRequestDto
	err := json.Unmarshal(body, &shortenDto)
	if err != nil {
		return c.SendString("Error")
	}
	return c.JSON(shortenDto)
}

func redirectHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.SendString(fmt.Sprintf("Redirect to %v", id))
}

func main() {
	fmt.Printf("This is a main %v", "gateway")

	port := os.Getenv("PORT")
	if port == "" {
		port = "3333"
	}

	gatewayService := shared.NewHttpService("gateway", port, false)
	gatewayService.Init()

	gatewayService.Routes("/shorten", shortenHandler, "POST")
	gatewayService.Routes("/redirect/:id", redirectHandler, "GET")

	gatewayService.Start()
}
