package tatnet

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithAPIKeySetsBearer(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"account_id":"a","key_id":"k","policy":[]}`))
	}))
	defer srv.Close()

	c, err := NewClientWithResponses(srv.URL, WithAPIKey("tn_live_test"))
	if err != nil {
		t.Fatalf("клиент: %v", err)
	}
	if _, err := c.AccountWhoamiWithResponse(context.Background()); err != nil {
		t.Fatalf("запрос: %v", err)
	}
	if got != "Bearer tn_live_test" {
		t.Errorf("заголовок = %q, want %q", got, "Bearer tn_live_test")
	}
}

// Пустой ключ обязан отвергаться на нашей стороне: иначе уходит «Bearer » и
// сервер отвечает 401, из которого не видно, что ключа просто нет.
func TestEmptyKeyIsRejectedBeforeTheRequest(t *testing.T) {
	if _, err := New("   "); err == nil {
		t.Fatal("пустой ключ принят конструктором")
	}
	c, err := NewClientWithResponses("http://127.0.0.1:1", WithAPIKey(""))
	if err != nil {
		t.Fatalf("клиент: %v", err)
	}
	if _, err := c.AccountWhoamiWithResponse(context.Background()); err == nil {
		t.Fatal("запрос с пустым ключом ушёл в сеть")
	}
}
