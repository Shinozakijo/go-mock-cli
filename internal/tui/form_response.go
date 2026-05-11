package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinozakijo/go-mock-cli/internal/cli"
	"github.com/shinozakijo/go-mock-cli/internal/model"
	"github.com/shinozakijo/go-mock-cli/internal/repository"
)

// ─── Messages ───────────────────────────────────

type responseCreatedMsg struct{ name string }

// ─── Field index ─────────────────────────────────

const (
	fieldResName = iota
	fieldResStatus
	fieldResDelay
	fieldResTotal
)

// ─── Model ───────────────────────────────────────

type FormResponseModel struct {
	responseRepo *repository.ResponseRepository
	route        model.Route
	inputs       []textinput.Model
	focusIndex   int
	status       string
	isError      bool
	program      *tea.Program
}

func NewFormResponseModel(
	responseRepo *repository.ResponseRepository,
	route model.Route,
	program *tea.Program,
) FormResponseModel {
	inputs := make([]textinput.Model, fieldResTotal)

	inputs[fieldResName] = textinput.New()
	inputs[fieldResName].Placeholder = "success"
	inputs[fieldResName].CharLimit = 100
	inputs[fieldResName].Width = 30
	inputs[fieldResName].Focus()

	inputs[fieldResStatus] = textinput.New()
	inputs[fieldResStatus].Placeholder = "200"
	inputs[fieldResStatus].CharLimit = 3
	inputs[fieldResStatus].Width = 10

	inputs[fieldResDelay] = textinput.New()
	inputs[fieldResDelay].Placeholder = "0"
	inputs[fieldResDelay].CharLimit = 6
	inputs[fieldResDelay].Width = 10

	return FormResponseModel{
		responseRepo: responseRepo,
		route:        route,
		inputs:       inputs,
		focusIndex:   0,
		program:      program,
	}
}

// ─── Commands ─────────────────────────────────────

func (m FormResponseModel) submit() tea.Cmd {
	name := strings.TrimSpace(m.inputs[fieldResName].Value())
	statusStr := strings.TrimSpace(m.inputs[fieldResStatus].Value())
	delayStr := strings.TrimSpace(m.inputs[fieldResDelay].Value())

	return func() tea.Msg {
		// validate
		if name == "" || strings.Contains(name, " ") {
			return errMsg{fmt.Errorf("name cannot be empty or contain spaces")}
		}

		statusCode, err := strconv.Atoi(statusStr)
		if err != nil || statusCode < 100 || statusCode > 599 {
			return errMsg{fmt.Errorf("invalid status code (100-599)")}
		}

		delayMs := 0
		if delayStr != "" {
			delayMs, err = strconv.Atoi(delayStr)
			if err != nil || delayMs < 0 {
				return errMsg{fmt.Errorf("delay must be >= 0")}
			}
		}

		// เปิด editor ให้กรอก body
		initial := []byte(`{
  "message": "ok"
}`)
		tmpFile, err := os.CreateTemp("", "mock-new-body-*.json")
		if err != nil {
			return errMsg{err}
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		tmpFile.Write(initial)
		tmpFile.Close()

		m.program.ReleaseTerminal()
		defer m.program.RestoreTerminal()

		if err := cli.OpenEditor(tmpPath); err != nil {
			return errMsg{fmt.Errorf("open editor: %w", err)}
		}

		body, err := cli.ReadJSONFile(tmpPath)
		if err != nil {
			return errMsg{err}
		}

		headers := json.RawMessage(`{"Content-Type":"application/json"}`)

		res, err := m.responseRepo.Create(
			context.Background(),
			m.route.ID,
			name,
			statusCode,
			body,
			headers,
			delayMs,
		)
		if err != nil {
			return errMsg{err}
		}

		return responseCreatedMsg{res.Name}
	}
}

// ─── Init / Update / View ─────────────────────────

func (m FormResponseModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m FormResponseModel) Update(msg tea.Msg) (FormResponseModel, tea.Cmd) {
	switch msg := msg.(type) {

	case errMsg:
		m.status = "❌ " + msg.err.Error()
		m.isError = true
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			m.focusIndex = (m.focusIndex + 1) % fieldResTotal
			return m, m.syncFocus()

		case "shift+tab", "up":
			m.focusIndex = (m.focusIndex - 1 + fieldResTotal) % fieldResTotal
			return m, m.syncFocus()

		case "enter":
			if m.focusIndex == fieldResTotal-1 {
				// field สุดท้าย → submit
				return m, m.submit()
			}
			m.focusIndex = (m.focusIndex + 1) % fieldResTotal
			return m, m.syncFocus()
		}
	}

	var cmds []tea.Cmd
	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m FormResponseModel) syncFocus() tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		if i == m.focusIndex {
			cmds[i] = m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
	return tea.Batch(cmds...)
}

func (m FormResponseModel) View() string {
	var sb strings.Builder

	routeLabel := fmt.Sprintf("Add Response  →  %s  %s",
		methodStyle(m.route.Method).Render(m.route.Method),
		m.route.Path,
	)
	sb.WriteString(titleStyle.Render(routeLabel) + "\n")
	sb.WriteString(strings.Repeat("─", 60) + "\n\n")

	labels := []string{"Name      ", "Status    ", "Delay(ms) "}
	hints := []string{
		"(no spaces, e.g. success_20000)",
		"(100-599)",
		"(0 = no delay)",
	}

	for i, label := range labels {
		focused := i == m.focusIndex
		prefix := "  "
		if focused {
			prefix = selectedRowStyle.Render("▶ ")
		}
		sb.WriteString(fmt.Sprintf("%s%s  %s  %s\n\n",
			prefix,
			sectionStyle.Render(label),
			m.inputs[i].View(),
			helpStyle.Render(hints[i]),
		))
	}

	sb.WriteString(helpStyle.Render("  (after delay → editor opens for body JSON)") + "\n\n")

	if m.status != "" {
		if m.isError {
			sb.WriteString(errorStyle.Render(m.status) + "\n")
		} else {
			sb.WriteString(successStyle.Render(m.status) + "\n")
		}
	}

	sb.WriteString(helpStyle.Render(
		"[tab] next  [shift+tab] prev  [enter] next/save  [esc] cancel",
	))

	return appStyle.Render(sb.String())
}
