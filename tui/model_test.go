package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewUsesDefaultOutputPath(t *testing.T) {
	model := New(func(Form) ([]byte, error) { return []byte("csr"), nil })
	if got := model.inputs[2].Value(); got != "./request.csr" {
		t.Fatalf("output path = %q, want ./request.csr", got)
	}
}

func TestWriteCSRPreservesExistingFileWhenOverwriteDeclined(t *testing.T) {
	path := filepath.Join(t.TempDir(), "request.csr")
	if err := os.WriteFile(path, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	err := WriteCSR(path, []byte("new"), func() bool { return false })
	if err == nil {
		t.Fatal("expected overwrite decline")
	}
	got, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != "old" {
		t.Fatalf("existing CSR = %q, want old", got)
	}
}

func TestNewWithCredentialsShowsResolvedCredentialSource(t *testing.T) {
	model := NewWithCredentials(func(Form) ([]byte, error) { return []byte("csr"), nil }, nil, "environment variables")
	view := model.View()
	if !strings.Contains(view, "Credentials: environment variables") {
		t.Fatalf("view = %q, want it to contain credential source line", view)
	}
}

func TestNewWithCredentialsOmitsSourceLineWhenEmpty(t *testing.T) {
	model := NewWithCredentials(func(Form) ([]byte, error) { return []byte("csr"), nil }, nil, "")
	view := model.View()
	if strings.Contains(view, "Credentials:") {
		t.Fatalf("view = %q, want no credential source line", view)
	}
}

func TestRequiredKeyIdentifierIsReported(t *testing.T) {
	model := New(func(Form) ([]byte, error) { return []byte("csr"), nil })
	model.inputs[0].Blur()
	model.focused = 2
	model.inputs[2].Focus()
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*Model)
	if model.status != "Key Identifier is required." {
		t.Fatalf("status = %q", model.status)
	}
}
