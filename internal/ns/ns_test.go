package ns_test

import (
	"testing"
	"time"

	"medops/internal/ns"
)

func TestWardRegistryAddAndFind(t *testing.T) {
	registry := ns.NewWardRegistry()
	at := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	hospital := ns.NewHospital("第一人民医院", "H001")
	if err := registry.AddHospital(hospital); err != nil {
		t.Fatal(err)
	}
	department := ns.NewDepartment(hospital.ID, "重症医学科", "D001")
	if err := registry.AddDepartment(department); err != nil {
		t.Fatal(err)
	}
	ward, err := ns.NewWard(department.ID, "ICU 病区", "W001", "icu", 16, at)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.AddWard(ward); err != nil {
		t.Fatal(err)
	}
	got, ok := registry.Ward(ward.ID)
	if !ok || got.Code != "W001" || got.Capacity != 16 {
		t.Fatalf("ward lookup mismatch: %+v ok=%v", got, ok)
	}
}

func TestWardRegistryRejectsDuplicateCode(t *testing.T) {
	registry := ns.NewWardRegistry()
	at := time.Now()
	hospital := ns.NewHospital("第二人民医院", "H002")
	_ = registry.AddHospital(hospital)
	department := ns.NewDepartment(hospital.ID, "心内科", "D002")
	_ = registry.AddDepartment(department)
	ward, err := ns.NewWard(department.ID, "CCU", "W003", "icu", 8, at)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.AddWard(ward); err != nil {
		t.Fatal(err)
	}
	duplicate, err := ns.NewWard(department.ID, "CCU 二号", "W003", "icu", 8, at)
	if err != nil {
		t.Fatal(err)
	}
	if err := registry.AddWard(duplicate); err == nil {
		t.Fatal("duplicate ward code accepted")
	}
}
