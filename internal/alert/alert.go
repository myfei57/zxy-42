package alert

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"medops/internal/device"
	"medops/internal/store"
)

var ErrAlertNotFound = errors.New("alert: alert not found")

// Alert is one raised operational alert.
type Alert struct {
	ID       string    `json:"id"`
	DeviceID string    `json:"device_id"`
	Kind     string    `json:"kind"`
	Severity string    `json:"severity"`
	Message  string    `json:"message"`
	RaisedAt time.Time `json:"raised_at"`
	Acked    bool      `json:"acked"`
	AckedAt  time.Time `json:"acked_at"`
}

type Service struct {
	fs     *store.FileStore
	alerts map[string]*Alert
	order  []string
}

func NewService(fs *store.FileStore) *Service {
	return &Service{fs: fs, alerts: map[string]*Alert{}}
}

func (s *Service) Raise(dev *device.Device, kind, severity, message string, at time.Time) (*Alert, error) {
	a := &Alert{
		ID:       uuid.NewString(),
		DeviceID: dev.ID,
		Kind:     kind,
		Severity: severity,
		Message:  message,
		RaisedAt: at,
	}
	s.alerts[a.ID] = a
	s.order = append(s.order, a.ID)
	if err := s.fs.WriteJSON("alerts/"+a.ID+".json", a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) Ack(id string, at time.Time) error {
	a, ok := s.alerts[id]
	if !ok {
		return ErrAlertNotFound
	}
	if a.Acked {
		return nil
	}
	a.Acked = true
	a.AckedAt = at
	return s.fs.WriteJSON("alerts/"+a.ID+".json", a)
}

func (s *Service) Get(id string) (*Alert, bool) {
	a, ok := s.alerts[id]
	return a, ok
}

func (s *Service) Open() []*Alert {
	var open []*Alert
	for _, a := range s.All() {
		if !a.Acked {
			open = append(open, a)
		}
	}
	return open
}

func (s *Service) All() []*Alert {
	alerts := make([]*Alert, 0, len(s.order))
	for _, id := range s.order {
		if a, ok := s.alerts[id]; ok {
			alerts = append(alerts, a)
		}
	}
	return alerts
}

func (s *Service) LoadAll() error {
	names, err := s.fs.List("alerts")
	if err != nil {
		return err
	}
	for _, name := range names {
		if len(name) < 5 || name[len(name)-5:] != ".json" {
			continue
		}
		id := name[:len(name)-5]
		var a Alert
		if err := s.fs.ReadJSON("alerts/"+name, &a); err != nil {
			return err
		}
		if _, exists := s.alerts[id]; !exists {
			s.alerts[id] = &a
			s.order = append(s.order, id)
		}
	}
	return nil
}
