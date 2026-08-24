package testcert

import (
	"example.com/railvolt/internal/model"
	"sort"
	"time"
)

func ScopeMatches(cert model.Certificate, area model.Area) bool {
	if cert.AreaID != area.ID || !cert.Passed {
		return false
	}
	need := append([]string(nil), area.Boundaries...)
	sort.Strings(need)
	have := append([]string(nil), cert.Scope...)
	sort.Strings(have)
	if len(need) != len(have) {
		return false
	}
	for i := range need {
		if need[i] != have[i] {
			return false
		}
	}
	return true
}

func Fresh(cert model.Certificate, now time.Time) bool { return cert.Current(now.Unix()) }
