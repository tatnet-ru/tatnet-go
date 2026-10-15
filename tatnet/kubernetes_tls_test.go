package tatnet

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestKubernetesTLSSettingsRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" || r.URL.Path != "/v1/projects/p/kubernetes-clusters/c/settings" {
			t.Errorf("request %s %s", r.Method, r.URL.Path)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body["tls_min_version"] != "VersionTLS13" {
			t.Errorf("body %v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"c","tls_min_version":"VersionTLS13"}`))
	}))
	defer srv.Close()
	client, err := NewClientWithResponses(srv.URL + "/v1")
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.KubernetesUpdateClusterSettingsWithResponse(context.Background(), "p", "c", KubernetesUpdateClusterSettingsJSONRequestBody{TlsMinVersion: UpdateK8sClusterSettingsRequestTlsMinVersionVersionTLS13})
	if err != nil {
		t.Fatal(err)
	}
	if response.JSON200 == nil || response.JSON200.TlsMinVersion == nil || *response.JSON200.TlsMinVersion != V1K8sClusterTlsMinVersionVersionTLS13 {
		t.Fatalf("response %+v", response)
	}
}
