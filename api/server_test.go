package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mustJson(t *testing.T, v interface{}) string {
	t.Helper()
	out, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

var api = apiConfig{}

func TestHealthHandler(t *testing.T) {
	t.Run("is ok", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "/v1/healthz", nil)
		response := httptest.NewRecorder()

		api.handlerHealth(response, request)

		if response.Code != 200 {
			t.Errorf("want 200, got %d", response.Code)
		}
		expected := mustJson(t, map[string]string{"status": "ok"})
		got := response.Body.String()
		if expected != got {
			t.Errorf("want %s, got %s", expected, got)
		}
	})
}

func TestErrHandler(t *testing.T) {
	t.Run("err format", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/v1/err", nil)
		resp := httptest.NewRecorder()

		api.handlerErr(resp, req)

		if resp.Code != 500 {
			t.Errorf("want 500, got %d", resp.Code)
		}
		expected := mustJson(t, map[string]string{"error": "Internal Server Error"})
		got := resp.Body.String()
		if expected != got {
			t.Errorf("want %s, got %s", expected, got)
		}
	})
}
