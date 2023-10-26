package xlg

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

const (
	requestTimeout   = 10 * time.Second
	maxRetryDuration = 300 * time.Second
	retryInterval    = 10 * time.Second
)

// HttpWriter represents an object for sending HTTP POST requests.
type HttpWriter struct {
	URL     string            // The URL to send the request to
	Headers map[string]string // Headers for authorization purposes
}

func (w HttpWriter) Write(p []byte) (n int, err error) {
	deadline := time.Now().Add(maxRetryDuration)
	for tries := 0; time.Now().Before(deadline); tries++ {
		err = send(w.URL, p, w.Headers)
		if err == nil {
			return len(p), nil
		}
		time.Sleep(retryInterval)
	}
	stderr.Printf("Write: %v\n", err)
	return 0, err
}

func send(url string, body []byte, headers map[string]string) (err error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		stderr.Printf("NewRequest: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	c := http.Client{Timeout: requestTimeout}
	resp, err := c.Do(req)
	if err != nil {
		stderr.Printf("Do: %v\n", err)
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := fmt.Errorf("expected 2xx, got %d", resp.StatusCode)
		stderr.Printf("send status check: %v\n", err)
		return err
	}
	return nil
}
