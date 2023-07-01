package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/HungTP-Play/lru/redirect/model"
	"github.com/HungTP-Play/lru/redirect/repo"
	"github.com/HungTP-Play/lru/shared"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var logger *shared.Logger
var rabbitmq *shared.RabbitMQ
var redirectRepo *repo.RedirectUrlRepo

func init() {

	logger = shared.NewLogger("redirect.log", 3, 1024, "info", "redirect")
	logger.Init()

	// Init Repo
	redirectRepo = repo.NewRedirectUrlRepo("")

	// Auto migrate
	redirectRepo.DB.Migrate(&model.RedirectUrl{})

	// Init rabbitmq
	rabbitmq = shared.NewRabbitMQ("")
	rabbitmq.Connect(10 * time.Second)

	logger.Info("Init done!!!")
}

func onGratefulShutDown() {
	fmt.Println("Shutting down...")
	redirectRepo.Close()
	rabbitmq.Close()
}

func redirectHandler(c *fiber.Ctx) error {
	var redirectRequest shared.RedirectRequest
	err := c.BodyParser(&redirectRequest)
	if err != nil {
		logger.Error("Cannot parse body", zap.Int("code", 400), zap.Error(err))
		return c.Status(400).JSON(map[string]interface{}{
			"error": "Cannot parse body",
		})
	}

	logger.Info("Redirect request", zap.String("id", redirectRequest.Id), zap.String("method", c.Method()), zap.String("path", c.Path()), zap.String("shorten", redirectRequest.Url))

	originalUrl, err := redirectRepo.GetRedirect(redirectRequest.Url)
	if err != nil {
		logger.Error("Cannot get redirect", zap.String("id", redirectRequest.Id), zap.Int("code", 500), zap.Error(err))
		return c.Status(500).JSON(map[string]interface{}{
			"error": "Internal server error",
		})
	}

	redirectResponse := shared.RedirectResponse{
		Url:         redirectRequest.Url,
		Id:          redirectRequest.Id,
		OriginalUrl: originalUrl,
	}

	go func() {
		analyticMessage := shared.AnalyticMessage{
			Id:        redirectRequest.Id,
			Url:       originalUrl,
			Shorten:   redirectRequest.Url,
			Type:      "redirect",
			Timestamp: time.Now().Unix(),
		}

		analyticQueue := os.Getenv("ANALYTIC_QUEUE")
		err := rabbitmq.Publish(analyticQueue, analyticMessage)
		if err != nil {
			logger.Error("Cannot publish analytic message", zap.String("id", redirectRequest.Id), zap.String("url", originalUrl), zap.String("shorten", redirectRequest.Url), zap.Error(err))
		}
	}()

	return c.Status(200).JSON(redirectResponse)
}

func redirectQueueHandler(msg []byte) error {
	innerLogger := shared.NewLogger("redirect.log", 3, 1024, "info", "redirect")
	innerLogger.Init()

	innerRepo := repo.NewRedirectUrlRepo("")

	var redirectMessage shared.RedirectMessage
	err := json.Unmarshal(msg, &redirectMessage)
	if err != nil {
		innerLogger.Error("Cannot unmarshal redirect message: %s", zap.Error(err))
		return err
	}

	innerLogger.Info("Receive add redirect message", zap.String("id", redirectMessage.Id), zap.String("url", redirectMessage.Url), zap.String("shorten", redirectMessage.Shorten))

	err = innerRepo.AddRedirect(redirectMessage)
	if err != nil {
		innerLogger.Error("Cannot add redirect", zap.String("id", redirectMessage.Id), zap.String("url", redirectMessage.Url), zap.String("shorten", redirectMessage.Shorten), zap.Error(err))
		return err
	}

	return nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "1111"
	}

	gatewayService := shared.NewHttpService("redirect", port, false)
	gatewayService.Init()

	gatewayService.Routes("/redirect", redirectHandler, "GET")

	redirectQueue := os.Getenv("REDIRECT_QUEUE")
	go func() {
		rabbitmq.Consume(redirectQueue, redirectQueueHandler, 9)
	}()

	gatewayService.Start(onGratefulShutDown)
}
