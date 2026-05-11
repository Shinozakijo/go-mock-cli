package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinozakijo/go-mock-cli/internal/cli"
	"github.com/shinozakijo/go-mock-cli/internal/model"
	"github.com/shinozakijo/go-mock-cli/internal/repository"
)

// ─── Messages ───────────────────────────────────

type responsesLoadedMsg struct{ responses []model.Response }
type responseActivatedMsg struct{ name string }
type responseDeletedMsg struct{ name string }
type responseUpdatedMsg struct{ name string }

// ─── Model ───────────────────────────────────────

type ResponseListModel struct {
	routeRepo    *repository.RouteRepository
	responseRepo *repository.ResponseRepository
	route        model.Route
	responses    []model.Response
	cursor       int
	status       string
	isError      bool
	loading      bool
}

func NewResponseListModel(
	routeRepo *repository.RouteRepository,
	responseRepo *repository.ResponseRepository,
	route model.Route,
) ResponseListModel {
	return ResponseListModel{
		routeRepo:    routeRepo,
		responseRepo: responseRepo,
		route:        route,
		loading:      true,
	}
}

// ─── Commands ─────────────────────────────────────

func (m ResponseListModel) loadResponses() tea.Cmd {
	return func() tea.Msg {
		responses, err := m.responseRepo.GetByRouteID(context.Background(), m.route.ID)
		if err != nil {
			return errMsg{err}
		}
		return responsesLoadedMsg{responses}
	}
}

func (m ResponseListModel) activateResponse() tea.Cmd {
	if len(m.responses) == 0 {
		return nil
	}
	res := m.responses[m.cursor]
	return func() tea.Msg {
		err := m.responseRepo.SetActive(context.Background(), m.route.ID, res.ID)
		if err != nil {
			return errMsg{err}
		}
		return responseActivatedMsg{res.Name}
	}
}

func (m ResponseListModel) deleteResponse() tea.Cmd {
	if len(m.responses) == 0 {
		return nil
	}
	res := m.responses[m.cursor]
	return func() tea.Msg {
		err := m.responseRepo.Delete(context.Background(), res.ID)
		if err != nil {
			return errMsg{err}
		}
		return responseDeletedMsg{res.Name}
	}
}

func (m ResponseListModel) editAllResponse(p *tea.Program) tea.Cmd {
	if len(m.responses) == 0 {
		return nil
	}
	res := m.responses[m.cursor]

	return func() tea.Msg {
		editable := cli.EditableResponse{
			Name:       res.Name,
			StatusCode: res.StatusCode,
			DelayMs:    res.DelayMs,
			Headers:    res.Headers,
			Body:       res.Body,
		}

		content, err := json.MarshalIndent(editable, "", "  ")
		if err != nil {
			return errMsg{err}
		}

		tmpFile, err := os.CreateTemp("", "mock-edit-all-*.json")
		if err != nil {
			return errMsg{err}
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		if _, err := tmpFile.Write(content); err != nil {
			tmpFile.Close()
			return errMsg{err}
		}
		tmpFile.Close()

		// หยุด TUI ชั่วคราวก่อนเปิด editor
		if p == nil {
			return errMsg{fmt.Errorf("tui program is not initialized")}
		}
		p.ReleaseTerminal()
		defer p.RestoreTerminal()

		if err := cli.OpenEditor(tmpPath); err != nil {
			return errMsg{fmt.Errorf("open editor: %w", err)}
		}

		updated, err := cli.ReadEditableResponseFile(tmpPath)
		if err != nil {
			return errMsg{err}
		}

		err = m.responseRepo.UpdateAll(
			context.Background(),
			res.ID,
			updated.Name,
			updated.StatusCode,
			updated.Headers,
			updated.Body,
			updated.DelayMs,
		)
		if err != nil {
			return errMsg{err}
		}

		return responseUpdatedMsg{updated.Name}
	}
}

// ─── Init / Update / View ─────────────────────────

func (m ResponseListModel) Init() tea.Cmd {
	return m.loadResponses()
}

func (m ResponseListModel) Update(msg tea.Msg, p *tea.Program) (ResponseListModel, tea.Cmd) {
	switch msg := msg.(type) {

	case responsesLoadedMsg:
		m.loading = false
		m.responses = msg.responses
		if m.cursor >= len(m.responses) {
			m.cursor = max(0, len(m.responses)-1)
		}
		return m, nil

	case responseActivatedMsg:
		m.status = fmt.Sprintf("✅ activated: %s", msg.name)
		m.isError = false
		return m, m.loadResponses()

	case responseDeletedMsg:
		m.status = fmt.Sprintf("✅ deleted: %s", msg.name)
		m.isError = false
		return m, m.loadResponses()

	case responseUpdatedMsg:
		m.status = fmt.Sprintf("✅ updated: %s", msg.name)
		m.isError = false
		return m, m.loadResponses()

	case errMsg:
		m.status = "❌ " + msg.err.Error()
		m.isError = true
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.responses)-1 {
				m.cursor++
			}
		case " ":
			return m, m.activateResponse()
		case "d":
			return m, m.deleteResponse()
		case "e":
			return m, m.editAllResponse(p)
		case "r":
			m.loading = true
			m.status = ""
			return m, m.loadResponses()
		}
	}

	return m, nil
}

func (m ResponseListModel) View() string {
	var sb strings.Builder

	routeLabel := fmt.Sprintf("%s  %s",
		methodStyle(m.route.Method).Render(m.route.Method),
		m.route.Path,
	)
	sb.WriteString(titleStyle.Render(routeLabel) + "\n")
	sb.WriteString(sectionStyle.Render("Responses") + "\n")
	sb.WriteString(strings.Repeat("─", 60) + "\n")

	if m.loading {
		sb.WriteString(normalRowStyle.Render("  loading...") + "\n")
	} else if len(m.responses) == 0 {
		sb.WriteString(normalRowStyle.Render("  no responses found") + "\n")
	} else {
		for i, res := range m.responses {
			cursor := "  "
			if i == m.cursor {
				cursor = "▶ "
			}

			badge := inactiveBadgeStyle.Render("✗")
			if res.IsActive {
				badge = activeBadgeStyle.Render("✔")
			}

			status := statusStyle(res.StatusCode).Render(fmt.Sprintf("%d", res.StatusCode))
			delay := fmt.Sprintf("%-8s", fmt.Sprintf("%dms", res.DelayMs))
			name := fmt.Sprintf("%-20s", res.Name)

			row := fmt.Sprintf("%s%s  %s  %s  %s", cursor, badge, name, status, delay)

			if i == m.cursor {
				sb.WriteString(selectedRowStyle.Render(row) + "\n")
			} else {
				sb.WriteString(normalRowStyle.Render(row) + "\n")
			}
		}
	}

	if m.status != "" {
		sb.WriteString("\n")
		if m.isError {
			sb.WriteString(errorStyle.Render(m.status) + "\n")
		} else {
			sb.WriteString(successStyle.Render(m.status) + "\n")
		}
	}

	sb.WriteString(helpStyle.Render(
		"[↑/k][↓/j] move  [space] activate  [e] edit-all  [a] add  [d] delete  [r] refresh  [←/esc] back  [q] quit",
	))

	return appStyle.Render(sb.String())
}
