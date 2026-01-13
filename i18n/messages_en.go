package i18n

// messagesEN contains English translations.
var messagesEN = map[string]string{
	// Root command
	"root.short": "Lightweight task management CLI + local web",
	"root.long":  "mini-kanban is a fast, lightweight task management tool with CLI and local web interface.",
	"root.flag.project": "project name (overrides env/config)",
	"root.flag.lang":    "language for help (en, ja)",

	// init command
	"init.use":   "init",
	"init.short": "Initialize configuration and database",
	"init.long":  "Initialize the mini-kanban configuration file and database to get started.",

	// add command
	"add.use":       "add <title>",
	"add.short":     "Add a new task",
	"add.long":      "Add a new task to the current project.",
	"add.flag.body": "task body/description",
	"add.flag.tag":  "tag (can be specified multiple times)",
	"add.flag.due":  "due date (YYYY-MM-DD or YYYY-MM-DD HH:MM)",
	"add.flag.ai":   "force AI assist on",
	"add.flag.noai": "force AI assist off",

	// ls command
	"ls.use":        "ls",
	"ls.short":      "List tasks",
	"ls.long":       "List tasks in the current project with optional filters.",
	"ls.flag.open":  "show only open tasks",
	"ls.flag.done":  "show only done tasks",
	"ls.flag.tag":   "filter by tag",
	"ls.flag.q":     "search query",
	"ls.flag.sort":  "sort by: created, updated, due",
	"ls.flag.limit": "max results",
	"ls.flag.json":  "output as NDJSON",

	// show command
	"show.use":   "show <task-id>",
	"show.short": "Show task details",
	"show.long":  "Display detailed information for a specific task.",

	// edit command
	"edit.use":           "edit <task-id>",
	"edit.short":         "Edit a task",
	"edit.long":          "Edit an existing task's properties.",
	"edit.flag.title":    "new title",
	"edit.flag.body":     "new body/description",
	"edit.flag.status":   "new status (open, done)",
	"edit.flag.due":      "new due date (YYYY-MM-DD or YYYY-MM-DD HH:MM)",
	"edit.flag.cleardue": "clear due date",
	"edit.flag.addtag":   "add tag",
	"edit.flag.rmtag":    "remove tag",

	// done command
	"done.use":   "done <task-id>",
	"done.short": "Mark task as done",
	"done.long":  "Mark a task as completed.",

	// undo command
	"undo.use":   "undo <task-id>",
	"undo.short": "Reopen a completed task",
	"undo.long":  "Reopen a task that was marked as done.",

	// rm command
	"rm.use":   "rm <task-id>",
	"rm.short": "Delete a task",
	"rm.long":  "Permanently delete a task.",

	// project command
	"project.use":        "project",
	"project.short":      "Manage projects",
	"project.long":       "Manage projects for organizing tasks.",
	"project.list.use":   "list",
	"project.list.short": "List all projects",

	// config command
	"config.use":        "config",
	"config.short":      "Manage configuration",
	"config.long":       "View and manage mini-kanban configuration.",
	"config.show.use":   "show",
	"config.show.short": "Show current configuration",
	"config.init.use":   "init",
	"config.init.short": "Initialize configuration file",

	// db command
	"db.use":          "db",
	"db.short":        "Database operations",
	"db.long":         "Perform database maintenance operations.",
	"db.path.use":     "path",
	"db.path.short":   "Show database path",

	// web command
	"web.use":       "web",
	"web.short":     "Start local web UI",
	"web.long":      "Start the local web server for the mini-kanban web interface.",
	"web.flag.port": "port to listen on",
	"web.flag.host": "host to bind to",

	// Web Interface
	"web.project":      "Project:",
	"web.search.ph":    "Search tasks...",
	"web.show_done":    "Show completed",
	"web.add.title":    "New task...",
	"web.add.tags":     "Tags (comma separated)",
	"web.add.btn":      "Add",
	"web.list.empty":   "No tasks found.",
	"web.status.open":  "Open",
	"web.status.done":  "Done",
	"web.overdue":      "Overdue",
	"web.created":      "Created",
	"web.footer":       "mini-kanban • Lightweight task management",
	"web.edit.save":    "Save",
	"web.edit.cancel":  "Cancel",
	"web.page.title":   "mini-kanban",

	// Kanban columns
	"web.kanban.todo":   "Todo",
	"web.kanban.doing":  "Doing",
	"web.kanban.review": "Review",
	"web.kanban.done":   "Done",

	// Due date presets
	"web.due.today":       "Today",
	"web.due.tomorrow":    "Tomorrow",
	"web.due.this_week":   "This week",
	"web.due.this_month":  "This month",
	"web.due.next_month":  "Next month",
	"web.due.none":        "No due date",
}
