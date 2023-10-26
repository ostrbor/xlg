package xlg

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_sendSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	body := []byte(`{"key": "value"}`)
	headers := map[string]string{"Authorization": "Bearer Token"}
	err := send(srv.URL, body, headers)

	if err != nil {
		t.Errorf("expected no error, but got %v", err)
	}
}

func Test_sendErrorStatusCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	body := []byte(`{"key": "value"}`)
	headers := map[string]string{"Authorization": "Bearer Token"}
	err := send(srv.URL, body, headers)

	expectedErr := "expected 2xx, got 500"
	if err == nil || err.Error() != expectedErr {
		t.Errorf("expected error '%s', but got '%v'", expectedErr, err)
	}
}

func Test_sendRequestError(t *testing.T) {
	url := "invalid-url"
	body := []byte(`{"key": "value"}`)
	headers := map[string]string{"Authorization": "Bearer Token"}
	err := send(url, body, headers)

	if err == nil {
		t.Error("expected an error, but got nil")
	}
}
