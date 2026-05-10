package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

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
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func (a *App) runServer() error {
	mockService := service.NewMockService(a.routeRepo, a.responseRepo)
	handler := server.NewHandler(mockService)
	srv := server.New(a.serverPort, handler)

	fmt.Printf("🌐 Mock server running on http://localhost:%s\n", a.serverPort)
	return srv.Start()
}

func (a *App) handleRoute(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing route command")
	}

	ctx := context.Background()

	switch args[0] {
	case "list":
		routes, err := a.routeRepo.GetAll(ctx)
		if err != nil {
			return err
		}

		if len(routes) == 0 {
			fmt.Println("no routes found")
			return nil
		}

		for _, r := range routes {
			fmt.Printf("%s | %s | %s | %s\n", r.ID, r.Method, r.Path, r.Description)
		}
		return nil

	case "add":
		if len(args) < 4 {
			return fmt.Errorf("usage: route add <METHOD> <PATH> <DESCRIPTION>")
		}

		method := args[1]
		path := args[2]
		description := args[3]

		route, err := a.routeRepo.Create(ctx, method, path, description)
		if err != nil {
			return err
		}

		fmt.Printf("✅ route created: %s | %s | %s | %s\n", route.ID, route.Method, route.Path, route.Description)
		return nil

	default:
		return fmt.Errorf("unknown route command: %s", args[0])
	}
}

func (a *App) handleResponse(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing response command")
	}

	ctx := context.Background()

	switch args[0] {
	case "list":
		if len(args) < 3 {
			return fmt.Errorf("usage: response list <METHOD> <PATH>")
		}

		method := args[1]
		path := args[2]

		route, err := a.routeRepo.GetByMethodAndPath(ctx, method, path)
		if err != nil {
			return fmt.Errorf("route not found for %s %s: %w", method, path, err)
		}

		responses, err := a.responseRepo.GetByRouteID(ctx, route.ID)
		if err != nil {
			return err
		}

		if len(responses) == 0 {
			fmt.Println("no responses found")
			return nil
		}

		fmt.Printf("responses for %s %s\n", route.Method, route.Path)
		for _, r := range responses {
			fmt.Printf("%s | %s | %d | active=%v | delay=%dms | body=%s\n",
				r.ID, r.Name, r.StatusCode, r.IsActive, r.DelayMs, string(r.Body))
		}
		return nil

	case "add":
		if len(args) < 6 {
			return fmt.Errorf("usage: response add <METHOD> <PATH> <NAME> <STATUS-CODE> <BODY-JSON>")
		}

		method := args[1]
		path := args[2]
		name := args[3]

		statusCode, err := strconv.Atoi(args[4])
		if err != nil {
			return fmt.Errorf("invalid status-code: %w", err)
		}

		body := json.RawMessage(args[5])
		if !json.Valid(body) {
			return fmt.Errorf("body must be valid JSON")
		}

		route, err := a.routeRepo.GetByMethodAndPath(ctx, method, path)
		if err != nil {
			return fmt.Errorf("route not found for %s %s: %w", method, path, err)
		}

		headers := json.RawMessage(`{"Content-Type":"application/json"}`)

		response, err := a.responseRepo.Create(ctx, route.ID, name, statusCode, body, headers, 0)
		if err != nil {
			return err
		}

		fmt.Printf("✅ response created for %s %s: %s | %s | %d\n",
			route.Method, route.Path, response.ID, response.Name, response.StatusCode)
		return nil

	case "add-file":
		if len(args) < 6 {
			return fmt.Errorf("usage: response add-file <METHOD> <PATH> <NAME> <STATUS-CODE> <JSON-FILE>")
		}

		method := args[1]
		path := args[2]
		name := args[3]

		statusCode, err := strconv.Atoi(args[4])
		if err != nil {
			return fmt.Errorf("invalid status-code: %w", err)
		}

		filePath := args[5]

		body, err := readJSONFile(filePath)
		if err != nil {
			return err
		}

		route, err := a.routeRepo.GetByMethodAndPath(ctx, method, path)
		if err != nil {
			return fmt.Errorf("route not found for %s %s: %w", method, path, err)
		}

		headers := json.RawMessage(`{"Content-Type":"application/json"}`)

		response, err := a.responseRepo.Create(ctx, route.ID, name, statusCode, body, headers, 0)
		if err != nil {
			return err
		}

		fmt.Printf("✅ response created from file for %s %s: %s | %s | %d\n",
			route.Method, route.Path, response.ID, response.Name, response.StatusCode)
		return nil

	case "activate":
		if len(args) < 4 {
			return fmt.Errorf("usage: response activate <METHOD> <PATH> <response-id>")
		}

		method := args[1]
		path := args[2]
		responseID := args[3]

		route, err := a.routeRepo.GetByMethodAndPath(ctx, method, path)
		if err != nil {
			return fmt.Errorf("route not found for %s %s: %w", method, path, err)
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

		responseID := args[1]

		response, err := a.responseRepo.GetByID(ctx, responseID)
		if err != nil {
			return err
		}

		tempFile, err := writeTempFile("mock-response-body", response.Body)
		if err != nil {
			return err
		}
		defer os.Remove(tempFile)

		fmt.Printf("opening editor for response: %s\n", response.ID)
		if err := openEditor(tempFile); err != nil {
			return fmt.Errorf("open editor: %w", err)
		}

		updatedBody, err := readJSONFile(tempFile)
		if err != nil {
			return fmt.Errorf("invalid edited JSON: %w", err)
		}

		if err := a.responseRepo.UpdateBody(ctx, response.ID, updatedBody); err != nil {
			return err
		}

		fmt.Println("✅ response body updated")
		return nil

	default:
		return fmt.Errorf("unknown response command: %s", args[0])
	}
}

func (a *App) printHelp() {
	fmt.Println("go-mock-cli commands:")
	fmt.Println("")
	fmt.Println("  serve")
	fmt.Println("      start mock HTTP server")
	fmt.Println("")
	fmt.Println("  route list")
	fmt.Println("      list all routes")
	fmt.Println("")
	fmt.Println("  route add <METHOD> <PATH> <DESCRIPTION>")
	fmt.Println("      create a route")
	fmt.Println("")
	fmt.Println("  response list <METHOD> <PATH>")
	fmt.Println("      list responses for an API route")
	fmt.Println("")
	fmt.Println("  response add <METHOD> <PATH> <NAME> <STATUS-CODE> <BODY-JSON>")
	fmt.Println("      create a response for an API route")
	fmt.Println("")
	fmt.Println("  response add-file <METHOD> <PATH> <NAME> <STATUS-CODE> <JSON-FILE>")
	fmt.Println("      create a response from JSON file")
	fmt.Println("")
	fmt.Println("  response activate <METHOD> <PATH> <response-id>")
	fmt.Println("      set active response for an API route")
	fmt.Println("")
	fmt.Println("  response edit <response-id>")
	fmt.Println("      open editor to edit response body")
}