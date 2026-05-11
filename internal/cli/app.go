package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/shinozakijo/go-mock-cli/internal/repository"
	"github.com/shinozakijo/go-mock-cli/internal/server"
	"github.com/shinozakijo/go-mock-cli/internal/service"
)

type App struct {
	routeRepo    *repository.RouteRepository
	responseRepo *repository.ResponseRepository
	serverPort   string
}

func NewApp(
	routeRepo *repository.RouteRepository,
	responseRepo *repository.ResponseRepository,
	serverPort string,
) *App {
	return &App{
		routeRepo:    routeRepo,
		responseRepo: responseRepo,
		serverPort:   serverPort,
	}
}

func (a *App) Run(args []string) error {
	if len(args) == 0 {
		a.printHelp()
		return nil
	}

	switch args[0] {
	case "serve":
		return a.runServer()
	case "route":
		return a.handleRoute(args[1:])
	case "response":
		return a.handleResponse(args[1:])
	case "help":
		a.printHelp()
		return nil
	default:
		return fmt.Errorf("unknown command: %s\nrun 'help' to see available commands", args[0])
	}
}

// ─────────────────────────────────────────────
// Server
// ─────────────────────────────────────────────

func (a *App) runServer() error {
	mockService := service.NewMockService(a.routeRepo, a.responseRepo)
	handler := server.NewHandler(mockService)
	srv := server.New(a.serverPort, handler)
	fmt.Printf("🌐 Mock server running on http://localhost:%s\n", a.serverPort)
	return srv.Start()
}

// ─────────────────────────────────────────────
// Route
// ─────────────────────────────────────────────

func (a *App) handleRoute(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing route command\n  try: route list | route add | route show | route delete")
	}

	ctx := context.Background()

	switch args[0] {
	case "list":
		routes, err := a.routeRepo.GetAll(ctx)
		if err != nil {
			return err
		}
		printRouteTable(routes)
		return nil

	case "add":
		if len(args) < 4 {
			return fmt.Errorf("usage: route add <METHOD> <PATH> <DESCRIPTION>")
		}

		method := strings.ToUpper(args[1])
		path := args[2]
		description := args[3]

		if err := validateMethod(method); err != nil {
			return err
		}
		if err := validatePath(path); err != nil {
			return err
		}

		route, err := a.routeRepo.Create(ctx, method, path, description)
		if err != nil {
			return err
		}

		fmt.Printf("✅ route created\n")
		printRouteDetail(*route)
		return nil

	case "show":
		if len(args) < 3 {
			return fmt.Errorf("usage: route show <METHOD> <PATH>")
		}

		method := strings.ToUpper(args[1])
		path := args[2]

		route, err := a.routeRepo.GetByMethodAndPath(ctx, method, path)
		if err != nil {
			return err
		}

		responses, err := a.responseRepo.GetByRouteID(ctx, route.ID)
		if err != nil {
			return err
		}

		printRouteDetail(*route)
		fmt.Printf(colorBold + "  Responses:\n" + colorReset)
		printResponseTable(responses)
		return nil

	case "delete":
		if len(args) < 3 {
			return fmt.Errorf("usage: route delete <METHOD> <PATH>")
		}

		method := strings.ToUpper(args[1])
		path := args[2]

		// เช็กก่อนว่ามีจริง
		if _, err := a.routeRepo.GetByMethodAndPath(ctx, method, path); err != nil {
			return fmt.Errorf("route not found: %s %s", method, path)
		}

		if err := a.routeRepo.DeleteByMethodAndPath(ctx, method, path); err != nil {
			return err
		}

		fmt.Printf("✅ route deleted: %s %s\n", method, path)
		fmt.Println(colorYellow + "  ⚠ all responses for this route have been deleted" + colorReset)
		return nil

	default:
		return fmt.Errorf("unknown route command: %s", args[0])
	}
}

// ─────────────────────────────────────────────
// Response
// ─────────────────────────────────────────────

