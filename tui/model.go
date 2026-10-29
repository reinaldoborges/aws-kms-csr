package tui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Form struct {
	KeyIdentifier   string
	CommonName      string
	OutputPath      string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Region          string
}

type Submit func(Form) ([]byte, error)

type Model struct {
	inputs           []textinput.Model
	focused          int
	status           string
	csr              []byte
	form             Form
	confirming       bool
	written          bool
	credentialNames  []string
	credentialSource string
	submit           Submit
	debugLogger      DebugLogger
	debugLogShown    bool
	quitting         bool
}

func New(submit Submit) *Model {
	return NewWithCredentials(submit, nil, "")
}

func NewWithCredentials(submit Submit, credentialNames []string, credentialSource string) *Model {
	return NewWithDebugLogger(submit, credentialNames, credentialSource, nil)
}

func NewWithDebugLogger(submit Submit, credentialNames []string, credentialSource string, debugLogger DebugLogger) *Model {
	labels := []string{"Key Identifier", "Subject Common Name", "Output Path"}
	placeholders := []string{"alias/my-signing-key", "example.com", "./request.csr"}
	for _, name := range credentialNames {
		labels = append(labels, name)
		placeholders = append(placeholders, credentialPlaceholder(name))
	}
	inputs := make([]textinput.Model, len(labels))
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Prompt = labels[i] + ": "
		inputs[i].Placeholder = placeholders[i]
		inputs[i].CharLimit = 512
	}
	inputs[2].SetValue("./request.csr")
	inputs[0].Focus()
	return &Model{
		inputs:           inputs,
		submit:           submit,
		credentialNames:  credentialNames,
		credentialSource: credentialSource,
		debugLogger:      debugLogger,
	}
}

func credentialPlaceholder(name string) string {
	if name == "Region" {
		return "sa-east-1"
	}
	return name
}

