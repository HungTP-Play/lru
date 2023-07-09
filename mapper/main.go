package main

import (
	"os"
	"strconv"
	"time"

	"github.com/HungTP-Play/lru/mapper/model"
	"github.com/HungTP-Play/lru/mapper/repo"
	"github.com/HungTP-Play/lru/shared"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var mapRepo *repo.UrlMappingRepo
var logger *shared.Logger
var rabbitmq *shared.RabbitMQ
var metrics *shared.Metrics
var requestPerSecond *prometheus.CounterVec
var TwoXXStatusCode *prometheus.GaugeVec
var FourXXStatusCode *prometheus.GaugeVec
var FiveXXStatusCode *prometheus.GaugeVec

var tracer *shared.Tracer

func init() {
	mapRepo = repo.NewUrlMappingRepo("")

	// Auto migrate
	mapRepo.DB.Migrate(&model.UrlMapping{})

	logger = shared.NewLogger("mapper.log", 3, 1024, "info", "mapper")
	logger.Init()

	// Init rabbitmq
	rabbitmq = shared.NewRabbitMQ("")
	rabbitmq.Connect(10 * time.Second)

	// Init metrics
	metrics = shared.NewMetrics()
	requestPerSecond = metrics.RegisterCounter("request_per_second", "Request per second", []string{"method", "path"})
	TwoXXStatusCode = metrics.RegisterGauge("status_code_2xx", "2xx status code", []string{"method", "path", "code"})
	FourXXStatusCode = metrics.RegisterGauge("status_code_4xx", "4xx status code", []string{"method", "path", "code"})
	FiveXXStatusCode = metrics.RegisterGauge("status_code_5xx", "5xx status code", []string{"method", "path", "code"})

	// Init tracer
	tracer = shared.NewTracer("mapper", "")
	tracer.Init()

	logger.Info("Init done!!!")
}

func onGratefulShutDown() {
	logger.Info("Shutting down...")
	mapRepo.DB.Close()
	rabbitmq.Close()
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

func mapHandler(c *fiber.Ctx) error {
	var mapUrlRequest shared.MapUrlRequest
	ctx := shared.GetParentContext(c)
	_, mapSpan := tracer.StartSpan("Map", ctx)
	defer mapSpan.End()

	body := c.Body()
	err := c.BodyParser(&mapUrlRequest)
	logger.Info("Map request", zap.String("id", mapUrlRequest.Id), zap.String("body", string(body)), zap.String("method", c.Method()), zap.String("path", c.Path()), zap.String("url", mapUrlRequest.Url))
	if err != nil {
		logger.Error("Cannot parse body", zap.String("id", mapUrlRequest.Id), zap.Int("code", 400), zap.Error(err))
		return c.Status(400).JSON(map[string]interface{}{
			"error": "Cannot parse body",
		})
	}

	ctx, mapUrlSpan := tracer.StartSpan("StoreDB", ctx)
	shortUrl, err := mapRepo.Map(mapUrlRequest)
	if err != nil {
		logger.Error("Cannot map url", zap.String("id", mapUrlRequest.Id), zap.Int("code", 500), zap.Error(err))
		mapSpan.End()
		return c.Status(500).JSON(map[string]interface{}{
			"error": "Internal server error",
		})
	}
	mapUrlSpan.End()

	mapUrlResponse := shared.MapUrlResponse{
		Url:       mapUrlRequest.Url,
		Shortened: shortUrl,
		Id:        mapUrlRequest.Id,
	}

	_, publishRedirectSpan := tracer.StartSpan("PublishRedirect", ctx)
	go func() {
		redirectQueue := os.Getenv("REDIRECT_QUEUE")
		redirectMessage := &shared.RedirectMessage{
			Id:      mapUrlRequest.Id,
			Url:     mapUrlRequest.Url,
			Shorten: shortUrl,
		}
		err = rabbitmq.Publish(redirectQueue, redirectMessage)
		if err != nil {
			publishRedirectSpan.End()
			logger.Error("Cannot publish redirect", zap.String("id", redirectMessage.Id), zap.Int("code", 500), zap.Error(err))
		}
		publishRedirectSpan.End()
		logger.Info("Publish redirect", zap.String("id", redirectMessage.Id), zap.String("shortUrl", shortUrl))
	}()

	// Publish to Analytic
	_, publishAnalyticSpan := tracer.StartSpan("PublishAnalytic", ctx)
	go func() {
		analyticQueue := os.Getenv("ANALYTIC_QUEUE")
		analyticMessage := &shared.AnalyticMessage{
			Id:        mapUrlRequest.Id,
			Url:       mapUrlRequest.Url,
			Shorten:   shortUrl,
			Type:      "map",
			Timestamp: time.Now().Unix(),
		}
		err = rabbitmq.Publish(analyticQueue, analyticMessage)
		if err != nil {
			publishAnalyticSpan.End()
			logger.Error("Cannot publish analytic", zap.String("id", mapUrlRequest.Id), zap.Int("code", 500), zap.Error(err))
		}
		publishAnalyticSpan.End()
		logger.Info("Publish analytic", zap.String("id", mapUrlRequest.Id), zap.String("shortUrl", shortUrl))
	}()

	logger.Info("Map response", zap.String("id", mapUrlRequest.Id), zap.Int("code", 200), zap.String("shortUrl", shortUrl))
	return c.Status(200).JSON(mapUrlResponse)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "1111"
	}

	mapperService := shared.NewHttpService("mapper", port, false)
	mapperService.Init()

	mapperService.Use(shared.ParentContextMiddleware)
	mapperService.Use(RequestPerSecondMiddleware)
	mapperService.Use(ResponseStatusCodeMiddleware)

	mapperService.Routes("/map", mapHandler, "POST")
	mapperService.Routes("/metrics", metricsHandler, "GET")

	mapperService.Start(onGratefulShutDown)
}
