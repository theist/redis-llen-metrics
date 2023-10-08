package main

// Metrics stuff.

type Metric struct {
	Name  string
	Value int
}

type MetricsStorage struct {
	Metrics []Metric
}

func (ms *MetricsStorage) AddMetric(metric Metric) {
	for cursor, m := range ms.Metrics {
		if m.Name == metric.Name {
			ms.Metrics[cursor] = metric
			return
		}
	}
	ms.Metrics = append(ms.Metrics, metric)
}

func (ms *MetricsStorage) ResetMetrics() {
	for cursor := range ms.Metrics {
		ms.Metrics[cursor].Value = 0
	}
}
