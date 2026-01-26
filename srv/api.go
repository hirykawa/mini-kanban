package srv

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	"mini-kanban/config"
	"mini-kanban/db"
	"mini-kanban/db/dbgen"
	"mini-kanban/util"
)

// corsMiddleware wraps a handler with CORS headers.
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

func (s *Server) handleAPIProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := dbgen.New(s.DB)

	projects, err := q.ListProjects(ctx)
	if err != nil {
		slog.Error("list projects", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func (s *Server) handleAPICreateProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := dbgen.New(s.DB)

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if _, err := q.GetProjectByName(ctx, req.Name); err == nil {
		http.Error(w, "project already exists", http.StatusConflict)
		return
	}

	project, err := q.CreateProject(ctx, req.Name)
	if err != nil {
		slog.Error("create project", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	s.broadcast("reload")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

func (s *Server) getProjectID(ctx context.Context, projectName string) (int64, error) {
	q := dbgen.New(s.DB)

	if projectName == "" {
		projectName = s.DefaultProject
	}

	project, err := q.GetProjectByName(ctx, projectName)
	if err != nil {
		// Try to create the project if it doesn't exist
		project, err = q.CreateProject(ctx, projectName)
		if err != nil {
			return 0, err
		}
	}
	return project.ID, nil
}

func (s *Server) handleAPIKanban(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := dbgen.New(s.DB)

	projectName := r.URL.Query().Get("project")
	if projectName == "" {
		projectName = s.DefaultProject
	}

	projectID, err := s.getProjectID(ctx, projectName)
	if err != nil {
		slog.Error("get project", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	tasks, err := q.ListTasks(ctx, projectID)
	if err != nil {
		slog.Error("list tasks", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	projects, err := q.ListProjects(ctx)
	if err != nil {
		slog.Error("list projects", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Task response matching frontend expectations
	type taskResponse struct {
		ID        int64    `json:"id"`
		Title     string   `json:"title"`
		Body      string   `json:"body"`
		Status    string   `json:"status"`
		DueAt     *int64   `json:"dueAt,omitempty"`
		Tags      []string `json:"tags"`
		CreatedAt int64    `json:"createdAt"`
		UpdatedAt int64    `json:"updatedAt"`
		IsOverdue bool     `json:"isOverdue"`
		Project   string   `json:"project"`
	}

	type projectResponse struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	// Group tasks by status
	todoTasks := []taskResponse{}
	doingTasks := []taskResponse{}
	reviewTasks := []taskResponse{}
	doneTasks := []taskResponse{}

	now := time.Now().Unix()

	// Batch fetch all tags for tasks
	taskIDs := make([]int64, len(tasks))
	for i, t := range tasks {
		taskIDs[i] = t.ID
	}
	tagMap, _ := db.GetTagsForTasks(ctx, q, taskIDs)

	for _, t := range tasks {
		isOverdue := t.DueAt != nil && *t.DueAt < now && t.Status != "done"

		// Get tags for this task from the batch
		tagNames := tagMap[t.ID]
		if tagNames == nil {
			tagNames = []string{}
		}

		resp := taskResponse{
			ID:        t.ID,
			Title:     t.Title,
			Body:      t.Body,
			Status:    t.Status,
			DueAt:     t.DueAt,
			Tags:      tagNames,
			CreatedAt: t.CreatedAt,
			UpdatedAt: t.UpdatedAt,
			IsOverdue: isOverdue,
			Project:   t.ProjectName,
		}

		switch t.Status {
		case "todo":
			todoTasks = append(todoTasks, resp)
		case "doing":
			doingTasks = append(doingTasks, resp)
		case "review":
			reviewTasks = append(reviewTasks, resp)
		case "done":
			doneTasks = append(doneTasks, resp)
		}
	}

	// Get top tags for project
	topTags, _ := q.TopTagsForProject(ctx, dbgen.TopTagsForProjectParams{
		ProjectID: projectID,
		Limit:     20,
	})
	tagList := make([]string, len(topTags))
	for i, t := range topTags {
		tagList[i] = t.Name
	}

	// Build project list
	projectList := make([]projectResponse, len(projects))
	for i, p := range projects {
		projectList[i] = projectResponse{ID: p.ID, Name: p.Name}
	}

	response := struct {
		Projects        []projectResponse `json:"projects"`
		CurrentProject  string            `json:"currentProject"`
		CurrentProjects []string          `json:"currentProjects"`
		TodoTasks       []taskResponse    `json:"todoTasks"`
		DoingTasks      []taskResponse    `json:"doingTasks"`
		ReviewTasks     []taskResponse    `json:"reviewTasks"`
		DoneTasks       []taskResponse    `json:"doneTasks"`
		Tags            []string          `json:"tags"`
	}{
		Projects:        projectList,
		CurrentProject:  projectName,
		CurrentProjects: []string{projectName},
		TodoTasks:       todoTasks,
		DoingTasks:      doingTasks,
		ReviewTasks:     reviewTasks,
		DoneTasks:       doneTasks,
		Tags:            tagList,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleAPICreateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := dbgen.New(s.DB)

	var req struct {
		Project string   `json:"project"`
		Title   string   `json:"title"`
		Body    string   `json:"body"`
		Tags    []string `json:"tags"`
		DueAt   string   `json:"dueAt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	projectID, err := s.getProjectID(ctx, req.Project)
	if err != nil {
		slog.Error("get project", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	params := dbgen.CreateTaskParams{
		ProjectID: projectID,
		Title:     req.Title,
		Body:      req.Body,
		Status:    "todo",
	}

	if req.DueAt != "" {
		dueDate, err := util.ParseDateTime(req.DueAt)
		if err != nil {
			http.Error(w, "invalid due_date", http.StatusBadRequest)
			return
		}
		dueUnix := dueDate.Unix()
		params.DueAt = &dueUnix
	}

	task, err := q.CreateTask(ctx, params)
	if err != nil {
		slog.Error("create task", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Add tags
	for _, tagName := range req.Tags {
		tag, err := q.GetOrCreateTag(ctx, tagName)
		if err != nil {
			slog.Error("get or create tag", "error", err)
			continue
		}
		_ = q.AddTagToTask(ctx, dbgen.AddTagToTaskParams{
			TaskID: task.ID,
			TagID:  tag.ID,
		})
	}

	s.broadcast("reload")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (s *Server) handleAPIUpdateTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := dbgen.New(s.DB)

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req struct {
		Project string `json:"project"`
		Title   string `json:"title"`
		Body    string `json:"body"`
		Status  string `json:"status"`
		DueAt   string `json:"dueAt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	projectID, err := s.getProjectID(ctx, req.Project)
	if err != nil {
		slog.Error("get project", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Update status if provided
	if req.Status != "" {
		_, err = q.UpdateTaskStatus(ctx, dbgen.UpdateTaskStatusParams{
			ID:        id,
			ProjectID: projectID,
			Status:    req.Status,
		})
		if err != nil {
			slog.Error("update task status", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	// Update other fields
	var clearDue interface{} = 0
	var dueAt *int64
	if req.DueAt == "" {
		clearDue = 1
	} else {
		dueDate, err := util.ParseDateTime(req.DueAt)
		if err != nil {
			http.Error(w, "invalid dueAt", http.StatusBadRequest)
			return
		}
		dueUnix := dueDate.Unix()
		dueAt = &dueUnix
	}

	task, err := q.UpdateTask(ctx, dbgen.UpdateTaskParams{
		ID:        id,
		ProjectID: projectID,
		Title:     &req.Title,
		Body:      &req.Body,
		ClearDue:  clearDue,
		DueAt:     dueAt,
	})
	if err != nil {
		slog.Error("update task", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	s.broadcast("reload")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (s *Server) handleAPIDeleteTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := dbgen.New(s.DB)

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	projectName := r.URL.Query().Get("project")
	projectID, err := s.getProjectID(ctx, projectName)
	if err != nil {
		slog.Error("get project", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if err := q.DeleteTask(ctx, dbgen.DeleteTaskParams{ID: id, ProjectID: projectID}); err != nil {
		slog.Error("delete task", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	s.broadcast("reload")
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAPIUpdateStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := dbgen.New(s.DB)

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req struct {
		Project string `json:"project"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	projectID, err := s.getProjectID(ctx, req.Project)
	if err != nil {
		slog.Error("get project", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	oldTask, err := q.GetTask(ctx, dbgen.GetTaskParams{
		ID:        id,
		ProjectID: projectID,
	})
	oldStatus := ""
	if err == nil {
		oldStatus = oldTask.Status
	}

	task, err := q.UpdateTaskStatus(ctx, dbgen.UpdateTaskStatusParams{
		ID:        id,
		ProjectID: projectID,
		Status:    req.Status,
	})
	if err != nil {
		slog.Error("update task status", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	s.broadcast("reload")

	if oldStatus != req.Status && (req.Status == "review" || req.Status == "done") {
		s.broadcastNotification(SSENotification{
			Type:      "status_change",
			TaskID:    id,
			TaskTitle: task.Title,
			NewStatus: req.Status,
			OldStatus: oldStatus,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (s *Server) handleAPIGetConfig(w http.ResponseWriter, r *http.Request) {
	cwd, err := os.Getwd()
	if err != nil {
		slog.Error("get cwd", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	pc, dir, err := config.LoadProjectConfig(cwd)
	if err != nil {
		slog.Error("load project config", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response := struct {
		ProjectName string `json:"projectName"`
		Context     string `json:"context"`
		ContextFile string `json:"contextFile"`
		ConfigPath  string `json:"configPath"`
	}{
		ConfigPath: filepath.Join(dir, ".mini-kanban.toml"),
	}

	if pc != nil {
		response.ProjectName = pc.Project.Name
		response.Context = pc.AI.Context
		response.ContextFile = pc.AI.ContextFile
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleAPIUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProjectName string `json:"projectName"`
		Context     string `json:"context"`
		ContextFile string `json:"contextFile"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		slog.Error("get cwd", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	pc, dir, err := config.LoadProjectConfig(cwd)
	if err != nil {
		slog.Error("load project config", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if pc == nil {
		pc = &config.ProjectConfig{}
	}

	// Update values
	if req.ProjectName != "" {
		pc.Project.Name = req.ProjectName
	}
	pc.AI.Context = req.Context
	pc.AI.ContextFile = req.ContextFile

	// Generate TOML content
	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(pc); err != nil {
		slog.Error("encode toml", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// Write to file
	configPath := filepath.Join(dir, ".mini-kanban.toml")
	if err := os.WriteFile(configPath, buf.Bytes(), 0644); err != nil {
		slog.Error("write config", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
