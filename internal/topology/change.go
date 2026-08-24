package topology

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"example.com/railvolt/internal/model"
	"github.com/google/uuid"
)

type ChangeKind string

const (
	ChangeSplit       ChangeKind = "split"
	ChangeMerge       ChangeKind = "merge"
	ChangeReturnGroup ChangeKind = "return_group"
)

type Change struct {
	ID           string            `json:"id"`
	Kind         ChangeKind        `json:"kind"`
	FromRevision string            `json:"from_revision"`
	ToRevision   string            `json:"to_revision"`
	AreasBefore  []string          `json:"areas_before"`
	AreasAfter   []string          `json:"areas_after"`
	Metadata     map[string]string `json:"metadata"`
	CreatedAt    time.Time         `json:"created_at"`
}

type ChangeLog struct {
	mu      sync.RWMutex
	changes []Change
}

func NewChangeLog() *ChangeLog {
	return &ChangeLog{changes: []Change{}}
}

func (l *ChangeLog) Record(kind ChangeKind, before, after model.Topology, metadata map[string]string) Change {
	change := Change{
		ID:           uuid.NewString(),
		Kind:         kind,
		FromRevision: before.Revision,
		ToRevision:   after.Revision,
		AreasBefore:  topologyAreaIDs(before),
		AreasAfter:   topologyAreaIDs(after),
		Metadata:     copyStrings(metadata),
		CreatedAt:    time.Now().UTC(),
	}
	l.mu.Lock()
	l.changes = append(l.changes, change)
	l.mu.Unlock()
	return change
}

func (l *ChangeLog) List() []Change {
	l.mu.RLock()
	defer l.mu.RUnlock()
	result := make([]Change, len(l.changes))
	for i, change := range l.changes {
		change.AreasBefore = append([]string(nil), change.AreasBefore...)
		change.AreasAfter = append([]string(nil), change.AreasAfter...)
		change.Metadata = copyStrings(change.Metadata)
		result[i] = change
	}
	return result
}

func (l *ChangeLog) Since(revision string) []Change {
	items := l.List()
	if revision == "" {
		return items
	}
	result := make([]Change, 0, len(items))
	include := false
	for _, item := range items {
		if item.FromRevision == revision || item.ToRevision == revision {
			include = true
		}
		if include {
			result = append(result, item)
		}
	}
	return result
}

func (l *ChangeLog) ValidateChain(activeRevision string) error {
	items := l.List()
	if len(items) == 0 {
		return nil
	}
	for index := 1; index < len(items); index++ {
		if items[index-1].ToRevision != items[index].FromRevision {
			return fmt.Errorf("topology change chain breaks between %s and %s", items[index-1].ID, items[index].ID)
		}
	}
	if items[len(items)-1].ToRevision != activeRevision {
		return fmt.Errorf("topology change chain ends at %s, active is %s", items[len(items)-1].ToRevision, activeRevision)
	}
	return nil
}

func (l *ChangeLog) AffectedAreas(fromRevision, toRevision string) []string {
	set := map[string]struct{}{}
	for _, change := range l.List() {
		if change.FromRevision == fromRevision || change.ToRevision == toRevision || len(set) > 0 {
			for _, area := range change.AreasBefore {
				set[area] = struct{}{}
			}
			for _, area := range change.AreasAfter {
				set[area] = struct{}{}
			}
		}
		if change.ToRevision == toRevision {
			break
		}
	}
	result := make([]string, 0, len(set))
	for area := range set {
		result = append(result, area)
	}
	sort.Strings(result)
	return result
}

func topologyAreaIDs(t model.Topology) []string {
	result := make([]string, 0, len(t.Areas))
	for id := range t.Areas {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
}

func copyStrings(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
