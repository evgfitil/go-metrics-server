package metrics

import "fmt"

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func (m Metrics) GetName() string {
	return m.ID
}

func (m Metrics) GetType() string {
	return m.MType
}

func (m Metrics) GetValueAsString() (string, error) {
	switch m.MType {
	case "counter":
		return fmt.Sprintf("%d", *m.Delta), nil
	case "gauge":

		return fmt.Sprintf("%g", *m.Value), nil
	}
	return "", fmt.Errorf("unsuported metrics type")
}

func NewGauge(name string, value float64) Metrics {
	return Metrics{
		ID:    name,
		MType: "gauge",
		Value: &value,
	}
}

func NewCounter(name string, value int64) Metrics {
	return Metrics{
		ID:    name,
		MType: "counter",
		Delta: &value,
	}
}
