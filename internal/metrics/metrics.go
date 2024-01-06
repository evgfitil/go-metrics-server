package metrics

import "fmt"

type Metric interface {
	GetName() string
	GetValueAsString() string
}

type Counter struct {
	Name  string
	Value int64
}

func (c Counter) GetName() string {
	return c.Name
}

func (c Counter) GetValueAsString() string {
	return fmt.Sprintf("%d", c.Value)
}

type Gauge struct {
	Name  string
	Value float64
}

func (g Gauge) GetName() string {
	return g.Name
}

func (g Gauge) GetValueAsString() string {
	return fmt.Sprintf("%f", g.Value)
}
