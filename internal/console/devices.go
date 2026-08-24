package console

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"medops/internal/device"
	"medops/internal/heart"
	"medops/internal/quota"
)

type registerDeviceRequest struct {
	Serial         string `json:"serial"`
	Model          string `json:"model"`
	Name           string `json:"name"`
	HospitalID     string `json:"hospital_id"`
	WardID         string `json:"ward_id"`
	Classification string `json:"classification"`
}

type beatRequest struct {
	Seq    int64  `json:"seq"`
	SentAt string `json:"sent_at"`
}

type wardMoveRequest struct {
	WardID string `json:"ward_id"`
}

type firmwareUpgradeRequest struct {
	Version string `json:"version"`
}

type reclassifyRequest struct {
	Classification string `json:"classification"`
}

type maintenanceBeginRequest struct {
	Reason string `json:"reason"`
}

func (a *API) handleRegisterDevice(w http.ResponseWriter, r *http.Request) {
	var req registerDeviceRequest
	if err := readBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	dev, err := device.Register(a.deps.Devices, a.deps.Namespaces, device.RegisterParams{
		Serial:         req.Serial,
		Model:          req.Model,
		Name:           req.Name,
		HospitalID:     req.HospitalID,
		WardID:         req.WardID,
		Classification: device.Classification(req.Classification),
	}, time.Now())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	a.audit("console", "device-register", dev.ID, dev.Serial)
	writeJSON(w, http.StatusCreated, dev)
}

func (a *API) handleListDevices(w http.ResponseWriter, r *http.Request) {
	ward := r.URL.Query().Get("ward")
	if ward != "" {
		writeJSON(w, http.StatusOK, a.deps.Devices.ByWard(ward))
		return
	}
	writeJSON(w, http.StatusOK, a.deps.Devices.All())
}

func (a *API) handleGetDevice(w http.ResponseWriter, r *http.Request) {
	dev, ok := a.deps.Devices.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	checkpoint, _ := a.deps.Heart.LastCheckpoint(dev.ID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"device":     dev,
		"elapsed":    dev.Clock.Elapsed().String(),
		"firmware":   dev.Firmware.Upgrades(),
		"checkpoint": checkpoint,
		"ward_moves": dev.WardHistory,
	})
}

func (a *API) handleDeviceBeat(w http.ResponseWriter, r *http.Request) {
	dev, ok := a.deps.Devices.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	var req beatRequest
	if err := readBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	sentAt, err := time.Parse(time.RFC3339, req.SentAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "sent_at must be RFC3339")
		return
	}
	now := time.Now()
	if err := a.deps.Quota.Consume("device:"+dev.ID, quota.KindHeartbeat, now); err != nil {
		writeError(w, http.StatusTooManyRequests, err.Error())
		return
	}
	verdict, err := a.deps.Heart.Record(dev, heart.NewBeat(dev.ID, req.Seq, sentAt, now))
	if err != nil {
		a.deps.Quota.Restore("device:"+dev.ID, quota.KindHeartbeat, now)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.deps.Devices.Save(dev); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, verdict)
}

func (a *API) handleWardMove(w http.ResponseWriter, r *http.Request) {
	dev, ok := a.deps.Devices.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	var req wardMoveRequest
	if err := readBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if _, ok := a.deps.Namespaces.Ward(req.WardID); !ok {
		writeError(w, http.StatusBadRequest, "ward not found")
		return
	}
	if err := dev.MoveWard(req.WardID, time.Now()); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.deps.Devices.Save(dev); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.audit("console", "device-ward-move", dev.ID, dev.WardID)
	writeJSON(w, http.StatusOK, dev)
}

func (a *API) handleFirmwareUpgrade(w http.ResponseWriter, r *http.Request) {
	dev, ok := a.deps.Devices.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	var req firmwareUpgradeRequest
	if err := readBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := dev.Firmware.Upgrade(req.Version, time.Now()); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.deps.Devices.Save(dev); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.audit("console", "device-firmware-upgrade", dev.ID, req.Version)
	writeJSON(w, http.StatusOK, dev)
}

func (a *API) handleReclassify(w http.ResponseWriter, r *http.Request) {
	dev, ok := a.deps.Devices.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	var req reclassifyRequest
	if err := readBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := dev.Reclassify(device.Classification(req.Classification), time.Now()); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.deps.Devices.Save(dev); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.audit("console", "device-reclassify", dev.ID, string(dev.Classification))
	writeJSON(w, http.StatusOK, dev)
}

func (a *API) handleMaintenanceBegin(w http.ResponseWriter, r *http.Request) {
	dev, ok := a.deps.Devices.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	var req maintenanceBeginRequest
	if err := readBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := dev.BeginMaintenance(req.Reason, time.Now()); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.deps.Devices.Save(dev); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.audit("console", "device-maintenance-begin", dev.ID, req.Reason)
	writeJSON(w, http.StatusOK, dev)
}

func (a *API) handleClockSync(w http.ResponseWriter, r *http.Request) {
	dev, ok := a.deps.Devices.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	now := time.Now()
	dev.Clock.Reanchor(now)
	if err := a.deps.Devices.Save(dev); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	a.audit("console", "device-clock-sync", dev.ID, now.Format(time.RFC3339))
	writeJSON(w, http.StatusOK, map[string]string{"clock": now.Format(time.RFC3339)})
}

func (a *API) handleRetireDevice(w http.ResponseWriter, r *http.Request) {
	dev, ok := a.deps.Devices.Get(chi.URLParam(r, "id"))
	if !ok {
		writeError(w, http.StatusNotFound, "device not found")
		return
	}
	if err := dev.Retire(time.Now()); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.deps.Devices.Save(dev); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_ = a.deps.FS.Remove("trend/" + dev.ID + ".json")
	a.audit("console", "device-retire", dev.ID, dev.Serial)
	writeJSON(w, http.StatusOK, dev)
}
