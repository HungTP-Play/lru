package main

import (
	"fmt"
	"os"

	"github.com/HungTP-Play/lru/shared"
	"github.com/gofiber/fiber/v2"
)

func mapHandler(c *fiber.Ctx) error {
	var mapUrlRequest shared.MapUrlRequest
	err := c.BodyParser(&mapUrlRequest)
	if err != nil {
		return c.Status(400).SendString("Format error")
	}

	return c.Status(200).JSON(mapUrlRequest)
}

func main() {
	fmt.Printf("This is a main %v", "gateway")

	port := os.Getenv("PORT")
	if port == "" {
		port = "1111"
	}

	gatewayService := shared.NewHttpService("mapper", port, false)
	gatewayService.Init()

	gatewayService.Routes("/map", mapHandler, "POST")

	gatewayService.Start()
}
