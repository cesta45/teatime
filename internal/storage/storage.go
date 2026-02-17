package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Reminder represents a missing summary that the user should write.
type Reminder struct {
	Category Category // e.g. CategoryWeekly
	Name     string   // e.g. "2025-W02"
	Label    string   // human-friendly, e.g. "Weekly summary for 2025-W02"
}

// Category represents a type of note (daily, weekly, monthly, quarterly, yearly).
type Category string

const (
	CategoryDaily     Category = "days"
	CategoryWeekly    Category = "weeks"
	CategoryMonthly   Category = "months"
	CategoryQuarterly Category = "quarters"
	CategoryYearly    Category = "years"
)

var AllCategories = []Category{
	CategoryDaily,
	CategoryWeekly,
	CategoryMonthly,
	CategoryQuarterly,
	CategoryYearly,
}

// Store handles all file system operations for teatime.
type Store struct {
	Root string // ~/.teatime
}

// New creates a new Store rooted at ~/.teatime.
// It ensures the root directory exists.
func New() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine home directory: %w", err)
	}
	root := filepath.Join(home, ".teatime")
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, fmt.Errorf("could not create teatime directory: %w", err)
	}
	return &Store{Root: root}, nil
}

// --- Projects ---

// ListProjects returns the names of all project directories, sorted alphabetically.
func (s *Store) ListProjects() ([]string, error) {
	entries, err := os.ReadDir(s.Root)
	if err != nil {
		return nil, fmt.Errorf("could not read teatime directory: %w", err)
	}
	var projects []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			projects = append(projects, e.Name())
		}
	}
	sort.Strings(projects)
	return projects, nil
}

// CreateProject creates a new project directory with all category subdirectories.
func (s *Store) CreateProject(name string) error {
	name = sanitizeName(name)
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	projectDir := filepath.Join(s.Root, name)
	for _, cat := range AllCategories {
		dir := filepath.Join(projectDir, string(cat))
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("could not create directory %s: %w", dir, err)
		}
	}
	return nil
}

// DeleteProject removes a project directory and all its contents.
func (s *Store) DeleteProject(name string) error {
	projectDir := filepath.Join(s.Root, name)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return fmt.Errorf("project %q does not exist", name)
	}
	return os.RemoveAll(projectDir)
}

// ProjectExists checks whether a project directory exists.
func (s *Store) ProjectExists(name string) bool {
	projectDir := filepath.Join(s.Root, name)
	info, err := os.Stat(projectDir)
	return err == nil && info.IsDir()
}

// --- Notes ---

// NoteFile represents a single markdown note file.
type NoteFile struct {
	Name     string   // filename without extension, e.g. "2025-01-15"
	Category Category // which category this belongs to
	Path     string   // full path on disk
}

// ListNotes returns all note files for a project in a given category,
// sorted by name descending (most recent first).
func (s *Store) ListNotes(project string, category Category) ([]NoteFile, error) {
	dir := filepath.Join(s.Root, project, string(category))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("could not ensure directory exists: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not read directory %s: %w", dir, err)
	}

	var notes []NoteFile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".md")
		notes = append(notes, NoteFile{
			Name:     name,
			Category: category,
			Path:     filepath.Join(dir, e.Name()),
		})
	}

	// Sort descending (most recent first)
	sort.Slice(notes, func(i, j int) bool {
		return notes[i].Name > notes[j].Name
	})

	return notes, nil
}

// ReadNote reads the content of a note file.
func (s *Store) ReadNote(project string, category Category, name string) (string, error) {
	path := s.notePath(project, category, name)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // return empty string for non-existent notes
		}
		return "", fmt.Errorf("could not read note: %w", err)
	}
	return string(data), nil
}

// WriteNote writes content to a note file, creating it if necessary.
func (s *Store) WriteNote(project string, category Category, name string, content string) error {
	dir := filepath.Join(s.Root, project, string(category))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("could not ensure directory exists: %w", err)
	}
	path := s.notePath(project, category, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("could not write note: %w", err)
	}
	return nil
}

// DeleteNote removes a note file.
func (s *Store) DeleteNote(project string, category Category, name string) error {
	path := s.notePath(project, category, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("note %q does not exist", name)
	}
	return os.Remove(path)
}

