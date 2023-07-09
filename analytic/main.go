package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/HungTP-Play/lru/analytic/model"
	"github.com/HungTP-Play/lru/analytic/repo"
	"github.com/HungTP-Play/lru/shared"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var logger *shared.Logger
var rabbitmq *shared.RabbitMQ
var metrics *shared.Metrics
var requestPerSecond *prometheus.CounterVec
var tracer *shared.Tracer
var analyticRepo *repo.AnalyticRepo

func init() {
	// Init logger
	logger = shared.NewLogger("analytic.log", 3, 1024, "info", "analytic")
	logger.Init()

	// Init analytic repo
	analyticRepo = repo.NewAnalyticRepo("")
	analyticRepo.DB.Migrate(&model.AnalyticRecord{})

	// Init rabbitmq
	rabbitmq = shared.NewRabbitMQ("")
	rabbitmq.Connect(10 * time.Second)

	// Init metrics
	metrics = shared.NewMetrics()
	requestPerSecond = metrics.RegisterCounter("request_per_second", "Request per second", []string{"method", "path"})

	// Init tracer
	tracer = shared.NewTracer("analytic", "")
	tracer.Init()
	logger.Info("Init done!!!")
}

func handleAnalytic(msg []byte, headers amqp091.Table) error {
	ctx := shared.ExtractAmqpTraceHeader(headers)
	ctx, span := tracer.StartSpan("handleAnalytic", ctx, trace.WithSpanKind(trace.SpanKindConsumer))
	defer span.End()
	metrics.IncCounter(requestPerSecond, "QUEUE", "analytic")
	var analytic shared.AnalyticMessage
	innerLogger := shared.NewLogger("analytic.log", 3, 1024, "info", "analytic")
	innerLogger.Init()

	innerRepo := repo.NewAnalyticRepo("")
	defer innerRepo.DB.Close()

	err := json.Unmarshal(msg, &analytic)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Cannot unmarshal analytic message")
		innerLogger.Error("Cannot unmarshal analytic message: %s", zap.Error(err))
		return err
	}

	_, updateDBSpan := tracer.StartSpan("updateDB", ctx, trace.WithSpanKind(trace.SpanKindInternal))
	if analytic.Type == "map" {
		// Create new record
		analyticRecord := model.AnalyticRecord{
			ShortUrl:      analytic.Shorten,
			OriginalUrl:   analytic.Url,
			RedirectCount: 0,
		}
		innerRepo.DB.Create(&analyticRecord)
		updateDBSpan.End()
	}

	if analytic.Type == "redirect" {
		// Increase redirect count
		innerRepo.IncAccessCount(analytic.Shorten)
		updateDBSpan.End()
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
	analyticRepo.Close()
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
