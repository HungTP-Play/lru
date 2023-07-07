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
var cacheClient *shared.CacheClient

var (
	defaultKeyCacheTime = 15 * time.Minute
)

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

	// Init cache
	cacheClient = shared.NewCacheClient(shared.RedisDefaultConfig())
	cacheClient.Connect()

	logger.Info("Init done!!!")
}

func onGratefulShutDown() {
	fmt.Println("Shutting down...")
	redirectRepo.Close()
	rabbitmq.Close()
	cacheClient.Close()
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

	// Check cache first => if not found => get from db
	// This called the cache-aside pattern
	var originalUrl string
	var redirectResponse shared.RedirectResponse
	originalUrl, err = cacheClient.Get(redirectRequest.Url)
	if err == nil {
		logger.Info("Cache hit", zap.String("key", redirectRequest.Url), zap.String("value", originalUrl))
		redirectResponse = shared.RedirectResponse{
			Url:         redirectRequest.Url,
			Id:          redirectRequest.Id,
			OriginalUrl: originalUrl,
		}
	} else {
		logger.Info("Cache miss", zap.String("key", redirectRequest.Url))
		logger.Error("Cannot get cache", zap.String("id", redirectRequest.Id), zap.String("key", redirectRequest.Url), zap.Error(err))

		originalUrl, err = redirectRepo.GetRedirect(redirectRequest.Url)
		if err != nil {
			logger.Error("Cannot get redirect", zap.String("id", redirectRequest.Id), zap.Int("code", 500), zap.Error(err))
			return c.Status(500).JSON(map[string]interface{}{
				"error": "Internal server error",
			})
		}

		redirectResponse = shared.RedirectResponse{
			Url:         redirectRequest.Url,
			Id:          redirectRequest.Id,
			OriginalUrl: originalUrl,
		}
	}

	// Send analytic message
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

	// Add to cache, use shorten as key, original url as value
	// This called the write-through cache pattern
	err = cacheClient.Set(redirectMessage.Shorten, redirectMessage.Url, defaultKeyCacheTime)
	if err != nil {
		innerLogger.Error("Cannot set cache", zap.String("id", redirectMessage.Id), zap.String("key", redirectMessage.Shorten), zap.String("value", redirectMessage.Url), zap.Error(err))
	}
	innerLogger.Info("Set cache", zap.String("key", redirectMessage.Shorten), zap.String("value", redirectMessage.Url))
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
