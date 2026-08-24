package model

import "time"

type Phase string

const (
	PhaseDraft           Phase = "draft"
	PhaseIsolating       Phase = "isolating"
	PhaseIsolated        Phase = "isolated"
	PhaseGrounded        Phase = "grounded"
	PhaseWorking         Phase = "working"
	PhaseReleasing       Phase = "releasing"
	PhaseReadyToEnergize Phase = "ready_to_energize"
	PhaseEnergized       Phase = "energized"
	PhaseFailed          Phase = "failed"
	PhaseCancelled       Phase = "cancelled"
)

type Area struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Feed        string   `json:"feed"`
	ReturnGroup string   `json:"return_group"`
	Boundaries  []string `json:"boundaries"`
}

type Topology struct {
	Revision     string              `json:"revision"`
	Areas        map[string]Area     `json:"areas"`
	SharedReturn map[string][]string `json:"shared_return"`
	PublishedAt  time.Time           `json:"published_at"`
}

type Permit struct {
	ID               string    `json:"id"`
	Worksite         string    `json:"worksite"`
	Owner            string    `json:"owner"`
	Generation       uint64    `json:"generation"`
	Phase            Phase     `json:"phase"`
	TopologyRevision string    `json:"topology_revision"`
	GroundLocked     bool      `json:"ground_locked"`
	EnergizeLocked   bool      `json:"energize_locked"`
	UpdatedAt        time.Time `json:"updated_at"`
	FailureReason    string    `json:"failure_reason,omitempty"`
}

type CommandState string

const (
	CommandPlanned   CommandState = "planned"
	CommandAccepted  CommandState = "accepted"
	CommandMoving    CommandState = "moving"
	CommandConfirmed CommandState = "confirmed"
	CommandFailed    CommandState = "failed"
	CommandCancelled CommandState = "cancelled"
)

type Command struct {
	ID        string       `json:"id"`
	PermitID  string       `json:"permit_id"`
	DeviceID  string       `json:"device_id"`
	Action    string       `json:"action"`
	State     CommandState `json:"state"`
	Epoch     uint64       `json:"epoch"`
	Error     string       `json:"error,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
}

type Certificate struct {
	ID        string    `json:"id"`
	AreaID    string    `json:"area_id"`
	Revision  string    `json:"revision"`
	Scope     []string  `json:"scope"`
	Passed    bool      `json:"passed"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Sample struct {
	AreaID     string    `json:"area_id"`
	Current    float64   `json:"current"`
	Voltage    float64   `json:"voltage"`
	ObservedAt time.Time `json:"observed_at"`
}

type TelemetryEvidence struct {
	ID          string    `json:"id"`
	AreaID      string    `json:"area_id"`
	Samples     []Sample  `json:"samples"`
	ZeroCurrent bool      `json:"zero_current"`
	Durable     bool      `json:"durable"`
	RecordedAt  time.Time `json:"recorded_at"`
}

type Receipt struct {
	PermitID   string    `json:"permit_id"`
	Worksite   string    `json:"worksite"`
	Owner      string    `json:"owner"`
	Generation uint64    `json:"generation"`
	Fencing    string    `json:"fencing"`
	AcceptedAt time.Time `json:"accepted_at"`
}

type Alarm struct {
	ID        string    `json:"id"`
	PermitID  string    `json:"permit_id"`
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	CommandID string    `json:"command_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Snapshot struct {
	Revision    uint64              `json:"revision"`
	Permits     []Permit            `json:"permits"`
	Commands    []Command           `json:"commands"`
	Topology    Topology            `json:"topology"`
	Evidence    []TelemetryEvidence `json:"evidence"`
	NextCommand int                 `json:"next_command"`
	Committed   bool                `json:"committed"`
}
