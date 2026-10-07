package tatnet

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProjectImageCataloguePreservesRegionBuild(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/projects/project-a/images" || r.URL.Query().Get("cluster_id") != "region-a" {
			t.Errorf("unexpected request: %s", r.URL)
		}
		if r.Header.Get("Authorization") != "Bearer tn_live_test" {
			t.Error("missing API key")
		}
		if r.URL.Query().Has("account_id") {
			t.Error("account must come from the API key")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"images":[{"id":"permitted-build","name":"RED OS 7.3","filename":"redos.qcow2"}],"families":[{"slug":"redos","name":"RED OS","kind":"os","os_family":"redos","versions":[{"version":"7.3","image_id":"newest-permitted","by_cluster":{"region-a":"permitted-build"}}]}]}`))
	}))
	defer srv.Close()
	c, err := NewClientWithResponses(srv.URL+"/v1", WithAPIKey("tn_live_test"))
	if err != nil {
		t.Fatal(err)
	}
	region := "region-a"
	response, err := c.VmsListImagesWithResponse(context.Background(), "project-a", &VmsListImagesParams{ClusterId: &region})
	if err != nil {
		t.Fatal(err)
	}
	if response.JSON200 == nil || len(response.JSON200.Images) != 1 || response.JSON200.Families == nil {
		t.Fatalf("invalid response: %+v", response)
	}
	version := (*(*response.JSON200.Families)[0].Versions)[0]
	if version.ByCluster == nil || (*version.ByCluster)[region] != "permitted-build" {
		t.Fatalf("lost regional build: %+v", version)
	}
}

func TestProjectImageCatalogueDoesNotTurnForbiddenIntoEmptyCatalogue(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"detail":"Access denied to project"}`))
	}))
	defer srv.Close()
	c, err := NewClientWithResponses(srv.URL+"/v1", WithAPIKey("tn_live_test"))
	if err != nil {
		t.Fatal(err)
	}
	response, err := c.VmsListImagesWithResponse(context.Background(), "other-project", nil)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode() != http.StatusForbidden || response.JSON200 != nil {
		t.Fatalf("denial was lost: %+v", response)
	}
}
