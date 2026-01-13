# Mini-Kanban

![Mini-Kanban](assets/mini-kanban.png)

**🌐 Language: English | [日本語](README.ja.md)**

---

A lightweight and fast task management CLI and local web application.

## Features

- 📝 Simple CLI commands for task management
- 🌐 Local Web UI (htmx-based)
- 🔍 Task search and filtering
- 📁 Project-based task management
- 💾 Offline-ready local storage with SQLite

## Installation

```bash
# Run directly with npx (no installation required)
npx mini-kanban

# Or install globally
npm install -g mini-kanban
```

## Initial Setup

For first-time use, we recommend running the `init` command to initialize the environment.

```bash
mini-kanban init
```

This will:
- Create configuration files (`~/.config/mini-kanban/config.toml`, etc.)
- Prepare the database directory

## Usage

### Adding Tasks

```bash
# Basic task addition
mini-kanban add "Task title"

# Add with notes
mini-kanban add "Task title" -n "Detailed notes"

# Add with status
mini-kanban add "Task title" -s doing
```

### Listing Tasks

```bash
# Show all tasks
mini-kanban ls

# Filter by status
mini-kanban ls -s todo
mini-kanban ls -s doing
mini-kanban ls -s done
```

### Showing Task Details

```bash
mini-kanban show <task-id>
```

### Editing Tasks

```bash
# Change title
mini-kanban edit <task-id> -t "New title"

# Change status
mini-kanban edit <task-id> -s doing

# Add notes
mini-kanban edit <task-id> -n "Note content"
```

### Completing / Undoing Tasks

```bash
# Mark as done
mini-kanban done <task-id>

# Undo completion
mini-kanban undo <task-id>
```

### Deleting Tasks

```bash
mini-kanban rm <task-id>
```

### Project Management

```bash
# List projects
mini-kanban project list

# Operate on a specific project
mini-kanban --project myproject add "Task"
```

### Web UI

```bash
# Start local web server (default: port 8000)
mini-kanban web

# Specify port
mini-kanban web -p 3000
```

Access the Web UI at `http://localhost:8000` in your browser.

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MINI_KANBAN_DB` | Database file path (takes priority) | (see below) |
| `MINI_KANBAN_PROJECT` | Default project name | `default` |
| `MINI_KANBAN_LANG` | Help language (`en`, `ja`) | (auto-detect) |

### Standard Paths and XDG Variables

The app follows standard XDG directory conventions.

- `XDG_CONFIG_HOME`: Configuration file location. Default: `~/.config/mini-kanban/`
- `XDG_DATA_HOME`: Database file location. Default: `~/.local/share/mini-kanban/tasks.db`

> [!NOTE]
> If `MINI_KANBAN_DB` is set, it takes priority as the database path.

### Internationalization (i18n)

The CLI supports help display in Japanese and English.

```bash
# Show help in Japanese
mini-kanban --lang ja --help

# Show help in English
mini-kanban --lang en --help
```

You can also set the `MINI_KANBAN_LANG` environment variable to permanently fix the language.

## Developer Information

For build instructions, directory structure, and release procedures, see the [Developer Guide](docs/DEVELOPMENT.md).

## License

MIT License
