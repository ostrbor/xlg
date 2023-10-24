// util file: Many utility functions handle nil input, enhancing code readability
// by eliminating the need for callers to perform nil checks.
package xlg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strings"
)

// fnName retrieves type of the provided function or interface value.
// It returns the name of a function when the input is a function, the value as a string
// when it's a string, and an error for unsupported types.
func fnName(fn any) string {
	if fn == nil {
		return "<xlg nil>"
	}

	k := reflect.TypeOf(fn).Kind()
	switch k {
	case reflect.String:
		return fn.(string)
	case reflect.Func:
		return runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	default:
		return fmt.Sprintf("<xlg unsupported: %s>", k)
	}
}

const hexLetters = "abcdef0123456789"

func uuid() string {
	b := make([]byte, 36)
	for i := range b {
		b[i] = hexLetters[rand.Intn(len(hexLetters))]
	}
	return string(b)
}

// Usually the size of rather big body is 2KB (example in testdata),
// limit is increased by around 2 times to be sure that most logs will not be truncated.
const bodyMaxBytes = 5 * 1 << 10

// truncate truncates the provided byte slice to bodyMaxBytes
func truncate(body []byte) []byte {
	if len(body) > bodyMaxBytes {
		b := body[:bodyMaxBytes]
		return append(b, []byte(fmt.Sprintf("...<xlg truncated %d bytes>", len(body)-len(b)))...)
	}
	return body
}

// head2str converts http.Header to a string.
//
// This function handles cases where the input http.Header (h) might be nil.
// While it's uncommon, h can be nil when constructing custom http.Request or http.Response objects,
// or when using setters like Req/Resp that accept http.Header, allowing callers to pass nil.
func head2str(h http.Header) string {
	if h == nil {
		return ""
	}
	b := new(bytes.Buffer)
	if err := h.Write(b); err != nil {
		stderr.Printf("head2str: %v\n", err)
	}
	return b.String()
}

// must be safe for use in header and url
const redacted = "...xlg_redacted..."

func redact(h http.Header) (c http.Header) {
	if h == nil {
		return nil
	}
	c = h.Clone()
	for k := range c {
		edit := true

		switch kl := strings.ToLower(k); {
		case kl == "authorization":
		case strings.Contains(kl, "password"):
		case strings.Contains(kl, "secret"):
		case strings.Contains(kl, "token"):
		case strings.Contains(kl, "key"):
		default:
			edit = false
		}

		if edit {
			c[k] = []string{redacted}
		}
	}
	return c
}

func redactURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	c := *u
	for k := range c.Query() {
		switch k {
		case "token", "password":
			c.RawQuery = strings.Replace(c.RawQuery, c.Query().Get(k), redacted, 1)
		}
	}
	return &c
}

// httpMsg returns a string representation of an HTTP transaction: request and optionally response.
// If the log includes only a request, the method and path are included in the message.
// If a response status is provided and is not zero, it is also included in the message.
func httpMsg(method, path string, status int) string {
	res := fmt.Sprintf("%s %s", method, path)
	if status != 0 {
		res += fmt.Sprintf(" [%d]", status)
	}
	return res
}

// compactJSON removes insignificant white spaces from the provided JSON input.
// This function reduces the size of testdata/req_body.json from 4KB to 2KB.
func compactJSON(input []byte) ([]byte, error) {
	if input == nil {
		return nil, nil
	}
	b := new(bytes.Buffer)
	err := json.Compact(b, input)
	return b.Bytes(), err
}
