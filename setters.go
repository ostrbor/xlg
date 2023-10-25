package xlg

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime/debug"
)

func (r Record) Fail(fn any, err error) Record {
	return r.Msg("FAIL " + fnName(fn)).Err(err)
}

func (r Record) Panic(fn, p any) Record {
	return r.Msg("PANIC "+fnName(fn)).Err(fmt.Errorf("%v", p)).Attrs("stack", debug.Stack())
}

func (r Record) Msg(m string) Record {
	r.Message = m
	return r
}

func (r Record) Ref(ref string) Record {
	r.Reference = ref
	return r
}

func (r Record) Err(e error) Record {
	if e != nil {
		r.Error = e.Error()
	} else {
		r.Error = "xlg_nil"
	}
	return r
}

func (r Record) User(u string) Record {
	r.Username = u
	return r
}

func (r Record) Attrs(args ...any) Record {
	if r.Attributes == nil {
		r.Attributes = make(map[string]string)
	}
	var key string
	for i, x := range args {
		switch (i + 1) % 2 {
		case 1:
			key = fmt.Sprintf("%v", x)
		case 0:
			switch val := x.(type) {
			default:
				// %#v is a Go-syntax representation of the value
				r.Attributes[key] = fmt.Sprintf("%#v", val)
			case string:
				// %#v for a string will add double quotes around the string
				// use separate case for a string to avoid double quotes around value
				r.Attributes[key] = val
			}
		}
	}
	return r
}

func (r Record) Request(req *http.Request) Record {
	if req == nil {
		return r
	}
	if req.Body == nil {
		return r.Req(req.Method, req.URL, req.Header, nil)
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		stderr.Printf("Request io.ReadAll: %v\n", err)
	}
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	return r.Req(req.Method, req.URL, req.Header, body)
}

func (r Record) Req(method string, url *url.URL, header http.Header, body []byte) Record {
	r.ReqMethod = method
	if url != nil {
		r.ReqURL = redactURL(url).String()
		r.ReqPath = url.Path
	}
	r.ReqHeader = head2str(redact(header))
	compactBody, err := compactJSON(body)
	if err == nil {
		body = compactBody
	}
	r.ReqBody = string(truncate(body))
	return r
}

func (r Record) Response(resp *http.Response) Record {
	if resp == nil {
		return r
	}
	if resp.Body == nil {
		return r.Resp(resp.StatusCode, resp.Header, nil)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		stderr.Printf("Response ReadAll: %v\n", err)
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(body))
	return r.Resp(resp.StatusCode, resp.Header, body)
}

func (r Record) Resp(status int, header http.Header, body []byte) Record {
	r.RespStatus = status
	r.RespHeader = head2str(redact(header))
	compactBody, err := compactJSON(body)
	if err == nil {
		body = compactBody
	}
	r.RespBody = string(truncate(body))
	return r
}
