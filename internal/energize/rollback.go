package energize

import (
	"example.com/railvolt/internal/alarm"
	"fmt"
)

func (s *Sequencer) ProtectionTrip(permitID, deviceID, commandID, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	op, ok := s.operations[permitID]
	if !ok {
		return fmt.Errorf("energize operation not found")
	}
	op.Failed = "energize aborted by protection"
	op.Complete = false
	s.operations[permitID] = op
	s.alarms.Trip(alarm.ProtectionTrip{PermitID: permitID, DeviceID: deviceID, CommandID: commandID, Reason: reason})
	_, err := s.permits.Rollback(permitID, op.Failed)
	return err
}
