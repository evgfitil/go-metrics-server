package metrics

type Metric interface {
	GetName() string
}

type Counter struct {
	Name  string
	Value int64
}

func (c Counter) GetName() string {
	return c.Name
}

type Gauge struct {
	Name  string
	Value float64
}

func (g Gauge) GetName() string {
	return g.Name
}
