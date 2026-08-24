package permit

import "example.com/railvolt/internal/model"

func (s *Service) Snapshot() []model.Permit { return s.List() }

func Restore(items []model.Permit) *Service {
	s := NewService()
	for _, p := range items {
		s.permits[p.ID] = p
	}
	return s
}
