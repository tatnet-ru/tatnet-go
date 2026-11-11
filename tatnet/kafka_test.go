package tatnet

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestKafkaResumePreservesHourlyBillingQuote(t *testing.T) {
	project := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	cluster := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/projects/"+project.String()+"/kafka-clusters/"+cluster.String()+"/resume" || r.Header.Get("Authorization") != "Bearer tn_live_test" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"hourly_cost":3.44773,"management_hourly_cost":0.71343,"monthly_cost":2516.8429}`))
	}))
	defer srv.Close()
	c, err := NewClientWithResponses(srv.URL, WithAPIKey("tn_live_test"))
	if err != nil {
		t.Fatal(err)
	}
	response, err := c.KafkaResumeClusterWithResponse(context.Background(), project, cluster)
	if err != nil || response.JSON200 == nil {
		t.Fatalf("resume: %v %+v", err, response)
	}
	quote := response.JSON200
	if quote.HourlyCost == nil || math.Abs(float64(*quote.HourlyCost)-3.44773) > 0.000001 || quote.ManagementHourlyCost == nil || quote.MonthlyCost == nil {
		t.Fatalf("billing quote lost: %+v", quote)
	}
}

func TestKafkaMonitoringAndResetPreserveScopeAndTimestampPrecision(t *testing.T) {
	project := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	cluster := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	node := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	base := "/projects/" + project.String() + "/kafka-clusters/" + cluster.String()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tn_live_test" {
			t.Error("authorization lost")
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case base + "/nodes/" + node.String() + "/metrics":
			if r.Method != "GET" || r.URL.Query().Get("metric") != "cpu" {
				t.Error("metric selection lost")
			}
			_, _ = w.Write([]byte(`{"metric":"cpu","series":[{"name":"cpu","data":[[1794424231.125,42.5]]}]}`))
		case base + "/password":
			if r.Method != "POST" {
				t.Error("wrong reset method")
			}
			_, _ = w.Write([]byte(`{"password":"one-time","applied":false}`))
		default:
			t.Fatalf("wrong scope: %s", r.URL)
		}
	}))
	defer srv.Close()
	c, err := NewClientWithResponses(srv.URL, WithAPIKey("tn_live_test"))
	if err != nil {
		t.Fatal(err)
	}
	metrics, err := c.KafkaNodeMetricsWithResponse(context.Background(), project, cluster, node, &KafkaNodeMetricsParams{Metric: Cpu})
	if err != nil || metrics.JSON200 == nil {
		t.Fatalf("metrics: %v %+v", err, metrics)
	}
	if got := metrics.JSON200.Series[0].Data[0][0]; got != 1794424231.125 {
		t.Fatalf("timestamp lost precision: %f", got)
	}
	reset, err := c.KafkaResetPasswordWithResponse(context.Background(), project, cluster)
	if err != nil || reset.JSON200 == nil || reset.JSON200.Password != "one-time" || reset.JSON200.Applied == nil || *reset.JSON200.Applied {
		t.Fatalf("reset: %v %+v", err, reset)
	}
}
