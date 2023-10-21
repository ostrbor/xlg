package xlg

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"
)

func testFn() {}

func TestRecord_Fail(t *testing.T) {
	err := errors.New("error")
	l := New().Fail(testFn, err)
	if l.Message != "FAIL xlg.testFn" {
		t.Errorf("expected msg '%s', got '%s'", "FAIL xlg.testFn", l.Message)
	}
	if l.Error != err.Error() {
		t.Errorf("expected error '%s', got '%s'", err.Error(), l.Error)
	}

	job := "job_name"
	l = New().Fail(job, err)
	if l.Message != "FAIL "+job {
		t.Errorf("expected msg '%s', got '%s'", "FAIL "+job, l.Message)
	}
}

func TestRecord_Panic(t *testing.T) {
	p := "panic"
	l := New().Panic(testFn, p)
	if l.Message != "PANIC xlg.testFn" {
		t.Errorf("expected msg '%s', got '%s'", "PANIC xlg.testFn", l.Message)
	}
	if l.Error != p {
		t.Errorf("expected error '%s', got '%s'", p, l.Error)
	}
	if _, ok := l.Attributes["stack"]; !ok {
		t.Errorf("expected stack trace in Attributes, but it is missing")
	}
}

func TestRecord_Msg(t *testing.T) {
	m := "message"
	l := New().Msg(m)
	if l.Message != m {
		t.Errorf("expected msg '%s', got '%s'", m, l.Message)
	}
}

func TestRecord_Ref(t *testing.T) {
	r := "reference"
	l := New().Ref(r)
	if l.Reference != r {
		t.Errorf("expected reference '%s', got '%s'", r, l.Reference)
	}
}

func TestRecord_Err(t *testing.T) {
	e := errors.New("error")
	l := New().Err(e)
	if l.Error != e.Error() {
		t.Errorf("expected error '%s', got '%s'", e.Error(), l.Error)
	}

	l = New().Err(nil)
	if l.Error != "<xlg nil>" {
		t.Errorf("expected error '%s', got '%s'", "<xlg nil>", l.Error)
	}
}

func TestRecord_User(t *testing.T) {
	u := "user"
	l := New().User(u)
	if l.Username != u {
		t.Errorf("expected user '%s', got '%s'", u, l.Username)
	}
}

func TestRecord_Attrs(t *testing.T) {
	t.Run("update Attributes with additional Attrs call", func(t *testing.T) {
		l := New().Attrs("key1", "value1").Attrs("key2", 42)
		if len(l.Attributes) != 2 {
			t.Errorf("expected 2 items in Attributes, but got %d", len(l.Attributes))
		}
		if l.Attributes["key1"] != "value1" {
			t.Errorf("expected 'key1' to map to 'value1', but got '%s'", l.Attributes["key1"])
		}
		if l.Attributes["key2"] != "42" {
			t.Errorf("expected 'key2' to map to '42', but got '%s'", l.Attributes["key2"])
		}
	})

	t.Run("empty Attributes", func(t *testing.T) {
		l := New().Attrs()
		if len(l.Attributes) != 0 {
			t.Errorf("expected 0 items in an empty Attributes, but got %d", len(l.Attributes))
		}
	})

	t.Run("mixed key and value types", func(t *testing.T) {
		l := New().Attrs("key1", "value1", "key2", 42, "key3", true)
		if len(l.Attributes) != 3 {
			t.Errorf("expected 3 items in Attributes, but got %d", len(l.Attributes))
		}
		if l.Attributes["key1"] != "value1" {
			t.Errorf("expected 'key1' to map to 'value1', but got '%s'", l.Attributes["key1"])
		}
		if l.Attributes["key2"] != "42" {
			t.Errorf("expected 'key2' to map to '42', but got '%s'", l.Attributes["key2"])
		}
		if l.Attributes["key3"] != "true" {
			t.Errorf("expected 'key3' to map to 'true', but got '%s'", l.Attributes["key3"])
		}
	})

	t.Run("odd number of key-value pairs", func(t *testing.T) {
		l := New().Attrs("key1", "value1", "key2")
		if len(l.Attributes) != 1 {
			t.Errorf("expected 1 item in Attributes, but got %d", len(l.Attributes))
		}
		if l.Attributes["key1"] != "value1" {
			t.Errorf("expected 'key1' to map to 'value1', but got '%s'", l.Attributes["key1"])
		}
		if _, ok := l.Attributes["key2"]; ok {
			t.Errorf("expected 'key2' to be missing from Attributes, but it exists")
		}
	})
}

