// Package srv implements the web server for mini-kanban.
package srv

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mini-kanban/db"
	"mini-kanban/db/dbgen"
	"mini-kanban/i18n"
	"mini-kanban/search"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

// Server handles HTTP requests for mini-kanban.
type Server struct {
	DB             *sql.DB
	DefaultProject string
	templates      *template.Template
}

// New creates a new server instance.
func New(dbPath, defaultProject string) (*Server, error) {
	database, err := db.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.RunMigrations(database); err != nil {
		database.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	// Parse templates with custom functions
	funcs := template.FuncMap{
		"formatDate":  formatDate,
		"formatShort": formatShort,
		"isOverdue":   isOverdue,
	}

	tmpl, err := template.New("").Funcs(funcs).ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	return &Server{
		DB:             database,
		DefaultProject: defaultProject,
		templates:      tmpl,
	}, nil
}

// Serve starts the HTTP server.
func (s *Server) Serve(addr string) error {
	mux := http.NewServeMux()

	// Pages
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("GET /partials/list", s.handlePartialList)
	mux.HandleFunc("GET /partials/kanban", s.handlePartialKanban)

	// API
	mux.HandleFunc("POST /tasks", s.handleCreateTask)
	mux.HandleFunc("POST /tasks/{id}/done", s.handleDone)
	mux.HandleFunc("POST /tasks/{id}/undo", s.handleUndo)
	mux.HandleFunc("POST /tasks/{id}/update", s.handleUpdate)
	mux.HandleFunc("POST /tasks/{id}/delete", s.handleDelete)
	mux.HandleFunc("POST /tasks/{id}/status", s.handleUpdateStatus)

	// Static files
	staticSub, _ := fs.Sub(staticFS, "static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	slog.Info("starting web server", "addr", addr)
	return http.ListenAndServe(addr, mux)
}

type pageData struct {
	Projects       []dbgen.Project
	CurrentProject string
	Tasks          []taskView
	// Kanban columns
	TodoTasks      []taskView
	DoingTasks     []taskView
	DoneTasks      []taskView
	Query          string
	ShowDone       bool
	Tags           []string
	DuePresets     []duePreset
	T              func(string) string
}

type duePreset struct {
	Value string
	Label string
}


type taskView struct {
	ID        int64
	Title     string
	Body      string
	Status    string
	DueAt     *int64
	Tags      []string
	CreatedAt int64
	IsDone    bool
	IsOverdue bool
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")
	if project == "" {
		project = s.DefaultProject
	}

	lang := i18n.DetectLanguageFromRequest(r)
	data, err := s.buildKanbanData(r.Context(), project, lang)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		slog.Error("render template", "error", err)
	}
}

func (s *Server) handlePartialList(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")
	if project == "" {
		project = s.DefaultProject
	}

	lang := i18n.DetectLanguageFromRequest(r)
	data, err := s.buildPageData(r.Context(), project, r.URL.Query(), lang)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "task_list.html", data); err != nil {
		slog.Error("render partial", "error", err)
	}
}

