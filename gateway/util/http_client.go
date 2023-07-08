package util

import (
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func GetHttpClient() *http.Client {
	return &http.Client{
		Timeout:   time.Second * 30,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
}
