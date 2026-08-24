package console

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"medops/internal/plan"
	"medops/internal/trend"
)

type adviseRequest struct {
	DeviceID  string `json:"device_id"`
	QCOverdue bool   `json:"qc_overdue"`
}

func (a *API) handleTrend(w http.ResponseWriter, r *http.Request) {
	dev, ok := a.deps.Devices.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	writeJSON(w, http.StatusOK, a.deps.Trend.History(dev))
}

func (a *API) handleTrendEvaluate(w http.ResponseWriter, r *http.Request) {
	dev, ok := a.deps.Devices.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	evaluation, err := a.deps.Trend.Evaluate(dev)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, evaluation)
}

func (a *API) handleAdvise(w http.ResponseWriter, r *http.Request) {
	var req adviseRequest
	if err := readBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	dev, ok := a.deps.Devices.Get(req.DeviceID)
	if !ok {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	var evaluation *trend.Evaluation
	if current, err := a.deps.Trend.Evaluate(dev); err == nil {
		evaluation = &current
	}
	due := req.QCOverdue
	for _, duePlan := range a.deps.QC.Due(time.Now()) {
		if duePlan.DeviceID == dev.ID {
			due = true
			break
		}
	}
	advice, err := a.deps.Plan.Advise(dev, evaluation, due, time.Now())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if advice == nil {
		writeJSON(w, http.StatusOK, map[string]bool{"recommended": false})
		return
	}
	a.audit("console", "maintenance-advise", advice.PlanID, dev.ID)
	writeJSON(w, http.StatusCreated, advice)
}

func (a *API) handleListMaintenancePlans(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.deps.Plan.All())
}

func (a *API) handleMaintenanceQueue(w http.ResponseWriter, r *http.Request) {
	items := plan.SortItems(plan.FromPlans(a.deps.Plan.All()))
	writeJSON(w, http.StatusOK, items)
}

func (a *API) handleBeginMaintenance(w http.ResponseWriter, r *http.Request) {
	plan, ok := a.deps.Plan.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "plan not found")
		return
	}
	dev, ok := a.deps.Devices.Get(plan.DeviceID)
	if !ok {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	if err := a.deps.Plan.Begin(plan, dev, plan.Reason, time.Now()); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.deps.Devices.Save(dev); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.audit("console", "maintenance-begin", plan.ID, dev.ID)
	writeJSON(w, http.StatusOK, plan)
}

func (a *API) handleCompleteMaintenance(w http.ResponseWriter, r *http.Request) {
	plan, ok := a.deps.Plan.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "plan not found")
		return
	}
	dev, ok := a.deps.Devices.Get(plan.DeviceID)
	if !ok {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	if err := a.deps.Plan.Complete(plan, dev, time.Now()); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.deps.Devices.Save(dev); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.audit("console", "maintenance-complete", plan.ID, dev.ID)
	writeJSON(w, http.StatusOK, plan)
}
