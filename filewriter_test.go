package xlg

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestFileWriter(t *testing.T) {
	tempDir := t.TempDir()
	writer := FileWriter{Dir: tempDir}
	msg := `{"message":"test"}`

	t.Run("log line has '- ' in the beginning", func(t *testing.T) {
		logLine := []byte(msg + "\n")
		n, err := writer.Write(logLine)
		if err != nil {
			t.Errorf("expected no error, but got %v", err)
		}
		if n != len(logLine)+2 {
			t.Errorf("expected %d bytes written, but wrote %d bytes", len(logLine)+2, n)
		}

		filePath := filepath.Join(tempDir, filename())
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("error reading log file: %v", err)
		}
		defer os.Remove(filePath)
		if string(fileContent) != "- "+msg+"\n" {
			t.Error("log file content does not match expected")
		}
	})

	t.Run("log line ends with one newline", func(t *testing.T) {
		logLine := []byte(msg)
		n, err := writer.Write(logLine)
		if err != nil {
			t.Errorf("expected no error, but got %v", err)
		}
		if n != len(logLine)+3 {
			t.Errorf("expected %d bytes written, but wrote %d bytes", len(logLine)+3, n)
		}

		filePath := filepath.Join(tempDir, filename())
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("error reading log file: %v", err)
		}
		defer os.Remove(filePath)
		if string(fileContent) != "- "+msg+"\n" {
			t.Error("log file content does not match expected")
		}

	})
}

// 15 000 ns/op
func Benchmark_appendToFile(b *testing.B) {
	pathname := path.Join(b.TempDir(), "test.log")
	for i := 0; i < b.N; i++ {
		if err := appendToFile(pathname, []byte("test")); err != nil {
			b.Fatal(err)
		}
	}
}

// 2 000 ns/op
func Benchmark_appendToOpenedFile(b *testing.B) {
	pathname := path.Join(b.TempDir(), "test.log")
	f, err := os.OpenFile(pathname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	check(err)
	defer f.Close()

	for i := 0; i < b.N; i++ {
		if _, err := f.Write([]byte("test")); err != nil {
			b.Fatal(err)
		}
	}
}
