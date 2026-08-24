package device

import (
	"errors"
	"strings"

	"medops/internal/store"
)

// Registry is the file-backed device registry.
type Registry struct {
	fs      *store.FileStore
	devices map[string]*Device
	order   []string
}

func NewRegistry(fs *store.FileStore) *Registry {
	return &Registry{fs: fs, devices: map[string]*Device{}}
}

func (r *Registry) Register(dev *Device) error {
	if _, exists := r.devices[dev.ID]; exists {
		return errors.New("device: already registered")
	}
	r.devices[dev.ID] = dev
	r.order = append(r.order, dev.ID)
	return r.Save(dev)
}

func (r *Registry) Save(dev *Device) error {
	return r.fs.WriteJSON("devices/"+dev.ID+".json", dev)
}

func (r *Registry) Get(id string) (*Device, bool) {
	dev, ok := r.devices[id]
	return dev, ok
}

func (r *Registry) All() []*Device {
	devices := make([]*Device, 0, len(r.order))
	for _, id := range r.order {
		if dev, ok := r.devices[id]; ok {
			devices = append(devices, dev)
		}
	}
	return devices
}

func (r *Registry) ByWard(wardID string) []*Device {
	var devices []*Device
	for _, dev := range r.All() {
		if dev.WardID == wardID {
			devices = append(devices, dev)
		}
	}
	return devices
}

func (r *Registry) LoadAll() error {
	names, err := r.fs.List("devices")
	if err != nil {
		return err
	}
	for _, name := range names {
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		id := strings.TrimSuffix(name, ".json")
		var dev Device
		if err := r.fs.ReadJSON("devices/"+name, &dev); err != nil {
			return err
		}
		if _, exists := r.devices[id]; !exists {
			r.devices[id] = &dev
			r.order = append(r.order, id)
		}
	}
	return nil
}
