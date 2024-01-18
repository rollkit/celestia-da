package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

// InstrumentReg stores the already registered instruments
//
//nolint:structcheck // generics
type InstrumentReg[T any, O any] struct {
	instruments   map[string]T
	mu            sync.Mutex
	newInstrument func(name string, options ...O) (T, error)
}

// GetInstrument registers a new instrument, otherwise returns the already created.
func (r *InstrumentReg[T, O]) GetInstrument(name string, options ...O) (T, error) {
	var err error
	r.mu.Lock()
	defer r.mu.Unlock()
	instrument, has := r.instruments[name]
	if !has {
		instrument, err = r.newInstrument(name, options...)
		if err != nil {
			return instrument, fmt.Errorf("unable to register metric %T %s: %w", r, name, err)
		}
		r.instruments[name] = instrument
	}

	return instrument, nil
}

var (
	// meter is the default meter
	meter metric.Meter //nolint:gochecknoglobals // private
	// meterOnce is used to init meter
	meterOnce sync.Once //nolint:gochecknoglobals // private
	// regInt64Counter stores Int64Counters
	regInt64Counter *InstrumentReg[metric.Int64Counter, metric.Int64CounterOption] //nolint:gochecknoglobals // private
	// regFloat64Counter stores Float64Counters
	regFloat64Counter *InstrumentReg[metric.Float64Counter, metric.Float64CounterOption] //nolint:gochecknoglobals // private
)

// GetMeter returns the default meter.
// Inits meter and InstrumentRegs (if needed)
func GetMeter(m metric.MeterProvider) metric.Meter {
	meterOnce.Do(func() {
		meter = m.Meter("github.com/pgillich/opentracing-example/internal/middleware", metric.WithInstrumentationVersion("0.1"))

		regInt64Counter = &InstrumentReg[metric.Int64Counter, metric.Int64CounterOption]{
			instruments:   map[string]metric.Int64Counter{},
			newInstrument: meter.Int64Counter,
		}
		regFloat64Counter = &InstrumentReg[metric.Float64Counter, metric.Float64CounterOption]{
			instruments:   map[string]metric.Float64Counter{},
			newInstrument: meter.Float64Counter,
		}
	})

	return meter
}

func Int64CounterGetInstrument(name string, options ...metric.Int64CounterOption) (metric.Int64Counter, error) {
	return regInt64Counter.GetInstrument(name, options...)
}

func Float64CounterGetInstrument(name string, options ...metric.Float64CounterOption) (metric.Float64Counter, error) {
	return regFloat64Counter.GetInstrument(name, options...)
}

// MetricTransport implements the http.RoundTripper interface and wraps
// outbound HTTP(S) requests with metrics.
type MetricTransport struct {
	rt http.RoundTripper

	meter       metric.Meter
	name        string
	description string
	baseAttrs   []attribute.KeyValue
}

// NewMetricTransport wraps the provided http.RoundTripper with one that
// meters metrics.
//
// If the provided http.RoundTripper is nil, http.DefaultTransport will be used
// as the base http.RoundTripper.
func NewMetricTransport(base http.RoundTripper, meter metric.Meter, name string,
	description string, attributes map[string]string) *MetricTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	baseAttrs := make([]attribute.KeyValue, 0, len(attributes))
	for aKey, aVal := range attributes {
		baseAttrs = append(baseAttrs, attribute.Key(aKey).String(aVal))
	}

	return &MetricTransport{
		rt:          base,
		meter:       meter,
		name:        name,
		description: description,
		baseAttrs:   baseAttrs,
	}
}

// RoundTrip meters outgoing request-response pair.
func (t *MetricTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx := r.Context()

	attempted, err := Int64CounterGetInstrument(t.name, metric.WithDescription(t.description))
	if err != nil {
		return nil, err
	}
	durationSum, err := Float64CounterGetInstrument(t.name+"_duration", metric.WithDescription(t.description+", duration sum"), metric.WithUnit("s"))
	if err != nil {
		return nil, err
	}
	beginTS := time.Now()
	var res *http.Response

	r = r.WithContext(ctx)
	res, err = t.rt.RoundTrip(r)

	elapsedSec := time.Since(beginTS).Seconds()
	attrs := make([]attribute.KeyValue, len(t.baseAttrs), len(t.baseAttrs)+6)
	copy(attrs, t.baseAttrs)
	opt := metric.WithAttributes(attrs...)
	attempted.Add(ctx, 1, opt)
	durationSum.Add(ctx, elapsedSec, opt)

	return res, err //nolint:wrapcheck // should not be changed
}

func setupMetrics(ctx context.Context, serviceName string) (*sdkmetric.MeterProvider, error) {
	exporter, err := otlpmetrichttp.New(
		ctx,
		otlpmetrichttp.WithEndpoint("localhost:4318"),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// labels/tags/resources that are common to all metrics.
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
		attribute.String("some-attribute", "some-value"),
	)

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(resource),
		sdkmetric.WithReader(
			// collects and exports metric data every 30 seconds.
			sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(30*time.Second)),
		),
	)
	GetMeter(mp)

	otel.SetMeterProvider(mp)

	return mp, nil
}
