package telemetry

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"example.com/railvolt/internal/model"
)

type WindowSummary struct {
	AreaID         string    `json:"area_id"`
	Start          time.Time `json:"start"`
	End            time.Time `json:"end"`
	Samples        int       `json:"samples"`
	MinimumCurrent float64   `json:"minimum_current"`
	MaximumCurrent float64   `json:"maximum_current"`
	AverageCurrent float64   `json:"average_current"`
	MaximumVoltage float64   `json:"maximum_voltage"`
	ZeroStable     bool      `json:"zero_stable"`
}

type WindowStore struct {
	mu       sync.RWMutex
	windows  map[string][]model.Sample
	capacity int
}

func NewWindowStore(capacity int) *WindowStore {
	if capacity < 1 {
		capacity = 1
	}
	return &WindowStore{windows: map[string][]model.Sample{}, capacity: capacity}
}

func (s *WindowStore) Add(sample model.Sample) error {
	if err := ValidateSample(sample); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	items := append(s.windows[sample.AreaID], sample)
	sort.Slice(items, func(i, j int) bool {
		return items[i].ObservedAt.Before(items[j].ObservedAt)
	})
	if len(items) > s.capacity {
		items = append([]model.Sample(nil), items[len(items)-s.capacity:]...)
	}
	s.windows[sample.AreaID] = items
	return nil
}

func (s *WindowStore) Samples(areaID string) []model.Sample {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]model.Sample(nil), s.windows[areaID]...)
}

func (s *WindowStore) Since(areaID string, since time.Time) []model.Sample {
	items := s.Samples(areaID)
	result := make([]model.Sample, 0, len(items))
	for _, item := range items {
		if !item.ObservedAt.Before(since) {
			result = append(result, item)
		}
	}
	return result
}

func (s *WindowStore) Summary(areaID string, since time.Time) (WindowSummary, error) {
	items := s.Since(areaID, since)
	if len(items) == 0 {
		return WindowSummary{}, fmt.Errorf("no telemetry samples in window")
	}
	last, _ := Last(items)
	summary := WindowSummary{
		AreaID:         areaID,
		Start:          items[0].ObservedAt,
		End:            last.ObservedAt,
		Samples:        len(items),
		MinimumCurrent: math.MaxFloat64,
		ZeroStable:     true,
	}
	var total float64
	for _, item := range items {
		if item.Current < summary.MinimumCurrent {
			summary.MinimumCurrent = item.Current
		}
		if item.Current > summary.MaximumCurrent {
			summary.MaximumCurrent = item.Current
		}
		if item.Voltage > summary.MaximumVoltage {
			summary.MaximumVoltage = item.Voltage
		}
		if math.Abs(item.Current) > 0.01 {
			summary.ZeroStable = false
		}
		total += item.Current
	}
	summary.AverageCurrent = total / float64(len(items))
	return summary, nil
}

func (s *WindowStore) Areas() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]string, 0, len(s.windows))
	for area := range s.windows {
		result = append(result, area)
	}
	sort.Strings(result)
	return result
}

func (s *WindowStore) StableZero(areaID string, duration time.Duration, now time.Time) bool {
	summary, err := s.Summary(areaID, now.Add(-duration))
	return err == nil && summary.ZeroStable && summary.Start.Before(now.Add(-duration/2))
}
