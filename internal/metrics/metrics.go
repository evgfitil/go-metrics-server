package metrics

import "fmt"

type Metric struct {
	ID    string `json:"id"`
	MType string `json:"type"`
	Value interface{}
}

func (m Metric) GetName() string {
	return m.ID
}

func (m Metric) GetType() string {
	return m.MType
}

func (m Metric) GetValueAsString() (string, error) {
	switch m.MType {
	case "counter":
		return fmt.Sprintf("%d", m.Value), nil
	case "gauge":
		return fmt.Sprintf("%g", m.Value), nil
	}
	return "", fmt.Errorf("unsuported metrics type")
}

func NewGauge(name string, value float64) Metric {
	return Metric{
		ID:    name,
		MType: "gauge",
		Value: value,
	}
}

func NewCounter(name string, value int64) Metric {
	return Metric{
		ID:    name,
		MType: "counter",
		Value: value,
	}
}
