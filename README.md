# 🍵 teatime

A terminal-based time-tracking journal built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea). Write daily work notes as markdown files, then distill them into weekly, monthly, quarterly, and yearly summaries. Made entirely with the use of AI.

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-green)

## Overview

Teatime is not a stopwatch — it's a **structured work journal**. You open the TUI, pick a project, and write free-text notes about what you did today. Entries are stored as plain markdown files organized by project and time period.

Summary files (weekly, monthly, quarterly, yearly) are **user-written**. The split-pane editor shows your lower-level entries as reference material so you can synthesize them into a summary — either by hand or with help from an LLM.

## Features

- **Multi-project support** — track work across as many projects as you need
- **Daily notes** — full-width editor for today's entry
- **Hierarchical summaries** — weekly, monthly, quarterly, and yearly summary files
- **Split-pane editor** — write summaries with reference entries visible alongside
- **Smart reminders** — automatically detects missing summaries for past periods
- **Interactive reminders** — press Enter on a reminder to jump straight into writing that summary
- **Plain markdown storage** — all data is human-readable files under `~/.teatime`
- **Keyboard-driven** — no mouse needed

## Installation

### From source

```sh
git clone https://github.com/cesta45/teatime.git
cd teatime
go build -o teatime .
```

Move the binary somewhere on your `$PATH`:

```sh
mv teatime ~/.local/bin/
```

### Run directly

```sh
go run .
```

## Usage

Launch the TUI:

```sh
teatime
```

The app opens in an alternate screen. Create a project, then start journaling.

### Typical workflow

1. **Start of day** — open teatime, select your project, press `e` to edit today's note
2. **Throughout the day** — jot down what you worked on, decisions made, blockers hit
3. **End of week** — press Enter on the "missing weekly summary" reminder, reference your daily entries in the right pane, and write a summary in the left pane
4. **End of month/quarter/year** — same flow, referencing the level below

## Keybindings

### Project List

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate projects |
| `Enter` | Select project |
| `n` | Create new project |
| `q` | Quit |

### Project View

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate menu & reminders |
| `Enter` | Select menu item or open reminder |
| `e` | Edit today's note |
| `d` | Browse daily notes |
| `w` | Browse weekly summaries |
| `m` | Browse monthly summaries |
| `Q` | Browse quarterly summaries |
| `y` | Browse yearly summaries |
| `b` | Back to project list |
| `q` | Quit |

### Note List

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate notes |
| `Enter` | Edit selected note |
| `n` | Create new note |
| `b` | Back |
| `q` | Quit |

### Edit Mode — Daily (full-width)

| Key | Action |
|-----|--------|
| `Esc` | Save and close |
| `Ctrl+C` | Discard changes and close |

### Edit Mode — Summary (split-pane)

| Key | Action |
|-----|--------|
| `Tab` | Switch focus between editor and reference pane |
| `↑` / `↓` | Scroll reference pane (when focused) |
| `Esc` | Save and close |
| `Ctrl+C` | Discard changes and close |

## Storage Layout

All data is stored under `~/.teatime/` as plain markdown files:

```
~/.teatime/
├── project-alpha/
│   ├── days/
│   │   ├── 2025-01-13.md
│   │   ├── 2025-01-14.md
│   │   └── 2025-01-15.md
│   ├── weeks/
│   │   └── 2025-W03.md
│   ├── months/
│   │   └── 2025-01.md
│   ├── quarters/
│   │   └── 2025-Q1.md
│   └── years/
│       └── 2025.md
└── another-project/
    └── ...
```

### File naming conventions

| Period | Directory | Format | Example |
|--------|-----------|--------|---------|
| Daily | `days/` | `YYYY-MM-DD.md` | `2025-01-15.md` |
| Weekly | `weeks/` | `YYYY-Www.md` | `2025-W03.md` |
| Monthly | `months/` | `YYYY-MM.md` | `2025-01.md` |
| Quarterly | `quarters/` | `YYYY-Qq.md` | `2025-Q1.md` |
| Yearly | `years/` | `YYYY.md` | `2025.md` |

### Reference pane content

When editing a summary, the reference pane shows the entries from the level below:

| Editing | Reference pane shows |
|---------|---------------------|
| Weekly summary | Daily entries for that week |
| Monthly summary | Weekly summaries for that month |
| Quarterly summary | Monthly summaries for that quarter |
| Yearly summary | Quarterly summaries for that year |

## Tech Stack

| Component | Library |
|-----------|---------|
| Language | [Go](https://go.dev/) |
| TUI framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) |
| Text input | [Bubbles](https://github.com/charmbracelet/bubbles) (textarea, viewport) |
| Styling | [Lip Gloss](https://github.com/charmbracelet/lipgloss) |
| Storage | Plain markdown files on disk |

## Project Structure

```
teatime/
├── main.go                      # Entry point
├── internal/
│   ├── storage/
│   │   └── storage.go           # File system operations, naming, reminders
│   └── tui/
│       ├── model.go             # Bubble Tea model, screens, and logic
│       └── styles.go            # Lip Gloss styles and layout constants
├── go.mod
└── go.sum
```

## Development

```sh
# Run
go run .

# Build
go build -o teatime .

# Vet
go vet ./...
```

## License

MIT
