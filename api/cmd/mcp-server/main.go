package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/LSUDOKOS/signal/internal/mcp"
)

func main() {
	port := os.Getenv("MCP_PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("starting signal MCP server on port %s", port)

	handler := &mcp.ToolHandler{}

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Tool definitions endpoint (for MCP client discovery)
	mux.HandleFunc("/tools", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mcp.ToolDefinitions())
	})

	// Tool execution endpoint
	mux.HandleFunc("/tools/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		toolName := r.URL.Path[len("/tools/"):]
		var args json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		var result map[string]interface{}
		var err error

		switch toolName {
		case "block_focus_time":
			result, err = handler.HandleBlockFocusTime(r.Context(), args)
		case "get_user_status":
			result, err = handler.HandleGetUserStatus(r.Context(), args)
		case "set_slack_status":
			result, err = handler.HandleSetSlackStatus(r.Context(), args)
		default:
			http.Error(w, fmt.Sprintf("unknown tool: %s", toolName), http.StatusNotFound)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	// SSE endpoint for streaming events
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		// Send initial connection event
		fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\"}\n\n")
		flusher.Flush()

		// Keep connection alive
		ctx := r.Context()
		<-ctx.Done()
	})

	// CORS middleware wrapper
	corsHandler := corsMiddleware(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: corsHandler,
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down MCP server...")
		server.Shutdown(context.Background())
	}()

	log.Printf("MCP server listening on :%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
