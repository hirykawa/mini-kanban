// Package model provides domain types and constants for mini-kanban.
package model

// TaskStatus represents the status of a task.
type TaskStatus string

const (
	StatusTodo   TaskStatus = "todo"
	StatusDoing  TaskStatus = "doing"
	StatusReview TaskStatus = "review"
	StatusDone   TaskStatus = "done"
)

// String returns the string representation of the status.
func (s TaskStatus) String() string {
	return string(s)
}

// IsValid returns true if the status is a valid task status.
func (s TaskStatus) IsValid() bool {
	switch s {
	case StatusTodo, StatusDoing, StatusReview, StatusDone:
		return true
	default:
		return false
	}
}

// AllStatuses returns all valid task statuses in order.
func AllStatuses() []TaskStatus {
	return []TaskStatus{StatusTodo, StatusDoing, StatusReview, StatusDone}
}

// OpenStatuses returns statuses that are considered "open" (not done).
func OpenStatuses() []TaskStatus {
	return []TaskStatus{StatusTodo, StatusDoing, StatusReview}
}
