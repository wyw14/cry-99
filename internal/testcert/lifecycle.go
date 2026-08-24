package testcert

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"example.com/railvolt/internal/model"
	"github.com/google/uuid"
)

type CertificateStatus string

const (
	CertificatePending  CertificateStatus = "pending"
	CertificatePassed   CertificateStatus = "passed"
	CertificateFailed   CertificateStatus = "failed"
	CertificateExpired  CertificateStatus = "expired"
	CertificateArchived CertificateStatus = "archived"
)

type Record struct {
	Certificate model.Certificate `json:"certificate"`
	Status      CertificateStatus `json:"status"`
	RequestedBy string            `json:"requested_by"`
	ApprovedBy  string            `json:"approved_by,omitempty"`
	Notes       []string          `json:"notes"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type Lifecycle struct {
	mu      sync.RWMutex
	records map[string]Record
}

func NewLifecycle() *Lifecycle {
	return &Lifecycle{records: map[string]Record{}}
}

func (l *Lifecycle) Request(areaID, revision string, scope []string, actor string, validity time.Duration) (Record, error) {
	if areaID == "" || revision == "" || actor == "" {
		return Record{}, fmt.Errorf("area, revision and requester are required")
	}
	if len(scope) == 0 {
		return Record{}, fmt.Errorf("certificate scope cannot be empty")
	}
	now := time.Now().UTC()
	record := Record{
		Certificate: model.Certificate{
			ID:        uuid.NewString(),
			AreaID:    areaID,
			Revision:  revision,
			Scope:     uniqueScope(scope),
			IssuedAt:  now,
			ExpiresAt: now.Add(validity),
		},
		Status:      CertificatePending,
		RequestedBy: actor,
		UpdatedAt:   now,
	}
	l.mu.Lock()
	l.records[record.Certificate.ID] = record
	l.mu.Unlock()
	return cloneRecord(record), nil
}

func (l *Lifecycle) Complete(id, approver string, passed bool, note string) (Record, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	record, ok := l.records[id]
	if !ok {
		return Record{}, fmt.Errorf("certificate request not found")
	}
	if record.Status != CertificatePending {
		return Record{}, fmt.Errorf("certificate request is already complete")
	}
	if approver == "" {
		return Record{}, fmt.Errorf("approver is required")
	}
	record.ApprovedBy = approver
	record.Certificate.Passed = passed
	if passed {
		record.Status = CertificatePassed
	} else {
		record.Status = CertificateFailed
	}
	if note != "" {
		record.Notes = append(record.Notes, note)
	}
	record.UpdatedAt = time.Now().UTC()
	l.records[id] = record
	return cloneRecord(record), nil
}

func (l *Lifecycle) ArchiveForRevision(revision, reason string) []Record {
	l.mu.Lock()
	defer l.mu.Unlock()
	result := []Record{}
	for id, record := range l.records {
		if record.Certificate.Revision == revision || record.Status == CertificateArchived {
			continue
		}
		record.Status = CertificateArchived
		if reason != "" {
			record.Notes = append(record.Notes, reason)
		}
		record.UpdatedAt = time.Now().UTC()
		l.records[id] = record
		result = append(result, cloneRecord(record))
	}
	return result
}

func (l *Lifecycle) Refresh(now time.Time) []Record {
	l.mu.Lock()
	defer l.mu.Unlock()
	changed := []Record{}
	for id, record := range l.records {
		if record.Status == CertificatePassed && now.After(record.Certificate.ExpiresAt) {
			record.Status = CertificateExpired
			record.UpdatedAt = now
			l.records[id] = record
			changed = append(changed, cloneRecord(record))
		}
	}
	return changed
}

func (l *Lifecycle) Get(id string) (Record, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	record, ok := l.records[id]
	return cloneRecord(record), ok
}

func (l *Lifecycle) List() []Record {
	l.mu.RLock()
	defer l.mu.RUnlock()
	result := make([]Record, 0, len(l.records))
	for _, record := range l.records {
		result = append(result, cloneRecord(record))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Certificate.IssuedAt.Before(result[j].Certificate.IssuedAt)
	})
	return result
}

func (l *Lifecycle) Usable(area model.Area, revision string, now time.Time) []model.Certificate {
	result := []model.Certificate{}
	for _, record := range l.List() {
		certificate := record.Certificate
		if record.Status != CertificatePassed || certificate.Revision != revision {
			continue
		}
		if !Fresh(certificate, now) || !ScopeMatches(certificate, area) {
			continue
		}
		result = append(result, certificate)
	}
	return result
}

func (l *Lifecycle) Summary() map[CertificateStatus]int {
	result := map[CertificateStatus]int{}
	for _, record := range l.List() {
		result[record.Status]++
	}
	return result
}

func uniqueScope(scope []string) []string {
	set := make(map[string]struct{}, len(scope))
	for _, item := range scope {
		if item != "" {
			set[item] = struct{}{}
		}
	}
	result := make([]string, 0, len(set))
	for item := range set {
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}

func cloneRecord(record Record) Record {
	record.Certificate.Scope = append([]string(nil), record.Certificate.Scope...)
	record.Notes = append([]string(nil), record.Notes...)
	return record
}
