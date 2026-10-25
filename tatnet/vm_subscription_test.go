package tatnet

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVMSubscriptionDesiredStateRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/projects/project/vms/vm/subscription" || r.Header.Get("Authorization") != "Bearer test" {
			t.Errorf("wrong request: %s", r.URL)
		}
		if r.Method == http.MethodPatch {
			var body V1VMSubscriptionUpdate
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body.PeriodType == nil || *body.PeriodType != N1m || body.AutoRenew == nil || *body.AutoRenew {
				t.Errorf("wrong billing settings: %+v", body)
			}
		} else if r.Method != http.MethodGet {
			t.Errorf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"resource_id":"vm","period_type":"1m","amount_per_period":"3474.00","currency":"RUB","status":"active","auto_renew":false}`))
	}))
	defer srv.Close()
	c, err := NewClientWithResponses(srv.URL+"/v1", WithAPIKey("test"))
	if err != nil {
		t.Fatal(err)
	}
	period, renew := N1m, false
	updated, err := c.VmsPatchVmSubscriptionWithResponse(context.Background(), "project", "vm", V1VMSubscriptionUpdate{PeriodType: &period, AutoRenew: &renew})
	if err != nil {
		t.Fatal(err)
	}
	if updated.JSON200 == nil || updated.JSON200.AmountPerPeriod != "3474.00" {
		t.Fatalf("unexpected response: %+v", updated)
	}
	read, err := c.VmsGetVmSubscriptionWithResponse(context.Background(), "project", "vm")
	if err != nil || read.JSON200 == nil || read.JSON200.PeriodType != "1m" {
		t.Fatalf("read: %+v, %v", read, err)
	}
}