// NoteExists checks whether a note file exists.
func (s *Store) NoteExists(project string, category Category, name string) bool {
	path := s.notePath(project, category, name)
	_, err := os.Stat(path)
	return err == nil
}

// --- Name generators ---

// TodayName returns today's date as a note name (e.g. "2025-01-15").
func TodayName() string {
	return time.Now().Format("2006-01-02")
}

// CurrentWeekName returns the current ISO week name (e.g. "2025-W03").
func CurrentWeekName() string {
	year, week := time.Now().ISOWeek()
	return fmt.Sprintf("%d-W%02d", year, week)
}

// CurrentMonthName returns the current month name (e.g. "2025-01").
func CurrentMonthName() string {
	return time.Now().Format("2006-01")
}

// CurrentQuarterName returns the current quarter name (e.g. "2025-Q1").
func CurrentQuarterName() string {
	now := time.Now()
	quarter := (now.Month()-1)/3 + 1
	return fmt.Sprintf("%d-Q%d", now.Year(), quarter)
}

// CurrentYearName returns the current year name (e.g. "2025").
func CurrentYearName() string {
	return time.Now().Format("2006")
}

// DefaultNameForCategory returns the default new-note name for a given category.
func DefaultNameForCategory(cat Category) string {
	switch cat {
	case CategoryDaily:
		return TodayName()
	case CategoryWeekly:
		return CurrentWeekName()
	case CategoryMonthly:
		return CurrentMonthName()
	case CategoryQuarterly:
		return CurrentQuarterName()
	case CategoryYearly:
		return CurrentYearName()
	default:
		return TodayName()
	}
}

// CategoryLabel returns a human-friendly label for a category.
func CategoryLabel(cat Category) string {
	switch cat {
	case CategoryDaily:
		return "Daily Notes"
	case CategoryWeekly:
		return "Weekly Notes"
	case CategoryMonthly:
		return "Monthly Notes"
	case CategoryQuarterly:
		return "Quarterly Notes"
	case CategoryYearly:
		return "Yearly Notes"
	default:
		return string(cat)
	}
}

// --- Reference Content ---

// GatherReferenceContent collects the content from the level below a summary
// and returns it as a single formatted string for display alongside the editor.
//
// Weekly  → daily entries for that week
// Monthly → weekly summaries for that month
// Quarterly → monthly summaries for that quarter
// Yearly  → quarterly summaries for that year
func (s *Store) GatherReferenceContent(project string, category Category, name string) (string, error) {
	switch category {
	case CategoryWeekly:
		return s.gatherDailyForWeek(project, name)
	case CategoryMonthly:
		return s.gatherWeeklyForMonth(project, name)
	case CategoryQuarterly:
		return s.gatherMonthlyForQuarter(project, name)
	case CategoryYearly:
		return s.gatherQuarterlyForYear(project, name)
	default:
		return "", nil
	}
}

// gatherDailyForWeek collects all daily entries that fall in the given ISO week.
// name is like "2025-W33".
func (s *Store) gatherDailyForWeek(project, name string) (string, error) {
	monday, err := mondayOfISOWeek(name)
	if err != nil {
		return "", fmt.Errorf("could not parse week %q: %w", name, err)
	}

	var parts []string
	for i := 0; i < 7; i++ {
		day := monday.AddDate(0, 0, i)
		dayName := day.Format("2006-01-02")
		content, err := s.ReadNote(project, CategoryDaily, dayName)
		if err != nil {
			return "", err
		}
		if content != "" {
			header := fmt.Sprintf("── %s (%s) ──", dayName, day.Weekday().String())
			parts = append(parts, header+"\n"+content)
		}
	}

	if len(parts) == 0 {
		return "(no daily entries for this week)", nil
	}
	return strings.Join(parts, "\n\n"), nil
}

