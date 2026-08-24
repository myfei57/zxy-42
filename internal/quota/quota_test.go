package quota_test

import (
	"testing"
	"time"

	"medops/internal/quota"
	"medops/internal/store"
)

func TestQuotaConsumeRestore(t *testing.T) {
	fs, err := store.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	policy := quota.Policy{HeartbeatLimit: 3, ResultLimit: 2, Window: time.Hour}
	service := quota.NewService(fs, policy)
	at := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		if err := service.Consume("device:dev-1", quota.KindHeartbeat, at); err != nil {
			t.Fatalf("consume %d failed: %v", i, err)
		}
	}
	if err := service.Consume("device:dev-1", quota.KindHeartbeat, at); err == nil {
		t.Fatal("consume beyond limit accepted")
	}
	service.Restore("device:dev-1", quota.KindHeartbeat, at)
	if err := service.Consume("device:dev-1", quota.KindHeartbeat, at); err != nil {
		t.Fatalf("consume after restore failed: %v", err)
	}
	entries, err := service.Ledger("device:dev-1", quota.KindHeartbeat)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 5 {
		t.Fatalf("expected 5 ledger entries, got %d", len(entries))
	}
}
