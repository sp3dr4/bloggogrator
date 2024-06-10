package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHealthHandler(t *testing.T) {
	t.Run("is ok", func(t *testing.T) {
		mockDbApi := new(MockedDbApi)
		testApi := apiConfig{DB: mockDbApi}
		req, err := http.NewRequest(http.MethodGet, "/v1/healthz", nil)
		require.NoError(t, err)
		rw := httptest.NewRecorder()

		testApi.handlerHealth(rw, req)

		require.Equal(t, http.StatusOK, rw.Code)
		expected := mustJSON(t, map[string]string{"status": "ok"})
		got := rw.Body.String()
		if expected != got {
			t.Errorf("want %s, got %s", expected, got)
		}
	})
}

func TestErrHandler(t *testing.T) {
	t.Run("err format", func(t *testing.T) {
		mockDbApi := new(MockedDbApi)
		testApi := apiConfig{DB: mockDbApi}
		req, _ := http.NewRequest(http.MethodGet, "/v1/err", nil)
		resp := httptest.NewRecorder()

		testApi.handlerErr(resp, req)

		if resp.Code != 500 {
			t.Errorf("want 500, got %d", resp.Code)
		}
		expected := mustJSON(t, map[string]string{"error": "Internal Server Error"})
		got := resp.Body.String()
		if expected != got {
			t.Errorf("want %s, got %s", expected, got)
		}
	})
}
