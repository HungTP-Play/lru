package main

import (
	"os"
	"time"

	"github.com/HungTP-Play/lru/mapper/model"
	"github.com/HungTP-Play/lru/mapper/repo"
	"github.com/HungTP-Play/lru/shared"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var mapRepo *repo.UrlMappingRepo
var logger *shared.Logger
var rabbitmq *shared.RabbitMQ

func init() {
	mapRepo = repo.NewUrlMappingRepo("")

	// Auto migrate
	mapRepo.DB.Migrate(&model.UrlMapping{})

	logger = shared.NewLogger("mapper.log", 3, 1024, "info", "mapper")
	logger.Init()

	// Init rabbitmq
	rabbitmq = shared.NewRabbitMQ("")
	rabbitmq.Connect(10 * time.Second)

	logger.Info("Init done!!!")
}

func onGratefulShutDown() {
	logger.Info("Shutting down...")
	mapRepo.DB.Close()
}

func mapHandler(c *fiber.Ctx) error {
	var mapUrlRequest shared.MapUrlRequest
	body := c.Body()
	err := c.BodyParser(&mapUrlRequest)
	logger.Info("Map request", zap.String("id", mapUrlRequest.Id), zap.String("body", string(body)), zap.String("method", c.Method()), zap.String("path", c.Path()), zap.String("url", mapUrlRequest.Url))
	if err != nil {
		logger.Error("Cannot parse body", zap.String("id", mapUrlRequest.Id), zap.Int("code", 400), zap.Error(err))
		return c.Status(400).JSON(map[string]interface{}{
			"error": "Cannot parse body",
		})
	}

	shortUrl, err := mapRepo.Map(mapUrlRequest)
	if err != nil {
		logger.Error("Cannot map url", zap.String("id", mapUrlRequest.Id), zap.Int("code", 500), zap.Error(err))
		return c.Status(500).JSON(map[string]interface{}{
			"error": "Internal server error",
		})
	}

	mapUrlResponse := shared.MapUrlResponse{
		Url:       mapUrlRequest.Url,
		Shortened: shortUrl,
		Id:        mapUrlRequest.Id,
	}

	// Publish to rabbitmq
	redirectQueue := os.Getenv("REDIRECT_QUEUE")
	err = rabbitmq.Publish(redirectQueue, mapUrlResponse)
	if err != nil {
		logger.Error("Cannot publish to rabbitmq", zap.String("id", mapUrlRequest.Id), zap.Int("code", 500), zap.Error(err))
		return c.Status(500).JSON(map[string]interface{}{
			"error": "Internal server error",
		})
	}
	logger.Info("Map response", zap.String("id", mapUrlRequest.Id), zap.Int("code", 200), zap.String("shortUrl", shortUrl))
	return c.Status(200).JSON(mapUrlResponse)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "1111"
	}

	gatewayService := shared.NewHttpService("mapper", port, false)
	gatewayService.Init()

	gatewayService.Routes("/map", mapHandler, "POST")

	gatewayService.Start(onGratefulShutDown)
}