func (a *App) handleResponse(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing response command\n  try: response list | response add | response show | response edit-all | response activate | response delete")
	}

	ctx := context.Background()

	switch args[0] {
	case "list":
		if len(args) < 3 {
			return fmt.Errorf("usage: response list <METHOD> <PATH>")
		}

		method := strings.ToUpper(args[1])
		path := args[2]

		route, err := a.routeRepo.GetByMethodAndPath(ctx, method, path)
		if err != nil {
			return fmt.Errorf("route not found for %s %s", method, path)
		}

		responses, err := a.responseRepo.GetByRouteID(ctx, route.ID)
		if err != nil {
			return err
		}

		fmt.Printf("\n  %s%s %s%s\n", colorBold, route.Method, route.Path, colorReset)
		printResponseTable(responses)
		return nil

	case "show":
		if len(args) < 2 {
			return fmt.Errorf("usage: response show <response-id>")
		}

		response, err := a.responseRepo.GetByID(ctx, args[1])
		if err != nil {
			return err
		}

		printResponseDetail(*response)
		return nil

	case "add":
		if len(args) < 6 {
			return fmt.Errorf("usage: response add <METHOD> <PATH> <NAME> <STATUS-CODE> <BODY-JSON>")
		}

		method := strings.ToUpper(args[1])
		path := args[2]
		name := args[3]

		statusCode, err := strconv.Atoi(args[4])
		if err != nil {
			return fmt.Errorf("invalid status-code: %w", err)
		}

		body := json.RawMessage(args[5])

		if err := validateMethod(method); err != nil {
			return err
		}
		if err := validatePath(path); err != nil {
			return err
		}
		if err := validateName(name); err != nil {
			return err
		}
		if err := validateStatusCode(statusCode); err != nil {
			return err
		}
		if !json.Valid(body) {
			return fmt.Errorf("body must be valid JSON")
		}

		route, err := a.routeRepo.GetByMethodAndPath(ctx, method, path)
		if err != nil {
			return fmt.Errorf("route not found for %s %s", method, path)
		}

		headers := json.RawMessage(`{"Content-Type":"application/json"}`)
		response, err := a.responseRepo.Create(ctx, route.ID, name, statusCode, body, headers, 0)
		if err != nil {
			return err
		}

		fmt.Printf("✅ response created\n")
		printResponseDetail(*response)
		return nil

	case "add-file":
		if len(args) < 6 {
			return fmt.Errorf("usage: response add-file <METHOD> <PATH> <NAME> <STATUS-CODE> <JSON-FILE>")
		}

		method := strings.ToUpper(args[1])
		path := args[2]
		name := args[3]

		statusCode, err := strconv.Atoi(args[4])
		if err != nil {
			return fmt.Errorf("invalid status-code: %w", err)
		}

		if err := validateMethod(method); err != nil {
			return err
		}
		if err := validatePath(path); err != nil {
			return err
		}
		if err := validateName(name); err != nil {
			return err
		}
		if err := validateStatusCode(statusCode); err != nil {
			return err
		}

		body, err := readJSONFile(args[5])
		if err != nil {
			return err
		}

		route, err := a.routeRepo.GetByMethodAndPath(ctx, method, path)
		if err != nil {
			return fmt.Errorf("route not found for %s %s", method, path)
		}

		headers := json.RawMessage(`{"Content-Type":"application/json"}`)
		response, err := a.responseRepo.Create(ctx, route.ID, name, statusCode, body, headers, 0)
		if err != nil {
			return err
		}

		fmt.Printf("✅ response created from file\n")
		printResponseDetail(*response)
		return nil

	case "activate":
		if len(args) < 4 {
			return fmt.Errorf("usage: response activate <METHOD> <PATH> <response-id>")
		}

		method := strings.ToUpper(args[1])
		path := args[2]
		responseID := args[3]

		route, err := a.routeRepo.GetByMethodAndPath(ctx, method, path)
		if err != nil {
			return fmt.Errorf("route not found for %s %s", method, path)
		}

		// เช็กว่า response อยู่ภายใต้ route นี้จริง
		if err := a.responseRepo.CheckExists(ctx, responseID); err != nil {
			return err
		}

		if err := a.responseRepo.SetActive(ctx, route.ID, responseID); err != nil {
			return err
		}

		fmt.Printf("✅ response activated for %s %s\n", route.Method, route.Path)
		return nil

	case "edit":
		if len(args) < 2 {
			return fmt.Errorf("usage: response edit <response-id>")
		}

		response, err := a.responseRepo.GetByID(ctx, args[1])
		if err != nil {
			return err
		}

		tempFile, err := writeTempFile("mock-body", response.Body)
		if err != nil {
			return err
		}
		defer os.Remove(tempFile)

		fmt.Printf("opening editor for body of: %s\n", response.Name)
		if err := openEditor(tempFile); err != nil {
			return fmt.Errorf("open editor: %w", err)
		}

		updated, err := readJSONFile(tempFile)
		if err != nil {
			return fmt.Errorf("invalid edited JSON: %w", err)
		}

		if err := a.responseRepo.UpdateBody(ctx, response.ID, updated); err != nil {
			return err
		}

		fmt.Println("✅ response body updated")
		return nil

	case "edit-all":
		if len(args) < 2 {
			return fmt.Errorf("usage: response edit-all <response-id>")
		}

		response, err := a.responseRepo.GetByID(ctx, args[1])
		if err != nil {
			return err
		}

		editable := EditableResponse{
			Name:       response.Name,
			StatusCode: response.StatusCode,
			DelayMs:    response.DelayMs,
			Headers:    response.Headers,
			Body:       response.Body,
		}

		content, err := json.MarshalIndent(editable, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal editable response: %w", err)
		}

		tempFile, err := writeTempFile("mock-edit-all", content)
		if err != nil {
			return err
		}
		defer os.Remove(tempFile)

		fmt.Printf("opening editor for full response: %s\n", response.Name)
		if err := openEditor(tempFile); err != nil {
			return fmt.Errorf("open editor: %w", err)
		}

		updated, err := readEditableResponseFile(tempFile)
		if err != nil {
			return err
		}

		if err := validateName(updated.Name); err != nil {
			return err
		}
		if err := validateStatusCode(updated.StatusCode); err != nil {
			return err
		}
		if err := validateDelay(updated.DelayMs); err != nil {
			return err
		}

		if err := a.responseRepo.UpdateAll(
			ctx,
			response.ID,
			updated.Name,
			updated.StatusCode,
			updated.Headers,
			updated.Body,
			updated.DelayMs,
		); err != nil {
			return err
		}

		fmt.Println("✅ response updated")
		return nil

	case "update-status":
		if len(args) < 3 {
			return fmt.Errorf("usage: response update-status <response-id> <status-code>")
		}

		statusCode, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid status-code: %w", err)
		}

		if err := validateStatusCode(statusCode); err != nil {
			return err
		}
		if err := a.responseRepo.CheckExists(ctx, args[1]); err != nil {
			return err
		}
		if err := a.responseRepo.UpdateStatusCode(ctx, args[1], statusCode); err != nil {
			return err
		}

		fmt.Println("✅ response status updated")
		return nil

	case "update-delay":
		if len(args) < 3 {
			return fmt.Errorf("usage: response update-delay <response-id> <delay-ms>")
		}

		delayMs, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid delay-ms: %w", err)
		}

		if err := validateDelay(delayMs); err != nil {
			return err
		}
		if err := a.responseRepo.CheckExists(ctx, args[1]); err != nil {
			return err
		}
		if err := a.responseRepo.UpdateDelay(ctx, args[1], delayMs); err != nil {
			return err
		}

		fmt.Println("✅ response delay updated")
		return nil

	case "delete":
		if len(args) < 2 {
			return fmt.Errorf("usage: response delete <response-id>")
		}

		if err := a.responseRepo.CheckExists(ctx, args[1]); err != nil {
			return err
		}
		if err := a.responseRepo.Delete(ctx, args[1]); err != nil {
			return err
		}

		fmt.Println("✅ response deleted")
		return nil

	default:
		return fmt.Errorf("unknown response command: %s", args[0])
	}
}