func TestRecord_Request(t *testing.T) {
	t.Run("nil Request", func(t *testing.T) {
		l := New().Request(nil)
		if l.ReqMethod != "" || l.ReqURL != "" || l.ReqPath != "" || l.ReqHeader != "" || l.ReqBody != "" {
			t.Errorf("expected empty request fields, but got %+v", l)
		}
	})

	t.Run("valid Request with body", func(t *testing.T) {
		b := []byte(`{"key":"value"}`)
		req, _ := http.NewRequest("POST", "https://example.com/path", bytes.NewBuffer(b))
		req.Header.Set("Content-Type", "application/json")

		l := New().Request(req)
		if l.ReqMethod != "POST" {
			t.Errorf("expected ReqMethod to be 'POST', but got '%s'", l.ReqMethod)
		}
		if l.ReqURL != "https://example.com/path" {
			t.Errorf("expected ReqURL to be 'https://example.com/path', but got '%s'", l.ReqURL)
		}
		if l.ReqPath != "/path" {
			t.Errorf("expected ReqPath to be '/path', but got '%s'", l.ReqPath)
		}
		expectedHeaders := "Content-Type: application/json\r\n"
		if l.ReqHeader != expectedHeaders {
			t.Errorf("expected ReqHeader to be:\n%s\nbut got:\n%s", expectedHeaders, l.ReqHeader)
		}
		if l.ReqBody != string(b) {
			t.Errorf("expected ReqBody to be '%s', but got '%s'", string(b), l.ReqBody)
		}
	})

	t.Run("valid Request without body", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example.com/path", nil)
		req.Header.Set("Content-Type", "application/json")
		l := New().Request(req)

		if l.ReqMethod != "GET" {
			t.Errorf("expected ReqMethod to be 'GET', but got '%s'", l.ReqMethod)
		}
		if l.ReqURL != "https://example.com/path" {
			t.Errorf("expected ReqURL to be 'https://example.com/path', but got '%s'", l.ReqURL)
		}
		if l.ReqPath != "/path" {
			t.Errorf("expected ReqPath to be '/path', but got '%s'", l.ReqPath)
		}
		expectedHeaders := "Content-Type: application/json\r\n"
		if l.ReqHeader != expectedHeaders {
			t.Errorf("expected ReqHeader to be:\n%s\nbut got:\n%s", expectedHeaders, l.ReqHeader)
		}
		if l.ReqBody != "" {
			t.Errorf("expected ReqBody to be an empty string, but got '%s'", l.ReqBody)
		}
	})

	t.Run("valid Request with empty Header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "https://example.com/path", nil)
		l := New().Request(req)
		if l.ReqHeader != "" {
			t.Errorf("expected ReqHeader to be an empty string, but got:\n%s", l.ReqHeader)
		}
	})

}

func TestRecord_Response(t *testing.T) {
	t.Run("nil Response", func(t *testing.T) {
		l := New().Response(nil)
		if l.RespStatus != 0 || l.RespHeader != "" || l.RespBody != "" {
			t.Errorf("expected empty response fields, but got %+v", l)
		}
	})

	t.Run("valid Response with body", func(t *testing.T) {
		b := []byte(`{"key":"value"}`)
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(bytes.NewBuffer(b)),
		}

		l := New().Response(resp)
		if l.RespStatus != 200 {
			t.Errorf("expected RespStatus to be 200, but got %d", l.RespStatus)
		}
		expectedHeaders := "Content-Type: application/json\r\n"
		if l.RespHeader != expectedHeaders {
			t.Errorf("expected RespHeader to be:\n%s\nbut got:\n%s", expectedHeaders, l.RespHeader)
		}
		if l.RespBody != string(b) {
			t.Errorf("expected RespBody to be '%s', but got '%s'", string(b), l.RespBody)
		}
	})

	t.Run("valid Response without body", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       nil,
		}

		l := New().Response(resp)
		if l.RespStatus != 200 {
			t.Errorf("expected RespStatus to be 200, but got %d", l.RespStatus)
		}
		expectedHeaders := "Content-Type: application/json\r\n"
		if l.RespHeader != expectedHeaders {
			t.Errorf("expected RespHeader to be:\n%s\nbut got:\n%s", expectedHeaders, l.RespHeader)
		}
		if l.RespBody != "" {
			t.Errorf("expected RespBody to be an empty string, but got '%s'", l.RespBody)
		}
	})

	t.Run("valid Response with empty Header", func(t *testing.T) {
		resp := &http.Response{StatusCode: http.StatusOK}
		l := New().Response(resp)
		if l.RespHeader != "" {
			t.Errorf("expected RespHeader to be an empty string, but got:\n%s", l.RespHeader)
		}
	})
}
