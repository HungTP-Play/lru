package main

import (
	"fmt"
	"os"

	"github.com/HungTP-Play/lru/mapper/model"
	"github.com/HungTP-Play/lru/mapper/repo"
	"github.com/HungTP-Play/lru/shared"
	"github.com/gofiber/fiber/v2"
)

var mapRepo *repo.UrlMappingRepo

func init() {
	mapRepo = repo.NewUrlMappingRepo("")

	// Auto migrate
	mapRepo.DB.Migrate(&model.UrlMapping{})
}

func mapHandler(c *fiber.Ctx) error {
	var mapUrlRequest shared.MapUrlRequest
	err := c.BodyParser(&mapUrlRequest)
	if err != nil {
		return c.Status(400).SendString("Format error")
	}

	shortUrl, err := mapRepo.Map(mapUrlRequest)
	if err != nil {
		return c.Status(500).SendString("Error")
	}

	mapUrlResponse := shared.MapUrlResponse{
		Url:       mapUrlRequest.Url,
		Shortened: shortUrl,
		Id:        mapUrlRequest.Id,
	}

	return c.Status(200).JSON(mapUrlResponse)
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
