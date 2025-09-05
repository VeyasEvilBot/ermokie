package models

import (
	"database/sql"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
	"github.com/stefanistkuhl/ermokie/pkg/models/styles"
)

var (
	cursorGlyph = styles.CursorGlyph
)

type selectionMsg []string
type cancelPickerMsg struct{}

type fuzzyFinder struct {
	title          string
	viewport       viewport.Model
	rows           []string
	cursor         int
	input          textinput.Model
	liveValue      string
	termW          int
	termH          int
	boxW           int
	multiMode      bool
	selection      []string
	styles         styles.Styles
	fromMainScreen bool
	db             *sql.DB
}

func newFuzzyFinder(data []string, multiMode bool, title string) (*fuzzyFinder, error) {
	ti := textinput.New()
	ti.Prompt = styles.CursorGlyph
	ti.Focus()

	vp := viewport.New(0, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder())

	f := &fuzzyFinder{
		title:     title,
		viewport:  vp,
		rows:      data,
		cursor:    0,
		input:     ti,
		termW:     0,
		termH:     0,
		boxW:      0,
		multiMode: multiMode,
		styles:    styles.DefaultStyles(),
	}
	return f, nil
}

func newFuzzyFinderWithTheme(data []string, multiMode bool, themeStyles styles.Styles, title string) (*fuzzyFinder, error) {
	return newFuzzyFinderWithThemeAndFlag(data, multiMode, themeStyles, title, false, nil)
}

