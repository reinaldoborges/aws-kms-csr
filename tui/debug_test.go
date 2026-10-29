package tui

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestFileDebugLoggerAppendsStageFailures(t *testing.T) {
	path := t.TempDir() + "/debug.log"
	logger := NewFileDebugLogger(path)

	if err := logger.Log("fetching KMS public key", errors.New("request id: abc123")); err != nil {
		t.Fatal(err)
	}
	if err := logger.Log("signing CSR with KMS", errors.New("timeout")); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(got)
	if !strings.Contains(text, "[fetching KMS public key] request id: abc123") {
		t.Fatalf("log = %q, want first stage failure", text)
	}
	if !strings.Contains(text, "[signing CSR with KMS] timeout") {
		t.Fatalf("log = %q, want appended second stage failure", text)
	}
}

func TestFileDebugLoggerDoesNotWriteCredentialValues(t *testing.T) {
	path := t.TempDir() + "/debug.log"
	logger := NewFileDebugLogger(path)

	if err := logger.Log("loading AWS configuration", errors.New("AccessKeyId=AKIASECRET SecretAccessKey=top-secret SessionToken=session-secret")); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(got)
	for _, secret := range []string{"AKIASECRET", "top-secret", "session-secret"} {
		if strings.Contains(text, secret) {
			t.Fatalf("log contains credential value %q: %q", secret, text)
		}
	}
}

type recordingDebugLogger struct {
	path  string
	calls []string
}

func (l *recordingDebugLogger) Path() string { return l.path }

func (l *recordingDebugLogger) Log(stage string, err error) error {
	l.calls = append(l.calls, stage+": "+err.Error())
	return nil
}

func TestModelReportsDebugLogPathOnceAfterStageFailure(t *testing.T) {
	logger := &recordingDebugLogger{path: "./aws-kms-csr-debug.log"}
	model := NewWithDebugLogger(func(Form) ([]byte, error) {
		return nil, errors.New("network timeout")
	}, nil, "", logger)
	model.inputs[0].SetValue("alias/example")
	model.inputs[1].SetValue("example.test")

	result := model.submitForm(Form{KeyIdentifier: "alias/example", CommonName: "example.test"})()
	updated, _ := model.Update(result)
	model = updated.(*Model)

	if len(logger.calls) != 1 || !strings.Contains(logger.calls[0], "generating CSR") {
		t.Fatalf("debug log calls = %#v", logger.calls)
	}
	if strings.Count(model.status, "Debug log: ./aws-kms-csr-debug.log") != 1 {
		t.Fatalf("status = %q, want one debug-log path", model.status)
	}

	result = model.submitForm(model.form)()
	updated, _ = model.Update(result)
	model = updated.(*Model)
	if strings.Count(model.status, "Debug log: ./aws-kms-csr-debug.log") != 1 {
		t.Fatalf("status = %q, want debug-log path shown once", model.status)
	}
}

func TestModelDoesNotLogOverwriteConfirmation(t *testing.T) {
	logger := &recordingDebugLogger{path: "./aws-kms-csr-debug.log"}
	model := NewWithDebugLogger(func(Form) ([]byte, error) {
		return nil, errors.New("unused")
	}, nil, "", logger)

	updated, _ := model.Update(submitResult{err: errors.New("output exists; press y to overwrite or n to choose another path")})
	model = updated.(*Model)

	if len(logger.calls) != 0 {
		t.Fatalf("debug log calls = %#v, want none", logger.calls)
	}
}
