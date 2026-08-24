package console

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"medops/internal/alert"
	"medops/internal/audit"
	"medops/internal/quota"
)

func (a *API) handleListAlerts(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("device_id")
	if deviceID != "" {
		writeJSON(w, http.StatusOK, a.deps.Alerts.History(deviceID))
		return
	}
	if r.URL.Query().Get("open") == "true" {
		writeJSON(w, http.StatusOK, a.deps.Alerts.Open())
		return
	}
	writeJSON(w, http.StatusOK, a.deps.Alerts.All())
}

func (a *API) handleAckAlert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := a.deps.Alerts.Ack(id, time.Now()); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	a.audit("console", "alert-ack", id, "")
	writeJSON(w, http.StatusOK, map[string]string{"acked": id})
}

func (a *API) handleQuota(w http.ResponseWriter, r *http.Request) {
	scope := r.URL.Query().Get("scope")
	kind := quota.Kind(r.URL.Query().Get("kind"))
	if scope == "" || kind == "" {
		writeError(w, http.StatusBadRequest, "scope and kind required")
		return
	}
	snapshot := a.deps.Quota.Snapshot(scope, kind)
	entries, err := a.deps.Quota.Ledger(scope, kind)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"quota":   snapshot,
		"entries": entries,
	})
}

func (a *API) handleNamespaces(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code != "" {
		ward, ok := a.deps.Namespaces.WardByCode(code)
		if !ok {
			writeError(w, http.StatusNotFound, "ward not found")
			return
		}
		writeJSON(w, http.StatusOK, ward)
		return
	}
	writeJSON(w, http.StatusOK, a.deps.Namespaces.Snapshot())
}

type alertEvaluateRequest struct {
	DeviceID string       `json:"device_id"`
	Kind     string       `json:"kind"`
	Value    float64      `json:"value"`
	Target   string       `json:"target"`
	Rules    []alert.Rule `json:"rules"`
}

func (a *API) handleAlertEvaluate(w http.ResponseWriter, r *http.Request) {
	var req alertEvaluateRequest
	if err := readBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	dev, ok := a.deps.Devices.Get(req.DeviceID)
	if !ok {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	raised, err := a.deps.Alerts.Evaluate(dev, req.Kind, req.Value, req.Rules, time.Now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	notifications := make([]alert.Notification, 0, len(raised))
	for _, raisedAlert := range raised {
		if notification, err := a.deps.Alerts.Notify(raisedAlert, req.Target, time.Now()); err == nil {
			notifications = append(notifications, notification)
		}
	}
	a.audit("console", "alert-evaluate", dev.ID, req.Kind)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"raised":        raised,
		"notifications": notifications,
	})
}

func (a *API) handleAudit(w http.ResponseWriter, r *http.Request) {
	filter := audit.Filter{
		Actor:  r.URL.Query().Get("actor"),
		Action: r.URL.Query().Get("action"),
	}
	records, err := a.deps.Audit.Filter(filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, records)
}
