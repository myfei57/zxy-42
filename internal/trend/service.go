package trend

import (
	"medops/internal/alert"
	"medops/internal/device"
	"medops/internal/store"
)

// Service owns the performance baselines and the classification ranges used
// for evaluation.
type Service struct {
	fs      *store.FileStore
	ranges  map[string]device.Classification
	icu     Range
	general Range
	alerts  *alert.Service
}

func NewService(fs *store.FileStore, icu, general Range, alerts *alert.Service) *Service {
	return &Service{
		fs:      fs,
		ranges:  map[string]device.Classification{},
		icu:     icu,
		general: general,
		alerts:  alerts,
	}
}
