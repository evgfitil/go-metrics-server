package storage

import (
	"encoding/json"
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"os"
)

type FileStorage struct {
	MemStorage
	file *os.File
}

func NewFileStorage(filename string) (*FileStorage, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		logger.Sugar.Fatalf("error open file: %v", err)
		return nil, err
	}

	fs := &FileStorage{
		MemStorage: MemStorage{
			metrics: make(map[string]*metrics.Metrics),
		},
		file: file,
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

	if err = json.Unmarshal(data, &f.metrics); err != nil {
		return err
	}

	for _, metric := range f.metrics {
		f.Update(metric)
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

	metricsMap := f.GetAllMetrics()
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
