package xlg

import (
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"
)

func TestFileWriter(t *testing.T) {
	tempDir := t.TempDir()
	writer := FileWriter{Dir: tempDir}

	logLine := []byte(`{"message":"test"}` + "\n")
	n, err := writer.Write(logLine)
	if err != nil {
		t.Errorf("expected no error, but got %v", err)
	}
	if n != len(logLine)+2 {
		t.Errorf("expected %d bytes written, but wrote %d bytes", len(logLine)+2, n)
	}

	fileName := time.Now().Format("2006-01-02")
	filePath := filepath.Join(tempDir, fileName)
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Errorf("error reading log file: %v", err)
	}
	if string(fileContent) != `- {"message":"test"}`+"\n" {
		t.Error("log file content does not match expected")
	}
}

// 15 000 ns/op
func Benchmark_appendToFile(b *testing.B) {
	dir := b.TempDir()
	pathname := path.Join(dir, "testFn.log")
	for i := 0; i < b.N; i++ {
		if err := appendToFile(pathname, []byte("test")); err != nil {
			b.Fatal(err)
		}
	}
}

// 2 000 ns/op
func Benchmark_appendToOpenedFile(b *testing.B) {
	dir := b.TempDir()
	pathname := path.Join(dir, "testFn.log")
	f, err := os.OpenFile(pathname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	check(err)
	defer f.Close()

	for i := 0; i < b.N; i++ {
		if _, err := f.Write([]byte("test")); err != nil {
			b.Fatal(err)
		}
	}
}
