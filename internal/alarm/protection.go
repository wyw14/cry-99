package alarm

import (
	"example.com/railvolt/internal/model"
	"fmt"
)

type ProtectionTrip struct{ PermitID, DeviceID, CommandID, Reason string }

func (s *Service) Trip(event ProtectionTrip) model.Alarm {
	message := fmt.Sprintf("energize aborted by protection device=%s reason=%s", event.DeviceID, event.Reason)
	return s.Record(event.PermitID, "protection_trip", message, event.CommandID)
}

func IsProtection(alarm model.Alarm) bool { return alarm.Code == "protection_trip" }
