package tatnet

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAppBranchUpdateRoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" || r.URL.Path != "/apps/control" || r.Header.Get("Authorization") != "Bearer tn_live_test" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if len(body) != 1 || body["branch"] != "main" {
			t.Fatalf("unexpected patch: %v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"control","name":"inference-control","project_id":"project","status":"active","branch":"main"}`))
	}))
	defer srv.Close()
	c, err := NewClientWithResponses(srv.URL, WithAPIKey("tn_live_test"))
	if err != nil {
		t.Fatal(err)
	}
	branch := "main"
	r, err := c.AppsUpdateAppByIdWithResponse(context.Background(), "control", V1AppUpdate{Branch: &branch})
	if err != nil || r.JSON200 == nil || r.JSON200.Branch == nil || *r.JSON200.Branch != branch {
		t.Fatalf("branch update: %+v %v", r, err)
	}
}
