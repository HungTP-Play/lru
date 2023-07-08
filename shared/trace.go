package shared

import (
	"context"
	"io"
	"net/http"
	"os"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

type Tracer struct {
	ServiceName  string
	CollectorURL string
	Ctx          context.Context
	Provider     *sdk.TracerProvider
	tracer       trace.Tracer
}

// NewExporter creates an exporter that just print the span data to stdout.
func NewExporter(w io.Writer) (sdk.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithWriter(w),
		// Use human-readable output.
		stdouttrace.WithPrettyPrint(),
		// Do not print timestamps for the demo.
		stdouttrace.WithoutTimestamps(),
	)
}

// NewResource returns a resource describing this application.
// Resource describes the entity for which a signals (metrics or traces) are collected.
func NewResource(serviceName string) *resource.Resource {
	r, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("v1.0.0"),
			attribute.String("environment", "local"),
		),
	)
	return r
}

// New creates a new tracer provider instance. => where the traces are sent to
func NewTracerProvider(serviceName string, collectorURL string) *sdk.TracerProvider {
	// Create the console exporter
	exporter, err := NewExporter(os.Stdout)
	if err != nil {
		panic(err)
	}

	// Create the trace provider with the exporter
	tp := sdk.NewTracerProvider(
		sdk.WithBatcher(exporter),
		sdk.WithResource(NewResource(serviceName)),
	)

	return tp
}

func GetDefaultCollectorURL() string {
	return os.Getenv("OTEL_ENDPOINT")
}

func NewTracer(serviceName string, collectorURL string) *Tracer {
	if collectorURL == "" {
		collectorURL = GetDefaultCollectorURL()
	}

	provider := NewTracerProvider(serviceName, collectorURL)
	return &Tracer{
		ServiceName:  serviceName,
		CollectorURL: collectorURL,
		Ctx:          context.Background(),
		Provider:     provider,
	}
}

func (t *Tracer) Init() {
	otel.SetTracerProvider(t.Provider)
	t.tracer = otel.Tracer(t.ServiceName)
}

func (t *Tracer) StartSpan(name string, ctx context.Context) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name)
}

func (t *Tracer) EndSpan(span trace.Span) {
	span.End()
}

func GetTraceHttpClient() *http.Client {
	return &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
}
