package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shinozakijo/go-mock-cli/internal/model"
)

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

func printRouteTable(routes []model.Route) {
	if len(routes) == 0 {
		fmt.Println(colorGray + "  no routes found" + colorReset)
		return
	}

	fmt.Println()
	fmt.Printf(colorBold+"  %-36s  %-8s  %-30s  %s\n"+colorReset,
		"ID", "METHOD", "PATH", "DESCRIPTION")
	fmt.Println("  " + strings.Repeat("─", 90))

	for _, r := range routes {
		methodColor := methodToColor(r.Method)
		fmt.Printf("  %s  %s%-8s%s  %-30s  %s\n",
			colorGray+r.ID+colorReset,
			methodColor, r.Method, colorReset,
			r.Path,
			r.Description,
		)
	}
	fmt.Println()
}

func printRouteDetail(route model.Route) {
	fmt.Println()
	fmt.Printf(colorBold + "  Route Detail\n" + colorReset)
	fmt.Println("  " + strings.Repeat("─", 50))
	fmt.Printf("  ID          : %s\n", colorGray+route.ID+colorReset)
	fmt.Printf("  Method      : %s%s%s\n", methodToColor(route.Method), route.Method, colorReset)
	fmt.Printf("  Path        : %s\n", colorCyan+route.Path+colorReset)
	fmt.Printf("  Description : %s\n", route.Description)
	fmt.Printf("  Created     : %s\n", colorGray+route.CreatedAt.Format("2006-01-02 15:04:05")+colorReset)
	fmt.Printf("  Updated     : %s\n", colorGray+route.UpdatedAt.Format("2006-01-02 15:04:05")+colorReset)
	fmt.Println()
}

func printResponseTable(responses []model.Response) {
	if len(responses) == 0 {
		fmt.Println(colorGray + "  no responses found" + colorReset)
		return
	}

	fmt.Println()
	fmt.Printf(colorBold+"  %-36s  %-20s  %-6s  %-8s  %s\n"+colorReset,
		"ID", "NAME", "STATUS", "DELAY", "ACTIVE")
	fmt.Println("  " + strings.Repeat("─", 85))

	for _, r := range responses {
		activeLabel := colorGray + "inactive" + colorReset
		if r.IsActive {
			activeLabel = colorGreen + "✔ active" + colorReset
		}

		statusColor := statusToColor(r.StatusCode)
		fmt.Printf("  %s  %-20s  %s%-6d%s  %-8s  %s\n",
			colorGray+r.ID+colorReset,
			r.Name,
			statusColor, r.StatusCode, colorReset,
			fmt.Sprintf("%dms", r.DelayMs),
			activeLabel,
		)
	}
	fmt.Println()
}

func printResponseDetail(r model.Response) {
	fmt.Println()
	fmt.Printf(colorBold + "  Response Detail\n" + colorReset)
	fmt.Println("  " + strings.Repeat("─", 50))
	fmt.Printf("  ID         : %s\n", colorGray+r.ID+colorReset)
	fmt.Printf("  Name       : %s\n", r.Name)

	statusColor := statusToColor(r.StatusCode)
	fmt.Printf("  Status     : %s%d%s\n", statusColor, r.StatusCode, colorReset)
	fmt.Printf("  Delay      : %dms\n", r.DelayMs)

	activeLabel := colorGray + "inactive" + colorReset
	if r.IsActive {
		activeLabel = colorGreen + "✔ active" + colorReset
	}
	fmt.Printf("  Active     : %s\n", activeLabel)
	fmt.Printf("  Created    : %s\n", colorGray+r.CreatedAt.Format("2006-01-02 15:04:05")+colorReset)
	fmt.Printf("  Updated    : %s\n", colorGray+r.UpdatedAt.Format("2006-01-02 15:04:05")+colorReset)

	// headers
	fmt.Printf("\n  Headers:\n")
	prettyPrint(r.Headers, "    ")

	// body
	fmt.Printf("\n  Body:\n")
	prettyPrint(r.Body, "    ")
	fmt.Println()
}

func prettyPrint(raw []byte, indent string) {
	if len(raw) == 0 {
		fmt.Println(indent + colorGray + "{}" + colorReset)
		return
	}

	var out interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		fmt.Println(indent + string(raw))
		return
	}

	formatted, err := json.MarshalIndent(out, indent, "  ")
	if err != nil {
		fmt.Println(indent + string(raw))
		return
	}

	fmt.Println(indent + colorCyan + string(formatted) + colorReset)
}

func methodToColor(method string) string {
	switch strings.ToUpper(method) {
	case "GET":
		return colorGreen
	case "POST":
		return colorYellow
	case "PUT":
		return colorCyan
	case "PATCH":
		return "\033[35m" // magenta
	case "DELETE":
		return colorRed
	default:
		return colorGray
	}
}

func statusToColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return colorGreen
	case code >= 300 && code < 400:
		return colorCyan
	case code >= 400 && code < 500:
		return colorYellow
	case code >= 500:
		return colorRed
	default:
		return colorGray
	}
}
