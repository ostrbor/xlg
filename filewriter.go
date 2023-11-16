package xlg

import (
	"os"
	"path"
	"sync"
	"time"
)

const (
	// Constant FileFormat is exported for use in xlg-agent to facilitate log file identification.
	FileFormat = "2006-01-02"

	// The presence of NotSentMark at the beginning of a line serves as an indicator
	// that the line has not been transmitted to the collector.
	NotSentMark = '-'
)

type FileWriter struct {
	mu sync.Mutex
	// Dir is a directory where log files are stored.
	Dir string
}

func (w *FileWriter) Write(p []byte) (n int, err error) {
	// Ensure the log line ends with a newline character.
	// This check adds a newline if it's missing to maintain consistent log format.
	if p[len(p)-1] != '\n' {
		p = append(p, '\n')
	}
	line := append([]byte{NotSentMark}, p...)

	pathname := path.Join(w.Dir, filename())
	w.mu.Lock()
	defer w.mu.Unlock()
	err = appendToFile(pathname, line)
	return len(line), err
}

// The file open and close operations in this func are fast enough,
// it takes only ~0.01ms to execute this func as per benchmark tests.
// For the sake of simplicity and code readability decided not to reuse the open file descriptor.
func appendToFile(pathname string, line []byte) (err error) {
	f, err := os.OpenFile(pathname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = f.Write(line)
	if err != nil {
		return err
	}
	// It's important to close the file to prevent 'too many open files' errors,
	// as there is a limit on open file descriptors.
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

func filename() string {
	return time.Now().Format(FileFormat)
}
