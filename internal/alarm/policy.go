package alarm

import "example.com/railvolt/internal/model"

func BlocksEnergize(items []model.Alarm, permitID string) bool {
	for _, item := range items {
		if item.PermitID == permitID && (IsProtection(item) || item.Code == "device_rejected") {
			return true
		}
	}
	return false
}

func Summary(items []model.Alarm) map[string]int {
	out := map[string]int{}
	for _, item := range items {
		out[item.Code]++
	}
	return out
}
