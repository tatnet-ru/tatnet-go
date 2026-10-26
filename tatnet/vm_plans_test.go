package tatnet

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestVMPlanCatalog(t *testing.T) {
	region := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	role := VmsListPlansParamsRoleK8s
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/v1/vms/plans" || r.URL.Query().Get("cluster_id") != region.String() || r.URL.Query().Get("role") != "k8s" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
		if r.Header.Get("Authorization") != "Bearer tn_live_test" {
			t.Error("missing key")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"plans":[{"id":"plan-4-8","name":"Standard CPU","vcpu":4,"mem":8192,"disk_size":40,"billing_plan_group_name":"Standard CPU","price_per_month":1200,"price_per_hour":null,"tags":["Standard CPU"],"suitable_for":["vm","k8s"]}]}`))
	}))
	defer srv.Close()
	c, err := NewClientWithResponses(srv.URL+"/v1", WithAPIKey("tn_live_test"))
	if err != nil {
		t.Fatal(err)
	}
	r, err := c.VmsListPlansWithResponse(context.Background(), &VmsListPlansParams{ClusterId: &region, Role: &role})
	if err != nil {
		t.Fatal(err)
	}
	if r.JSON200 == nil || len(r.JSON200.Plans) != 1 {
		t.Fatalf("invalid catalog: %+v", r)
	}
	p := r.JSON200.Plans[0]
	if p.Id != "plan-4-8" || p.Mem != 8192 || p.PricePerHour != nil || p.PricePerMonth == nil || *p.PricePerMonth != 1200 {
		t.Fatalf("lost plan UUID, MiB units or nullable prices: %+v", p)
	}
}
