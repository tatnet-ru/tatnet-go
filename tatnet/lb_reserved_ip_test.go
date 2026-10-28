package tatnet

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReservedLBIPContract(t *testing.T) {
	const fid = "8c10620b-6af6-4ae2-b4bd-e69eef8fa0e8"
	response := `{"id":"lb","name":"web","node_count":1,"status":"active","vip":"185.152.80.101","floating_ip_id":"` + fid + `"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tn_live_test" {
			t.Error("missing API key")
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost:
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if body["floating_ip_id"] != fid {
				t.Errorf("reservation lost in request: %+v", body)
			}
			w.WriteHeader(201)
			_, _ = w.Write([]byte(response))
		case r.URL.Path == "/v1/projects/p/load-balancers":
			_, _ = w.Write([]byte(`{"data":[` + response + `],"count":1,"limit":50,"offset":0}`))
		default:
			_, _ = w.Write([]byte(response))
		}
	}))
	defer srv.Close()
	c, err := NewClientWithResponses(srv.URL+"/v1", WithAPIKey("tn_live_test"))
	if err != nil {
		t.Fatal(err)
	}
	id := fid
	created, err := c.LoadBalancersCreateLbWithResponse(context.Background(), "p", V1LBCreate{Name: "web", FloatingIpId: &id})
	if err != nil {
		t.Fatal(err)
	}
	if created.JSON201 == nil || created.JSON201.FloatingIpId == nil || *created.JSON201.FloatingIpId != fid {
		t.Fatalf("create lost reservation: %+v", created)
	}
	list, err := c.LoadBalancersListLbsWithResponse(context.Background(), "p", nil)
	if err != nil {
		t.Fatal(err)
	}
	if list.JSON200 == nil || len(list.JSON200.Data) != 1 || list.JSON200.Data[0].FloatingIpId == nil || *list.JSON200.Data[0].FloatingIpId != fid {
		t.Fatalf("list lost reservation: %+v", list)
	}
	detail, err := c.LoadBalancersGetLbWithResponse(context.Background(), "p", "lb")
	if err != nil {
		t.Fatal(err)
	}
	if detail.JSON200 == nil || detail.JSON200.FloatingIpId == nil || *detail.JSON200.FloatingIpId != fid {
		t.Fatalf("detail lost reservation: %+v", detail)
	}
}

func TestReservedLBIPOptionalAndFloatingIPAttachment(t *testing.T) {
	for _, body := range []string{`{"id":"lb"}`, `{"id":"lb","floating_ip_id":null}`} {
		var lb LoadBalancerResponse
		if err := json.Unmarshal([]byte(body), &lb); err != nil {
			t.Fatal(err)
		}
		if lb.FloatingIpId != nil {
			t.Fatalf("automatic VIP must not imply reservation: %+v", lb)
		}
	}
	var allocation V1FloatingIP
	if err := json.Unmarshal([]byte(`{"id":"fip","address":"185.152.80.101","cluster_id":"r","status":"attached","target_kind":"load_balancer","load_balancer_id":"lb","load_balancer_project_id":"p"}`), &allocation); err != nil {
		t.Fatal(err)
	}
	if allocation.LoadBalancerId == nil || *allocation.LoadBalancerId != "lb" || allocation.LoadBalancerProjectId == nil || *allocation.LoadBalancerProjectId != "p" {
		t.Fatalf("attachment lost: %+v", allocation)
	}
	body, err := json.Marshal(V1LBCreate{Name: "web"})
	if err != nil {
		t.Fatal(err)
	}
	var request map[string]any
	if err := json.Unmarshal(body, &request); err != nil {
		t.Fatal(err)
	}
	if _, present := request["floating_ip_id"]; present {
		t.Fatal("absent reservation must be omitted")
	}
}

func TestRetainCurrentLBIPContract(t *testing.T) {
	for _, status := range []int{200, 409, 503} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" || r.URL.Path != "/v1/projects/p/load-balancers/lb/retain-ip" || r.Header.Get("Authorization") != "Bearer tn_live_test" {
					t.Fatalf("unexpected request %s %s", r.Method, r.URL)
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if len(body) != 1 || body["expected_vip"] != "198.51.100.42" {
					t.Fatalf("wrong body %v", body)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"id":"lb","vip":"198.51.100.42","floating_ip_id":"allocation","ip_billing_included":true,"can_retain_current_ip":false}`))
			}))
			defer srv.Close()
			c, err := NewClientWithResponses(srv.URL+"/v1", WithAPIKey("tn_live_test"))
			if err != nil {
				t.Fatal(err)
			}
			res, err := c.LoadBalancersRetainLbIpWithResponse(context.Background(), "p", "lb", RetainLBIPRequest{ExpectedVip: "198.51.100.42"})
			if err != nil || res.StatusCode() != status {
				t.Fatalf("response %v %v", res, err)
			}
			if status == 200 && (res.JSON200 == nil || res.JSON200.FloatingIpId == nil || *res.JSON200.FloatingIpId != "allocation" || res.JSON200.IpBillingIncluded == nil || !*res.JSON200.IpBillingIncluded) {
				t.Fatalf("reservation lost: %+v", res)
			}
		})
	}
}
