package xlg

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"testing"
)

func TestRedact(t *testing.T) {
	type test struct {
		name     string
		input    http.Header
		expected http.Header
	}

	tests := []test{
		{name: "nil header", input: nil, expected: nil},
		{name: "empty header", input: make(http.Header), expected: make(http.Header)},
		{name: "redact authorization header",
			input:    http.Header{"Authorization": []string{"Bearer token123"}},
			expected: http.Header{"Authorization": []string{"<xlg redacted>"}},
		},
		{name: "redact password and secret headers",
			input:    http.Header{"X-Password": []string{"mypassword"}, "X-Secret": []string{"mysecret"}},
			expected: http.Header{"X-Password": []string{"<xlg redacted>"}, "X-Secret": []string{"<xlg redacted>"}},
		},
		{name: "do not redact other headers",
			input:    http.Header{"Content-Type": []string{"application/json"}},
			expected: http.Header{"Content-Type": []string{"application/json"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := redact(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("expected: %v\ngot: %v", tc.expected, result)
			}
		})
	}
}

func TestHead2Str(t *testing.T) {
	type test struct {
		name     string
		header   http.Header
		expected string
	}

	tests := []test{
		{name: "nil header", header: nil, expected: ""},
		{name: "empty header", header: make(http.Header), expected: ""},
		{
			name: "non-empty header",
			header: http.Header{
				"Content-Type": []string{"application/json"},
				"User-Agent":   []string{"Mozilla/5.0"},
			},
			expected: "Content-Type: application/json\r\nUser-Agent: Mozilla/5.0\r\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := head2str(tc.header)
			if result != tc.expected {
				t.Errorf("expected: %s\ngot: %s", tc.expected, result)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	t.Run("body shorter than max", func(t *testing.T) {
		input := []byte("short body")
		result := truncate(input)
		if !bytes.Equal(result, input) {
			t.Errorf("expected body to remain unchanged, got: %s", string(result))
		}
	})

	t.Run("body equal to max", func(t *testing.T) {
		input := make([]byte, bodyMaxBytes)
		result := truncate(input)
		if !bytes.Equal(result, input) {
			t.Errorf("expected body to remain unchanged, got: %s", string(result))
		}
	})

	t.Run("body longer than max", func(t *testing.T) {
		input := make([]byte, bodyMaxBytes+10)
		result := truncate(input)
		expectedLen := bodyMaxBytes + len(fmt.Sprintf("...<xlg truncated %d bytes>", 10))
		if len(result) != expectedLen {
			t.Errorf("expected truncated body length to be %d, got: %d", expectedLen, len(result))
		}
	})
}

func TestFnName(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		result := fnName(nil)
		expected := "<xlg nil>"
		if result != expected {
			t.Errorf("expected: %s, got: %s", expected, result)
		}
	})

	t.Run("string input", func(t *testing.T) {
		result := fnName("testFn string")
		expected := "testFn string"
		if result != expected {
			t.Errorf("expected: %s, got: %s", expected, result)
		}
	})

	t.Run("function input", func(t *testing.T) {
		testFunc := func() {}
		result := fnName(testFunc)
		expected := runtime.FuncForPC(reflect.ValueOf(testFunc).Pointer()).Name()
		if result != expected {
			t.Errorf("expected: %s, got: %s", expected, result)
		}
	})

	t.Run("unsupported type input", func(t *testing.T) {
		var unsupportedType struct{}
		result := fnName(unsupportedType)
		expected := fmt.Sprintf("<xlg unsupported: %s>", reflect.TypeOf(unsupportedType).Kind())
		if result != expected {
			t.Errorf("expected: %s, got: %s", expected, result)
		}
	})
}

func TestHTTPMsg(t *testing.T) {
	type test struct {
		name     string
		method   string
		path     string
		status   int
		expected string
	}

	tests := []test{
		{
			name:   "without status",
			method: "GET", path: "/api", status: 0,
			expected: "GET /api",
		},
		{
			name:   "with status",
			method: "POST", path: "/users", status: 201,
			expected: "POST /users [201]",
		},
		{
			name:   "empty path",
			method: "PUT", path: "", status: 204,
			expected: "PUT  [204]",
		},
		{
			name:   "status 0",
			method: "DELETE", path: "/items", status: 0,
			expected: "DELETE /items",
		},
		{
			name:   "negative status",
			method: "GET", path: "/negative", status: -1,
			expected: "GET /negative [-1]",
		},
		{
			name:   "status below 100",
			method: "POST", path: "/low", status: 99,
			expected: "POST /low [99]",
		},
		{
			name:   "status above 599",
			method: "GET", path: "/high", status: 600,
			expected: "GET /high [600]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := httpMsg(tc.method, tc.path, tc.status)
			if result != tc.expected {
				t.Errorf("expected: %s, got: %s", tc.expected, result)
			}
		})
	}
}

func TestCompactJSON(t *testing.T) {
	t.Run("valid JSON input", func(t *testing.T) {
		input := []byte(`{"fnName": "John", "age": 30}`)
		expected := []byte(`{"fnName":"John","age":30}`)
		result, err := compactJSON(input)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !bytes.Equal(result, expected) {
			t.Errorf("expected: %s, got: %s", expected, result)
		}
	})

	t.Run("nil input", func(t *testing.T) {
		result, err := compactJSON(nil)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected: nil, got: %s", result)
		}
	})

	t.Run("invalid XML input", func(t *testing.T) {
		input := []byte(`<fnName>John</fnName><age>30</age>`)
		_, err := compactJSON(input)

		if err == nil {
			t.Errorf("expected error, but got nil")
		}
	})
}

func TestRedactURL(t *testing.T) {
	type test struct {
		name     string
		input    string
		expected string
	}
	tests := []test{
		{
			name:     "redact token and password",
			input:    "https://example.com/api?token=secret&password=secure&other=value",
			expected: "https://example.com/api?token=xlg_redacted&password=xlg_redacted&other=value",
		},
		{
			name:     "no sensitive parameters",
			input:    "https://example.com/api?name=John&age=30",
			expected: "https://example.com/api?name=John&age=30",
		},
		{
			name:     "empty url",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input, _ := url.Parse(tc.input)
			expected, _ := url.Parse(tc.expected)
			result := redactURL(input)
			if result.String() != expected.String() {
				t.Fatalf("expected: %s, got: %s", expected.String(), result.String())
			}
		})
	}

	t.Run("nil", func(t *testing.T) {
		result := redactURL(nil)
		if result != nil {
			t.Fatalf("expected: nil, got: %s", result.String())
		}
	})
}
