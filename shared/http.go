package shared

import (
	"context"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
)

type HttpService struct {
	Name    string `json:"name"`
	Port    string `json:"port"`
	Prefork bool   `json:"prefork"`
	App     *fiber.App
	AppCtx  context.Context
}

func NewHttpService(name string, port string, prefork bool) *HttpService {
	return &HttpService{
		Name:    name,
		Port:    port,
		Prefork: prefork,
		AppCtx:  context.Background(),
	}
}

// Add middleware to the application stack
func (h *HttpService) Use(args ...interface{}) {
	h.App.Use(args...)
}

func (h *HttpService) Init() {
	h.App = fiber.New(fiber.Config{
		Prefork: h.Prefork,
	})
}

func (h *HttpService) UseMiddleware(args ...interface{}) {
	h.App.Use(args...)
}

func (h *HttpService) Routes(path string, handler fiber.Handler, method string) {
	switch method {
	case "GET":
		h.App.Get(path, handler)
	case "POST":
		h.App.Post(path, handler)
	case "PUT":
		h.App.Put(path, handler)
	case "DELETE":
		h.App.Delete(path, handler)
	default:
		h.App.Get(path, handler)
	}
}

func (h *HttpService) CtxAdd(key string, value interface{}) {
	h.AppCtx = context.WithValue(h.AppCtx, key, value)
}

func (h *HttpService) CtxGet(key string) interface{} {
	return h.AppCtx.Value(key)
}

func (h *HttpService) Start(onGratefulShutDown func()) error {
	port := fmt.Sprintf(":%s", h.Port)

	// Do prepare for gratefully shutdown
	shutdownChan := make(chan os.Signal, 1)
	go func() {
		<-shutdownChan
		fmt.Println("Shutting down the server...")
		h.App.Shutdown()
	}()

	err := h.App.Listen(port)
	if err != nil {
		return err
	}

	// Clean up
	close(shutdownChan)

	onGratefulShutDown()
	return nil
}
