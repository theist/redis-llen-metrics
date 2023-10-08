package main

// Test for ResetMetrics.
func (ms *MetricsStorage) TestResetMetrics() {
	ms.AddMetric(Metric{Name: "test", Value: 1})
	ms.ResetMetrics()
	for _, m := range ms.Metrics {
		if m.Value != 0 {
			panic("ResetMetrics failed")
		}
	}
}
