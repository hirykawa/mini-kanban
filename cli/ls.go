package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"mini-kanban/db"
	"mini-kanban/db/dbgen"
	"mini-kanban/model"
	"mini-kanban/search"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List tasks",
	Long:  "List tasks in the current project with optional filters.",
	RunE:  runLs,
}

var (
	lsOpen  bool
	lsDone  bool
	lsTags  []string
	lsQuery string
	lsSort  string
	lsLimit int
	lsJSON  bool
)

func init() {
	lsCmd.Flags().BoolVar(&lsOpen, "open", false, "show only open tasks")
	lsCmd.Flags().BoolVar(&lsDone, "done", false, "show only done tasks")
	lsCmd.Flags().StringArrayVar(&lsTags, "tag", nil, "filter by tag")
	lsCmd.Flags().StringVarP(&lsQuery, "q", "q", "", "search query")
	lsCmd.Flags().StringVar(&lsSort, "sort", "created", "sort by: created, updated, due")
	lsCmd.Flags().IntVar(&lsLimit, "limit", 50, "max results")
	lsCmd.Flags().BoolVar(&lsJSON, "json", false, "output as NDJSON")
}

func runLs(cmd *cobra.Command, args []string) error {
	cc, err := NewCmdContext()
	if err != nil {
		return err
	}
	defer cc.Close()

	ctx := cc.Context()

	// Determine status filter
	var tasks []dbgen.ListTasksRow

	// Default: show open only unless --done specified
	if lsDone && !lsOpen {
		rows, err := cc.Queries.ListTasksDone(ctx, cc.ProjectID)
		if err != nil {
			return fmt.Errorf("list done tasks: %w", err)
		}
		for _, r := range rows {
			tasks = append(tasks, dbgen.ListTasksRow(r))
		}
	} else if lsOpen || (!lsOpen && !lsDone) {
		// Default to open
		rows, err := cc.Queries.ListTasksOpen(ctx, cc.ProjectID)
		if err != nil {
			return fmt.Errorf("list open tasks: %w", err)
		}
		for _, r := range rows {
			tasks = append(tasks, dbgen.ListTasksRow(r))
		}
	}

	// Apply search if query provided
	if lsQuery != "" {
		parsed := search.ParseQuery(lsQuery)

		// If free text, do FTS or LIKE search
		if parsed.FreeText != "" {
			if search.ShouldUseLike(parsed.FreeText) {
				searchText := parsed.FreeText
				rows, err := cc.Queries.SearchTasksLike(ctx, dbgen.SearchTasksLikeParams{
					ProjectID: cc.ProjectID,
					Column2:   &searchText,
					Column3:   &searchText,
				})
				if err != nil {
					return fmt.Errorf("search (like): %w", err)
				}
				tasks = filterBySearchLike(tasks, rows)
			} else {
				rows, err := db.SearchTasksFTS(ctx, cc.DB, cc.ProjectID, parsed.FreeText)
				if err != nil {
					return fmt.Errorf("search (fts): %w", err)
				}
				tasks = filterBySearchFTS(tasks, rows)
			}
		}

		// Apply tag filters from query
		if len(parsed.Tags) > 0 {
			lsTags = append(lsTags, parsed.Tags...)
		}

		// Apply status from query
		if parsed.Status != "" {
			if parsed.Status == "open" {
				lsOpen = true
				lsDone = false
			} else if parsed.Status == "done" {
				lsDone = true
				lsOpen = false
			}
		}
	}

	// Filter by tags
	if len(lsTags) > 0 {
		tasks, err = filterByTags(cc.Queries, tasks, lsTags)
		if err != nil {
			return err
		}
	}

	// Apply limit
	if lsLimit > 0 && len(tasks) > lsLimit {
		tasks = tasks[:lsLimit]
	}

	// Output
	if lsJSON {
		return outputJSON(cc.Queries, tasks)
	}
	return outputTable(cc.Queries, tasks)
}

