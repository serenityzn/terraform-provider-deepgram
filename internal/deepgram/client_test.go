package deepgram_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/serenityzn/terraform-provider-deepgram/internal/deepgram"
)

// newTestServer starts an httptest server and returns it along with a client
// pointed at it.
func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *deepgram.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv, deepgram.NewClient("test-api-key", srv.URL)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// ---------------------------------------------------------------------------
// CreateKey
// ---------------------------------------------------------------------------

func TestCreateKey_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Token test-api-key" {
			t.Errorf("unexpected Authorization header: %s", r.Header.Get("Authorization"))
		}

		writeJSON(w, http.StatusOK, deepgram.CreateKeyResponse{
			APIKeyID: "key-123",
			Key:      "secret-abc",
			Comment:  "test key",
			Scopes:   []string{"usage:read"},
		})
	})

	resp, err := client.CreateKey(context.Background(), "proj-1", deepgram.CreateKeyRequest{
		Comment: "test key",
		Scopes:  []string{"usage:read"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.APIKeyID != "key-123" {
		t.Errorf("expected api_key_id=key-123, got %s", resp.APIKeyID)
	}
	if resp.Key != "secret-abc" {
		t.Errorf("expected key=secret-abc, got %s", resp.Key)
	}
}

func TestCreateKey_NonOKStatus(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "bad request"})
	})

	_, err := client.CreateKey(context.Background(), "proj-1", deepgram.CreateKeyRequest{
		Comment: "test",
		Scopes:  []string{"usage:read"},
	})
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
}

// ---------------------------------------------------------------------------
// GetKey
// ---------------------------------------------------------------------------

func TestGetKey_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"member": map[string]any{
				"member_id":  "member-1",
				"email":      "user@example.com",
				"first_name": "John",
				"last_name":  "Doe",
			},
			"api_key": map[string]any{
				"api_key_id": "key-123",
				"comment":    "test key",
				"scopes":     []string{"usage:read"},
			},
		})
	})

	resp, err := client.GetKey(context.Background(), "proj-1", "key-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.APIKey.APIKeyID != "key-123" {
		t.Errorf("expected api_key_id=key-123, got %s", resp.APIKey.APIKeyID)
	}
	if resp.Member.Email != "user@example.com" {
		t.Errorf("expected email=user@example.com, got %s", resp.Member.Email)
	}
}

func TestGetKey_NotFound(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	resp, err := client.GetKey(context.Background(), "proj-1", "missing-key")
	if err != nil {
		t.Fatalf("expected nil error for 404, got: %v", err)
	}
	if resp != nil {
		t.Errorf("expected nil response for 404, got: %+v", resp)
	}
}

func TestGetKey_ServerError(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "server error"})
	})

	_, err := client.GetKey(context.Background(), "proj-1", "key-123")
	if err == nil {
		t.Fatal("expected error for 500 status, got nil")
	}
}

// ---------------------------------------------------------------------------
// ListKeys
// ---------------------------------------------------------------------------

func TestListKeys_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"api_keys": []map[string]any{
				{
					"member": map[string]any{"member_id": "m1", "email": "a@example.com"},
					"api_key": map[string]any{
						"api_key_id": "key-1",
						"comment":    "first",
						"scopes":     []string{"usage:read"},
						"created":    "2026-01-01T00:00:00Z",
					},
				},
				{
					"member": map[string]any{"member_id": "m2", "email": "b@example.com"},
					"api_key": map[string]any{
						"api_key_id": "key-2",
						"comment":    "second",
						"scopes":     []string{"keys:write"},
						"created":    "2026-02-01T00:00:00Z",
					},
				},
			},
		})
	})

	resp, err := client.ListKeys(context.Background(), "proj-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.APIKeys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(resp.APIKeys))
	}
	if resp.APIKeys[0].APIKey.APIKeyID != "key-1" {
		t.Errorf("expected first key id=key-1, got %s", resp.APIKeys[0].APIKey.APIKeyID)
	}
}

func TestListKeys_Empty(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"api_keys": []any{}})
	})

	resp, err := client.ListKeys(context.Background(), "proj-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.APIKeys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(resp.APIKeys))
	}
}

// ---------------------------------------------------------------------------
// DeleteKey
// ---------------------------------------------------------------------------

func TestDeleteKey_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "Success"})
	})

	err := client.DeleteKey(context.Background(), "proj-1", "key-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteKey_NonOKStatus(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusForbidden, map[string]string{"message": "forbidden"})
	})

	err := client.DeleteKey(context.Background(), "proj-1", "key-123")
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
}

// ---------------------------------------------------------------------------
// Authorization header
// ---------------------------------------------------------------------------

func TestClient_AuthorizationHeader(t *testing.T) {
	var gotAuth string
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		writeJSON(w, http.StatusOK, map[string]any{"api_keys": []any{}})
	})

	_, _ = client.ListKeys(context.Background(), "proj-1")

	if gotAuth != "Token test-api-key" {
		t.Errorf("expected Authorization header 'Token test-api-key', got %q", gotAuth)
	}
}
