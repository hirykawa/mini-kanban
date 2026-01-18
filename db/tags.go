package db

import (
	"context"

	"mini-kanban/db/dbgen"
)

// TagMap is a map of task ID to tag names.
type TagMap map[int64][]string

// GetTagsForTasks fetches tags for multiple tasks in a single query and returns a map.
// This eliminates N+1 query problems when listing tasks with tags.
func GetTagsForTasks(ctx context.Context, q *dbgen.Queries, taskIDs []int64) (TagMap, error) {
	if len(taskIDs) == 0 {
		return make(TagMap), nil
	}

	rows, err := q.ListTagsForTasks(ctx, taskIDs)
	if err != nil {
		return nil, err
	}

	result := make(TagMap)
	for _, row := range rows {
		result[row.TaskID] = append(result[row.TaskID], row.Name)
	}

	return result, nil
}
