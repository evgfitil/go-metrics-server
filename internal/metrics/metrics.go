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
	case "Counter":
		return fmt.Sprintf("%d", m.Value), nil
	case "Gauge":
		return fmt.Sprintf("%g", m.Value), nil
	}
	return "", fmt.Errorf("unsuported metrics type")
}
