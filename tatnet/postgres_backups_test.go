package tatnet

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostgresCreateBackupParsesSingleBackup(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != "POST" || r.URL.Path != "/v1/projects/project/pg-clusters/cluster/backups" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("missing API authorization")
		}
		var body CreatePgBackupRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Label != "restore check" {
			t.Errorf("backup label lost: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"backup-uuid","kind":"manual","backup_type":"full","label":"restore check","status":"requested","size_bytes":null,"pgbackrest_label":null,"created_at":"2026-10-11T07:35:19Z","completed_at":null,"expires_at":null,"error":null}`))
	}))
	defer server.Close()
	client, err := NewClientWithResponses(server.URL+"/v1", WithAPIKey("test-key"))
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.PostgresCreateBackupWithResponse(context.Background(), "project", "cluster", CreatePgBackupRequest{Label: "restore check"})
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode() != http.StatusCreated || response.JSON201 == nil {
		t.Fatal("missing typed 201 response")
	}
	backup := response.JSON201
	if backup.Id != "backup-uuid" || backup.Kind != "manual" || backup.BackupType != "full" || backup.Status != "requested" || backup.Label != "restore check" {
		t.Errorf("single-backup response lost: %+v", backup)
	}
	if calls != 1 {
		t.Errorf("backup creation repeated %d times", calls)
	}
}
