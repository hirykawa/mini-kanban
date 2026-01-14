package srv

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"mini-kanban/db/dbgen"
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
	for _, t := range tasks {
		isOverdue := t.DueAt != nil && *t.DueAt < now && t.Status != "done"

		// Get tags for this task
		tags, _ := q.ListTagsForTask(ctx, t.ID)
		tagNames := make([]string, len(tags))
		for i, tag := range tags {
			tagNames[i] = tag.Name
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
		Projects       []projectResponse `json:"projects"`
		CurrentProject string            `json:"currentProject"`
		TodoTasks      []taskResponse    `json:"todoTasks"`
		DoingTasks     []taskResponse    `json:"doingTasks"`
		ReviewTasks    []taskResponse    `json:"reviewTasks"`
		DoneTasks      []taskResponse    `json:"doneTasks"`
		Tags           []string          `json:"tags"`
	}{
		Projects:       projectList,
		CurrentProject: projectName,
		TodoTasks:      todoTasks,
		DoingTasks:     doingTasks,
		ReviewTasks:    reviewTasks,
		DoneTasks:      doneTasks,
		Tags:           tagList,
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
		dueDate, err := parseDateTimeWeb(req.DueAt)
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
		DueDate string `json:"due_date"`
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
	if req.DueDate == "" {
		clearDue = 1
	} else {
		dueDate, err := parseDateTimeWeb(req.DueDate)
		if err != nil {
			http.Error(w, "invalid due_date", http.StatusBadRequest)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}
