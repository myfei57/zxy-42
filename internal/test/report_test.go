package test_test

import (
	"testing"

	"medops/internal/store"
)

func TestReportWriteRead(t *testing.T) {
	fs, err := store.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	reports := store.NewReportStore(fs)
	payload := map[string]interface{}{"summary": "2/2 passed"}
	rel, err := reports.Write("plan-1", "dev-1", payload)
	if err != nil {
		t.Fatal(err)
	}
	if rel != "reports/plan-1/dev-1.json" {
		t.Fatalf("unexpected report path: %s", rel)
	}
	data, err := reports.Read("plan-1", "dev-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("empty report content")
	}
	names, err := reports.List("plan-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "dev-1" {
		t.Fatalf("report list mismatch: %v", names)
	}
}
