# Teatime — CLI Time Tracking Journal

A terminal-based (TUI) journaling tool for logging daily work across projects, built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea). Entries are stored as plain markdown files organized by project and date.

---

## Concept

Teatime is not a stopwatch-style time tracker. Instead, it's a **structured work journal**. You open the TUI, pick a project, and write free-text notes about what you did. Entries are saved as markdown files, one per day per project.

Summary files (weekly, monthly, quarterly, yearly) are **user-written** — the intended workflow is to copy your daily entries, paste them into an LLM, and ask it to distill the highlights. You then paste the result back as a summary file.

---

## Storage Layout

All data lives under `~/.teatime/`.

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
├── project-beta/
│   ├── days/
│   │   └── ...
│   └── ...
└── .config.yaml          # (future) optional config
```

### File Naming

| Granularity | Path                                      | Example              |
|-------------|-------------------------------------------|----------------------|
| Daily       | `<project>/days/YYYY-MM-DD.md`            | `days/2025-01-15.md` |
| Weekly      | `<project>/weeks/YYYY-Www.md`             | `weeks/2025-W03.md`  |
| Monthly     | `<project>/months/YYYY-MM.md`             | `months/2025-01.md`  |
| Quarterly   | `<project>/quarters/YYYY-Qq.md`           | `quarters/2025-Q1.md`|

| Yearly      | `<project>/years/YYYY.md`                 | `years/2025.md`      |

---

## Markdown File Formats

### Daily Entry

```markdown
# 2025-01-15

Free-form text goes here. The user writes whatever they want.

- Worked on the API redesign
- Had a meeting with the team about deployment
- Fixed that annoying CSS bug
```

### Weekly / Monthly / Quarterly / Yearly Summaries

These are user-written files. The user copies daily entries, feeds them to an LLM to extract highlights, and pastes the result. There is no enforced format — they're just markdown files.

```markdown
# Week 3 — 2025-01-13 to 2025-01-19

Key accomplishments this week:
- Completed the API redesign and merged to main
- Set up CI pipeline for staging deploys
- Fixed 4 bugs from the backlog
```

---

## TUI Design

The TUI is launched by running `teatime` with no arguments.

### Screens

#### 1. Project List (home screen)

The landing screen. Shows all projects and lets you pick one, or create a new one.

```
┌─────────────────────────────────────┐
│  🍵 teatime                        │
│                                     │
│  Projects:                          │
│                                     │
│  > project-alpha                    │
│    project-beta                     │
│    personal-site                    │
│                                     │
│  [n] new project  [q] quit          │
└─────────────────────────────────────┘
```

#### 2. Project View (split layout)

After selecting a project, the screen splits:
- **Left pane**: menu with shortcuts
- **Right pane**: today's daily note (read-only preview)

```
┌──────────────────┬───────────────────────────────────┐
│  project-alpha   │  📅 2025-01-15                    │
│                  │                                    │
│  [e] edit today  │  - Worked on the API redesign     │
│  [d] daily notes │  - Had a meeting about deployment │
│  [w] weekly      │  - Fixed that annoying CSS bug    │
│  [m] monthly     │                                    │
│  [Q] quarterly   │                                    │
│  [y] yearly      │                                    │
│                  │                                    │
│  [b] back        │                                    │
│  [q] quit        │                                    │
└──────────────────┴───────────────────────────────────┘
```

#### 3. Edit Mode

A text area (using Bubble Tea's `textarea` component) for writing or editing the current note.

```
┌──────────────────────────────────────┐
│  project-alpha — 2025-01-15 [edit]  │
│                                      │
│  - Worked on the API redesign        │
│  - Had a meeting about deployment    │
│  - Fixed that annoying CSS bug       │
│  - Started writing tests for the     │
│    new endpoints█                    │
│                                      │
│                                      │
│  [esc] save & close  [ctrl+c] abort │
└──────────────────────────────────────┘
```

#### 4. Note List

When choosing a category (daily/weekly/monthly/quarterly/yearly), show a list of existing files. Selecting one opens it for viewing or editing.

```
┌──────────────────┬───────────────────────────────────┐
│  daily notes     │  📅 2025-01-15                    │
│                  │                                    │
│  > 2025-01-15    │  - Worked on the API redesign     │
│    2025-01-14    │  - Had a meeting about deployment │
│    2025-01-13    │  - Fixed that annoying CSS bug    │
│    2025-01-10    │                                    │
│    2025-01-09    │                                    │
│                  │                                    │
│  [e] edit        │                                    │
│  [n] new note    │                                    │
│  [b] back        │                                    │
└──────────────────┴───────────────────────────────────┘
```

---

## Tech Stack

| Component      | Choice                                                               |
|----------------|----------------------------------------------------------------------|
| Language        | Go                                                                  |
| TUI framework  | [Bubble Tea](https://github.com/charmbracelet/bubbletea)            |
| Text input      | [Bubble Tea textarea](https://github.com/charmbracelet/bubbles)     |
| Styling         | [Lip Gloss](https://github.com/charmbracelet/lipgloss)             |
| Storage         | Plain markdown files on disk                                        |

---

## Tasks

### Phase 1 — Foundation

- [ ] Initialize Go module (`go mod init github.com/user/teatime`)
- [ ] Set up project structure (`cmd/`, `internal/`, `main.go`)
- [ ] Create the storage layer
  - [ ] Initialize `~/.teatime/` on first run
  - [ ] Create / list / delete projects (folders)
  - [ ] Read / write markdown files (days, weeks, months, quarters, years)
  - [ ] List files per category (sorted, most recent first)
- [ ] Build the TUI skeleton with Bubble Tea
  - [ ] Project list screen (home)
  - [ ] Navigation between screens

### Phase 2 — Core Editing & Viewing

- [ ] Project view screen (split layout)
  - [ ] Left pane: menu with shortcuts
  - [ ] Right pane: today's daily note preview
- [ ] Edit mode
  - [ ] Integrate `textarea` bubble for free-text editing
  - [ ] Save on esc, discard on ctrl+c
  - [ ] Auto-create the daily file if it doesn't exist
- [ ] Note list screen
  - [ ] List files for any category (daily/weekly/monthly/quarterly/yearly)
  - [ ] Preview selected note in right pane
  - [ ] Open selected note for editing

### Phase 3 — Summary Workflow

- [ ] Create new summary files (weekly/monthly/quarterly/yearly)
  - [ ] Auto-name based on current date
  - [ ] Open in edit mode for pasting content
- [ ] Navigate between summaries in list view

### Phase 4 — Polish

- [ ] Styling with Lip Gloss (colors, borders, spacing)
- [ ] Responsive layout (adapt to terminal size)
- [ ] Empty states (no projects yet, no entries today, etc.)
- [ ] Error handling and user feedback
- [ ] Key-binding help bar at the bottom of each screen

### Future Ideas

- [ ] Configurable storage path (`.config.yaml`)
- [ ] Search across entries (`/` key)
- [ ] Tags / labels for entries
- [ ] Export to a single markdown or PDF
- [ ] Git auto-commit on save
- [ ] Clipboard integration (copy daily entries for LLM pasting)