func (s *Server) buildPageData(ctx context.Context, projectName string, query map[string][]string, lang string) (*pageData, error) {
	q := dbgen.New(s.DB)

	// Get all projects
	projects, err := q.ListProjects(ctx)
	if err != nil {
		return nil, err
	}

	// Ensure project exists
	proj, err := q.GetProjectByName(ctx, projectName)
	if err == sql.ErrNoRows {
		proj, err = q.CreateProject(ctx, projectName)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Get query params
	searchQuery := ""
	if qs := query["q"]; len(qs) > 0 {
		searchQuery = qs[0]
	}
	showDone := false
	if sd := query["done"]; len(sd) > 0 && sd[0] == "1" {
		showDone = true
	}

	// Get tasks
	var tasks []dbgen.ListTasksRow
	if showDone {
		rows, err := q.ListTasksDone(ctx, proj.ID)
		if err != nil {
			return nil, err
		}
		for _, r := range rows {
			tasks = append(tasks, dbgen.ListTasksRow(r))
		}
	} else {
		rows, err := q.ListTasksOpen(ctx, proj.ID)
		if err != nil {
			return nil, err
		}
		for _, r := range rows {
			tasks = append(tasks, dbgen.ListTasksRow(r))
		}
	}

	// Apply search
	if searchQuery != "" {
		parsed := search.ParseQuery(searchQuery)
		if parsed.FreeText != "" {
			if search.ShouldUseLike(parsed.FreeText) {
				searchText := parsed.FreeText
				rows, err := q.SearchTasksLike(ctx, dbgen.SearchTasksLikeParams{
					ProjectID: proj.ID,
					Column2:   &searchText,
					Column3:   &searchText,
				})
				if err != nil {
					return nil, err
				}
				matchIDs := make(map[int64]bool)
				for _, r := range rows {
					matchIDs[r.ID] = true
				}
				var filtered []dbgen.ListTasksRow
				for _, t := range tasks {
					if matchIDs[t.ID] {
						filtered = append(filtered, t)
					}
				}
				tasks = filtered
			} else {
				rows, err := db.SearchTasksFTS(ctx, s.DB, proj.ID, parsed.FreeText)
				if err != nil {
					return nil, err
				}
				matchIDs := make(map[int64]bool)
				for _, r := range rows {
					matchIDs[r.ID] = true
				}
				var filtered []dbgen.ListTasksRow
				for _, t := range tasks {
					if matchIDs[t.ID] {
						filtered = append(filtered, t)
					}
				}
				tasks = filtered
			}
		}
	}

	// Convert to view models
	var taskViews []taskView
	for _, t := range tasks {
		tags, _ := q.ListTagsForTask(ctx, t.ID)
		tagNames := make([]string, len(tags))
		for i, tag := range tags {
			tagNames[i] = tag.Name
		}

		tv := taskView{
			ID:        t.ID,
			Title:     t.Title,
			Body:      t.Body,
			Status:    t.Status,
			DueAt:     t.DueAt,
			Tags:      tagNames,
			CreatedAt: t.CreatedAt,
			IsDone:    t.Status == "done",
		}
		if t.DueAt != nil && (t.Status == "todo" || t.Status == "doing") {
			tv.IsOverdue = *t.DueAt < time.Now().Unix()
		}
		taskViews = append(taskViews, tv)
	}

	// Get all tags for filter
	allTags, _ := q.TopTagsForProject(ctx, dbgen.TopTagsForProjectParams{
		ProjectID: proj.ID,
		Limit:     20,
	})
	tagNames := make([]string, len(allTags))
	for i, t := range allTags {
		tagNames[i] = t.Name
	}

	return &pageData{
		Projects:       projects,
		CurrentProject: projectName,
		Tasks:          taskViews,
		Query:          searchQuery,
		ShowDone:       showDone,
		Tags:           tagNames,
		T: func(key string) string {
			return i18n.Translate(lang, key)
		},
	}, nil
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	project := r.FormValue("project")
	title := strings.TrimSpace(r.FormValue("title"))
	tagsStr := r.FormValue("tags")
	dueStr := r.FormValue("due")

	if title == "" {
		http.Error(w, "Title required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	q := dbgen.New(s.DB)

	proj, err := q.GetProjectByName(ctx, project)
	if err != nil {
		http.Error(w, "Project not found", http.StatusBadRequest)
		return
	}

	var dueAt *int64
	if dueStr != "" {
		t, err := parseDateTimeWeb(dueStr)
		if err == nil {
			unix := t.Unix()
			dueAt = &unix
		}
	}

	task, err := q.CreateTask(ctx, dbgen.CreateTaskParams{
		ProjectID: proj.ID,
		Title:     title,
		Body:      "",
		Status:    "todo",
		DueAt:     dueAt,
	})
	if err != nil {
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	// Handle tags
	if tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		for _, tagName := range tags {
			tagName = strings.TrimSpace(tagName)
			if tagName == "" {
				continue
			}
			tag, err := q.GetOrCreateTag(ctx, tagName)
			if err != nil {
				continue
			}
			q.AddTagToTask(ctx, dbgen.AddTagToTaskParams{
				TaskID: task.ID,
				TagID:  tag.ID,
			})
		}
	}

	// Return updated list via htmx
	w.Header().Set("HX-Trigger", "taskCreated")
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) handleDone(w http.ResponseWriter, r *http.Request) {
	s.toggleStatus(w, r, true)
}

func (s *Server) handleUndo(w http.ResponseWriter, r *http.Request) {
	s.toggleStatus(w, r, false)
}

func (s *Server) toggleStatus(w http.ResponseWriter, r *http.Request, done bool) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	project := r.URL.Query().Get("project")
	ctx := r.Context()
	q := dbgen.New(s.DB)

	proj, err := q.GetProjectByName(ctx, project)
	if err != nil {
		http.Error(w, "Project not found", http.StatusBadRequest)
		return
	}

	if done {
		_, err = q.MarkTaskDone(ctx, dbgen.MarkTaskDoneParams{
			ID:        id,
			ProjectID: proj.ID,
		})
	} else {
		_, err = q.MarkTaskOpen(ctx, dbgen.MarkTaskOpenParams{
			ID:        id,
			ProjectID: proj.ID,
		})
	}

	if err != nil {
		http.Error(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", "taskUpdated")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	project := r.FormValue("project")
	title := strings.TrimSpace(r.FormValue("title"))

	ctx := r.Context()
	q := dbgen.New(s.DB)

	proj, err := q.GetProjectByName(ctx, project)
	if err != nil {
		http.Error(w, "Project not found", http.StatusBadRequest)
		return
	}

	params := dbgen.UpdateTaskParams{
		ID:        id,
		ProjectID: proj.ID,
	}
	if title != "" {
		params.Title = &title
	}

	_, err = q.UpdateTask(ctx, params)
	if err != nil {
		http.Error(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", "taskUpdated")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	project := r.URL.Query().Get("project")
	ctx := r.Context()
	q := dbgen.New(s.DB)

	proj, err := q.GetProjectByName(ctx, project)
	if err != nil {
		http.Error(w, "Project not found", http.StatusBadRequest)
		return
	}

	err = q.DeleteTask(ctx, dbgen.DeleteTaskParams{
		ID:        id,
		ProjectID: proj.ID,
	})
	if err != nil {
		http.Error(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", "taskDeleted")
	w.WriteHeader(http.StatusOK)
}

func parseDateTimeWeb(s string) (time.Time, error) {
	loc := time.Local
	formats := []string{
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, s, loc); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse date: %s", s)
}

func formatDate(unix int64) string {
	return time.Unix(unix, 0).Local().Format("2006-01-02 15:04")
}

func formatShort(unix int64) string {
	return time.Unix(unix, 0).Local().Format("01/02")
}

func isOverdue(unix *int64) bool {
	if unix == nil {
		return false
	}
	return *unix < time.Now().Unix()
}

// JSON response helpers
type jsonResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	ID      int64  `json:"id,omitempty"`
}

func jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(jsonResponse{Success: false, Message: message})
}

func (s *Server) handleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	project := r.FormValue("project")
	status := r.FormValue("status")

	// Validate status
	if status != "todo" && status != "doing" && status != "done" {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	q := dbgen.New(s.DB)

	proj, err := q.GetProjectByName(ctx, project)
	if err != nil {
		http.Error(w, "Project not found", http.StatusBadRequest)
		return
	}

	_, err = q.UpdateTaskStatus(ctx, dbgen.UpdateTaskStatusParams{
		Status:    status,
		ID:        id,
		ProjectID: proj.ID,
	})
	if err != nil {
		http.Error(w, "Failed to update task status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", "taskUpdated")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handlePartialKanban(w http.ResponseWriter, r *http.Request) {
	project := r.URL.Query().Get("project")
	if project == "" {
		project = s.DefaultProject
	}

	lang := i18n.DetectLanguageFromRequest(r)
	data, err := s.buildKanbanData(r.Context(), project, lang)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "kanban.html", data); err != nil {
		slog.Error("render kanban partial", "error", err)
	}
}

func (s *Server) buildKanbanData(ctx context.Context, projectName string, lang string) (*pageData, error) {
	q := dbgen.New(s.DB)

	projects, err := q.ListProjects(ctx)
	if err != nil {
		return nil, err
	}

	proj, err := q.GetProjectByName(ctx, projectName)
	if err == sql.ErrNoRows {
		proj, err = q.CreateProject(ctx, projectName)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Get tasks by status
	todoRows, err := q.ListTasksTodo(ctx, proj.ID)
	if err != nil {
		return nil, err
	}
	doingRows, err := q.ListTasksDoing(ctx, proj.ID)
	if err != nil {
		return nil, err
	}
	doneRows, err := q.ListTasksDone(ctx, proj.ID)
	if err != nil {
		return nil, err
	}

	// Convert to task views
	todoTasks := convertToTaskViewsTodo(ctx, q, todoRows)
	doingTasks := convertToTaskViewsDoing(ctx, q, doingRows)
	doneTasks := convertToTaskViewsDone(ctx, q, doneRows)

	// Get due presets
	duePresets := getDuePresets(lang)

	return &pageData{
		Projects:       projects,
		CurrentProject: projectName,
		TodoTasks:      todoTasks,
		DoingTasks:     doingTasks,
		DoneTasks:      doneTasks,
		DuePresets:     duePresets,
		T: func(key string) string {
			return i18n.Translate(lang, key)
		},
	}, nil
}

func convertToTaskViewsTodo(ctx context.Context, q *dbgen.Queries, rows []dbgen.ListTasksTodoRow) []taskView {
	var views []taskView
	for _, t := range rows {
		tags, _ := q.ListTagsForTask(ctx, t.ID)
		tagNames := make([]string, len(tags))
		for i, tag := range tags {
			tagNames[i] = tag.Name
		}
		tv := taskView{
			ID:        t.ID,
			Title:     t.Title,
			Body:      t.Body,
			Status:    t.Status,
			DueAt:     t.DueAt,
			Tags:      tagNames,
			CreatedAt: t.CreatedAt,
			IsDone:    false,
		}
		if t.DueAt != nil {
			tv.IsOverdue = *t.DueAt < time.Now().Unix()
		}
		views = append(views, tv)
	}
	return views
}

func convertToTaskViewsDoing(ctx context.Context, q *dbgen.Queries, rows []dbgen.ListTasksDoingRow) []taskView {
	var views []taskView
	for _, t := range rows {
		tags, _ := q.ListTagsForTask(ctx, t.ID)
		tagNames := make([]string, len(tags))
		for i, tag := range tags {
			tagNames[i] = tag.Name
		}
		tv := taskView{
			ID:        t.ID,
			Title:     t.Title,
			Body:      t.Body,
			Status:    t.Status,
			DueAt:     t.DueAt,
			Tags:      tagNames,
			CreatedAt: t.CreatedAt,
			IsDone:    false,
		}
		if t.DueAt != nil {
			tv.IsOverdue = *t.DueAt < time.Now().Unix()
		}
		views = append(views, tv)
	}
	return views
}

func convertToTaskViewsDone(ctx context.Context, q *dbgen.Queries, rows []dbgen.ListTasksDoneRow) []taskView {
	var views []taskView
	for _, t := range rows {
		tags, _ := q.ListTagsForTask(ctx, t.ID)
		tagNames := make([]string, len(tags))
		for i, tag := range tags {
			tagNames[i] = tag.Name
		}
		tv := taskView{
			ID:        t.ID,
			Title:     t.Title,
			Body:      t.Body,
			Status:    t.Status,
			DueAt:     t.DueAt,
			Tags:      tagNames,
			CreatedAt: t.CreatedAt,
			IsDone:    true,
		}
		views = append(views, tv)
	}
	return views
}

func getDuePresets(lang string) []duePreset {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	tomorrow := today.AddDate(0, 0, 1)
	
	// End of this week (Sunday)
	daysUntilSunday := (7 - int(now.Weekday())) % 7
	if daysUntilSunday == 0 {
		daysUntilSunday = 7
	}
	endOfWeek := today.AddDate(0, 0, daysUntilSunday)
	
	// End of this month
	endOfMonth := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, now.Location())
	
	// End of next month
	endOfNextMonth := time.Date(now.Year(), now.Month()+2, 0, 23, 59, 59, 0, now.Location())

	presets := []duePreset{
		{Value: today.Format("2006-01-02"), Label: i18n.Translate(lang, "web.due.today")},
		{Value: tomorrow.Format("2006-01-02"), Label: i18n.Translate(lang, "web.due.tomorrow")},
		{Value: endOfWeek.Format("2006-01-02"), Label: i18n.Translate(lang, "web.due.this_week")},
		{Value: endOfMonth.Format("2006-01-02"), Label: i18n.Translate(lang, "web.due.this_month")},
		{Value: endOfNextMonth.Format("2006-01-02"), Label: i18n.Translate(lang, "web.due.next_month")},
	}
	return presets
}
