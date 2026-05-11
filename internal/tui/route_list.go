package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinozakijo/go-mock-cli/internal/model"
	"github.com/shinozakijo/go-mock-cli/internal/repository"
)

type routesLoadedMsg struct{ routes []model.Route }
type routeDeletedMsg struct{ method, path string }
type errMsg struct{ err error }

type RouteListModel struct {
	routeRepo    *repository.RouteRepository
	responseRepo *repository.ResponseRepository
	routes       []model.Route
	cursor       int
	status       string
	isError      bool
	loading      bool
	serverMgr    *ServerManager
}

func NewRouteListModel(
	routeRepo *repository.RouteRepository,
	responseRepo *repository.ResponseRepository,
	serverMgr *ServerManager,
) RouteListModel {
	return RouteListModel{
		routeRepo:    routeRepo,
		responseRepo: responseRepo,
		loading:      true,
		serverMgr:    serverMgr,
	}
}

func (m RouteListModel) loadRoutes() tea.Cmd {
	return func() tea.Msg {
		routes, err := m.routeRepo.GetAll(context.Background())
		if err != nil {
			return errMsg{err}
		}
		return routesLoadedMsg{routes}
	}
}

func (m RouteListModel) deleteRoute() tea.Cmd {
	if len(m.routes) == 0 {
		return nil
	}
	route := m.routes[m.cursor]
	return func() tea.Msg {
		err := m.routeRepo.DeleteByMethodAndPath(
			context.Background(), route.Method, route.Path,
		)
		if err != nil {
			return errMsg{err}
		}
		return routeDeletedMsg{route.Method, route.Path}
	}
}

func (m RouteListModel) Init() tea.Cmd {
	return m.loadRoutes()
}

func (m RouteListModel) Update(msg tea.Msg) (RouteListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case routesLoadedMsg:
		m.loading = false
		m.routes = msg.routes
		if m.cursor >= len(m.routes) {
			m.cursor = max(0, len(m.routes)-1)
		}
		return m, nil

	case routeDeletedMsg:
		m.status = fmt.Sprintf("✅ deleted %s %s", msg.method, msg.path)
		m.isError = false
		return m, m.loadRoutes()

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
			if m.cursor < len(m.routes)-1 {
				m.cursor++
			}
		case "r":
			m.loading = true
			m.status = ""
			return m, m.loadRoutes()
		case "d":
			if len(m.routes) > 0 {
				return m, m.deleteRoute()
			}
		}
	}
	return m, nil
}

func (m RouteListModel) View() string {
	var sb strings.Builder

	// title + server status
	serverBadge := ""
	if m.serverMgr != nil {
		if m.serverMgr.IsRunning() {
			serverBadge = "  " + successStyle.Render(fmt.Sprintf("🟢 :%s", m.serverMgr.Port()))
		} else {
			serverBadge = "  " + helpStyle.Render("🔴 stopped")
		}
	}
	sb.WriteString(titleStyle.Render("🔧 go-mock-cli") + serverBadge + "\n")
	sb.WriteString(sectionStyle.Render("Routes") + "\n")
	sb.WriteString(strings.Repeat("─", 60) + "\n")

	if m.loading {
		sb.WriteString(normalRowStyle.Render("  loading...") + "\n")
	} else if len(m.routes) == 0 {
		sb.WriteString(normalRowStyle.Render("  no routes — press [a] to add one") + "\n")
	} else {
		for i, route := range m.routes {
			cursor := "  "
			if i == m.cursor {
				cursor = "▶ "
			}

			method := methodStyle(route.Method).Render(route.Method)
			path := fmt.Sprintf("%-30s", route.Path)
			desc := route.Description
			row := fmt.Sprintf("%s%s  %s  %s", cursor, method, path, desc)

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
		"[↑/k][↓/j] move  [enter] responses  [a] add route  [d] delete  [s] toggle server  [r] refresh  [q] quit",
	))

	return appStyle.Render(sb.String())
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
