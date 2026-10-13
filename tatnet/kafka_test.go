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
