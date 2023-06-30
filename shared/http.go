package shared

import (
	"context"
	"fmt"

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

func (h *HttpService) Start() error {
	port := fmt.Sprintf(":%s", h.Port)
	return h.App.Listen(port)
}
