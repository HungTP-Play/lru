package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/HungTP-Play/lru/shared"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var logger *shared.Logger
var rabbitmq *shared.RabbitMQ
var metrics *shared.Metrics
var requestPerSecond *prometheus.CounterVec

func init() {
	// Init logger
	logger = shared.NewLogger("analytic.log", 3, 1024, "info", "analytic")
	logger.Init()

	// Init rabbitmq
	rabbitmq = shared.NewRabbitMQ("")
	rabbitmq.Connect(10 * time.Second)

	// Init metrics
	metrics = shared.NewMetrics()
	requestPerSecond = metrics.RegisterCounter("request_per_second", "Request per second", []string{"method", "path"})

	logger.Info("Init done!!!")
}

func handleAnalytic(msg []byte) error {
	metrics.IncCounter(requestPerSecond, "QUEUE", "analytic")
	var analytic shared.AnalyticMessage
	innerLogger := shared.NewLogger("analytic.log", 3, 1024, "info", "analytic")
	innerLogger.Init()

	err := json.Unmarshal(msg, &analytic)
	if err != nil {
		innerLogger.Error("Cannot unmarshal analytic message: %s", zap.Error(err))
		return err
	}

	innerLogger.Info("Analytic message: %s", zap.String("id", analytic.Id), zap.String("url", analytic.Url), zap.String("shorten", analytic.Shorten), zap.String("type", analytic.Type))

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
}

func main() {

	analyticService := shared.NewHttpService("analytic", "4444", false)
	analyticService.Init()

	analyticService.Routes("/metrics", metricsHandler, "GET")

	analyticQueue := os.Getenv("ANALYTIC_QUEUE")
	go func() {
		rabbitmq.Consume(analyticQueue, handleAnalytic, 9)
	}()

	analyticService.Start(onGratefulShutDown)
}
