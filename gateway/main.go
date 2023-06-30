package main

import (
	"fmt"
	"os"

	"github.com/HungTP-Play/lru/shared"
	"github.com/gofiber/fiber/v2"
)

func main() {
	fmt.Printf("This is a main %v", "gateway")

	port := os.Getenv("PORT")
	if port == "" {
		port = "3333"
	}

	gatewayService := shared.NewHttpService("gateway", port, false)
	gatewayService.Init()

	gatewayService.Routes("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	}, "GET")

	gatewayService.Start()
}
