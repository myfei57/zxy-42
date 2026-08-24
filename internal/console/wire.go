package console

import (
	"time"

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

// WireDeps builds every service over a shared file store and namespace tree.
func WireDeps(fs *store.FileStore, namespaces *ns.WardRegistry) (Deps, error) {
	thresholds := qc.NewThresholdService([]qc.ThresholdTable{
		{
			FirmwareVersion: "v1.0.0",
			Thresholds: []qc.Threshold{
				{Parameter: "spo2", Min: 90, Max: 100},
				{Parameter: "pr", Min: 50, Max: 120},
			},
		},
		{
			FirmwareVersion: "v1.1.0",
			Thresholds: []qc.Threshold{
				{Parameter: "spo2", Min: 92, Max: 100},
				{Parameter: "pr", Min: 45, Max: 115},
			},
		},
	})
	methods := qc.NewMethodService([]qc.Method{
		{FirmwareVersion: "v1.0.0", Name: "legacy-spo2", Parameters: []string{"spo2"}},
		{FirmwareVersion: "v1.1.0", Name: "continuous-spo2", Parameters: []string{"spo2", "pr"}},
	})

	reagents := test.NewReagentRegistry()
	reagent := test.NewReagent("质控液 A", "LOT-A", time.Now().Add(30*24*time.Hour), time.Now())
	if err := reagents.Add(reagent); err != nil {
		return Deps{}, err
	}

	reports := store.NewReportStore(fs)
	batches := store.NewBatchStore(fs)
	checkpoints := store.NewCheckpointStore(fs)
	calibrations := qc.NewCalibrationRegistry(fs)

	alerts := alert.NewService(fs)
	heartService := heart.NewService(heart.DefaultPolicy().Window(), checkpoints)
	devices := device.NewRegistry(fs)
	schedule := qc.NewScheduleService(fs, thresholds, methods, calibrations, batches)
	trendService := trend.NewService(
		fs,
		trend.Range{Classification: device.ClassificationICU, Min: 94, Max: 100},
		trend.Range{Classification: device.ClassificationGeneral, Min: 90, Max: 100},
		alerts,
	)
	tests := test.NewService(reagents, calibrations, thresholds, reports, trendService)
	plans := plan.NewService(fs)
	quotas := quota.NewService(fs, quota.DefaultPolicy())
	audits := audit.NewRecorder(fs)

	if err := devices.LoadAll(); err != nil {
		return Deps{}, err
	}
	if err := schedule.LoadAll(); err != nil {
		return Deps{}, err
	}
	if err := calibrations.LoadAll(); err != nil {
		return Deps{}, err
	}
	if err := plans.LoadAll(); err != nil {
		return Deps{}, err
	}
	if err := alerts.LoadAll(); err != nil {
		return Deps{}, err
	}

	return Deps{
		FS:           fs,
		Namespaces:   namespaces,
		Devices:      devices,
		Heart:        heartService,
		HeartPolicy:  heart.DefaultPolicy(),
		QC:           schedule,
		Thresholds:   thresholds,
		Methods:      methods,
		Calibrations: calibrations,
		Tests:        tests,
		Reagents:     reagents,
		Trend:        trendService,
		Plan:         plans,
		Alerts:       alerts,
		Quota:        quotas,
		Audit:        audits,
		Reports:      reports,
		Batches:      batches,
	}, nil
}
