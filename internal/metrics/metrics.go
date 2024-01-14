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

func (m Metric) GetValueAsString() (string, error) {
	switch m.Type {
	case "counter":
		return fmt.Sprintf("%d", m.Value), nil
	case "gauge":
		return fmt.Sprintf("%g", m.Value), nil
	}
	return "", fmt.Errorf("unsuported metrics type")
}
