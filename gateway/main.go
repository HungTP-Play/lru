package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/HungTP-Play/lru/gateway/dto"
	"github.com/HungTP-Play/lru/gateway/util"
	"github.com/HungTP-Play/lru/shared"
	"github.com/gofiber/fiber/v2"
)

func shortenHandler(c *fiber.Ctx) error {
	body := c.Body()
	var shortenDto dto.ShortenRequestDto
	err := json.Unmarshal(body, &shortenDto)
	if err != nil {
		return c.SendString("Error")
	}

	mapUrlRequest := shared.MapUrlRequest{
		Id:  util.GenUUID(),
		Url: shortenDto.Url,
	}

	httpClient := util.GetHttpClient()
	mapperUrl := util.GetMapperUrl()

	var mapUrlResponse shared.MapUrlResponse
	resp, err := httpClient.R().SetBody(mapUrlRequest).SetSuccessResult(&mapUrlResponse).Post(fmt.Sprintf("%v/map", mapperUrl))
	if err != nil {
		return c.SendString("Error")
	}

	if resp.GetStatusCode() != 200 {
		return c.SendString("Error")
	}

	return c.Status(200).JSON(mapUrlResponse)
}

func redirectHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.SendString(fmt.Sprintf("Redirect to %v", id))
}

func main() {
	fmt.Printf("This is a main %v", "gateway")

	port := os.Getenv("PORT")
	if port == "" {
		port = "3333"
	}

	gatewayService := shared.NewHttpService("gateway", port, false)
	gatewayService.Init()

	gatewayService.Routes("/shorten", shortenHandler, "POST")
	gatewayService.Routes("/redirect/:id", redirectHandler, "GET")

	gatewayService.Start()
}
