package journal

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"example.com/railvolt/internal/model"
	"github.com/google/uuid"
)

type ComponentRevision struct {
	Name     string `json:"name"`
	Revision uint64 `json:"revision"`
	Digest   string `json:"digest"`
}

type Checkpoint struct {
	ID          string              `json:"id"`
	Revision    uint64              `json:"revision"`
	Components  []ComponentRevision `json:"components"`
	Snapshot    model.Snapshot      `json:"snapshot"`
	Committed   bool                `json:"committed"`
	CreatedAt   time.Time           `json:"created_at"`
	CommittedAt time.Time           `json:"committed_at,omitempty"`
}

type Checkpoints struct {
	mu      sync.RWMutex
	items   map[string]Checkpoint
	ordered []string
}

func NewCheckpoints() *Checkpoints {
	return &Checkpoints{items: map[string]Checkpoint{}, ordered: []string{}}
}

func (c *Checkpoints) Prepare(snapshot model.Snapshot, components []ComponentRevision) (Checkpoint, error) {
	if snapshot.Revision == 0 {
		return Checkpoint{}, fmt.Errorf("snapshot revision is required")
	}
	if len(components) == 0 {
		return Checkpoint{}, fmt.Errorf("component revisions are required")
	}
	seen := map[string]struct{}{}
	for _, component := range components {
		if component.Name == "" || component.Revision == 0 || component.Digest == "" {
			return Checkpoint{}, fmt.Errorf("component revision is incomplete")
		}
		if _, exists := seen[component.Name]; exists {
			return Checkpoint{}, fmt.Errorf("duplicate component %s", component.Name)
		}
		seen[component.Name] = struct{}{}
	}
	checkpoint := Checkpoint{
		ID:         uuid.NewString(),
		Revision:   snapshot.Revision,
		Components: append([]ComponentRevision(nil), components...),
		Snapshot:   snapshot,
		CreatedAt:  time.Now().UTC(),
	}
	c.mu.Lock()
	c.items[checkpoint.ID] = checkpoint
	c.ordered = append(c.ordered, checkpoint.ID)
	c.mu.Unlock()
	return cloneCheckpoint(checkpoint), nil
}

func (c *Checkpoints) Commit(id string) (Checkpoint, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	checkpoint, ok := c.items[id]
	if !ok {
		return Checkpoint{}, fmt.Errorf("checkpoint not found")
	}
	if checkpoint.Committed {
		return cloneCheckpoint(checkpoint), nil
	}
	for _, component := range checkpoint.Components {
		if component.Revision != checkpoint.Revision {
			return Checkpoint{}, fmt.Errorf("component %s revision %d differs from checkpoint %d", component.Name, component.Revision, checkpoint.Revision)
		}
	}
	checkpoint.Committed = true
	checkpoint.CommittedAt = time.Now().UTC()
	checkpoint.Snapshot.Committed = true
	c.items[id] = checkpoint
	return cloneCheckpoint(checkpoint), nil
}

func (c *Checkpoints) Get(id string) (Checkpoint, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	checkpoint, ok := c.items[id]
	return cloneCheckpoint(checkpoint), ok
}

func (c *Checkpoints) LatestCommitted() (Checkpoint, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for index := len(c.ordered) - 1; index >= 0; index-- {
		checkpoint := c.items[c.ordered[index]]
		if checkpoint.Committed {
			return cloneCheckpoint(checkpoint), true
		}
	}
	return Checkpoint{}, false
}

func (c *Checkpoints) List() []Checkpoint {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]Checkpoint, 0, len(c.ordered))
	for _, id := range c.ordered {
		result = append(result, cloneCheckpoint(c.items[id]))
	}
	return result
}

func (c *Checkpoints) CompleteRevisions() []uint64 {
	set := map[uint64]struct{}{}
	for _, checkpoint := range c.List() {
		if checkpoint.Committed {
			set[checkpoint.Revision] = struct{}{}
		}
	}
	result := make([]uint64, 0, len(set))
	for revision := range set {
		result = append(result, revision)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func (c *Checkpoints) Recover() (model.Snapshot, error) {
	checkpoint, ok := c.LatestCommitted()
	if !ok {
		return model.Snapshot{}, fmt.Errorf("no committed recovery checkpoint")
	}
	consistent, err := RecoverSingleRevision(checkpoint.Snapshot, checkpoint.Snapshot)
	if err != nil {
		return model.Snapshot{}, err
	}
	return Recover(consistent)
}

func cloneCheckpoint(checkpoint Checkpoint) Checkpoint {
	checkpoint.Components = append([]ComponentRevision(nil), checkpoint.Components...)
	checkpoint.Snapshot.Permits = append([]model.Permit(nil), checkpoint.Snapshot.Permits...)
	checkpoint.Snapshot.Commands = append([]model.Command(nil), checkpoint.Snapshot.Commands...)
	checkpoint.Snapshot.Evidence = append([]model.TelemetryEvidence(nil), checkpoint.Snapshot.Evidence...)
	return checkpoint
}
