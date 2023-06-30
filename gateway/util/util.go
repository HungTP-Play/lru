package util

import (
	"fmt"
	"os"

	"github.com/lithammer/shortuuid/v4"
)

func GetMapperUrl() string {
	host := os.Getenv("MAPPER_HOST")
	if host != "" {
		host = "mapper"
	}

	port := os.Getenv("MAPPER_PORT")
	if port == "" {
		port = "1111"
	}

	return fmt.Sprintf("http://%s:%s", host, port)
}

func GetRedirectUrl() string {
	host := os.Getenv("REDIRECT_HOST")
	if host != "" {
		host = "redirect"
	}

	port := os.Getenv("REDIRECT_PORT")
	if port == "" {
		port = "2222"
	}

	return fmt.Sprintf("http://%s:%s", host, port)
}

func GenUUID() string {
	return shortuuid.New()
}
