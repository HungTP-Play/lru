package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/HungTP-Play/lru/gateway/dto"
	"github.com/HungTP-Play/lru/gateway/util"
	"github.com/HungTP-Play/lru/shared"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

var logger *shared.Logger
var metrics *shared.Metrics
var requestPerSecond *prometheus.CounterVec
var TwoXXStatusCode *prometheus.GaugeVec
var FourXXStatusCode *prometheus.GaugeVec
var FiveXXStatusCode *prometheus.GaugeVec
var tracer *shared.Tracer

func init() {

	logger = shared.NewLogger("gateway.log", 3, 1024, "info", "gateway")
	logger.Init()

	// Init metrics
	metrics = shared.NewMetrics()
	requestPerSecond = metrics.RegisterCounter("request_per_second", "Request per second", []string{"method", "path"})
	TwoXXStatusCode = metrics.RegisterGauge("status_code_2xx", "2xx status code", []string{"method", "path", "code"})
	FourXXStatusCode = metrics.RegisterGauge("status_code_4xx", "4xx status code", []string{"method", "path", "code"})
	FiveXXStatusCode = metrics.RegisterGauge("status_code_5xx", "5xx status code", []string{"method", "path", "code"})

	// Init tracer
	tracer = shared.NewTracer("gateway", "")
	tracer.Init()

	logger.Info("Init done!!!")
}

func RequestCounterMiddleware(c *fiber.Ctx) error {
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
		logger.Info("ResponseStatusCodeMiddleware", zap.Int("code", statusCode))
		metrics.IncGauge(FourXXStatusCode, c.Method(), c.Path(), strconv.Itoa(statusCode))
	}
	if statusCode >= 500 {
		metrics.IncGauge(FiveXXStatusCode, c.Method(), c.Path(), strconv.Itoa(statusCode))
	}
	return nil
}

func onGratefulShutDown() {
	logger.Info("Shutting down...")
}

func shortenHandler(c *fiber.Ctx) error {
	shortenCtx, shortenSpan := tracer.StartSpan("ShortenHandler", tracer.Ctx)
	defer shortenSpan.End()

	requestID := util.GenUUID()
	body := c.Body()
	var shortenDto dto.ShortenRequestDto
	logger.Info("RequestShorten", zap.String("id", requestID), zap.String("body", string(body)), zap.String("method", c.Method()), zap.String("path", c.Path()), zap.String("url", shortenDto.Url))
	err := json.Unmarshal(body, &shortenDto)

	if err != nil {
		logger.Error("CannotParseBody", zap.Int("code", 400), zap.Error(err))
		return c.Status(400).JSON(map[string]interface{}{
			"error": "Cannot parse body",
		})
	}

	if shortenDto.Url == "" {
		logger.Error("EmptyString", zap.Int("code", 400), zap.Error(err))
		return c.Status(400).JSON(map[string]interface{}{
			"error": "Cannot parse body",
		})
	}

	mapUrlRequest := shared.MapUrlRequest{
		Id:  requestID,
		Url: shortenDto.Url,
	}

	httpClient := util.GetHttpClient()
	mapperUrl := util.GetMapperUrl()

	var mapUrlResponse shared.MapUrlResponse
	logger.Info("SendToMapper", zap.String("id", requestID), zap.String("url", mapUrlRequest.Url))

	_, mapperCallSpan := tracer.StartSpan("SendToMapper", shortenCtx)
	defer mapperCallSpan.End()
	mapperUrl = fmt.Sprintf("%v/map", mapperUrl)
	reqBody, _ := json.Marshal(mapUrlRequest)
	resp, err := httpClient.Post(mapperUrl, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Error("CannotSendToMapper", zap.String("id", requestID), zap.Int("code", 500), zap.Error(err))
		return c.Status(500).JSON(map[string]interface{}{
			"error": "Internal server error",
		})
	}

	json.NewDecoder(resp.Body).Decode(&mapUrlResponse)

	if resp.StatusCode >= 500 {
		logger.Error("MapperResultError__ServerError", zap.String("id", requestID), zap.Int("code", resp.StatusCode), zap.Error(err))
		return c.Status(resp.StatusCode).JSON(map[string]interface{}{
			"error": "Internal server error",
		})
	}

	if resp.StatusCode >= 400 {
		logger.Error("MapperResultError__ClientError", zap.String("id", requestID), zap.Int("code", resp.StatusCode), zap.Error(err))
		return c.Status(resp.StatusCode).JSON(map[string]interface{}{
			"error": "Bad request",
		})
	}

	logger.Info("ShortenUrl", zap.String("id", requestID), zap.Int("code", 200), zap.String("url", mapUrlResponse.Url))
	return c.Status(200).JSON(mapUrlResponse)
}

