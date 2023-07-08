package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/HungTP-Play/lru/redirect/model"
	"github.com/HungTP-Play/lru/redirect/repo"
	"github.com/HungTP-Play/lru/shared"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var logger *shared.Logger
var rabbitmq *shared.RabbitMQ
var redirectRepo *repo.RedirectUrlRepo
var cacheClient *shared.CacheClient
var metrics *shared.Metrics
var requestPerSecond *prometheus.CounterVec
var TwoXXStatusCode *prometheus.GaugeVec
var FourXXStatusCode *prometheus.GaugeVec
var FiveXXStatusCode *prometheus.GaugeVec
var tracer *shared.Tracer

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

	// Init metrics
	metrics = shared.NewMetrics()
	requestPerSecond = metrics.RegisterCounter("request_per_second", "Request per second", []string{"method", "path"})
	TwoXXStatusCode = metrics.RegisterGauge("status_code_2xx", "2xx status code", []string{"method", "path", "code"})
	FourXXStatusCode = metrics.RegisterGauge("status_code_4xx", "4xx status code", []string{"method", "path", "code"})
	FiveXXStatusCode = metrics.RegisterGauge("status_code_5xx", "5xx status code", []string{"method", "path", "code"})

	// Init tracer
	tracer = shared.NewTracer("redirect", "")
	tracer.Init()

	logger.Info("Init done!!!")
}

func RequestPerSecondMiddleware(c *fiber.Ctx) error {
	metrics.IncCounter(requestPerSecond, c.Method(), c.Path())
	return c.Next()
}

func ResponseStatusCodeMiddleware(c *fiber.Ctx) error {
	c.Next()
	statusCode := c.Response().StatusCode()
	if statusCode >= 200 && statusCode < 300 {
		metrics.IncGauge(TwoXXStatusCode, c.Method(), c.Path(), strconv.Itoa(statusCode))
	}

	if statusCode >= 400 && statusCode < 500 {
		metrics.IncGauge(FourXXStatusCode, c.Method(), c.Path(), strconv.Itoa(statusCode))
	}

	if statusCode >= 500 {
		metrics.IncGauge(FiveXXStatusCode, c.Method(), c.Path(), strconv.Itoa(statusCode))
	}

	return nil
}

func metricsHandler(c *fiber.Ctx) error {
	metrics, err := metrics.GetPrometheusMetrics()
	if err != nil {
		return c.Status(500).SendString("Failed to collect metrics")
	}
	return c.Type("text/plain").SendString(metrics)
}

func onGratefulShutDown() {
	fmt.Println("Shutting down...")
	redirectRepo.Close()
	rabbitmq.Close()
	cacheClient.Close()
}

func redirectHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	redirectCtx, redirectSpan := tracer.StartSpan("RedirectHandler", ctx)
	defer redirectSpan.End()
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
	redirectCtx, cacheSpan := tracer.StartSpan("GetCache", redirectCtx)
	originalUrl, err = cacheClient.Get(redirectRequest.Url)
	cacheSpan.End()

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
		_, dbSpan := tracer.StartSpan("GetRedirect", redirectCtx)

		originalUrl, err = redirectRepo.GetRedirect(redirectRequest.Url)
		if err != nil {
			logger.Error("Cannot get redirect", zap.String("id", redirectRequest.Id), zap.Int("code", 500), zap.Error(err))
			dbSpan.End()
			return c.Status(500).JSON(map[string]interface{}{
				"error": "Internal server error",
			})
		}

		redirectResponse = shared.RedirectResponse{
			Url:         redirectRequest.Url,
			Id:          redirectRequest.Id,
			OriginalUrl: originalUrl,
		}
		dbSpan.End()
	}

	// Send analytic message
	_, analyticSpan := tracer.StartSpan("SendAnalytic", redirectCtx)
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
			analyticSpan.End()
			logger.Error("Cannot publish analytic message", zap.String("id", redirectRequest.Id), zap.String("url", originalUrl), zap.String("shorten", redirectRequest.Url), zap.Error(err))
		}
		analyticSpan.End()
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

	redirectService := shared.NewHttpService("redirect", port, false)
	redirectService.Init()

	otelfiberOpts := []otelfiber.Option{
		otelfiber.WithTracerProvider(tracer.Provider),
	}
	redirectService.Use(otelfiber.Middleware(otelfiberOpts...))
	redirectService.Use(RequestPerSecondMiddleware)
	redirectService.Use(ResponseStatusCodeMiddleware)

	redirectService.Routes("/redirect", redirectHandler, "GET")
	redirectService.Routes("/metrics", metricsHandler, "GET")

	redirectQueue := os.Getenv("REDIRECT_QUEUE")
	go func() {
		rabbitmq.Consume(redirectQueue, redirectQueueHandler, 9)
	}()

	redirectService.Start(onGratefulShutDown)
}
