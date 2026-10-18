package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogEntry represents a single structured JSON log record.
type LogEntry struct {
	Time         string `json:"time"`
	Level        string `json:"level"`
	Method       string `json:"method"`
	Path         string `json:"path"`
	Status       int    `json:"status"`
	Duration     string `json:"duration"`
	ErrorMessage string `json:"errorMessage"`
}

// Logger writes JSON logs to a file and stdout.
type Logger struct {
	file *os.File
	out  io.Writer
	mu   sync.Mutex
}

// New initializes the logger service writing to logFilePath.
func New(logFilePath string) (*Logger, error) {
	dir := filepath.Dir(logFilePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Writes to both the log file and stdout (for Docker / console visibility)
	mw := io.MultiWriter(os.Stdout, file)

	return &Logger{
		file: file,
		out:  mw,
	}, nil
}

// Close closes the underlying log file.
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Log writes a single LogEntry as a JSON line.
func (l *Logger) Log(entry LogEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	_, _ = l.out.Write(append(data, '\n'))
}

// responseCapture intercepts the response to capture status code and error messages.
type responseCapture struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
	body        bytes.Buffer
}

func (rc *responseCapture) WriteHeader(code int) {
	if rc.wroteHeader {
		return
	}
	rc.wroteHeader = true
	rc.statusCode = code
	rc.ResponseWriter.WriteHeader(code)
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	if !rc.wroteHeader {
		rc.WriteHeader(http.StatusOK)
	}
	if rc.statusCode >= 400 {
		rc.body.Write(b)
	}
	return rc.ResponseWriter.Write(b)
}

// Middleware automatically logs all HTTP requests in JSON format.
func (l *Logger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rc := &responseCapture{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rc, r)

		duration := time.Since(start)

		level := "INFO"
		var errorMsg string

		if rc.statusCode >= 500 {
			level = "ERROR"
		} else if rc.statusCode >= 400 {
			level = "WARN"
		}

		if rc.statusCode >= 400 && rc.body.Len() > 0 {
			var errResp struct {
				Error   string `json:"error"`
				Message string `json:"message"`
			}
			raw := rc.body.Bytes()
			if err := json.Unmarshal(raw, &errResp); err == nil {
				if errResp.Error != "" {
					errorMsg = errResp.Error
				} else if errResp.Message != "" {
					errorMsg = errResp.Message
				} else {
					errorMsg = string(bytes.TrimSpace(raw))
				}
			} else {
				errorMsg = string(bytes.TrimSpace(raw))
			}
		}

		l.Log(LogEntry{
			Time:         start.UTC().Format(time.RFC3339),
			Level:        level,
			Method:       r.Method,
			Path:         r.URL.Path,
			Status:       rc.statusCode,
			Duration:     duration.Round(time.Microsecond).String(),
			ErrorMessage: errorMsg,
		})
	})
}
