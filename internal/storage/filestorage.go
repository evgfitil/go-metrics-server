package storage

import (
	"encoding/json"
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"os"
	"sync"
	"time"
)

type FileStorage struct {
	file          *os.File
	memStorage    *MemStorage
	mu            sync.RWMutex
	storeInterval time.Duration
	saveSignal    chan struct{}
}

func NewFileStorage(filename string, memStorage *MemStorage, storeInteval time.Duration, saveSignal chan struct{}) (*FileStorage, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		logger.Sugar.Fatalf("error open file: %v", err)
		return nil, err
	}

	fs := &FileStorage{
		file:          file,
		memStorage:    memStorage,
		storeInterval: storeInteval,
		saveSignal:    saveSignal,
	}

	go fs.Saving()
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

	var metricsCache map[string]*metrics.Metrics
	if err := json.Unmarshal(data, &metricsCache); err != nil {
		return err
	}

	for _, metric := range metricsCache {
		f.memStorage.Update(metric)
	}
	return nil
}

func (f *FileStorage) SaveMetrics() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.file.Truncate(0); err != nil {
		return err
	}

	if _, err := f.file.Seek(0, 0); err != nil {
		return err
	}

	metricsMap := f.memStorage.GetAllMetrics()
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

func (f *FileStorage) Close() error {
	if f.file != nil {
		if err := f.SaveMetrics(); err != nil {
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

func (f *FileStorage) Saving() {
	var ticker *time.Ticker
	var tickerC <-chan time.Time
	if f.storeInterval > 0 {
		ticker = time.NewTicker(f.storeInterval)
		defer ticker.Stop()
		tickerC = ticker.C
	}

	for {
		select {
		case <-f.saveSignal:
			if err := f.SaveMetrics(); err != nil {
				logger.Sugar.Errorf("error saving metrics: %v", err)
			}
		case <-tickerC:
			if ticker != nil {
				if err := f.SaveMetrics(); err != nil {
					logger.Sugar.Errorf("error saving metrics: %v", err)
				}
			}
		}
	}
}
