package xlg

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func TestWriteAddsSource(t *testing.T) {
	buf := new(bytes.Buffer)
	writer = buf
	defer func() { writer = nil }()

	New().Write()
	var l Record
	check(json.Unmarshal(buf.Bytes(), &l))

	if !strings.HasSuffix(l.Source.File, "xlg_test.go") {
		t.Errorf("expected file %q, got %q", "xlg_test.go", l.Source.File)
	}
	if l.Source.Line == 0 {
		t.Error("expected non-zero line")
	}
	if l.Source.Func != "xlg.TestWriteAddsSource" {
		t.Errorf("expected func %q, got %q", "xlg.TestWriteAddsSource", l.Source.Func)
	}
}

func Test_writeOccurredWithin(t *testing.T) {
	message := "message"
	err := "error"
	period := "1s"

	res := writeOccurredWithin(period, message, err)
	if res {
		t.Error("Test case 1: expected false for the first write, but got true")
	}

	res = writeOccurredWithin(period, message, err)
	if !res {
		t.Error("Test case 2: expected true for the second write of the same log, but got false")
	}

	time.Sleep(2 * time.Second)
	res = writeOccurredWithin(period, message, err)
	if res {
		t.Error("Test case 3: expected false for the third write after the period has elapsed, but got true")
	}

	invalidPeriod := "invalid"
	res = writeOccurredWithin(invalidPeriod, message, err)
	if res {
		t.Error("Test case 4: expected false for an invalid period, but got true")
	}
}

func TestWrite_httpMsg(t *testing.T) {
	buf := new(bytes.Buffer)
	writer = buf
	defer func() { writer = nil }()

	rec := Record{
		Message:    "",
		ReqMethod:  "GET",
		ReqPath:    "/test",
		RespStatus: 200,
	}
	rec.Write()
	var l Record
	check(json.Unmarshal(buf.Bytes(), &l))

	if l.Message == "" || l.Message == "xlog_empty" {
		t.Error("expected Message field to be populated, but it's still empty or 'xlog_empty'")
	}
}

func TestWriteEncoder(t *testing.T) {
	buf := new(bytes.Buffer)
	writer = buf
	defer func() { writer = nil }()

	rec := Record{Message: "<test>"}
	rec.Write()

	if !strings.Contains(buf.String(), "<test>") {
		t.Errorf("want %q in log, got %q", "<test>", buf.String())
	}
	if !strings.HasSuffix(buf.String(), "\n") {
		t.Errorf("want newline at the end, got %q", buf.String())
	}
}