// ─────────────────────────────────────────────
// Help
// ─────────────────────────────────────────────

func (a *App) printHelp() {
	fmt.Println()
	fmt.Println(colorBold + "  go-mock-cli" + colorReset)
	fmt.Println()
	fmt.Println(colorBold + "  Server" + colorReset)
	fmt.Println("    serve                                                     start mock HTTP server")
	fmt.Println()
	fmt.Println(colorBold + "  Route" + colorReset)
	fmt.Println("    route list                                                list all routes")
	fmt.Println("    route add <METHOD> <PATH> <DESCRIPTION>                   create a route")
	fmt.Println("    route show <METHOD> <PATH>                                show route + responses")
	fmt.Println("    route delete <METHOD> <PATH>                              delete route + all responses")
	fmt.Println()
	fmt.Println(colorBold + "  Response" + colorReset)
	fmt.Println("    response list <METHOD> <PATH>                             list responses for a route")
	fmt.Println("    response show <response-id>                               show response detail + body")
	fmt.Println("    response add <METHOD> <PATH> <NAME> <STATUS> <JSON>       create response")
	fmt.Println("    response add-file <METHOD> <PATH> <NAME> <STATUS> <FILE>  create response from file")
	fmt.Println("    response activate <METHOD> <PATH> <response-id>           set active response")
	fmt.Println("    response edit <response-id>                               edit response body in editor")
	fmt.Println("    response edit-all <response-id>                           edit full response in editor")
	fmt.Println("    response update-status <response-id> <status-code>        update status code")
	fmt.Println("    response update-delay <response-id> <delay-ms>            update delay (ms)")
	fmt.Println("    response delete <response-id>                             delete response")
	fmt.Println()
}
