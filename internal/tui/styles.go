package tui

import "github.com/charmbracelet/lipgloss"

var (
	// layout
	appStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// title bar
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("6")).
			PaddingBottom(1)

	// header ของแต่ละ section
	sectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15"))

	// row ปกติ
	normalRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	// row ที่ cursor อยู่
	selectedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("6")).
				Bold(true)

	// active badge
	activeBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("2")).
				Bold(true)

	// inactive badge
	inactiveBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))

	// status bar ด้านล่าง
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			PaddingTop(1)

	// error message
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)

	// success message
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Bold(true)

	// method colors
	methodGetStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Bold(true).
			Width(7)

	methodPostStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")).
			Bold(true).
			Width(7)

	methodPutStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).
			Bold(true).
			Width(7)

	methodPatchStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("5")).
				Bold(true).
				Width(7)

	methodDeleteStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("1")).
				Bold(true).
				Width(7)

	// status code colors
	status2xxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Width(6)

	status3xxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).
			Width(6)

	status4xxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")).
			Width(6)

	status5xxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Width(6)
)

func methodStyle(method string) lipgloss.Style {
	switch method {
	case "GET":
		return methodGetStyle
	case "POST":
		return methodPostStyle
	case "PUT":
		return methodPutStyle
	case "PATCH":
		return methodPatchStyle
	case "DELETE":
		return methodDeleteStyle
	default:
		return normalRowStyle.Copy().Width(7)
	}
}

func statusStyle(code int) lipgloss.Style {
	switch {
	case code >= 200 && code < 300:
		return status2xxStyle
	case code >= 300 && code < 400:
		return status3xxStyle
	case code >= 400 && code < 500:
		return status4xxStyle
	default:
		return status5xxStyle
	}
}
