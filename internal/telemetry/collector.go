package telemetry

import (
	"example.com/railvolt/internal/model"
	"sync"
	"time"
)

type Collector struct {
	mu      sync.RWMutex
	samples map[string][]model.Sample
}

func NewCollector() *Collector { return &Collector{samples: map[string][]model.Sample{}} }

func (c *Collector) Add(area string, current, voltage float64) model.Sample {
	c.mu.Lock()
	defer c.mu.Unlock()
	sample := model.Sample{AreaID: area, Current: current, Voltage: voltage, ObservedAt: time.Now().UTC()}
	c.samples[area] = append(c.samples[area], sample)
	return sample
}

func (c *Collector) Samples(area string) []model.Sample {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]model.Sample(nil), c.samples[area]...)
}

func (c *Collector) Zero(area string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	values := c.samples[area]
	if len(values) == 0 {
		return false
	}
	for _, sample := range values {
		if sample.Current != 0 {
			return false
		}
	}
	return true
}
