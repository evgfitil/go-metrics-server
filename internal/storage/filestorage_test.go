package storage

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
)

func TestNewFileStorage(t *testing.T) {
	file, err := os.CreateTemp("", "metrics")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	fs, err := NewFileStorage(file.Name(), 10)
	assert.NoError(t, err)
	assert.NotNil(t, fs)
}

func TestSaveMetrics(t *testing.T) {
	file, err := os.CreateTemp("", "metrics")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	fs, err := NewFileStorage(file.Name(), 10)
	assert.NoError(t, err)
	assert.NotNil(t, fs)

	err = fs.SaveMetrics(context.Background())
	assert.NoError(t, err)
}

func TestClose(t *testing.T) {
	file, err := os.CreateTemp("", "metrics")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	fs, err := NewFileStorage(file.Name(), 10)
	assert.NoError(t, err)
	assert.NotNil(t, fs)

	err = fs.Close()
	assert.NoError(t, err)
}

func TestFileStorage_Ping(t *testing.T) {
	tests := []struct {
		name string
		m    *FileStorage
		want error
	}{
		{name: "Ping FileStorage", m: &FileStorage{}, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Ping(context.Background()); got != tt.want {
				t.Errorf("Ping() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadMetrics(t *testing.T) {
	logger.InitLogger()
	t.Run("SuccessfulLoad", func(t *testing.T) {
		tmpfile, err := os.CreateTemp("", "metrics*.json")
		assert.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		metricValue := 1.1
		expectedMetrics := map[string]*metrics.Metrics{
			"metric1": {ID: "metric1", Value: &metricValue},
			"metric2": {ID: "metric2", Value: &metricValue},
		}

		data, err := json.Marshal(expectedMetrics)
		assert.NoError(t, err)

		_, err = tmpfile.Write(data)
		assert.NoError(t, err)

		err = tmpfile.Close()
		assert.NoError(t, err)

		fs, err := NewFileStorage(tmpfile.Name(), 0)
		assert.NoError(t, err)

		err = fs.LoadMetrics()
		assert.NoError(t, err)

		fs.mu.Lock()
		defer fs.mu.Unlock()

		for id, expectedMetric := range expectedMetrics {
			actualMetric, exists := fs.metrics[id]
			assert.True(t, exists, "Metric %s not found", id)
			assert.Equal(t, expectedMetric.ID, actualMetric.ID)
			assert.Equal(t, expectedMetric.Value, actualMetric.Value)
		}
	})

	// File does not exist
	t.Run("FileDoesNotExist", func(t *testing.T) {
		fs, err := NewFileStorage("nonexistent_file.json", 0)
		assert.NoError(t, err)

		err = fs.LoadMetrics()
		assert.NoError(t, err)
	})

	// Empty file
	t.Run("EmptyFile", func(t *testing.T) {
		tmpfile, err := os.CreateTemp("", "metrics*.json")
		assert.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		fs, err := NewFileStorage(tmpfile.Name(), 0)
		assert.NoError(t, err)

		err = fs.LoadMetrics()
		assert.NoError(t, err)
	})

	// Invalid JSON format
	t.Run("InvalidJSON", func(t *testing.T) {
		tmpfile, err := os.CreateTemp("", "metrics*.json")
		assert.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		invalidJSON := []byte(`{"metric1": {"id": "metric1", "value": "invalid"}}`)
		_, err = tmpfile.Write(invalidJSON)
		assert.NoError(t, err)

		err = tmpfile.Close()
		assert.NoError(t, err)

		fs, err := NewFileStorage(tmpfile.Name(), 0)
		assert.NoError(t, err)

		err = fs.LoadMetrics()
		assert.Error(t, err)
	})
}