func redirectHandler(c *fiber.Ctx) error {
	redirectCtx, redirectSpan := tracer.StartSpan("RedirectHandler", tracer.Ctx)
	defer redirectSpan.End()
	var body map[string]string
	requestId := util.GenUUID()
	err := c.BodyParser(&body)
	if err != nil {
		logger.Error("CannotParseBody", zap.Int("code", 400), zap.Error(err))
		return c.Status(400).JSON(map[string]interface{}{
			"error": "Cannot parse body",
		})
	}

	redirectRequest := shared.RedirectRequest{
		Id:  requestId,
		Url: body["url"],
	}

	httpClient := util.GetHttpClient()
	mapperUrl := util.GetRedirectUrl()

	var redirectResponse shared.RedirectResponse
	logger.Info("SendToRedirect", zap.String("id", requestId), zap.String("url", redirectRequest.Url))

	_, redirectCallSpan := tracer.StartSpan("SendToRedirect", redirectCtx)
	defer redirectCallSpan.End()

	mapperUrl = fmt.Sprintf("%v/redirect", mapperUrl)
	reqBody, _ := json.Marshal(redirectRequest)
	url, _ := url.Parse(mapperUrl)
	resp, err := httpClient.Do(&http.Request{
		Method: "GET",
		URL:    url,
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: io.NopCloser(bytes.NewReader(reqBody)),
	})

	_ = json.NewDecoder(resp.Body).Decode(&redirectResponse)

	if err != nil {
		logger.Error("CannotSendToRedirect", zap.String("id", requestId), zap.Int("code", 500), zap.Error(err))
		return c.Status(500).JSON(map[string]interface{}{
			"error": "Internal server error",
		})
	}

	if resp.StatusCode >= 500 {
		logger.Error("RedirectResultError__ServerError", zap.String("id", requestId), zap.Int("code", resp.StatusCode), zap.Error(err))
		return c.Status(resp.StatusCode).JSON(map[string]interface{}{
			"error": "Internal server error",
		})
	}

	if resp.StatusCode >= 400 {
		logger.Error("RedirectResultError__ClientError", zap.String("id", requestId), zap.Int("code", resp.StatusCode), zap.Error(err))
		return c.Status(resp.StatusCode).JSON(map[string]interface{}{
			"error": "Bad request",
		})
	}

	logger.Info("RedirectUrl", zap.String("id", requestId), zap.Int("code", 200), zap.String("url", redirectResponse.Url))
	return c.Status(200).JSON(redirectResponse)

}
func metricsHandler(c *fiber.Ctx) error {
	metrics, err := metrics.GetPrometheusMetrics()
	if err != nil {
		return c.Status(500).SendString("Failed to collect metrics")
	}
	return c.Type("text/plain").SendString(metrics)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3333"
	}

	gatewayService := shared.NewHttpService("gateway", port, false)
	gatewayService.Init()

	gatewayService.Use(RequestCounterMiddleware)
	gatewayService.Use(ResponseStatusCodeMiddleware)

	gatewayService.Routes("/shorten", shortenHandler, "POST")
	gatewayService.Routes("/redirect", redirectHandler, "GET")
	gatewayService.Routes("/metrics", metricsHandler, "GET")

	gatewayService.Start(onGratefulShutDown)
}
