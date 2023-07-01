package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/HungTP-Play/lru/shared"
	"go.uber.org/zap"
)

var logger *shared.Logger
var rabbitmq *shared.RabbitMQ

func init() {
	// Init logger
	logger := shared.NewLogger("analytic.log", 3, 1024, "info", "redirect")
	logger.Init()

	// Init rabbitmq
	rabbitmq = shared.NewRabbitMQ("")
	rabbitmq.Connect(10 * time.Second)

	logger.Info("Init done!!!")
}

func handleAnalytic(msg []byte) {
	var analytic shared.AnalyticMessage
	err := json.Unmarshal(msg, &analytic)
	if err != nil {
		logger.Error("Cannot unmarshal analytic message: %s", zap.Error(err))
		return
	}

	logger.Info("Analytic message: %s", zap.String("id", analytic.Id), zap.String("url", analytic.Url), zap.String("shorten", analytic.Shorten), zap.String("type", analytic.Type))
}

func main() {
	// Start consumer
	forever := make(chan bool)
	analyticQueue := os.Getenv("ANALYTIC_QUEUE")
	rabbitmq.Consume(analyticQueue, handleAnalytic, 9)
	<-forever
}
