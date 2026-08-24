package telemetry

import (
	"example.com/railvolt/internal/model"
	"fmt"
)

func ValidateSample(sample model.Sample) error {
	if sample.AreaID == "" {
		return fmt.Errorf("sample area missing")
	}
	if sample.Current < 0 || sample.Voltage < 0 {
		return fmt.Errorf("sample values invalid")
	}
	return nil
}

func Last(samples []model.Sample) (model.Sample, bool) {
	if len(samples) == 0 {
		return model.Sample{}, false
	}
	return samples[len(samples)-1], true
}
