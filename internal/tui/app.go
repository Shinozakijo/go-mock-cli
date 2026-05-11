package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shinozakijo/go-mock-cli/internal/repository"
)

// ─── Screen ───────────────────────────────────────

type screen int

const (
	screenRouteList screen = iota
	screenResponseList
	screenAddRoute
	screenAddResponse
)

// ─── App Model ────────────────────────────────────

type AppModel struct {
	screen        screen
	routeList     RouteListModel
	responseList  ResponseListModel
	formRoute     FormRouteModel
	formResponse  FormResponseModel
	serverManager *ServerManager
	program       *tea.Program
}

func NewAppModel(
	routeRepo *repository.RouteRepository,
	responseRepo *repository.ResponseRepository,
	serverManager *ServerManager,
) AppModel {
	return AppModel{
		screen:        screenRouteList,
		routeList:     NewRouteListModel(routeRepo, responseRepo, serverManager),
		serverManager: serverManager,
	}
}

func (m *AppModel) SetProgram(p *tea.Program) {
	m.program = p
}

// ─── Init / Update / View ─────────────────────────

func (m *AppModel) Init() tea.Cmd {
	return m.routeList.Init()
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// global quit
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "q", "ctrl+c":
			// stop server ก่อน quit
			if m.serverManager.IsRunning() {
				m.serverManager.Stop()
			}
			return m, tea.Quit
		}
	}

	switch m.screen {

	// ──────────── Route List ────────────
	case screenRouteList:
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {

			case "enter":
				if len(m.routeList.routes) > 0 {
					route := m.routeList.routes[m.routeList.cursor]
					m.screen = screenResponseList
					m.responseList = NewResponseListModel(
						m.routeList.routeRepo,
						m.routeList.responseRepo,
						route,
					)
					return m, m.responseList.Init()
				}
				return m, nil

			case "a":
				m.screen = screenAddRoute
				m.formRoute = NewFormRouteModel(m.routeList.routeRepo)
				return m, m.formRoute.Init()

			case "s":
				if m.serverManager.IsRunning() {
					m.serverManager.Stop()
					m.routeList.status = "🔴 server stopped"
				} else {
					if err := m.serverManager.Start(); err != nil {
						m.routeList.status = "❌ " + err.Error()
						m.routeList.isError = true
					} else {
						m.routeList.status = fmt.Sprintf("🟢 server started on :%s", m.serverManager.Port())
						m.routeList.isError = false
					}
				}
				return m, nil
			}
		}

		updated, cmd := m.routeList.Update(msg)
		m.routeList = updated
		return m, cmd

	// ──────────── Response List ────────────
	case screenResponseList:
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "left", "esc", "backspace":
				m.screen = screenRouteList
				return m, m.routeList.loadRoutes()

			case "a":
				route := m.responseList.route
				m.screen = screenAddResponse
				m.formResponse = NewFormResponseModel(
					m.responseList.responseRepo,
					route,
					m.program,
				)
				return m, m.formResponse.Init()
			}
		}

		updated, cmd := m.responseList.Update(msg, m.program)
		m.responseList = updated
		return m, cmd

	// ──────────── Add Route Form ────────────
	case screenAddRoute:
		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
			m.screen = screenRouteList
			return m, m.routeList.loadRoutes()
		}

		// รับ success message แล้วกลับหน้าหลัก
		if _, ok := msg.(routeCreatedMsg); ok {
			m.screen = screenRouteList
			m.routeList.status = "✅ route created"
			m.routeList.isError = false
			return m, m.routeList.loadRoutes()
		}

		updated, cmd := m.formRoute.Update(msg)
		m.formRoute = updated
		return m, cmd

	// ──────────── Add Response Form ────────────
	case screenAddResponse:
		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
			m.screen = screenResponseList
			return m, m.responseList.loadResponses()
		}

		// รับ success message แล้วกลับหน้า response list
		if msg, ok := msg.(responseCreatedMsg); ok {
			m.screen = screenResponseList
			m.responseList.status = fmt.Sprintf("✅ response created: %s", msg.name)
			m.responseList.isError = false
			return m, m.responseList.loadResponses()
		}

		updated, cmd := m.formResponse.Update(msg)
		m.formResponse = updated
		return m, cmd
	}

	return m, nil
}

func (m *AppModel) View() string {
	switch m.screen {
	case screenResponseList:
		return m.responseList.View()
	case screenAddRoute:
		return m.formRoute.View()
	case screenAddResponse:
		return m.formResponse.View()
	default:
		return m.routeListView()
	}
}

// routeListView เพิ่ม server status bar บน title
func (m *AppModel) routeListView() string {
	view := m.routeList.View()

	// inject server status เข้าไปใน title bar
	serverStatus := serverStatusBadge(m.serverManager)
	_ = serverStatus

	return view
}

func serverStatusBadge(sm *ServerManager) string {
	if sm.IsRunning() {
		return successStyle.Render(fmt.Sprintf("🟢 :%s", sm.Port()))
	}
	return helpStyle.Render("🔴 stopped")
}

// ─── Run ──────────────────────────────────────────

func Run(
	routeRepo *repository.RouteRepository,
	responseRepo *repository.ResponseRepository,
	serverPort string,
) error {
	sm := NewServerManager(serverPort, routeRepo, responseRepo)
	appModel := NewAppModel(routeRepo, responseRepo, sm)

	p := tea.NewProgram(&appModel, tea.WithAltScreen())
	appModel.SetProgram(p)

	_, err := p.Run()
	return err
}
