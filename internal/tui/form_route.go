package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinozakijo/go-mock-cli/internal/repository"
)

// ─── Messages ───────────────────────────────────

type routeCreatedMsg struct{ method, path string }

// ─── Field index ─────────────────────────────────

const (
	fieldMethod = iota
	fieldPath
	fieldDescription
	fieldRouteTotal
)

// ─── Model ───────────────────────────────────────

type FormRouteModel struct {
	routeRepo  *repository.RouteRepository
	inputs     []textinput.Model
	focusIndex int
	status     string
	isError    bool
}

func NewFormRouteModel(routeRepo *repository.RouteRepository) FormRouteModel {
	inputs := make([]textinput.Model, fieldRouteTotal)

	inputs[fieldMethod] = textinput.New()
	inputs[fieldMethod].Placeholder = "GET"
	inputs[fieldMethod].CharLimit = 10
	inputs[fieldMethod].Width = 10
	inputs[fieldMethod].Focus()

	inputs[fieldPath] = textinput.New()
	inputs[fieldPath].Placeholder = "/api/example"
	inputs[fieldPath].CharLimit = 255
	inputs[fieldPath].Width = 40

	inputs[fieldDescription] = textinput.New()
	inputs[fieldDescription].Placeholder = "describe this route"
	inputs[fieldDescription].CharLimit = 200
	inputs[fieldDescription].Width = 40

	return FormRouteModel{
		routeRepo:  routeRepo,
		inputs:     inputs,
		focusIndex: 0,
	}
}

// ─── Commands ─────────────────────────────────────

func (m FormRouteModel) submit() tea.Cmd {
	method := strings.ToUpper(strings.TrimSpace(m.inputs[fieldMethod].Value()))
	path := strings.TrimSpace(m.inputs[fieldPath].Value())
	desc := strings.TrimSpace(m.inputs[fieldDescription].Value())

	return func() tea.Msg {
		if method == "" {
			return errMsg{fmt.Errorf("method is required")}
		}
		if path == "" || !strings.HasPrefix(path, "/") {
			return errMsg{fmt.Errorf("path must start with /")}
		}

		allowed := map[string]bool{
			"GET": true, "POST": true, "PUT": true,
			"PATCH": true, "DELETE": true,
		}
		if !allowed[method] {
			return errMsg{fmt.Errorf("invalid method: %s", method)}
		}

		_, err := m.routeRepo.Create(context.Background(), method, path, desc)
		if err != nil {
			return errMsg{err}
		}

		return routeCreatedMsg{method, path}
	}
}

// ─── Init / Update / View ─────────────────────────

func (m FormRouteModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m FormRouteModel) Update(msg tea.Msg) (FormRouteModel, tea.Cmd) {
	switch msg := msg.(type) {

	case errMsg:
		m.status = "❌ " + msg.err.Error()
		m.isError = true
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			m.focusIndex = (m.focusIndex + 1) % fieldRouteTotal
			return m, m.syncFocus()

		case "shift+tab", "up":
			m.focusIndex = (m.focusIndex - 1 + fieldRouteTotal) % fieldRouteTotal
			return m, m.syncFocus()

		case "enter":
			if m.focusIndex == fieldRouteTotal-1 {
				return m, m.submit()
			}
			m.focusIndex = (m.focusIndex + 1) % fieldRouteTotal
			return m, m.syncFocus()
		}
	}

	// update input ที่ focus อยู่
	var cmds []tea.Cmd
	for i := range m.inputs {
		var cmd tea.Cmd
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m FormRouteModel) syncFocus() tea.Cmd {
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

func (m FormRouteModel) View() string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("Add Route") + "\n")
	sb.WriteString(strings.Repeat("─", 60) + "\n\n")

	labels := []string{"Method     ", "Path       ", "Description"}
	for i, label := range labels {
		focused := i == m.focusIndex
		prefix := "  "
		if focused {
			prefix = selectedRowStyle.Render("▶ ")
		}
		sb.WriteString(fmt.Sprintf("%s%s  %s\n\n",
			prefix,
			sectionStyle.Render(label),
			m.inputs[i].View(),
		))
	}

	if m.status != "" {
		if m.isError {
			sb.WriteString(errorStyle.Render(m.status) + "\n")
		} else {
			sb.WriteString(successStyle.Render(m.status) + "\n")
		}
	}

	sb.WriteString(helpStyle.Render(
		"[tab] next  [shift+tab] prev  [enter] save  [esc] cancel",
	))

	return appStyle.Render(sb.String())
}
