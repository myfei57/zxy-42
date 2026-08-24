package store_test

import (
	"testing"
	"time"

	"medops/internal/store"
)

func TestFileStoreJSONRoundtrip(t *testing.T) {
	fs, err := store.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	type sample struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	want := sample{Name: "monitor", Count: 3}
	if err := fs.WriteJSON("a/b.json", want); err != nil {
		t.Fatal(err)
	}
	var got sample
	if err := fs.ReadJSON("a/b.json", &got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("roundtrip mismatch: got %+v want %+v", got, want)
	}
}

func TestExpiredBoundary(t *testing.T) {
	bound := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	if store.Expired(bound, bound.Add(-time.Hour)) {
		t.Fatal("moment before the bound must not be expired")
	}
	if !store.Expired(bound, bound) {
		t.Fatal("moment at the bound must count as expired")
	}
	if !store.Expired(bound, bound.Add(time.Hour)) {
		t.Fatal("moment after the bound must be expired")
	}
}

func TestCheckpointRoundtrip(t *testing.T) {
	fs, err := store.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	checkpoints := store.NewCheckpointStore(fs)
	at := time.Date(2026, 8, 1, 10, 30, 0, 0, time.UTC)
	if err := checkpoints.Write("dev-1", 42, at); err != nil {
		t.Fatal(err)
	}
	cp, ok := checkpoints.Read("dev-1")
	if !ok || cp.LastBeatSeq != 42 || !cp.LastBeatAt.Equal(at) {
		t.Fatalf("checkpoint mismatch: %+v ok=%v", cp, ok)
	}
}
