package metrics

import "fmt"

type Metric struct {
	Name  string
	Type  string
	Value interface{}
}

func (m Metric) GetName() string {
	return m.Name
}

func (m Metric) GetType() string {
	return m.Type
}

func (m Metric) GetValueAsString() (string, error) {
	switch m.Type {
	case "counter":
		return fmt.Sprintf("%d", m.Value), nil
	case "gauge":
		return fmt.Sprintf("%g", m.Value), nil
	}
	return "", fmt.Errorf("unsuported metrics type")
}

func NewGauge(name string, value float64) Metric {
	return Metric{
		Name:  name,
		Type:  "gauge",
		Value: value,
	}
}

func NewCounter(name string, value int64) Metric {
	return Metric{
		Name:  name,
		Type:  "counter",
		Value: value,
	}
}
