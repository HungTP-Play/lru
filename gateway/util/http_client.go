package util

import (
	"time"

	"github.com/imroc/req/v3"
)

func GetHttpClient() *req.Client {
	return req.SetTimeout(5*time.Second).SetUserAgent("Golang").SetCommonHeader("Content-Type", "application/json")
}
