// Package srv implements the web server for mini-kanban.
package srv

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"mini-kanban/db"
	"mini-kanban/web"
)

// Server handles HTTP requests for mini-kanban.
type Server struct {
	DB             *sql.DB
	DefaultProject string
	DevMode        bool

	// Real-time updates
	clients   map[chan string]bool
	clientsMu sync.Mutex
}

// New creates a new server instance.
func New(dbPath, defaultProject string, devMode bool) (*Server, error) {
	database, err := db.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.RunMigrations(database); err != nil {
		database.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return &Server{
		DB:             database,
		DefaultProject: defaultProject,
		DevMode:        devMode,
		clients:        make(map[chan string]bool),
	}, nil
}

// Serve starts the HTTP server.
func (s *Server) Serve(addr string) error {
	mux := http.NewServeMux()

	// JSON API for React frontend (most specific patterns first)
	mux.HandleFunc("GET /api/config", corsMiddleware(s.handleAPIGetConfig))
	mux.HandleFunc("PUT /api/config", corsMiddleware(s.handleAPIUpdateConfig))
	mux.HandleFunc("OPTIONS /api/config", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {}))
	mux.HandleFunc("GET /api/projects", corsMiddleware(s.handleAPIProjects))
	mux.HandleFunc("GET /api/kanban", corsMiddleware(s.handleAPIKanban))
	mux.HandleFunc("POST /api/tasks", corsMiddleware(s.handleAPICreateTask))
	mux.HandleFunc("PUT /api/tasks/{id}", corsMiddleware(s.handleAPIUpdateTask))
	mux.HandleFunc("DELETE /api/tasks/{id}", corsMiddleware(s.handleAPIDeleteTask))
	mux.HandleFunc("PATCH /api/tasks/{id}/status", corsMiddleware(s.handleAPIUpdateStatus))
	mux.HandleFunc("OPTIONS /api/tasks", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {}))
	mux.HandleFunc("OPTIONS /api/tasks/{id}", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {}))
	mux.HandleFunc("OPTIONS /api/tasks/{id}/status", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {}))

	// AI assist API
	mux.HandleFunc("POST /api/ai/assist", corsMiddleware(s.handleAPIAIAssist))
	mux.HandleFunc("OPTIONS /api/ai/assist", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {}))

	// SSE for real-time updates
	mux.HandleFunc("GET /events", s.handleEvents)
	mux.HandleFunc("POST /internal/notify", s.handleNotify)

	// Serve static files
	var distFS fs.FS
	if s.DevMode {
		slog.Info("running in dev mode, serving from web/dist")
		distFS = os.DirFS("web/dist")
	} else {
		var err error
		distFS, err = fs.Sub(web.DistFS, "dist")
		if err != nil {
			return fmt.Errorf("get dist fs: %w", err)
		}
	}

	fileServer := http.FileServer(http.FS(distFS))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve index.html for SPA routes
		path := r.URL.Path
		if path != "/" && path != "" {
			// Try to serve the file; if not found, serve index.html
			if _, err := fs.Stat(distFS, path[1:]); err != nil {
				r.URL.Path = "/"
			}
		}
		fileServer.ServeHTTP(w, r)
	})

	slog.Info("starting web server", "addr", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	clientChan := make(chan string, 1)

	s.clientsMu.Lock()
	s.clients[clientChan] = true
	s.clientsMu.Unlock()

	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, clientChan)
		s.clientsMu.Unlock()
		close(clientChan)
	}()

	// Signal client is connected
	fmt.Fprintf(w, "data: connected\n\n")
	w.(http.Flusher).Flush()

	for {
		select {
		case msg := <-clientChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (s *Server) broadcast(msg string) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	for client := range s.clients {
		select {
		case client <- msg:
		default:
			// Client channel full, skip
		}
	}
}

type SSENotification struct {
	Type      string `json:"type"`
	TaskID    int64  `json:"taskId"`
	TaskTitle string `json:"taskTitle"`
	NewStatus string `json:"newStatus,omitempty"`
	OldStatus string `json:"oldStatus,omitempty"`
}

type SSEMessage struct {
	Action       string           `json:"action"`
	Notification *SSENotification `json:"notification,omitempty"`
}

func (s *Server) broadcastNotification(notification SSENotification) {
	msg := SSEMessage{
		Action:       "notify",
		Notification: &notification,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("marshal notification", "error", err)
		return
	}
	s.broadcast(string(data))
}

func (s *Server) handleNotify(w http.ResponseWriter, r *http.Request) {
	s.broadcast("reload")
	w.WriteHeader(http.StatusOK)
}
