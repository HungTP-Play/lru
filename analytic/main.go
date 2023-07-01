package main

import (
	"os"
	"time"

	"github.com/HungTP-Play/lru/shared"
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
	logger.Info(string(msg))
}

func main() {
	// Start consumer
	forever := make(chan bool)
	analyticQueue := os.Getenv("ANALYTIC_QUEUE")
	rabbitmq.Consume(analyticQueue, handleAnalytic, 9)
	<-forever
}
