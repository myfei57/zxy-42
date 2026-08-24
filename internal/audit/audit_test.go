package audit_test

import (
	"testing"
	"time"

	"medops/internal/audit"
	"medops/internal/store"
)

func TestAuditAppendAndFilter(t *testing.T) {
	fs, err := store.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	recorder := audit.NewRecorder(fs)
	at := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
	if _, err := recorder.Append("console", "device-register", "dev-1", "SN-001", at); err != nil {
		t.Fatal(err)
	}
	if _, err := recorder.Append("console", "device-beat", "dev-1", "seq=5", at.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	records, err := recorder.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	filtered, err := recorder.Filter(audit.Filter{Action: "device-register"})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 || filtered[0].Target != "dev-1" {
		t.Fatalf("filter mismatch: %+v", filtered)
	}
}
