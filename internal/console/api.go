package console

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"medops/internal/alert"
	"medops/internal/audit"
	"medops/internal/device"
	"medops/internal/heart"
	"medops/internal/ns"
	"medops/internal/plan"
	"medops/internal/qc"
	"medops/internal/quota"
	"medops/internal/store"
	"medops/internal/test"
	"medops/internal/trend"
)

// Deps carries every service the console handlers need.
type Deps struct {
	FS           *store.FileStore
	Namespaces   *ns.WardRegistry
	Devices      *device.Registry
	Heart        *heart.Service
	HeartPolicy  heart.Policy
	QC           *qc.ScheduleService
	Thresholds   *qc.ThresholdService
	Methods      *qc.MethodService
	Calibrations *qc.CalibrationRegistry
	Tests        *test.Service
	Reagents     *test.ReagentRegistry
	Trend        *trend.Service
	Plan         *plan.Service
	Alerts       *alert.Service
	Quota        *quota.Service
	Audit        *audit.Recorder
	Reports      *store.ReportStore
	Batches      *store.BatchStore
}

// API is the HTTP control plane.
type API struct {
	deps   Deps
	router chi.Router
}

func NewAPI(deps Deps) *API {
	api := &API{deps: deps}
	api.router = chi.NewRouter()
	api.routes()
	return api
}

func (a *API) Router() http.Handler { return a.router }

func (a *API) routes() {
	a.router.Get("/healthz", a.handleHealth)
	a.router.Get("/api/namespaces", a.handleNamespaces)
	a.router.Route("/api/devices", func(r chi.Router) {
		r.Post("/", a.handleRegisterDevice)
		r.Get("/", a.handleListDevices)
		r.Get("/{id}", a.handleGetDevice)
		r.Post("/{id}/beat", a.handleDeviceBeat)
		r.Post("/{id}/ward-move", a.handleWardMove)
		r.Post("/{id}/firmware-upgrade", a.handleFirmwareUpgrade)
		r.Post("/{id}/reclassify", a.handleReclassify)
		r.Post("/{id}/maintenance-begin", a.handleMaintenanceBegin)
		r.Post("/{id}/retire", a.handleRetireDevice)
		r.Post("/{id}/clock-sync", a.handleClockSync)
		r.Get("/{id}/trend", a.handleTrend)
		r.Post("/{id}/trend/evaluate", a.handleTrendEvaluate)
	})
	a.router.Route("/api/qc", func(r chi.Router) {
		r.Post("/plans", a.handleCreatePlan)
		r.Get("/plans", a.handleListPlans)
		r.Get("/plans/{id}", a.handleGetPlan)
		r.Post("/plans/{id}/activate", a.handleActivatePlan)
		r.Post("/plans/{id}/suspend", a.handleSuspendPlan)
		r.Post("/plans/{id}/advance", a.handleAdvancePlan)
		r.Post("/plans/{id}/calibrate", a.handleRunCalibration)
		r.Post("/plans/{id}/dispatch", a.handleDispatchPlan)
		r.Post("/plans/{id}/commit", a.handleCommitPlan)
		r.Post("/plans/{id}/recover", a.handleRecoverPlan)
	})
	a.router.Route("/api/test", func(r chi.Router) {
		r.Post("/run", a.handleRunTest)
		r.Post("/sample", a.handleRunSample)
		r.Get("/reports", a.handleListReports)
	})
	a.router.Route("/api/maintenance", func(r chi.Router) {
		r.Post("/advise", a.handleAdvise)
		r.Get("/plans", a.handleListMaintenancePlans)
		r.Get("/queue", a.handleMaintenanceQueue)
		r.Post("/plans/{id}/begin", a.handleBeginMaintenance)
		r.Post("/plans/{id}/complete", a.handleCompleteMaintenance)
	})
	a.router.Route("/api/alerts", func(r chi.Router) {
		r.Get("/", a.handleListAlerts)
		r.Post("/evaluate", a.handleAlertEvaluate)
		r.Post("/{id}/ack", a.handleAckAlert)
	})
	a.router.Get("/api/quota", a.handleQuota)
	a.router.Get("/api/audit", a.handleAudit)
}

func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":             "ok",
		"time":               time.Now().Format(time.RFC3339),
		"heartbeat_interval": a.deps.HeartPolicy.Interval.String(),
	})
}

func (a *API) audit(actor, action, target, detail string) {
	_, _ = a.deps.Audit.Append(actor, action, target, detail, time.Now())
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func readBody(r *http.Request, target interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}
