package api

import (
	"os"
	"path/filepath"
	"time"

	"example.com/railvolt/internal/alarm"
	"example.com/railvolt/internal/crew"
	"example.com/railvolt/internal/earth"
	"example.com/railvolt/internal/energize"
	"example.com/railvolt/internal/isolation"
	"example.com/railvolt/internal/journal"
	"example.com/railvolt/internal/permit"
	"example.com/railvolt/internal/switchgear"
	"example.com/railvolt/internal/telemetry"
	"example.com/railvolt/internal/testcert"
	"example.com/railvolt/internal/topology"
)

type Application struct {
	Topology        *topology.Manager
	Changes         *topology.ChangeLog
	Certificates    *testcert.Registry
	CertificateFlow *testcert.Lifecycle
	Crew            *crew.Manager
	Roster          *crew.Roster
	Permits         *permit.Service
	PermitHistory   *permit.History
	Commands        *switchgear.Queue
	Controller      *switchgear.Client
	Evidence        *switchgear.EvidenceStore
	Fleet           *switchgear.Fleet
	Plans           *isolation.Registry
	Isolation       *isolation.Service
	Telemetry       *telemetry.Collector
	TelemetryWindow *telemetry.WindowStore
	TelemetryLog    *telemetry.EvidenceStore
	Grounds         *earth.Service
	GroundChecklist *earth.Checklist
	GroundView      *earth.Aggregate
	GroundLock      *earth.Lock
	Alarms          *alarm.Service
	Eligibility     *energize.Eligibility
	Readiness       *energize.Readiness
	Sequencer       *energize.Sequencer
	Runner          *energize.Runner
	Journal         *journal.Store
	Sealer          *journal.Sealer
	Checkpoints     *journal.Checkpoints
	DataDir         string
	WebDir          string
}

func NewApplication(dataDir, webDir string) (*Application, error) {
	if dataDir == "" {
		dataDir = "data"
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}
	top := topology.NewManager()
	changes := topology.NewChangeLog()
	certs := testcert.NewRegistry()
	certificateFlow := testcert.NewLifecycle()
	crews := crew.NewManager()
	roster := crew.NewRoster()
	permits := permit.NewService()
	permitHistory := permit.NewHistory()
	commands := switchgear.NewQueue()
	controller := switchgear.NewClient()
	deviceEvidence := switchgear.NewEvidenceStore()
	fleet := switchgear.NewFleet()
	plans := isolation.NewRegistry()
	isolationService := isolation.NewService(plans, commands)
	collector := telemetry.NewCollector()
	window := telemetry.NewWindowStore(256)
	telemetryLog := telemetry.NewEvidenceStore()
	grounds := earth.NewService(collector, telemetryLog)
	groundChecklist := earth.NewChecklist()
	groundView := earth.NewAggregate()
	groundLock := earth.NewLock()
	alarms := alarm.NewService()
	eligibility := energize.NewEligibility(top, certs, groundView, alarms)
	readiness := energize.NewReadiness(eligibility, top)
	sequencer := energize.NewSequencer(commands, controller, permits, alarms)
	runner := energize.NewRunner(sequencer)
	store := journal.NewStore(filepath.Join(dataDir, "events.json"))
	checkpoints := journal.NewCheckpoints()
	if err := store.Load(); err != nil {
		return nil, err
	}
	app := &Application{Topology: top, Changes: changes, Certificates: certs, CertificateFlow: certificateFlow, Crew: crews, Roster: roster, Permits: permits, PermitHistory: permitHistory, Commands: commands, Controller: controller, Evidence: deviceEvidence, Fleet: fleet, Plans: plans, Isolation: isolationService, Telemetry: collector, TelemetryWindow: window, TelemetryLog: telemetryLog, Grounds: grounds, GroundChecklist: groundChecklist, GroundView: groundView, GroundLock: groundLock, Alarms: alarms, Eligibility: eligibility, Readiness: readiness, Sequencer: sequencer, Runner: runner, Journal: store, Sealer: journal.NewSealer(), Checkpoints: checkpoints, DataDir: dataDir, WebDir: webDir}
	app.seed()
	return app, nil
}

func (a *Application) seed() {
	t := a.Topology.Current()
	for _, area := range t.Areas {
		a.Certificates.Issue(area.ID, t.Revision, area.Boundaries, true, 8*time.Hour)
		request, _ := a.CertificateFlow.Request(area.ID, t.Revision, area.Boundaries, "system", 8*time.Hour)
		a.CertificateFlow.Complete(request.Certificate.ID, "duty-controller", true, "initial commissioning evidence")
		sample := a.Telemetry.Add(area.ID, 0, 0)
		a.TelemetryWindow.Add(sample)
		a.GroundLock.Set(area.ID, true)
		a.Fleet.Register(switchgear.Device{ID: area.ID + "-CB", Kind: "breaker", AreaID: area.ID, Position: "open", RemoteEnabled: true})
		a.GroundChecklist.Begin(area.ID+"-ground", area.ID)
	}
	a.GroundView.Rebuild(t, a.Grounds.Records())
	session := a.Crew.Handover("WS-1", "day", 1)
	a.Roster.Start(session.Worksite, session.Owner, []crew.Member{{Name: "Duty operator", Role: "controller", Qualification: "traction-power", ValidUntil: time.Now().UTC().Add(24 * time.Hour)}})
	p := a.Permits.Create(session.Worksite, session.Owner, t.Revision)
	a.Journal.RecordPermit(p)
}

func (a *Application) RefreshGroundView() {
	a.GroundView.Rebuild(a.Topology.Current(), a.Grounds.Records())
}