// gatherWeeklyForMonth collects all weekly summaries whose ISO week overlaps
// with the given month. name is like "2025-08".
func (s *Store) gatherWeeklyForMonth(project, name string) (string, error) {
	t, err := time.Parse("2006-01", name)
	if err != nil {
		return "", fmt.Errorf("could not parse month %q: %w", name, err)
	}
	year, month := t.Year(), t.Month()

	// Find all weeks that have at least one day in this month.
	// Walk every day of the month and collect unique ISO weeks.
	seen := make(map[string]bool)
	var weekNames []string
	daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, t.Location()).Day()
	for d := 1; d <= daysInMonth; d++ {
		day := time.Date(year, month, d, 0, 0, 0, 0, t.Location())
		wy, wn := day.ISOWeek()
		wk := fmt.Sprintf("%d-W%02d", wy, wn)
		if !seen[wk] {
			seen[wk] = true
			weekNames = append(weekNames, wk)
		}
	}

	var parts []string
	for _, wk := range weekNames {
		content, err := s.ReadNote(project, CategoryWeekly, wk)
		if err != nil {
			return "", err
		}
		header := fmt.Sprintf("── %s ──", wk)
		if content != "" {
			parts = append(parts, header+"\n"+content)
		} else {
			parts = append(parts, header+"\n(no summary)")
		}
	}

	if len(parts) == 0 {
		return "(no weekly summaries for this month)", nil
	}
	return strings.Join(parts, "\n\n"), nil
}

// gatherMonthlyForQuarter collects the 3 monthly summaries for the given quarter.
// name is like "2025-Q3".
func (s *Store) gatherMonthlyForQuarter(project, name string) (string, error) {
	var year, q int
	_, err := fmt.Sscanf(name, "%d-Q%d", &year, &q)
	if err != nil {
		return "", fmt.Errorf("could not parse quarter %q: %w", name, err)
	}
	startMonth := (q-1)*3 + 1 // Q1→1, Q2→4, Q3→7, Q4→10

	var parts []string
	for i := 0; i < 3; i++ {
		m := time.Month(startMonth + i)
		monthName := fmt.Sprintf("%d-%02d", year, m)
		content, err := s.ReadNote(project, CategoryMonthly, monthName)
		if err != nil {
			return "", err
		}
		header := fmt.Sprintf("── %s (%s) ──", monthName, m.String())
		if content != "" {
			parts = append(parts, header+"\n"+content)
		} else {
			parts = append(parts, header+"\n(no summary)")
		}
	}

	return strings.Join(parts, "\n\n"), nil
}

// gatherQuarterlyForYear collects the 4 quarterly summaries for the given year.
// name is like "2025".
func (s *Store) gatherQuarterlyForYear(project, name string) (string, error) {
	var year int
	_, err := fmt.Sscanf(name, "%d", &year)
	if err != nil {
		return "", fmt.Errorf("could not parse year %q: %w", name, err)
	}

	var parts []string
	for q := 1; q <= 4; q++ {
		qName := fmt.Sprintf("%d-Q%d", year, q)
		content, err := s.ReadNote(project, CategoryQuarterly, qName)
		if err != nil {
			return "", err
		}
		header := fmt.Sprintf("── %s ──", qName)
		if content != "" {
			parts = append(parts, header+"\n"+content)
		} else {
			parts = append(parts, header+"\n(no summary)")
		}
	}

	return strings.Join(parts, "\n\n"), nil
}

// mondayOfISOWeek parses an ISO week string like "2025-W33" and returns
// the Monday of that week.
func mondayOfISOWeek(name string) (time.Time, error) {
	var year, week int
	_, err := fmt.Sscanf(name, "%d-W%d", &year, &week)
	if err != nil {
		return time.Time{}, err
	}

	// Jan 4 is always in ISO week 1.
	jan4 := time.Date(year, 1, 4, 0, 0, 0, 0, time.Local)
	// Find the Monday of week 1.
	weekday := jan4.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	mondayWeek1 := jan4.AddDate(0, 0, -int(weekday-time.Monday))
	// Offset to the target week.
	monday := mondayWeek1.AddDate(0, 0, (week-1)*7)
	return monday, nil
}

// --- Reminders ---

