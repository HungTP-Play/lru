package shared

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/propagation"
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
func (h *HttpService) Use(middleware interface{}, paths ...string) {
	if len(paths) == 0 {
		h.App.Use(middleware)
	} else {
		for _, path := range paths {
			h.App.Use(path, middleware)
		}
	}
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

func ParentContextMiddleware(c *fiber.Ctx) error {
	// Extract the parent span context from the HTTP headers
	propagator := propagation.TraceContext{}

	headers := c.GetReqHeaders()
	var httpHeaders http.Header = make(http.Header)
	for k, v := range headers {
		httpHeaders.Set(k, v)
	}
	// Extract the parent context from the HTTP headers
	parentCtx := propagator.Extract(context.Background(), propagation.HeaderCarrier(httpHeaders))

	// Store the parent context in Fiber's context storage
	c.Locals("parentCtx", parentCtx)

	return c.Next()
}

func GetParentContext(c *fiber.Ctx) context.Context {
	return c.Locals("parentCtx").(context.Context)
}
