package tatnet

import (
	"context"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNATDisableTypedRetentionAndAllocationGuard(t *testing.T) {
	id := uuid.MustParse("8c10620b-6af6-4ae2-b4bd-e69eef8fa0e8")
	for _, retain := range []bool{false, true} {
		t.Run(map[bool]string{false: "release", true: "retain"}[retain], func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "DELETE" || r.URL.Path != "/v1/vpcs/vpc/nat-gateway" {
					t.Errorf("unexpected request %s %s", r.Method, r.URL)
				}
				if r.URL.Query().Get("expected_fip_id") != id.String() {
					t.Error("allocation guard missing")
				}
				want := "false"
				if retain {
					want = "true"
				}
				if r.URL.Query().Get("retain_ip") != want {
					t.Errorf("retention=%q, want %q", r.URL.Query().Get("retain_ip"), want)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write([]byte(`{"enabled":false,"status":"detaching"}`))
			}))
			defer srv.Close()
			c, err := NewClientWithResponses(srv.URL + "/v1")
			if err != nil {
				t.Fatal(err)
			}
			res, err := c.NetworkingDisableNatGatewayWithResponse(context.Background(), "vpc", &NetworkingDisableNatGatewayParams{ExpectedFipId: &id, RetainIp: &retain})
			if err != nil || res.StatusCode() != http.StatusAccepted {
				t.Fatalf("disable: %v %v", res, err)
			}
		})
	}
}
