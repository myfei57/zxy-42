package device

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"medops/internal/heart"
)

var (
	ErrSerialExists          = errors.New("device: serial already registered")
	ErrWardRequired          = errors.New("device: ward is required")
	ErrUnknownStatus         = errors.New("device: unknown status")
	ErrUnknownClassification = errors.New("device: unknown classification")
	ErrLockoutActive         = errors.New("device: maintenance lockout is active")
	ErrInvalidTransition     = errors.New("device: invalid status transition")
	ErrStaleBeat             = errors.New("device: stale heartbeat sequence")
	ErrDeviceOffline         = errors.New("device: heartbeat outside online window")
	ErrUnknownRisk           = errors.New("device: unknown risk level")
)

type Status string

const (
	StatusRegistered  Status = "registered"
	StatusInUse       Status = "in-use"
	StatusMaintenance Status = "maintenance"
	StatusRetired     Status = "retired"
)

type Classification string

const (
	ClassificationICU     Classification = "icu"
	ClassificationGeneral Classification = "general"
)

// WardMove records a device transfer between wards.
type WardMove struct {
	FromWard string    `json:"from_ward"`
	ToWard   string    `json:"to_ward"`
	MovedAt  time.Time `json:"moved_at"`
}

// Device is the aggregate root of the medical device lifecycle: registration,
// firmware, risk, classification, ward placement and maintenance state.
type Device struct {
	ID              string         `json:"id"`
	Serial          string         `json:"serial"`
	Model           string         `json:"model"`
	Name            string         `json:"name"`
	HospitalID      string         `json:"hospital_id"`
	WardID          string         `json:"ward_id"`
	Status          Status         `json:"status"`
	Classification  Classification `json:"classification"`
	Firmware        *Firmware      `json:"firmware"`
	Risk            Risk           `json:"risk"`
	Lockout         *Lockout       `json:"lockout"`
	Clock           *Clock         `json:"clock"`
	RegisteredAt    time.Time      `json:"registered_at"`
	StatusChangedAt time.Time      `json:"status_changed_at"`
	LastHeartbeat   time.Time      `json:"last_heartbeat"`
	LastBeatSeq     int64          `json:"last_beat_seq"`
	WardHistory     []WardMove     `json:"ward_history"`
	ReclassifiedAt  time.Time      `json:"reclassified_at"`
}

func NewDevice(serial, model, name, hospitalID, wardID string, classification Classification, at time.Time) (*Device, error) {
	if serial == "" || model == "" {
		return nil, errors.New("device: serial and model are required")
	}
	if wardID == "" {
		return nil, ErrWardRequired
	}
	if classification != ClassificationICU && classification != ClassificationGeneral {
		return nil, ErrUnknownClassification
	}
	return &Device{
		ID:              uuid.NewString(),
		Serial:          serial,
		Model:           model,
		Name:            name,
		HospitalID:      hospitalID,
		WardID:          wardID,
		Status:          StatusRegistered,
		Classification:  classification,
		Firmware:        NewFirmware("v1.0.0", at),
		Risk:            RiskLow,
		Lockout:         NewLockout(),
		Clock:           NewClock(at),
		RegisteredAt:    at,
		StatusChangedAt: at,
	}, nil
}

func (d *Device) DeviceID() string { return d.ID }

func (d *Device) Ward() string { return d.WardID }

func (d *Device) CurrentClassification() Classification { return d.Classification }

func (d *Device) LastSeq() int64 { return d.LastBeatSeq }

func (d *Device) LiveClock() heart.Clock { return d.Clock }
