package tui

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

const DefaultDebugLogPath = "./aws-kms-csr-debug.log"

type DebugLogger interface {
	Log(stage string, err error) error
	Path() string
}

type FileDebugLogger struct {
	path string
}

func NewFileDebugLogger(path string) *FileDebugLogger {
	return &FileDebugLogger{path: path}
}

func (l *FileDebugLogger) Path() string {
	return l.path
}

func (l *FileDebugLogger) Log(stage string, err error) error {
	if err == nil {
		return nil
	}
	detail := redactCredentialValues(fmt.Sprintf("%+v", err))
	entry := fmt.Sprintf("%s [%s] %s\n", time.Now().Format(time.RFC3339), stage, detail)
	file, openErr := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if openErr != nil {
		return openErr
	}
	defer file.Close()
	_, writeErr := file.WriteString(entry)
	return writeErr
}

var credentialValuePattern = regexp.MustCompile(`(?i)(access[_ ]?key[_ ]?id|secret[_ ]?access[_ ]?key|session[_ ]?token)\s*[:=]\s*[^\s,;]+`)

func redactCredentialValues(detail string) string {
	return credentialValuePattern.ReplaceAllStringFunc(detail, func(match string) string {
		name := match[:strings.IndexAny(match, ":=")]
		return name + "=[REDACTED]"
	})
}
