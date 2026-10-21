package tatnet

import (
	"encoding/json"
	"testing"
)

func TestVMCreateHasStoragePolicyWithoutPrepaidAllocation(t *testing.T) {
	enabled := true
	retention := 14
	body := V1VMCreate{Name: "vm", Hostname: "vm", BackupPolicy: &VMBackupPolicyRequest{Enabled: &enabled, RetentionDays: &retention}}
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	var data map[string]any
	if err = json.Unmarshal(encoded, &data); err != nil {
		t.Fatal(err)
	}
	if _, exists := data["backup_count"]; exists {
		t.Fatal("retired allocation in VM create")
	}
	policy, ok := data["backup_policy"].(map[string]any)
	if !ok || policy["enabled"] != true || policy["retention_days"] != float64(14) {
		t.Fatalf("missing storage policy: %s", encoded)
	}
}
