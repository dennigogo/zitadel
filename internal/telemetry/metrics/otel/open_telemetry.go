package otel

import (
	"context"
	"net/http"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/export/metric/aggregation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"

	caos_errs "github.com/dennigogo/zitadel/internal/errors"
	"github.com/dennigogo/zitadel/internal/telemetry/metrics"
	otel_resource "github.com/dennigogo/zitadel/internal/telemetry/otel"
)

type Metrics struct {
	Exporter          *prometheus.Exporter
	Meter             metric.Meter
	Counters          sync.Map
	UpDownSumObserver sync.Map
	ValueObservers    sync.Map
}

func NewMetrics(meterName string) (metrics.Metrics, error) {
	resource, err := otel_resource.ResourceWithService()
	if err != nil {
		return nil, err
	}
	exporter, err := prometheus.New(
		prometheus.Config{},
		controller.New(
			processor.NewFactory(
				selector.NewWithHistogramDistribution(),
				aggregation.CumulativeTemporalitySelector(),
				processor.WithMemory(true),
			),
			controller.WithResource(resource),
		),
	)
	if err != nil {
		return &Metrics{}, err
	}
	return &Metrics{
		Exporter: exporter,
		Meter:    exporter.MeterProvider().Meter(meterName),
	}, nil
}

func (m *Metrics) GetExporter() http.Handler {
	return m.Exporter
}

func (m *Metrics) GetMetricsProvider() metric.MeterProvider {
	return m.Exporter.MeterProvider()
}

func (m *Metrics) RegisterCounter(name, description string) error {
	if _, exists := m.Counters.Load(name); exists {
		return nil
	}
	counter := metric.Must(m.Meter).NewInt64Counter(name, metric.WithDescription(description))
	m.Counters.Store(name, counter)
	return nil
}

func (m *Metrics) AddCount(ctx context.Context, name string, value int64, labels map[string]attribute.Value) error {
	counter, exists := m.Counters.Load(name)
	if !exists {
		return caos_errs.ThrowNotFound(nil, "METER-4u8fs", "Errors.Metrics.Counter.NotFound")
	}
	counter.(metric.Int64Counter).Add(ctx, value, MapToKeyValue(labels)...)
	return nil
}

func (m *Metrics) RegisterUpDownSumObserver(name, description string, callbackFunc metric.Int64ObserverFunc) error {
	if _, exists := m.UpDownSumObserver.Load(name); exists {
		return nil
	}
	sumObserver := metric.Must(m.Meter).NewInt64UpDownCounterObserver(
		name, callbackFunc, metric.WithDescription(description))

	m.UpDownSumObserver.Store(name, sumObserver)
	return nil
}

func (m *Metrics) RegisterValueObserver(name, description string, callbackFunc metric.Int64ObserverFunc) error {
	if _, exists := m.UpDownSumObserver.Load(name); exists {
		return nil
	}
	sumObserver := metric.Must(m.Meter).NewInt64GaugeObserver(
		name, callbackFunc, metric.WithDescription(description))

	m.UpDownSumObserver.Store(name, sumObserver)
	return nil
}

func MapToKeyValue(labels map[string]attribute.Value) []attribute.KeyValue {
	if labels == nil {
		return nil
	}
	keyValues := make([]attribute.KeyValue, 0, len(labels))
	for key, value := range labels {
		keyValues = append(keyValues, attribute.KeyValue{
			Key:   attribute.Key(key),
			Value: value,
		})
	}
	return keyValues
}
