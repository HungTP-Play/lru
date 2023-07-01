package main

import (
	"fmt"
	"os"
	"time"

	"github.com/HungTP-Play/lru/shared"
	"github.com/gofiber/fiber/v2"
)

var logger *shared.Logger
var rabbitmq *shared.RabbitMQ

func init() {

	logger = shared.NewLogger("redirect.log", 3, 1024, "info", "redirect")
	logger.Init()

	// Init rabbitmq
	rabbitmq = shared.NewRabbitMQ("")
	rabbitmq.Connect(10 * time.Second)

	logger.Info("Init done!!!")
}

func onGratefulShutDown() {
	fmt.Println("Shutting down...")
}

func redirectHandler(c *fiber.Ctx) error {
	return c.SendString("Hello, World 👋!")
}

func redirectQueueHandler(msg []byte) {
	fmt.Println(string(msg))
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
