package storage

import (
	"context"
	"encoding/json"
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"os"
	"time"
)

type FileStorage struct {
	MemStorage
	file          *os.File
	storeInterval int
}

func NewFileStorage(filename string, storeInterval int) (*FileStorage, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		logger.Sugar.Fatalf("error open file: %v", err)
		return nil, err
	}

	fs := &FileStorage{
		MemStorage: MemStorage{
			metrics: make(map[string]*metrics.Metrics),
		},
		file:          file,
		storeInterval: storeInterval,
	}
	return fs, nil
}

func (f *FileStorage) LoadMetrics() error {
	data, err := os.ReadFile(f.file.Name())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if len(data) == 0 {
		logger.Sugar.Infoln("empty file, continue without loading saved data")
		return nil
	}

	var loadedMetrics map[string]*metrics.Metrics
	if err = json.Unmarshal(data, &loadedMetrics); err != nil {
		logger.Sugar.Errorf("error reading metrics from file: %v", err)
		return err
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	for id, metric := range loadedMetrics {
		f.metrics[id] = metric
	}
	return nil
}

func (f *FileStorage) Update(ctx context.Context, metric *metrics.Metrics) {
	f.mu.Lock()
	defer f.mu.Unlock()
	switch metric.MType {
	case "counter":
		if oldMetric, ok := f.metrics[metric.ID]; ok {
			if oldDelta := oldMetric.Delta; oldDelta != nil {
				newDelta := *metric.Delta + *oldDelta
				newMetric := &metrics.Metrics{
					ID:    metric.ID,
					MType: metric.MType,
					Delta: &newDelta,
				}
				f.metrics[metric.ID] = newMetric
			}
		} else {
			f.metrics[metric.ID] = metric
		}
	case "gauge":
		f.metrics[metric.ID] = metric
	}

	if f.storeInterval == 0 {
		f.mu.Unlock()
		err := f.SaveMetrics(ctx)
		if err != nil {
			logger.Sugar.Error("error write data")
		}
		f.mu.Lock()
	}
}

func (f *FileStorage) SaveMetrics(ctx context.Context) error {
	if err := f.file.Truncate(0); err != nil {
		return err
	}

	if _, err := f.file.Seek(0, 0); err != nil {
		return err
	}

	metricsMap := f.GetAllMetrics(ctx)
	data, err := json.Marshal(metricsMap)
	if err != nil {
		logger.Sugar.Errorf("error marshaling: %v", err)
		return err
	}

	if _, err := f.file.Write(data); err != nil {
		logger.Sugar.Errorf("error writing data to a file: %v", err)
		return err
	}
	if err := f.file.Sync(); err != nil {
		logger.Sugar.Errorf("error syncing file: %v", err)
		return err
	}

	return nil
}

func (f *FileStorage) AsyncSave() {
	go func() {
		interval := time.Duration(f.storeInterval) * time.Second
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := f.SaveMetrics(context.TODO()); err != nil {
				logger.Sugar.Errorf("error during async save: %v", err)
			}
		}
	}()
}

func (f *FileStorage) Close() error {
	if f.file != nil {
		if err := f.SaveMetrics(context.TODO()); err != nil {
			logger.Sugar.Errorf("error when writing when closing: %v", err)
			f.file.Close()
			return err
		}
		err := f.file.Close()
		f.file = nil
		return err
	}

	return nil
}

func (f *FileStorage) Ping(_ context.Context) error {
	return nil
}

func (f *FileStorage) UpdateMetrics(ctx context.Context, metrics []*metrics.Metrics) error {
	for _, metric := range metrics {
		f.Update(ctx, metric)
	}
	return nil
}
