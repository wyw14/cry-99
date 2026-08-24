package isolation

import "example.com/railvolt/internal/model"

func Complete(p Plan) bool {
	if p.Cancelled || len(p.Steps) == 0 {
		return false
	}
	for _, step := range p.Steps {
		if step.State != model.CommandConfirmed {
			return false
		}
	}
	return true
}

func Recoverable(p Plan) bool {
	if !p.Cancelled {
		return true
	}
	for _, step := range p.Steps {
		if step.State == model.CommandConfirmed || step.State == model.CommandMoving {
			return true
		}
	}
	return false
}
