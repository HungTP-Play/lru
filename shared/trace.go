package shared

import (
	"context"
	"net/http"
	"os"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
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
func NewExporter(collectorURL string, ctx context.Context) (sdk.SpanExporter, error) {
	exporter, err := otlptrace.New(ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(collectorURL),
			otlptracegrpc.WithInsecure(),
		),
	)

	return exporter, err
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
func NewTracerProvider(serviceName string, collectorURL string, ctx context.Context) *sdk.TracerProvider {
	// Create the console exporter
	exporter, err := NewExporter(collectorURL, ctx)
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

	context := context.Background()
	provider := NewTracerProvider(serviceName, collectorURL, context)
	return &Tracer{
		ServiceName:  serviceName,
		CollectorURL: collectorURL,
		Ctx:          context,
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

func InjectPropagationHeader(ctx context.Context, req *http.Request) {
	propagator := propagation.TraceContext{}
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))
}
