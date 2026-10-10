package tatnet

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRRsetWritesCarryExplicitTTLAndRevision(t *testing.T) {
	for _, op := range []string{"create", "replace", "delete"} {
		for _, inherit := range []bool{false, true} {
			t.Run(op+map[bool]string{false: "/explicit", true: "/inherited"}[inherit], func(t *testing.T) {
				revision := strings.Repeat("a", 64)
				calls := 0
				s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls++
					if r.Header.Get("Authorization") != "Bearer test-key" {
						t.Error("auth lost")
					}
					w.Header().Set("Content-Type", "application/json")
					if op == "delete" {
						if r.Method != "DELETE" || r.URL.Query().Get("expected_revision") != revision {
							t.Error("conditional delete lost")
						}
						w.WriteHeader(204)
						return
					}
					var body map[string]json.RawMessage
					if e := json.NewDecoder(r.Body).Decode(&body); e != nil {
						t.Fatal(e)
					}
					ttl, exists := body["ttl"]
					if !exists || string(ttl) != map[bool]string{false: "300", true: "null"}[inherit] {
						t.Errorf("explicit TTL/null lost: %s", ttl)
					}
					if op == "replace" {
						if r.Method != "PUT" || string(body["expected_revision"]) != `"`+revision+`"` {
							t.Error("revision lost")
						}
						w.WriteHeader(409)
						_, _ = w.Write([]byte(`{"detail":"dns_rrset_changed"}`))
						return
					}
					if r.Method != "POST" || string(body["records"]) != `["one","two"]` {
						t.Error("create body lost")
					}
					w.WriteHeader(201)
					_, _ = w.Write([]byte(`{"id":"rrset","zone_id":"zone","name":"test.example.com.","type":"TXT","ttl":null,"records":["one","two"],"managed":false,"revision":"` + revision + `"}`))
				}))
				defer s.Close()
				c, e := NewClientWithResponses(s.URL+"/v1", WithAPIKey("test-key"))
				if e != nil {
					t.Fatal(e)
				}
				var ttl *int
				if !inherit {
					v := 300
					ttl = &v
				}
				ctx := context.Background()
				switch op {
				case "create":
					r, e := c.DnsCreateDnsRrsetWithResponse(ctx, "zone", V1DNSRRsetCreate{Name: "test.example.com.", Type: "TXT", Records: []string{"one", "two"}, Ttl: ttl})
					if e != nil || r.StatusCode() != 201 || r.JSON201 == nil || r.JSON201.Revision != revision || r.JSON201.Ttl != nil {
						t.Fatal("create response lost", e)
					}
				case "replace":
					r, e := c.DnsReplaceDnsRrsetWithResponse(ctx, "zone", "rrset", V1DNSRRsetReplace{ExpectedRevision: revision, Ttl: ttl, Records: []string{"one"}})
					if e != nil || r.StatusCode() != 409 || r.JSON200 != nil {
						t.Fatal("conflict hidden", e)
					}
				case "delete":
					r, e := c.DnsDeleteDnsRrsetWithResponse(ctx, "zone", "rrset", &DnsDeleteDnsRrsetParams{ExpectedRevision: revision})
					if e != nil || r.StatusCode() != 204 {
						t.Fatal("delete failed", e)
					}
				}
				if calls != 1 {
					t.Fatal("unexpected retry", calls)
				}
			})
		}
	}
}
