package journal

import (
	"example.com/railvolt/internal/model"
	"time"
)

func (s *Store) RecordCommand(command model.Command) error {
	return s.Append("command", map[string]any{"at": time.Now().UTC(), "command": command})
}

func (s *Store) RecordPermit(permit model.Permit) error { return s.Append("permit", permit) }

func (s *Store) RecordAlarm(alarm model.Alarm) error { return s.Append("alarm", alarm) }
