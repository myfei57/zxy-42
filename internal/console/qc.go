package console

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"medops/internal/qc"
)

type createPlanRequest struct {
	DeviceID   string `json:"device_id"`
	CycleDays  int    `json:"cycle_days"`
	ReagentLot string `json:"reagent_lot"`
}

type commitRequest struct {
	DeviceIDs []string `json:"device_ids"`
	Parameter string   `json:"parameter"`
	Value     float64  `json:"value"`
}

type testRunRequest struct {
	PlanID    string  `json:"plan_id"`
	DeviceID  string  `json:"device_id"`
	Parameter string  `json:"parameter"`
	Value     float64 `json:"value"`
}

type sampleRunRequest struct {
	PlanID   string  `json:"plan_id"`
	DeviceID string  `json:"device_id"`
	Value    float64 `json:"value"`
}

func (a *API) handleCreatePlan(w http.ResponseWriter, r *http.Request) {
	var req createPlanRequest
	if err := readBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	dev, ok := a.deps.Devices.Get(req.DeviceID)
	if !ok {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	plan, err := a.deps.QC.Create(dev, req.CycleDays, req.ReagentLot, time.Now())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.audit("console", "qc-plan-create", plan.ID, dev.ID)
	writeJSON(w, http.StatusCreated, plan)
}

func (a *API) handleListPlans(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.deps.QC.All())
}

func (a *API) handleGetPlan(w http.ResponseWriter, r *http.Request) {
	plan, ok := a.deps.QC.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "plan not found")
		return
	}
	calibration, _ := a.deps.Calibrations.Get(plan.ID)
	response := map[string]interface{}{
		"plan":        plan,
		"calibration": calibration,
	}
	if calibration != nil {
		response["remaining_steps"] = calibration.RemainingSteps()
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *API) handleActivatePlan(w http.ResponseWriter, r *http.Request) {
	plan, ok := a.deps.QC.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "plan not found")
		return
	}
	if err := a.deps.QC.Activate(plan, time.Now()); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.audit("console", "qc-plan-activate", plan.ID, "")
	writeJSON(w, http.StatusOK, plan)
}

func (a *API) handleAdvancePlan(w http.ResponseWriter, r *http.Request) {
	plan, ok := a.deps.QC.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "plan not found")
		return
	}
	if err := a.deps.QC.AdvanceCycle(plan, time.Now()); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.audit("console", "qc-plan-advance", plan.ID, "")
	writeJSON(w, http.StatusOK, plan)
}

func (a *API) handleSuspendPlan(w http.ResponseWriter, r *http.Request) {
	plan, ok := a.deps.QC.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "plan not found")
		return
	}
	if err := a.deps.QC.Suspend(plan); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.audit("console", "qc-plan-suspend", plan.ID, "")
	writeJSON(w, http.StatusOK, plan)
}

func (a *API) handleRunCalibration(w http.ResponseWriter, r *http.Request) {
	plan, ok := a.deps.QC.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "plan not found")
		return
	}
	cal, ok := a.deps.Calibrations.Get(plan.ID)
	if !ok {
		writeError(w, http.StatusNotFound, "calibration not found")
		return
	}
	if err := cal.RunSequence(cal.Steps, time.Now()); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.deps.Calibrations.Put(cal); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.audit("console", "qc-calibrate", plan.ID, strconv.Itoa(len(cal.Steps)))
	writeJSON(w, http.StatusOK, cal)
}

func (a *API) handleDispatchPlan(w http.ResponseWriter, r *http.Request) {
	plan, ok := a.deps.QC.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "plan not found")
		return
	}
	dev, ok := a.deps.Devices.Get(plan.DeviceID)
	if !ok {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	dispatch, err := a.deps.QC.Dispatch(plan, dev, time.Now())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.audit("console", "qc-dispatch", plan.ID, dispatch.TargetWard)
	writeJSON(w, http.StatusOK, dispatch)
}

func (a *API) handleCommitPlan(w http.ResponseWriter, r *http.Request) {
	plan, ok := a.deps.QC.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "plan not found")
		return
	}
	var req commitRequest
	if err := readBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	at := time.Now()
	items := make([]qc.BatchItem, 0, len(req.DeviceIDs))
	for _, deviceID := range req.DeviceIDs {
		dev, ok := a.deps.Devices.Get(deviceID)
		if !ok {
			writeError(w, http.StatusNotFound, "device not found: "+deviceID)
			return
		}
		method, err := a.deps.Methods.Resolve(plan, dev)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		verdict, err := a.deps.Thresholds.Evaluate(plan, dev, req.Parameter, req.Value)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		result := qc.NewResult(plan.ID, deviceID, req.Parameter, req.Value, verdict, method.Name, plan.ReagentLot, at)
		items = append(items, qc.BatchItem{DeviceID: deviceID, Result: result})
	}
	outcome, err := a.deps.QC.CommitBatch(plan, qc.NewBatch(items))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.audit("console", "qc-commit", plan.ID, strconv.Itoa(outcome.Committed))
	writeJSON(w, http.StatusOK, outcome)
}

func (a *API) handleRecoverPlan(w http.ResponseWriter, r *http.Request) {
	plan, ok := a.deps.QC.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "plan not found")
		return
	}
	missing, err := a.deps.QC.RecoverPending(plan)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	paths := make([]string, 0, len(missing))
	for _, id := range missing {
		paths = append(paths, a.deps.Batches.Path(plan.ID, id))
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"missing": missing, "paths": paths})
}

func (a *API) handleRunTest(w http.ResponseWriter, r *http.Request) {
	var req testRunRequest
	if err := readBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	plan, ok := a.deps.QC.Get(req.PlanID)
	if !ok {
		writeError(w, http.StatusNotFound, "plan not found")
		return
	}
	dev, ok := a.deps.Devices.Get(req.DeviceID)
	if !ok {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	method, err := a.deps.Methods.Resolve(plan, dev)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := a.deps.Tests.Run(plan, dev, method, req.Parameter, req.Value, time.Now())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if _, err := a.deps.Tests.GenerateReport(plan, dev, []*qc.Result{result}, time.Now()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.audit("console", "qc-test-run", plan.ID, dev.ID)
	writeJSON(w, http.StatusCreated, result)
}

func (a *API) handleRunSample(w http.ResponseWriter, r *http.Request) {
	var req sampleRunRequest
	if err := readBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	plan, ok := a.deps.QC.Get(req.PlanID)
	if !ok {
		writeError(w, http.StatusNotFound, "plan not found")
		return
	}
	dev, ok := a.deps.Devices.Get(req.DeviceID)
	if !ok {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	calibration, ok := a.deps.Calibrations.Get(plan.ID)
	if !ok {
		writeError(w, http.StatusNotFound, "calibration not found")
		return
	}
	method, err := a.deps.Methods.Resolve(plan, dev)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	result, err := a.deps.Tests.RunSample(plan, dev, calibration, method, req.Value, time.Now())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.audit("console", "qc-sample-run", plan.ID, dev.ID)
	writeJSON(w, http.StatusCreated, result)
}

func (a *API) handleListReports(w http.ResponseWriter, r *http.Request) {
	planID := r.URL.Query().Get("plan_id")
	if planID == "" {
		writeError(w, http.StatusBadRequest, "plan_id required")
		return
	}
	names, err := a.deps.Reports.List(planID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	paths := make([]string, 0, len(names))
	for _, name := range names {
		paths = append(paths, a.deps.Reports.Path(planID, name))
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"names": names, "paths": paths})
}