// CheckMissingSummaries scans all daily entries for a project and finds every
// past period (week, month, quarter, year) that has entries but no corresponding
// summary file. This catches ALL missing summaries, not just the immediately
// previous period.
func (s *Store) CheckMissingSummaries(project string) ([]Reminder, error) {
	notes, err := s.ListNotes(project, CategoryDaily)
	if err != nil {
		return nil, err
	}
	if len(notes) == 0 {
		return nil, nil
	}

	// Parse all daily entry dates
	var dates []time.Time
	for _, n := range notes {
		d, err := time.Parse("2006-01-02", n.Name)
		if err != nil {
			continue
		}
		dates = append(dates, d)
	}
	if len(dates) == 0 {
		return nil, nil
	}

	now := time.Now()
	currentWeekYear, currentWeek := now.ISOWeek()
	currentMonth := now.Format("2006-01")
	currentQuarter := quarterName(now)
	currentYear := now.Format("2006")

	// Collect unique periods that have daily entries
	weeks := make(map[string]bool)
	months := make(map[string]bool)
	quarters := make(map[string]bool)
	years := make(map[string]bool)

	for _, d := range dates {
		wy, wn := d.ISOWeek()
		weekKey := fmt.Sprintf("%d-W%02d", wy, wn)
		monthKey := d.Format("2006-01")
		quarterKey := quarterName(d)
		yearKey := d.Format("2006")

		// Only consider past periods, not the current one
		if weekKey != fmt.Sprintf("%d-W%02d", currentWeekYear, currentWeek) {
			weeks[weekKey] = true
		}
		if monthKey != currentMonth {
			months[monthKey] = true
		}
		if quarterKey != currentQuarter {
			quarters[quarterKey] = true
		}
		if yearKey != currentYear {
			years[yearKey] = true
		}
	}

	var reminders []Reminder

	// Check each period for a missing summary
	for name := range weeks {
		if !s.NoteExists(project, CategoryWeekly, name) {
			reminders = append(reminders, Reminder{
				Category: CategoryWeekly,
				Name:     name,
				Label:    "Weekly summary for " + name,
			})
		}
	}
	for name := range months {
		if !s.NoteExists(project, CategoryMonthly, name) {
			reminders = append(reminders, Reminder{
				Category: CategoryMonthly,
				Name:     name,
				Label:    "Monthly summary for " + name,
			})
		}
	}
	for name := range quarters {
		if !s.NoteExists(project, CategoryQuarterly, name) {
			reminders = append(reminders, Reminder{
				Category: CategoryQuarterly,
				Name:     name,
				Label:    "Quarterly summary for " + name,
			})
		}
	}
	for name := range years {
		if !s.NoteExists(project, CategoryYearly, name) {
			reminders = append(reminders, Reminder{
				Category: CategoryYearly,
				Name:     name,
				Label:    "Yearly summary for " + name,
			})
		}
	}

	// Sort reminders by category order, then by name descending (most recent first)
	sort.Slice(reminders, func(i, j int) bool {
		ci := categoryOrder(reminders[i].Category)
		cj := categoryOrder(reminders[j].Category)
		if ci != cj {
			return ci < cj
		}
		return reminders[i].Name > reminders[j].Name
	})

	return reminders, nil
}

// quarterName returns the quarter name for a given time (e.g. "2025-Q3").
// Q1 = Jan-Mar, Q2 = Apr-Jun, Q3 = Jul-Sep, Q4 = Oct-Dec.
func quarterName(t time.Time) string {
	q := (t.Month()-1)/3 + 1
	return fmt.Sprintf("%d-Q%d", t.Year(), q)
}

// categoryOrder returns a sort key so reminders are grouped by category.
func categoryOrder(c Category) int {
	switch c {
	case CategoryWeekly:
		return 0
	case CategoryMonthly:
		return 1
	case CategoryQuarterly:
		return 2
	case CategoryYearly:
		return 3
	default:
		return 4
	}
}

// --- Helpers ---

func (s *Store) notePath(project string, category Category, name string) string {
	return filepath.Join(s.Root, project, string(category), name+".md")
}

// sanitizeName cleans up a project name: lowercase, replace spaces with hyphens,
// remove anything that isn't alphanumeric, hyphen, or underscore.
func sanitizeName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	var clean strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			clean.WriteRune(r)
		}
	}
	return clean.String()
}
