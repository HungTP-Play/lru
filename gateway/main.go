package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/HungTP-Play/lru/gateway/dto"
	"github.com/HungTP-Play/lru/gateway/util"
	"github.com/HungTP-Play/lru/shared"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var logger *shared.Logger

func init() {

	logger = shared.NewLogger("gateway.log", 3, 1024, "info", "gateway")
	logger.Init()

	logger.Info("Init done!!!")
}

func onGratefulShutDown() {
	logger.Info("Shutting down...")
}

func shortenHandler(c *fiber.Ctx) error {
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
	resp, err := httpClient.R().SetBody(mapUrlRequest).SetSuccessResult(&mapUrlResponse).Post(fmt.Sprintf("%v/map", mapperUrl))
	if err != nil {
		logger.Error("CannotSendToMapper", zap.String("id", requestID), zap.Int("code", 500), zap.Error(err))
		return c.Status(500).JSON(map[string]interface{}{
			"error": "Internal server error",
		})
	}

	if resp.GetStatusCode() >= 500 {
		logger.Error("MapperResultError__ServerError", zap.String("id", requestID), zap.Int("code", resp.GetStatusCode()), zap.Error(err))
		return c.Status(resp.GetStatusCode()).JSON(map[string]interface{}{
			"error": "Internal server error",
		})
	}

	if resp.GetStatusCode() >= 400 {
		logger.Error("MapperResultError__ClientError", zap.String("id", requestID), zap.Int("code", resp.GetStatusCode()), zap.Error(err))
		return c.Status(resp.GetStatusCode()).JSON(map[string]interface{}{
			"error": "Bad request",
		})
	}

	logger.Info("ShortenUrl", zap.String("id", requestID), zap.Int("code", 200), zap.String("url", mapUrlResponse.Url))
	return c.Status(200).JSON(mapUrlResponse)
}

func redirectHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.SendString(fmt.Sprintf("Redirect to %v", id))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3333"
	}

	gatewayService := shared.NewHttpService("gateway", port, false)
	gatewayService.Init()

	gatewayService.Routes("/shorten", shortenHandler, "POST")
	gatewayService.Routes("/redirect/:id", redirectHandler, "GET")

	gatewayService.Start(onGratefulShutDown)
}
