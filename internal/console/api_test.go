package console_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"medops/internal/console"
	"medops/internal/ns"
	"medops/internal/store"
)

func TestHealthz(t *testing.T) {
	fs, err := store.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	namespaces := ns.NewWardRegistry()
	if err := ns.Seed(namespaces); err != nil {
		t.Fatal(err)
	}
	deps, err := console.WireDeps(fs, namespaces)
	if err != nil {
		t.Fatal(err)
	}
	api := console.NewAPI(deps)
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()
	api.Router().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("healthz status %d", recorder.Code)
	}
}

func TestRegisterDeviceEndpoint(t *testing.T) {
	fs, err := store.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	namespaces := ns.NewWardRegistry()
	if err := ns.Seed(namespaces); err != nil {
		t.Fatal(err)
	}
	ward, ok := namespaces.WardByCode("W002")
	if !ok {
		t.Fatal("seeded ward missing")
	}
	deps, err := console.WireDeps(fs, namespaces)
	if err != nil {
		t.Fatal(err)
	}
	api := console.NewAPI(deps)
	body := map[string]interface{}{
		"serial":         "SN-100",
		"model":          "M-9",
		"name":           "监护仪",
		"hospital_id":    "any",
		"ward_id":        ward.ID,
		"classification": "general",
	}
	payload, _ := json.Marshal(body)
	request := httptest.NewRequest(http.MethodPost, "/api/devices", bytes.NewReader(payload))
	recorder := httptest.NewRecorder()
	api.Router().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("register status %d body=%s", recorder.Code, recorder.Body.String())
	}
	var registered map[string]interface{}
	if err := json.Unmarshal(recorder.Body.Bytes(), &registered); err != nil {
		t.Fatal(err)
	}
	if registered["serial"] != "SN-100" {
		t.Fatalf("registered serial mismatch: %v", registered["serial"])
	}
}