func newFuzzyFinderWithThemeAndFlag(data []string, multiMode bool, themeStyles styles.Styles, title string, fromMainScreen bool, db *sql.DB) (*fuzzyFinder, error) {
	ti := textinput.New()
	ti.Prompt = styles.CursorGlyph
	ti.Focus()
	vp := viewport.New(0, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(themeStyles.Colors.Accent)

	f := &fuzzyFinder{
		title:          title,
		viewport:       vp,
		rows:           data,
		cursor:         0,
		input:          ti,
		termW:          0,
		termH:          0,
		boxW:           0,
		multiMode:      multiMode,
		styles:         themeStyles,
		fromMainScreen: fromMainScreen,
		db:             db,
	}
	return f, nil
}

func (f *fuzzyFinder) refreshContent() {
	contentWidth := max(f.viewport.Width-2, 0)
	pattern := f.liveValue
	matches := fuzzy.Find(pattern, f.rows)
	results := getMatches(matches)

	if pattern == "" {
		lines := make([]string, 0, len(f.rows)+1)
		for i, row := range f.rows {
			prefix := "  "
			if i == f.cursor {
				prefix = cursorGlyph
			}
			marker := " "
			if slices.Contains(f.selection, row) {
				marker = "│"
			}
			line := marker + " " + prefix + row
			padded := lipgloss.NewStyle().Width(contentWidth).Render(line)
			if i == f.cursor {
				lines = append(lines, f.styles.Title.Render(padded))
			} else if slices.Contains(f.selection, row) {
				lines = append(lines, lipgloss.NewStyle().Foreground(f.styles.Colors.Highlight).Render(padded))
			} else {
				lines = append(lines, f.styles.BodyText.Render(padded))
			}
		}
		f.viewport.SetContent(strings.Join(lines, "\n"))
	} else {
		if f.cursor >= len(results) {
			f.cursor = max(0, len(results)-1)
		}
		lines := make([]string, 0, len(results)+1)
		for i, row := range results {
			prefix := "  "
			if i == f.cursor {
				prefix = cursorGlyph
			}
			marker := " "
			if slices.Contains(f.selection, row) {
				marker = "│"
			}
			line := marker + " " + prefix + row
			padded := lipgloss.NewStyle().Width(contentWidth).Render(line)
			if i == f.cursor {
				lines = append(lines, f.styles.Title.Render(padded))
			} else if slices.Contains(f.selection, row) {
				lines = append(lines, lipgloss.NewStyle().Foreground(f.styles.Colors.Highlight).Render(padded))
			} else {
				lines = append(lines, f.styles.BodyText.Render(padded))
			}
		}
		f.viewport.SetContent(strings.Join(lines, "\n"))
	}

	top := f.viewport.YOffset
	bottom := top + f.viewport.Height - 1
	if f.cursor < top {
		f.viewport.YOffset = f.cursor
	} else if f.cursor > bottom {
		f.viewport.YOffset = f.cursor - f.viewport.Height + 1
	}
}

func (f *fuzzyFinder) Init() tea.Cmd {
	return textinput.Blink
}

func (f *fuzzyFinder) SetStyles(themeStyles styles.Styles) {
	f.styles = themeStyles
	f.viewport.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(f.styles.Colors.Accent)
}

func (f *fuzzyFinder) helpView() string {
	if f.multiMode {
		return f.styles.Hint.Render("  ↑/↓: Navigate • <tab>: Toggle selection • <enter>: Confirm & quit • Control c/<esc>: Quit\n")
	}
	return f.styles.Hint.Render("  ↑/↓: Navigate • <enter>: Select & quit • Control c/<esc>: Quit\n")
}

func (f *fuzzyFinder) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.termW = msg.Width
		f.termH = msg.Height
		f.boxW = int(float64(msg.Width) * 0.5)
		// Use the same calculation as the working code
		promptCells := lipgloss.Width(f.input.Prompt)
		f.input.Width = max(f.boxW-promptCells-3, 0)
		f.viewport.Width = f.boxW
		f.viewport.Style = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(f.styles.Colors.Accent)
		f.refreshContent()
		return f, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return f, func() tea.Msg { return cancelPickerMsg{} }

		case "up":
			// Move within the currently shown list (filtered or full)
			var maxIdx int
			if f.liveValue == "" {
				maxIdx = len(f.rows) - 1
			} else {
				matches := fuzzy.Find(f.liveValue, f.rows)
				results := getMatches(matches)
				maxIdx = len(results) - 1
			}
			if maxIdx >= 0 && f.cursor > 0 {
				f.cursor--
			}
			f.refreshContent()
			return f, nil

		case "down":
			var maxIdx int
			if f.liveValue == "" {
				maxIdx = len(f.rows) - 1
			} else {
				matches := fuzzy.Find(f.liveValue, f.rows)
				results := getMatches(matches)
				maxIdx = len(results) - 1
			}
			if maxIdx >= 0 && f.cursor < maxIdx {
				f.cursor++
			}
			f.refreshContent()
			return f, nil

		case "enter":
			var choice string
			if f.liveValue == "" {
				if len(f.rows) > 0 && f.cursor >= 0 && f.cursor < len(f.rows) {
					choice = f.rows[f.cursor]
				}
			} else {
				matches := fuzzy.Find(f.liveValue, f.rows)
				results := getMatches(matches)
				if len(results) > 0 && f.cursor >= 0 && f.cursor < len(results) {
					choice = results[f.cursor]
				}
			}

			if choice != "" && !slices.Contains(f.selection, choice) {
				f.selection = append(f.selection, choice)
			}

			return f, func() tea.Msg { return selectionMsg(deduplicate(f.selection)) }

		case "tab":
			if f.multiMode {
				if len(f.rows) > 0 && f.cursor >= 0 && f.cursor < len(f.rows) {
					row := f.rows[f.cursor]
					if slices.Contains(f.selection, row) {
						f.selection = slices.DeleteFunc(f.selection, func(s string) bool {
							return s == row
						})
					} else {
						f.selection = append(f.selection, row)
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	f.input, cmd = f.input.Update(msg)
	f.liveValue = f.input.Value()
	f.refreshContent()
	return f, cmd
}

func (f *fuzzyFinder) ViewBox() string {
	gap := "\n"
	if f.title != "" {
		title := f.styles.Title.Render(f.title)
		boxedInput := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(f.styles.Colors.Accent).
			Render(f.input.View())
		return title + gap + boxedInput + gap + f.viewport.View() + gap + f.helpView()
	}
	boxedInput := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(f.styles.Colors.Accent).
		Render(f.input.View())
	return boxedInput + gap + f.viewport.View() + gap + f.helpView()
}

func (f *fuzzyFinder) View() string {
	gap := "\n"
	if f.title != "" {
		title := f.styles.Title.Render(f.title)
		boxedInput := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(f.styles.Colors.Accent).
			Render(f.input.View())
		content := title + gap + boxedInput + gap + f.viewport.View() + gap + f.helpView()
		return lipgloss.Place(
			max(f.termW, 0),
			max(f.termH, 0),
			lipgloss.Center,
			lipgloss.Center,
			content,
		)

	}
	boxedInput := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(f.styles.Colors.Accent).
		Render(f.input.View())
	content := boxedInput + gap + f.viewport.View() + gap + f.helpView()
	return lipgloss.Place(
		max(f.termW, 0),
		max(f.termH, 0),
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func getMatches(matches []fuzzy.Match) []string {
	var results = []string{}
	for _, match := range matches {
		results = append(results, match.Str)
	}
	return results
}

func deduplicate(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

func NewFuzzyFinder(input []string, multiMode bool, title string) []string {
	var result []string
	model, err := newFuzzyFinder(input, multiMode, title)

	if err != nil {
		fmt.Println("Could not initialize Bubble Tea model:", err)
		os.Exit(1)
	}

	a, err := tea.NewProgram(model).Run()
	if err != nil {
		fmt.Println("Bummer, there's been an error:", err)
		os.Exit(1)
	}

	final := a.(*fuzzyFinder)

	result = deduplicate(final.selection)
	return result
}

func NewFuzzyFinderWithTheme(input []string, multiMode bool, themeStyles styles.Styles, title string) []string {
	return NewFuzzyFinderWithThemeAndFlag(input, multiMode, themeStyles, title, false, nil)
}

func NewFuzzyFinderWithThemeAndFlag(
	input []string,
	multiMode bool,
	themeStyles styles.Styles,
	title string,
	fromMainScreen bool,
	db *sql.DB,
) []string {
	var result []string
	model, err := newFuzzyFinderWithThemeAndFlag(
		input, multiMode, themeStyles, title, fromMainScreen, db,
	)
	if err != nil {
		fmt.Println("Could not initialize Bubble Tea model:", err)
		return []string{}
	}

	var a tea.Model
	if fromMainScreen {
		a, err = tea.NewProgram(model).Run()
	} else {
		a, err = tea.NewProgram(model, tea.WithAltScreen()).Run()
	}
	if err != nil {
		return []string{}
	}

	final := a.(*fuzzyFinder)
	result = deduplicate(final.selection)
	return result
}
