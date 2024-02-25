package agentcore

import (
	"context"
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"math/rand"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var (
	memStats runtime.MemStats
)

func collectRuntimeMetrics(m *runtime.MemStats) []metrics.Metrics {
	runtimeMetrics := []metrics.Metrics{
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

	return runtimeMetrics
}

func collectSystemMetrics() []metrics.Metrics {
	vmStat, err := mem.VirtualMemoryWithContext(context.TODO())
	if err != nil {
		logger.Sugar.Errorf("error retrieving mem system metrics: %v", err)
	}
	systemMetrics := []metrics.Metrics{
		metrics.NewGauge("TotalMemory", float64(vmStat.Total)),
		metrics.NewGauge("FreeMemory", float64(vmStat.Free)),
	}

	cpuPercentages, err := cpu.PercentWithContext(context.TODO(), 0, true)
	if err != nil {
		logger.Sugar.Errorf("error retrieving cpu system metrics: %v", err)
	}

	for i, cpuPercent := range cpuPercentages {
		metricName := "CPUutilization" + strconv.Itoa(i+1)
		systemMetrics = append(systemMetrics, metrics.NewGauge(metricName, cpuPercent))
	}

	return systemMetrics
}

func aggregateChannels(channels ...<-chan []metrics.Metrics) <-chan []metrics.Metrics {
	var wg sync.WaitGroup

	outCh := make(chan []metrics.Metrics)

	multiplex := func(channel <-chan []metrics.Metrics) {
		defer wg.Done()
		for i := range channel {
			outCh <- i
		}
	}

	wg.Add(len(channels))
	for _, channel := range channels {
		go multiplex(channel)
	}

	go func() {
		wg.Wait()
		close(outCh)
	}()

	return outCh
}

func StartCollector(metricsChan chan<- []metrics.Metrics, pollInterval time.Duration) {
	pollCollectorChan := make(chan []metrics.Metrics)
	runtimeCollectorChan := make(chan []metrics.Metrics)
	systemCollectorChan := make(chan []metrics.Metrics)

	go func() {
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		for range ticker.C {

			go func() {
				pollCollectorChan <- []metrics.Metrics{metrics.NewCounter("PollCount", 1)}
			}()

			go func() {
				runtime.ReadMemStats(&memStats)
				runtimeCollectorChan <- collectRuntimeMetrics(&memStats)
			}()

			go func() {
				systemCollectorChan <- collectSystemMetrics()
			}()
		}
	}()

	for output := range aggregateChannels(pollCollectorChan, runtimeCollectorChan, systemCollectorChan) {
		metricsChan <- output
	}
}
