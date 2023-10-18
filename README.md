# xlg

# Features:

- simple (~700 lines of code)
- reliable (saves logs to disk, network independent)
- concise API (one line to log event)

# Examples:

```go
package main

// Discard logs in tests
xlg.SetOutput(os.Discard)

// Set file writer for logging
xlg.SetOutput(xlg.FileWriter{Dir: "/log/dir"})

// Set http writer for sending logs to http server
xlg.SetOutput(xlg.HttpWriter{
	URL: "http://localhost:8080/logserver", 
	Headers: make(map[string]string){"Authorization", "Bearer token"}}),
})

// Log about failed function call
xlg.Failed(validate, err).Attrs("input", input).Write()

// Log request and corresponding response
xlg.Request(req).Response(resp).Write()

// Log request/response redacted
xlg.Req(method, url, reqHeadRedacted, reqBodyRedacted).Resp(code, respHeadRedacted, respBodyRedacted).Write()

// Log summary of job
xlg.Msg("job succeeded").Attrs("durationSeconds", durationSeconds).Write()

// Limit logging of operation in loop to once per minute
xlg.Failed(function, err).WriteOnceIn("1m")

// Pre-create a 'user' log record for streamlined logging.
lg := xlg.User("user")
if err != nil {
	// no need to set user, because lg already has it
    lg.Failed(function, err).Write()
}
// no need to set user, because lg already has it
lg.Msg("job succeeded").Write()

```