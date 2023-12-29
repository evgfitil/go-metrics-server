package main

import (
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"runtime"
)

func collectMetrics(m *runtime.MemStats) []metrics.Metric {
	collectedMetrics := []metrics.Metric{
		metrics.Gauge{Name: "Alloc", Value: float64(m.Alloc)},
		metrics.Gauge{Name: "BuckHashSys", Value: float64(m.BuckHashSys)},
		metrics.Gauge{Name: "Frees", Value: float64(m.Frees)},
		metrics.Gauge{Name: "GCCPUFraction", Value: m.GCCPUFraction},
		metrics.Gauge{Name: "GCSys", Value: float64(m.GCSys)},
		metrics.Gauge{Name: "HeapAlloc", Value: float64(m.HeapAlloc)},
		metrics.Gauge{Name: "HeapIdle", Value: float64(m.HeapIdle)},
		metrics.Gauge{Name: "HeapInuse", Value: float64(m.HeapInuse)},
		metrics.Gauge{Name: "HeapObjects", Value: float64(m.HeapObjects)},
		metrics.Gauge{Name: "HeapReleased", Value: float64(m.HeapReleased)},
		metrics.Gauge{Name: "HeapSys", Value: float64(m.HeapSys)},
		metrics.Gauge{Name: "LastGC", Value: float64(m.LastGC)},
		metrics.Gauge{Name: "Lookups", Value: float64(m.Lookups)},
		metrics.Gauge{Name: "MCacheInuse", Value: float64(m.MCacheInuse)},
		metrics.Gauge{Name: "MCacheSys", Value: float64(m.MCacheSys)},
		metrics.Gauge{Name: "MSpanInuse", Value: float64(m.MSpanInuse)},
		metrics.Gauge{Name: "MSpanSys", Value: float64(m.MSpanSys)},
		metrics.Gauge{Name: "Mallocs", Value: float64(m.Mallocs)},
		metrics.Gauge{Name: "NextGC", Value: float64(m.NextGC)},
		metrics.Gauge{Name: "NumForcedGC", Value: float64(m.NumForcedGC)},
		metrics.Gauge{Name: "NumGC", Value: float64(m.NumGC)},
		metrics.Gauge{Name: "OtherSys", Value: float64(m.OtherSys)},
		metrics.Gauge{Name: "PauseTotalNs", Value: float64(m.PauseTotalNs)},
		metrics.Gauge{Name: "StackInuse", Value: float64(m.StackInuse)},
		metrics.Gauge{Name: "Sys", Value: float64(m.Sys)},
		metrics.Gauge{Name: "TotalAlloc", Value: float64(m.TotalAlloc)},
	}

	return collectedMetrics
}

func main() {}
