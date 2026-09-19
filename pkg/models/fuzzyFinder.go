package models

import (
	"fmt"
	"slices"
	"strings"

	"codeberg.org/veya/ermokie/pkg/db"
	"codeberg.org/veya/ermokie/pkg/models/styles"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
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
	db             *db.Store
}

func newFuzzyFinderWithThemeAndFlag(data []string, multiMode bool, themeStyles styles.Styles, title string, fromMainScreen bool, db *db.Store) (*fuzzyFinder, error) {
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
		rows:           deduplicate(data),
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

func (f *fuzzyFinder) visibleRows() []string {
	if f.liveValue == "" {
		return f.rows
	}
	return getMatches(fuzzy.Find(f.liveValue, f.rows))
}

func (f *fuzzyFinder) currentChoice() (string, bool) {
	rows := f.visibleRows()
	if f.cursor < 0 || f.cursor >= len(rows) {
		return "", false
	}
	return rows[f.cursor], true
}

func (f *fuzzyFinder) toggleCurrent() {
	choice, ok := f.currentChoice()
	if !ok {
		return
	}
	if slices.Contains(f.selection, choice) {
		f.selection = slices.DeleteFunc(f.selection, func(value string) bool { return value == choice })
		return
	}
	f.selection = append(f.selection, choice)
}

func (f *fuzzyFinder) SetStyles(themeStyles styles.Styles) {
	f.styles = themeStyles
	f.viewport.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(f.styles.Colors.Accent)
}

func (f *fuzzyFinder) helpView() string {
	up := GlobalKeybindingManager.GetKeysForAction(ActionUp)[0]
	down := GlobalKeybindingManager.GetKeysForAction(ActionDown)[0]
	confirm := GlobalKeybindingManager.GetKeysForAction(ActionConfirm)[0]
	cancel := GlobalKeybindingManager.GetKeysForAction(ActionCancel)[0]
	if f.multiMode {
		selectKey := GlobalKeybindingManager.GetKeysForAction(ActionSelect)[0]
		return f.styles.Hint.Render(fmt.Sprintf("  %s/%s: Navigate • %s: Toggle • %s: Confirm • %s: Cancel\n", up, down, selectKey, confirm, cancel))
	}
	return f.styles.Hint.Render(fmt.Sprintf("  %s/%s: Navigate • %s: Select • %s: Cancel\n", up, down, confirm, cancel))
}

func (f *fuzzyFinder) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.termW = max(msg.Width, 1)
		f.termH = max(msg.Height, 1)
		f.boxW = max(int(float64(msg.Width)*0.5), 24)
		promptCells := lipgloss.Width(f.input.Prompt)
		f.input.Width = max(f.boxW-promptCells-3, 1)
		f.viewport.Width = f.boxW
		f.viewport.Height = max(min(20, msg.Height-8), 3)
		f.viewport.Style = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(f.styles.Colors.Accent)
		f.refreshContent()
		return f, nil

	case selectionMsg:
		if !f.fromMainScreen {
			return f, tea.Quit
		}

	case tea.KeyMsg:
		if MatchesCancel(msg) || msg.String() == "ctrl+c" {
			if f.fromMainScreen {
				return f, func() tea.Msg { return cancelPickerMsg{} }
			}
			return f, tea.Quit
		}

		if MatchesFuzzyUp(msg) {
			if f.cursor > 0 {
				f.cursor--
			}
			f.refreshContent()
			return f, nil
		}

		if MatchesFuzzyDown(msg) {
			maxIdx := len(f.visibleRows()) - 1
			if maxIdx >= 0 && f.cursor < maxIdx {
				f.cursor++
			}
			f.refreshContent()
			return f, nil
		}

		if MatchesConfirm(msg) {
			choice, ok := f.currentChoice()
			if !ok && (!f.multiMode || len(f.selection) == 0) {
				return f, nil
			}
			if f.multiMode {
				if ok && !slices.Contains(f.selection, choice) {
					f.selection = append(f.selection, choice)
				}
				return f, func() tea.Msg { return selectionMsg(slices.Clone(f.selection)) }
			}
			f.selection = []string{choice}
			return f, func() tea.Msg { return selectionMsg(slices.Clone(f.selection)) }
		}

		if f.multiMode && MatchesSelect(msg) {
			f.toggleCurrent()
			f.refreshContent()
			return f, nil
		}

		previous := f.input.Value()
		var cmd tea.Cmd
		f.input, cmd = f.input.Update(msg)
		f.liveValue = f.input.Value()
		if f.liveValue != previous {
			f.cursor = 0
		}
		f.refreshContent()
		return f, cmd
	}

	return f, nil
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

func NewFuzzyFinderWithTheme(input []string, multiMode bool, themeStyles styles.Styles, title string) []string {
	return NewFuzzyFinderWithThemeAndFlag(input, multiMode, themeStyles, title, false, nil)
}

func NewFuzzyFinderWithThemeAndFlag(
	input []string,
	multiMode bool,
	themeStyles styles.Styles,
	title string,
	fromMainScreen bool,
	db *db.Store,
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
