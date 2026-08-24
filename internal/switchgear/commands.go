package switchgear

import (
	"example.com/railvolt/internal/model"
	"github.com/google/uuid"
	"sync"
	"time"
)

type Queue struct {
	mu       sync.Mutex
	commands []model.Command
	epoch    uint64
}

func NewQueue() *Queue { return &Queue{epoch: 1} }

func (q *Queue) Plan(permitID, deviceID, action string) model.Command {
	q.mu.Lock()
	defer q.mu.Unlock()
	c := model.Command{ID: uuid.NewString(), PermitID: permitID, DeviceID: deviceID, Action: action, State: model.CommandPlanned, Epoch: q.epoch, CreatedAt: time.Now().UTC()}
	q.commands = append(q.commands, c)
	return c
}

func (q *Queue) SetState(id string, state model.CommandState, reason string) (model.Command, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i := range q.commands {
		if q.commands[i].ID == id {
			q.commands[i].State, q.commands[i].Error = state, reason
			return q.commands[i], true
		}
	}
	return model.Command{}, false
}

func (q *Queue) All() []model.Command {
	q.mu.Lock()
	defer q.mu.Unlock()
	return append([]model.Command(nil), q.commands...)
}

func (q *Queue) Confirmed() []model.Command {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := []model.Command{}
	for _, c := range q.commands {
		if c.State == model.CommandConfirmed {
			out = append(out, c)
		}
	}
	return out
}

func (q *Queue) Pending() []model.Command {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := []model.Command{}
	for _, c := range q.commands {
		if c.State != model.CommandConfirmed && c.State != model.CommandFailed && c.State != model.CommandCancelled {
			out = append(out, c)
		}
	}
	return out
}

func (q *Queue) CancelPending() []model.Command {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := []model.Command{}
	for i := range q.commands {
		if q.commands[i].State != model.CommandConfirmed {
			q.commands[i].State = model.CommandCancelled
			out = append(out, q.commands[i])
		}
	}
	return out
}

func (q *Queue) RestoreBoundary(commands []model.Command, next int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.commands = append([]model.Command(nil), commands...)
	q.epoch++
	if next > len(q.commands) {
		next = len(q.commands)
	}
	for i := range q.commands {
		if i < next && q.commands[i].State == model.CommandPlanned {
			q.commands[i].State = model.CommandConfirmed
		}
	}
}
