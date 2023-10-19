package xlg

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"sync"
	"time"
)

var (
	writer io.Writer = os.Stdout
	stderr           = log.New(os.Stderr, "xlg: ", log.Flags())
)

func SetOutput(w io.Writer) {
	writer = w
}

func New() Record {
	return newRecord()
}

func Msg(m string) Record {
	return newRecord().Msg(m)
}

// Request creates a Record based on an HTTP request. There is no separate Response constructor,
// as it is common practice to log both the request and response together to provide comprehensive
// context in log entries.
func Request(r *http.Request) Record {
	return newRecord().Request(r)
}

// Req allows you to create a Record with a redacted request body/headers/url.
// This is useful, for example, when handling requests that contain large files or
// sensitive information that should not be included in the logged data.
func Req(method string, url *url.URL, header http.Header, body []byte) Record {
	return newRecord().Req(method, url, header, body)
}

func Fail(fn any, err error) Record {
	return newRecord().Fail(fn, err)
}

func Panic(fn, p any) Record {
	return newRecord().Panic(fn, p)
}

func newRecord() Record {
	host, _ := os.Hostname()
	return Record{
		Hostname:    host,
		Environment: os.Getenv("ENVIRONMENT"),
		Reference:   uuid(),
	}
}

type Record struct {
	// Hostname is obtained from the kernel by default.
	// You can override it using the 'hostname' field in docker-compose.
	Hostname string `json:"host,omitempty"`

	// Environment is a variable used to specify the environment type, such as development, testing, or production.
	// You can set it by defining an 'ENVIRONMENT' environment variable in your docker-compose.yaml file.
	//
	// Example docker-compose.yaml:
	//
	// services:
	//   my_service:
	//     environment:
	//       ENVIRONMENT: prod
	//
	Environment string `json:"env,omitempty"`

	// Reference is a unique identifier for a log record.
	// It can be provided to API clients for issue resolution and additional information retrieval.
	// Additionally, it serves to link logs across different services collaborating to respond to a request.
	Reference string `json:"ref,omitempty"`

	Message string `json:"msg"`
	Error   string `json:"err,omitempty"`

	// Username represents the user who initiated the request.
	// While this field is common, it is not included in the Attributes and is provided as a separate field.
	Username string `json:"user,omitempty"`

	// Attributes holds supplementary information related to the log record.
	// This data is indexed in the collector database to facilitate fast searching.
	Attributes map[string]string `json:"attributes,omitempty"`

	ReqMethod string `json:"req_method,omitempty"`
	ReqURL    string `json:"req_url,omitempty"`

	// ReqPath is derived from ReqURL.
	// This field is separately sent and stored in the collector database to optimize fast searches.
	// This optimization is necessary because PostgreSQL lacks built-in URI parsing support,
	// and statistics are typically based on grouping by path rather than the full URI.
	ReqPath string `json:"req_path,omitempty"`

	// ReqHeader represents header section in the HTTP request.
	ReqHeader string `json:"req_header,omitempty"`

	// ReqBody represents the body of the HTTP request.
	// While it is typically in JSON format, it can also be plain text or XML.
	ReqBody string `json:"req_body,omitempty"`

	RespStatus int     `json:"resp_status,omitempty"`
	RespHeader string  `json:"resp_header,omitempty"`
	RespBody   string  `json:"resp_body,omitempty"`
	Source     *Source `json:"source,omitempty"`
}

type Source struct {
	Func string `json:"func"`
	File string `json:"file"`
	Line int    `json:"line"`
}

type key struct {
	msg string
	err string
}

var (
	cache = make(map[key]time.Time)
	mu    sync.Mutex
)

func (r Record) Write() {
	r.WriteOnceIn("")
}

func (r Record) WriteOnceIn(period string) {
	if r.Message == "" {
		if r.ReqMethod != "" && r.ReqPath != "" {
			r.Message = httpMsg(r.ReqMethod, r.ReqPath, r.RespStatus)
		} else {
			r.Message = "<xlog empty>"
		}
	}

	if period != "" && writeOccurredWithin(period, r.Message, r.Error) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(2, pcs[:]) // skip [Callers, Write]
	fs := runtime.CallersFrames([]uintptr{pcs[0]})
	f, _ := fs.Next()
	r.Source = &Source{
		Func: f.Function,
		File: f.File,
		Line: f.Line,
	}

	// encoder does not escape html (<, >, &)
	// encoder adds newline at the end
	enc := json.NewEncoder(writer)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(r); err != nil {
		stderr.Println("failed to encode record:", err)
		return
	}
}

func writeOccurredWithin(period, message, error string) bool {
	d, err := time.ParseDuration(period)
	if err != nil {
		return false
	}
	mu.Lock()
	defer mu.Unlock()
	if lastWrite, ok := cache[key{message, error}]; ok {
		if time.Since(lastWrite) < d {
			return true
		}
	}
	cache[key{message, error}] = time.Now()
	return false
}
