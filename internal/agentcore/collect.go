package agentcore

import (
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"math/rand"
	"runtime"
)

type MetricInterface interface {
	GetName() string
	GetType() string
	GetValueAsString() (string, error)
}

func CollectMetrics(m *runtime.MemStats) []MetricInterface {
	collectedMetrics := []MetricInterface{
		metrics.NewGauge("Alloc", float64(m.Alloc)),
		metrics.NewGauge("BuckHashSys", float64(m.BuckHashSys)),
		metrics.NewGauge("Frees", float64(m.Frees)),
		metrics.NewGauge("GCCPUFraction", m.GCCPUFraction),
		metrics.NewGauge("GCSys", float64(m.GCSys)),
		metrics.NewGauge("HeapAlloc", float64(m.HeapAlloc)),
		metrics.NewGauge("HeapIdle", float64(m.HeapIdle)),
		metrics.NewGauge("HeapInuse", float64(m.HeapInuse)),
		metrics.NewGauge("HeapObjects", float64(m.HeapObjects)),
		metrics.NewGauge("HeapReleased", float64(m.HeapReleased)),
		metrics.NewGauge("HeapSys", float64(m.HeapSys)),
		metrics.NewGauge("LastGC", float64(m.LastGC)),
		metrics.NewGauge("Lookups", float64(m.Lookups)),
		metrics.NewGauge("MCacheInuse", float64(m.MCacheInuse)),
		metrics.NewGauge("MCacheSys", float64(m.MCacheSys)),
		metrics.NewGauge("MSpanInuse", float64(m.MSpanInuse)),
		metrics.NewGauge("MSpanSys", float64(m.MSpanSys)),
		metrics.NewGauge("Mallocs", float64(m.Mallocs)),
		metrics.NewGauge("NextGC", float64(m.NextGC)),
		metrics.NewGauge("NumForcedGC", float64(m.NumForcedGC)),
		metrics.NewGauge("NumGC", float64(m.NumGC)),
		metrics.NewGauge("OtherSys", float64(m.OtherSys)),
		metrics.NewGauge("PauseTotalNs", float64(m.PauseTotalNs)),
		metrics.NewGauge("StackInuse", float64(m.StackInuse)),
		metrics.NewGauge("StackSys", float64(m.StackSys)),
		metrics.NewGauge("Sys", float64(m.Sys)),
		metrics.NewGauge("TotalAlloc", float64(m.TotalAlloc)),
		metrics.NewGauge("RandomValue", rand.Float64()),
	}

	return collectedMetrics
}
