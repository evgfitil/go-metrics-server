package storage

import (
	"encoding/json"
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
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

func NewFileStorage(filename string, memStorage *MemStorage, storeInteval time.Duration) (*FileStorage, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		logger.Sugar.Fatalf("error open file: %v", err)
		return nil, err
	}

	fs := &FileStorage{
		file:          file,
		memStorage:    memStorage,
		storeInterval: storeInteval,
		saveSignal:    make(chan struct{}),
	}

	go fs.Saving()
	return fs, nil
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
	if err := f.SaveMetrics(); err != nil {
		logger.Sugar.Errorf("error when writing when closing: %v", err)
	}
	return f.file.Close()
}

func (f *FileStorage) Saving() {
	var ticker *time.Ticker
	if f.storeInterval > 0 {
		ticker = time.NewTicker(f.storeInterval)
		defer ticker.Stop()
	}

	for {
		select {
		case <-f.saveSignal:
			if err := f.SaveMetrics(); err != nil {
				logger.Sugar.Errorf("error saving metrics: %v", err)
			}
		case <-ticker.C:
			if ticker != nil {
				if err := f.SaveMetrics(); err != nil {
					logger.Sugar.Errorf("error saving metrics: %v", err)
				}
			}
		}
	}
}