func (m *Model) Init() tea.Cmd { return textinput.Blink }

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "r":
			if m.status != "" && !m.confirming && m.form.KeyIdentifier != "" {
				return m, m.submitForm(m.form)
			}
		case "y":
			if m.confirming {
				m.confirming = false
				return m, m.writeCSR(true)
			}
		case "n":
			if m.confirming {
				m.confirming = false
				m.status = "Overwrite declined. Edit Output Path and press Enter to retry."
				m.focused = 2
				m.inputs[2].Focus()
				return m, nil
			}
		case "tab", "shift+tab":
			direction := 1
			if msg.String() == "shift+tab" {
				direction = -1
			}
			m.inputs[m.focused].Blur()
			m.focused = (m.focused + direction + len(m.inputs)) % len(m.inputs)
			m.inputs[m.focused].Focus()
			return m, nil
		case "ctrl+s":
			if m.confirming {
				return m, nil
			}
			if len(m.csr) > 0 {
				return m, m.writeCSR(false)
			}
			return m, m.runSubmit()
		case "enter":
			if m.confirming {
				return m, nil
			}
			if len(m.csr) > 0 {
				return m, m.writeCSR(false)
			}
			if m.focused == len(m.inputs)-1 {
				return m, m.runSubmit()
			}
		}
	case submitResult:
		if msg.err != nil {
			m.status = msg.err.Error()
			m.recordDebugFailure(msg.err)
			var actionable interface{ FocusKeyIdentifier() bool }
			if errors.As(msg.err, &actionable) && actionable.FocusKeyIdentifier() {
				m.focused = 0
				m.inputs[0].Focus()
			}
			if strings.Contains(m.status, "output exists") {
				m.confirming = true
			}
			if strings.Contains(m.status, "writing CSR") {
				m.focused = 2
				m.inputs[2].Focus()
			}
			return m, nil
		}
		m.csr = msg.csr
		if msg.written {
			m.written = true
			m.status = "CSR written successfully. Press Esc to quit."
			m.quitting = true
			return m, tea.Quit
		} else {
			m.form.OutputPath = strings.TrimSpace(m.inputs[2].Value())
			m.status = "CSR generated. Press Enter to write it, or Esc to quit."
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

func (m *Model) recordDebugFailure(err error) {
	if m.debugLogger == nil {
		return
	}
	if strings.Contains(err.Error(), "output exists") {
		return
	}
	stage, detail := debugFailure(err)
	if stage == "" || detail == nil {
		return
	}
	if m.debugLogger.Log(stage, detail) == nil {
		if !m.debugLogShown {
			pathMessage := "Debug log: " + m.debugLogger.Path()
			m.status += "\n" + pathMessage
			m.debugLogShown = true
		}
	}
}

type debugStageError interface {
	DebugStage() string
	DebugErr() error
}

func debugFailure(err error) (string, error) {
	var staged debugStageError
	if errors.As(err, &staged) {
		return staged.DebugStage(), staged.DebugErr()
	}
	if strings.HasPrefix(err.Error(), "writing CSR:") || strings.HasPrefix(err.Error(), "checking output path:") {
		return "writing CSR", err
	}
	return "generating CSR", err
}

type submitResult struct {
	csr     []byte
	err     error
	written bool
}

func (m *Model) runSubmit() tea.Cmd {
	m.csr = nil
	form := Form{
		KeyIdentifier: strings.TrimSpace(m.inputs[0].Value()),
		CommonName:    strings.TrimSpace(m.inputs[1].Value()),
		OutputPath:    strings.TrimSpace(m.inputs[2].Value()),
	}
	for i, name := range m.credentialNames {
		value := strings.TrimSpace(m.inputs[3+i].Value())
		switch name {
		case "Access Key ID":
			form.AccessKeyID = value
		case "Secret Access Key":
			form.SecretAccessKey = value
		case "Session Token":
			form.SessionToken = value
		case "Region":
			form.Region = value
		}
	}
	if form.KeyIdentifier == "" {
		m.status = "Key Identifier is required."
		m.focused = 0
		m.inputs[0].Focus()
		return nil
	}
	if form.CommonName == "" {
		m.status = "Subject Common Name is required."
		m.focused = 1
		m.inputs[1].Focus()
		return nil
	}
	if form.OutputPath == "" {
		m.status = "Output Path is required."
		m.focused = 2
		m.inputs[2].Focus()
		return nil
	}
	for i, name := range m.credentialNames {
		if strings.TrimSpace(m.inputs[3+i].Value()) == "" {
			m.status = name + " is required."
			m.focused = 3 + i
			m.inputs[m.focused].Focus()
			return nil
		}
	}
	m.form = form
	return m.submitForm(form)
}

func (m *Model) submitForm(form Form) tea.Cmd {
	m.form = form
	return func() tea.Msg {
		csr, err := m.submit(form)
		return submitResult{csr: csr, err: err}
	}
}

func (m *Model) writeCSR(force bool) tea.Cmd {
	return func() tea.Msg {
		if !force {
			if _, err := os.Stat(m.form.OutputPath); err == nil {
				return submitResult{err: fmt.Errorf("output exists; press y to overwrite or n to choose another path")}
			}
		}
		if err := WriteCSR(m.form.OutputPath, m.csr, func() bool { return force }); err != nil {
			return submitResult{err: err}
		}
		return submitResult{csr: m.csr, written: true}
	}
}

func (m *Model) View() string {
	if m.quitting {
		return ""
	}
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Render("AWS KMS CSR") + "\n\n")
	for i := range m.inputs {
		b.WriteString(m.inputs[i].View() + "\n")
		if message := m.validationMessage(i); message != "" {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("  "+message) + "\n")
		}
	}

	if m.credentialSource != "" {
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("Credentials: "+m.credentialSource) + "\n")
	}
	if m.status != "" {
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(m.status) + "\n")
	}
	if m.confirming {
		b.WriteString("\nOverwrite existing file? y/n")
	}
	if len(m.csr) > 0 {
		b.WriteString("\n\n" + string(m.csr))
	}
	b.WriteString("\nTab/Shift+Tab navigate • Enter submits • r retries • Esc quits")
	return b.String()
}

func (m *Model) validationMessage(index int) string {
	if m.inputs[index].Value() != "" {
		return ""
	}
	switch index {
	case 0:
		return "Key Identifier is required."
	case 1:
		return "Subject Common Name is required."
	case 2:
		return "Output Path is required."
	default:
		return m.credentialNames[index-3] + " is required."
	}
}

func WriteCSR(path string, csr []byte, confirm func() bool) error {
	path = filepath.Clean(path)
	if _, err := os.Stat(path); err == nil && !confirm() {
		return fmt.Errorf("overwrite declined")
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("checking output path: %w", err)
	}
	if err := os.WriteFile(path, csr, 0600); err != nil {
		return fmt.Errorf("writing CSR: %w", err)
	}
	return nil
}
