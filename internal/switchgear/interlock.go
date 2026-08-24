package switchgear

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"example.com/railvolt/internal/model"
)

type Device struct {
	ID               string    `json:"id"`
	Kind             string    `json:"kind"`
	AreaID           string    `json:"area_id"`
	Position         string    `json:"position"`
	RemoteEnabled    bool      `json:"remote_enabled"`
	LocalLock        bool      `json:"local_lock"`
	ProtectionLocked bool      `json:"protection_locked"`
	Epoch            uint64    `json:"epoch"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type InterlockDecision struct {
	DeviceID string   `json:"device_id"`
	Action   string   `json:"action"`
	Allowed  bool     `json:"allowed"`
	Reasons  []string `json:"reasons"`
	Epoch    uint64   `json:"epoch"`
}

type Fleet struct {
	mu      sync.RWMutex
	devices map[string]Device
}

func NewFleet() *Fleet {
	return &Fleet{devices: map[string]Device{}}
}

func (f *Fleet) Register(device Device) error {
	if device.ID == "" || device.AreaID == "" || device.Kind == "" {
		return fmt.Errorf("device identity, kind and area are required")
	}
	if device.Position == "" {
		device.Position = "unknown"
	}
	if device.Epoch == 0 {
		device.Epoch = 1
	}
	device.UpdatedAt = time.Now().UTC()
	f.mu.Lock()
	defer f.mu.Unlock()
	if existing, ok := f.devices[device.ID]; ok && existing.Epoch >= device.Epoch {
		return fmt.Errorf("device epoch must advance")
	}
	f.devices[device.ID] = device
	return nil
}

func (f *Fleet) UpdatePosition(deviceID, position string, epoch uint64) (Device, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	device, ok := f.devices[deviceID]
	if !ok {
		return Device{}, fmt.Errorf("device not found")
	}
	if epoch < device.Epoch {
		return Device{}, fmt.Errorf("stale device epoch")
	}
	device.Position = position
	device.Epoch = epoch
	device.UpdatedAt = time.Now().UTC()
	f.devices[deviceID] = device
	return device, nil
}

func (f *Fleet) SetLocks(deviceID string, local, protection bool) (Device, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	device, ok := f.devices[deviceID]
	if !ok {
		return Device{}, fmt.Errorf("device not found")
	}
	device.LocalLock = local
	device.ProtectionLocked = protection
	device.Epoch++
	device.UpdatedAt = time.Now().UTC()
	f.devices[deviceID] = device
	return device, nil
}

func (f *Fleet) Decide(command model.Command) InterlockDecision {
	f.mu.RLock()
	defer f.mu.RUnlock()
	decision := InterlockDecision{DeviceID: command.DeviceID, Action: command.Action}
	device, ok := f.devices[command.DeviceID]
	if !ok {
		decision.Reasons = append(decision.Reasons, "device_not_registered")
		return decision
	}
	decision.Epoch = device.Epoch
	if !device.RemoteEnabled {
		decision.Reasons = append(decision.Reasons, "remote_control_disabled")
	}
	if device.LocalLock {
		decision.Reasons = append(decision.Reasons, "local_lock_active")
	}
	if device.ProtectionLocked {
		decision.Reasons = append(decision.Reasons, "protection_lock_active")
	}
	if command.Epoch < device.Epoch {
		decision.Reasons = append(decision.Reasons, "command_epoch_stale")
	}
	if command.Action == "open" && device.Position == "open" {
		decision.Reasons = append(decision.Reasons, "already_open")
	}
	if command.Action == "close" && device.Position == "closed" {
		decision.Reasons = append(decision.Reasons, "already_closed")
	}
	decision.Allowed = len(decision.Reasons) == 0
	return decision
}

func (f *Fleet) Get(deviceID string) (Device, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	device, ok := f.devices[deviceID]
	return device, ok
}

func (f *Fleet) List() []Device {
	f.mu.RLock()
	defer f.mu.RUnlock()
	result := make([]Device, 0, len(f.devices))
	for _, device := range f.devices {
		result = append(result, device)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].AreaID == result[j].AreaID {
			return result[i].ID < result[j].ID
		}
		return result[i].AreaID < result[j].AreaID
	})
	return result
}

func (f *Fleet) AreaState(areaID string) map[string]int {
	counts := map[string]int{}
	for _, device := range f.List() {
		if device.AreaID == areaID {
			counts[device.Position]++
		}
	}
	return counts
}
