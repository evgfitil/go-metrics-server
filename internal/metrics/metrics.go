package metrics

import "fmt"

// Metrics represents a metric with its ID, type, and value.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

// GetName returns the name (ID) of the metric.
func (m Metrics) GetName() string {
	return m.ID
}

// GetType returns the type of the metric.
func (m Metrics) GetType() string {
	return m.MType
}

// GetValueAsString returns the value of the metric as a string.
// It returns an error if the metric type is unsupported.
func (m Metrics) GetValueAsString() (string, error) {
	switch m.MType {
	case "counter":
		return fmt.Sprintf("%d", *m.Delta), nil
	case "gauge":

		return fmt.Sprintf("%g", *m.Value), nil
	}
	return "", fmt.Errorf("unsuported metrics type")
}

// NewGauge creates a new gauge metric with the specified name and value.
func NewGauge(name string, value float64) Metrics {
	return Metrics{
		ID:    name,
		MType: "gauge",
		Value: &value,
	}
}

// NewCounter creates a new counter metric with the specified name and value.
func NewCounter(name string, value int64) Metrics {
	return Metrics{
		ID:    name,
		MType: "counter",
		Delta: &value,
	}
}
