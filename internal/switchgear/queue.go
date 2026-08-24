package switchgear

import (
	"example.com/railvolt/internal/model"
	"fmt"
)

func (q *Queue) Begin(id string) (model.Command, error) {
	c, ok := q.SetState(id, model.CommandAccepted, "")
	if !ok {
		return model.Command{}, fmt.Errorf("command not found")
	}
	return c, nil
}

func (q *Queue) Move(id string) (model.Command, error) {
	c, ok := q.SetState(id, model.CommandMoving, "")
	if !ok {
		return model.Command{}, fmt.Errorf("command not found")
	}
	return c, nil
}

func (q *Queue) Confirm(id string) (model.Command, error) {
	c, ok := q.SetState(id, model.CommandConfirmed, "")
	if !ok {
		return model.Command{}, fmt.Errorf("command not found")
	}
	return c, nil
}

func (q *Queue) Fail(id, reason string) (model.Command, error) {
	c, ok := q.SetState(id, model.CommandFailed, reason)
	if !ok {
		return model.Command{}, fmt.Errorf("command not found")
	}
	return c, nil
}
