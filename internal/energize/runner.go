package energize

import (
	"context"
	"sync"
)

type Runner struct {
	sequence *Sequencer
	mu       sync.Mutex
	running  map[string]context.CancelFunc
}

func NewRunner(sequence *Sequencer) *Runner {
	return &Runner{sequence: sequence, running: map[string]context.CancelFunc{}}
}

func (r *Runner) Run(ctx context.Context, permitID string) <-chan error {
	done := make(chan error, 1)
	child, cancel := context.WithCancel(ctx)
	r.mu.Lock()
	r.running[permitID] = cancel
	r.mu.Unlock()
	go func() {
		defer close(done)
		defer r.forget(permitID)
		for {
			op, err := r.sequence.Step(child, permitID)
			if err != nil {
				done <- err
				return
			}
			if op.Complete || op.Failed != "" {
				done <- nil
				return
			}
		}
	}()
	return done
}

func (r *Runner) Cancel(permitID string) {
	r.mu.Lock()
	cancel := r.running[permitID]
	r.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}
func (r *Runner) forget(id string) { r.mu.Lock(); defer r.mu.Unlock(); delete(r.running, id) }