func filterBySearchLike(current []dbgen.ListTasksRow, searchResults []dbgen.SearchTasksLikeRow) []dbgen.ListTasksRow {
	matchIDs := make(map[int64]bool)
	for _, row := range searchResults {
		matchIDs[row.ID] = true
	}

	var filtered []dbgen.ListTasksRow
	for _, t := range current {
		if matchIDs[t.ID] {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func filterBySearchFTS(current []dbgen.ListTasksRow, searchResults []db.TaskRow) []dbgen.ListTasksRow {
	matchIDs := make(map[int64]bool)
	for _, row := range searchResults {
		matchIDs[row.ID] = true
	}

	var filtered []dbgen.ListTasksRow
	for _, t := range current {
		if matchIDs[t.ID] {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func filterByTags(q *dbgen.Queries, tasks []dbgen.ListTasksRow, tags []string) ([]dbgen.ListTasksRow, error) {
	ctx := context.Background()
	var filtered []dbgen.ListTasksRow

	for _, t := range tasks {
		taskTags, err := q.ListTagsForTask(ctx, t.ID)
		if err != nil {
			return nil, err
		}

		tagNames := make(map[string]bool)
		for _, tt := range taskTags {
			tagNames[tt.Name] = true
		}

		// Check all required tags are present
		hasAll := true
		for _, required := range tags {
			if !tagNames[required] {
				hasAll = false
				break
			}
		}
		if hasAll {
			filtered = append(filtered, t)
		}
	}
	return filtered, nil
}

type taskOutput struct {
	ID        int64    `json:"id"`
	Project   string   `json:"project"`
	Status    string   `json:"status"`
	Title     string   `json:"title"`
	Body      string   `json:"body,omitempty"`
	Tags      []string `json:"tags"`
	DueAt     *string  `json:"dueAt,omitempty"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
	StartedAt *string  `json:"startedAt,omitempty"`
	DoneAt    *string  `json:"doneAt,omitempty"`
}

func outputJSON(q *dbgen.Queries, tasks []dbgen.ListTasksRow) error {
	ctx := context.Background()
	enc := json.NewEncoder(os.Stdout)

	// Batch fetch all tags for tasks
	taskIDs := make([]int64, len(tasks))
	for i, t := range tasks {
		taskIDs[i] = t.ID
	}
	tagMap, _ := db.GetTagsForTasks(ctx, q, taskIDs)

	for _, t := range tasks {
		tagNames := tagMap[t.ID]
		if tagNames == nil {
			tagNames = []string{}
		}

		out := taskOutput{
			ID:        t.ID,
			Project:   t.ProjectName,
			Status:    t.Status,
			Title:     t.Title,
			Body:      t.Body,
			Tags:      tagNames,
			CreatedAt: formatUnix(t.CreatedAt),
			UpdatedAt: formatUnix(t.UpdatedAt),
		}
		if t.DueAt != nil {
			s := formatUnix(*t.DueAt)
			out.DueAt = &s
		}
		if t.StartedAt != nil {
			s := formatUnix(*t.StartedAt)
			out.StartedAt = &s
		}
		if t.DoneAt != nil {
			s := formatUnix(*t.DoneAt)
			out.DoneAt = &s
		}

		enc.Encode(out)
	}
	return nil
}

func outputTable(q *dbgen.Queries, tasks []dbgen.ListTasksRow) error {
	ctx := context.Background()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATUS\tDUE\tTITLE\tTAGS")

	// Batch fetch all tags for tasks
	taskIDs := make([]int64, len(tasks))
	for i, t := range tasks {
		taskIDs[i] = t.ID
	}
	tagMap, _ := db.GetTagsForTasks(ctx, q, taskIDs)

	for _, t := range tasks {
		tagNames := tagMap[t.ID]
		if tagNames == nil {
			tagNames = []string{}
		}

		due := "-"
		if t.DueAt != nil {
			due = formatUnixShort(*t.DueAt)
		}

		status := "○"
		if t.Status == model.StatusDone.String() {
			status = "✓"
		}

		tagsStr := strings.Join(tagNames, ", ")
		title := t.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", t.ID, status, due, title, tagsStr)
	}

	return w.Flush()
}

func formatUnix(unix int64) string {
	return time.Unix(unix, 0).Local().Format(time.RFC3339)
}

func formatUnixShort(unix int64) string {
	return time.Unix(unix, 0).Local().Format("2006-01-02")
}
