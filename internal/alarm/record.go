package alarm

import (
	"example.com/railvolt/internal/model"
	"github.com/google/uuid"
	"sync"
	"time"
)

type Service struct {
	mu     sync.RWMutex
	alarms []model.Alarm
}

func NewService() *Service { return &Service{} }

func (s *Service) Record(permitID, code, message, commandID string) model.Alarm {
	s.mu.Lock()
	defer s.mu.Unlock()
	a := model.Alarm{ID: uuid.NewString(), PermitID: permitID, Code: code, Message: message, CommandID: commandID, CreatedAt: time.Now().UTC()}
	s.alarms = append(s.alarms, a)
	return a
}

func (s *Service) List() []model.Alarm {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]model.Alarm(nil), s.alarms...)
}

func (s *Service) ForPermit(id string) []model.Alarm {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []model.Alarm{}
	for _, a := range s.alarms {
		if a.PermitID == id {
			out = append(out, a)
		}
	}
	return out
}
