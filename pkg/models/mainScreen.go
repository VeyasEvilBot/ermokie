package models

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
	"github.com/stefanistkuhl/ermokie/pkg/config"
	"github.com/stefanistkuhl/ermokie/pkg/db"
	"github.com/stefanistkuhl/ermokie/pkg/db/sqlc"
	"github.com/stefanistkuhl/ermokie/pkg/models/styles"
	"github.com/stefanistkuhl/ermokie/pkg/types"
)

type errMsg error

type rowData struct {
	Name    string
	Type    string // "run" or "split"
	ID      int
	Hits    int
	HitPB   int
	HitDiff int
}

type model struct {
	s               styles.Styles
	spinner         spinner.Model
	vp              viewport.Model
	cursor          int
	rows            []rowData
	runName         string
	runID           int
	splitID         int
	runAttempts     int
	termW           int
	termH           int
	db              *db.Store
	activeRun       *types.Run
	cfg             config.Config
	quitting        bool
	err             error
	selection       []string
	showSelection   bool
	showFuzzyPicker bool
	fuzzyOptions    []string
	picker          *fuzzyFinder
	showHelpMenu    bool
	showError       bool
}

type initDataMsg struct {
	activeRun *types.Run
	err       error
}

var QuitKeys = GetQuitKeys()

func clipCells(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= w {
		return s
	}
	tail := "…"
	return truncate.StringWithTail(s, uint(w), tail)
}

func centerNumber(num int, width int) string {
	numStr := fmt.Sprintf("%d", num)
	numWidth := len(numStr)
	if numWidth >= width {
		return numStr
	}

	totalSpaces := width - numWidth
	leftSpaces := totalSpaces / 2
	rightSpaces := totalSpaces - leftSpaces

	return fmt.Sprintf("%*s%s%*s", leftSpaces, "", numStr, rightSpaces, "")
}

func centerSignedNumber(num int, width int) string {
	numStr := fmt.Sprintf("%+d", num)
	numWidth := len(numStr)
	if numWidth >= width {
		return numStr
	}

	totalSpaces := width - numWidth
	leftSpaces := totalSpaces / 2
	rightSpaces := totalSpaces - leftSpaces

	return fmt.Sprintf("%*s%s%*s", leftSpaces, "", numStr, rightSpaces, "")
}

func centerText(text string, width int) string {
	textWidth := len(text)
	if textWidth >= width {
		return text
	}

	totalSpaces := width - textWidth
	leftSpaces := totalSpaces / 2
	rightSpaces := totalSpaces - leftSpaces

	return fmt.Sprintf("%*s%s%*s", leftSpaces, "", text, rightSpaces, "")
}

func initialModel(theme styles.Styles, db *db.Store) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	vp := viewport.New(20, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.Colors.Accent)
	s.Style = lipgloss.NewStyle().Foreground(theme.Colors.Title)
	cfg, loadCfgErr := config.LoadConfig()
	if loadCfgErr != nil {
		log.Fatal("if this gets printed i messed up")
	}
	return model{
		s:       theme,
		vp:      vp,
		cursor:  0,
		spinner: s,
		cfg:     cfg,
		db:      db,
		termW:   0,
		termH:   0,
	}
}

func (m *model) buildListWithSplits() {
	splits, fetchSplitsErr := m.db.GetSplitsByRunID(context.Background(), int64(m.runID))
	if fetchSplitsErr != nil {
		m.err = fetchSplitsErr
		m.showError = true
		return
	}
	width := m.vp.Width
	if width <= 0 {
		if m.termW > 0 {
			width = m.termW
		} else {
			width = 80
		}
	}

	lineStyle := lipgloss.NewStyle().
		Width(width - 2).
		PaddingLeft(1)

	lines := make([]string, 0, len(splits))
	for _, split := range splits {
		lines = append(lines, lineStyle.Render(split.Name))
	}
	m.rows = make([]rowData, len(splits))
	for i, s := range splits {
		m.rows[i] = rowData{
			Name:    s.Name,
			Type:    "split",
			ID:      int(s.ID),
			Hits:    int(s.HitCount.Int64),
			HitPB:   int(s.PbHitCount.Int64),
			HitDiff: int(s.Diff.Int64),
		}
	}
	m.setCursorToActiveSplit()
	m.clampCursor()
	m.rebuildList()
}

func (m *model) buildListWithRuns() {
	width := m.vp.Width
	if width <= 0 {
		if m.termW > 0 {
			width = m.termW
		} else {
			width = 80
		}
	}

	lineStyle := lipgloss.NewStyle().
		Width(width - 2).
		PaddingLeft(1)

	names, err := m.db.GetAllRunNames(context.Background())
	if err != nil {
		m.err = err
		m.showError = true
		return
	}

	lines := make([]string, 0, len(names))
	for _, name := range names {
		lines = append(lines, lineStyle.Render(name.Name))
	}
	m.rows = make([]rowData, len(names))
	for i, n := range names {
		m.rows[i] = rowData{
			Name: n.Name,
			Type: "run",
			ID:   int(n.ID),
		}
	}
	m.clampCursor()
	m.rebuildList()
}

func (m *model) rebuildList() {
	width := m.vp.Width
	if width <= 0 {
		if m.termW > 0 {
			width = m.termW
		} else {
			width = 80
		}
	}
	lineStyle := lipgloss.NewStyle().
		Width(width - 2).
		PaddingLeft(1)

	sel := lipgloss.NewStyle().
		Foreground(m.s.Colors.Highlight).
		Bold(true)

	r, err := m.db.GetActiveRun(context.Background())
	if err != nil && err != sql.ErrNoRows {
		m.err = err
		m.showError = true
	} else if err == nil {
		aRun := types.Run{
			ID:          int(r.ID),
			Name:        r.Name,
			Game:        sql.NullString{String: r.Game.String, Valid: true},
			Category:    sql.NullString{String: r.Category.String, Valid: true},
			Attempts:    int(r.Attempts.Int64),
			ActiveSplit: int(r.ActiveSplit.Int64),
		}

		m.activeRun = &aRun
	}
	leftPad := 1
	contentWidth := (width - 2) - leftPad
	lines := make([]string, 0, len(m.rows)+2)
	isSplit := false

	for i, row := range m.rows {
		isCursor := i == m.cursor

		if m.activeRun == nil && isCursor && row.Type == "run" {
			m.runID = row.ID
			run, err := m.db.GetRunByID(context.Background(), int64(m.runID))
			if err != nil {
				m.err = err
				m.showError = true
				return
			}
			m.runName = run.Name
			m.runID = int(run.ID)
		}
		switch row.Type {
		case "split":
			if row.Type == "split" {
				isSplit = true
			}

			stats := ""
			if m.cfg.General.ShowDiff {
				stats = fmt.Sprintf("│ %4d │ %2d │ %+4d │", row.Hits, row.HitPB, row.HitDiff)
			} else {
				stats = fmt.Sprintf("│ %4d │ %2d │", row.Hits, row.HitPB)
			}
			statsW := lipgloss.Width(stats)
			if statsW > contentWidth {
				stats = clipCells(stats, contentWidth)
				statsW = lipgloss.Width(stats)
			}

			leftSpace := contentWidth - statsW - 1
			leftSpace = max(leftSpace, 0)

			leftText := row.Name
			if isCursor {
				leftText = sel.Render(styles.CursorGlyph + leftText)
			}

			leftClipped := clipCells(leftText, leftSpace)
			leftCell := lipgloss.NewStyle().Width(leftSpace).Render(leftClipped)

			statsText := stats
			if isCursor {
				statsText = sel.Render(stats)
			}

			line := lipgloss.JoinHorizontal(lipgloss.Top, leftCell, " ", statsText)
			lines = append(lines, lineStyle.Render(line))

		case "run":
			content := row.Name
			if isCursor {
				content = sel.Render(styles.CursorGlyph + content)
			}

			lines = append(lines, lineStyle.Render(content))
		default:
			content := row.Name
			if isCursor {
				content = sel.Render(styles.CursorGlyph + content)
			}

			lines = append(lines, lineStyle.Render(content))
		}
	}
	if isSplit {
		totals, err := m.db.GetSplitTotalsByRunID(context.Background(), int64(m.runID))
		if err != nil {
			m.err = err
			m.showError = true
			return
		}

		summaryStats := ""
		if m.cfg.General.ShowDiff {
			summaryStats = fmt.Sprintf("│ %4d │ %2d │ %+4d │", int(totals.TotalHits.Float64), int(totals.TotalPb.Float64), int(totals.TotalDiff.Float64))
		} else {
			summaryStats = fmt.Sprintf("│ %4d │ %2d │", int(totals.TotalHits.Float64), int(totals.TotalPb.Float64))
		}

		summaryLabels := ""
		if m.cfg.General.ShowDiff {
			summaryLabels = fmt.Sprintf("│ %4s │ %2s │ %4s │", "Hits", "PB", "Diff")
		} else {
			summaryLabels = fmt.Sprintf("│ %4s │ %2s │", "Hits", "PB")
		}

		separatorLine := strings.Repeat("─", contentWidth)
		separatorStyle := lipgloss.NewStyle().Foreground(m.s.Colors.Accent)
		lines = append(lines, lineStyle.Render(separatorStyle.Render(separatorLine)))

		totalLabel := "Total:"
		totalLabelW := lipgloss.Width(totalLabel)
		summaryStatsW := lipgloss.Width(summaryStats)
		summaryLabelsW := lipgloss.Width(summaryLabels)

		totalLeftSpace := contentWidth - totalLabelW - summaryStatsW
		totalLeftSpace = max(totalLeftSpace, 0)

		labelsLeftSpace := contentWidth - summaryLabelsW
		labelsLeftSpace = max(labelsLeftSpace, 0)

		totalSpacer := ""
		if totalLeftSpace > 0 {
			totalSpacer = lipgloss.NewStyle().Width(totalLeftSpace).Render("")
		}

		labelsSpacer := ""
		if labelsLeftSpace > 0 {
			labelsSpacer = lipgloss.NewStyle().Width(labelsLeftSpace).Render("")
		}

		totalLine := lipgloss.JoinHorizontal(lipgloss.Top, totalLabel, totalSpacer, summaryStats)
		labelsLine := lipgloss.JoinHorizontal(lipgloss.Top, labelsSpacer, summaryLabels)

		highlightStyle := lipgloss.NewStyle().
			Foreground(m.s.Colors.Highlight).
			Bold(true)

		lines = append(lines, lineStyle.Render(highlightStyle.Render(totalLine)))
		lines = append(lines, lineStyle.Render(highlightStyle.Render(labelsLine)))
	}

	m.vp.SetContent(strings.Join(lines, "\n"))
}

func (m *model) refreshContent() {
	if m.activeRun == nil {
		m.buildListWithRuns()
	} else {
		m.buildListWithSplits()
	}
}

func (m *model) refreshAfterPicker() {
	r, err := m.db.GetActiveRun(context.Background())
	if err != nil && err != sql.ErrNoRows {
		m.err = err
		m.showError = true
		m.showError = true
		return
	}
	arun := &types.Run{
		ID:          int(r.ID),
		Name:        r.Name,
		Game:        sql.NullString{String: r.Game.String, Valid: true},
		Category:    sql.NullString{String: r.Category.String, Valid: true},
		Attempts:    int(r.Attempts.Int64),
		ActiveSplit: int(r.ActiveSplit.Int64),
	}
	m.activeRun = arun
	if m.activeRun != nil {
		m.runID = int(r.ID)
		m.runName = r.Name
		m.buildListWithSplits()
	} else {
		m.runID = 0
		m.runName = ""
		m.buildListWithRuns()
	}
}

func (m *model) renderHelpMenu() string {
	currentSet := GlobalKeybindingManager.GetCurrentKeybindingSet()
	setDescription := GlobalKeybindingManager.GetKeybindingSetDescription(currentSet)

	var content strings.Builder

	title := m.s.Title.Render("Help - Keybindings")
	content.WriteString(title + "\n\n")

	keymapInfo := m.s.Heading.Render("Current Keymap: " + currentSet)
	content.WriteString(keymapInfo + "\n")
	content.WriteString(m.s.BodyText.Render(setDescription) + "\n\n")

	content.WriteString(m.s.Heading.Render("Keybindings:") + "\n")

	actionDescriptions := map[KeybindingAction]string{
		ActionUp:               "Navigate up",
		ActionDown:             "Navigate down",
		ActionLeft:             "Navigate left",
		ActionRight:            "Navigate right",
		ActionSelect:           "Select item",
		ActionConfirm:          "Confirm selection",
		ActionCancel:           "Cancel/Go back",
		ActionQuit:             "Quit application",
		ActionOpenGameSwitcher: "Open game switcher",
		ActionJumpToTop:        "Jump to top of list",
		ActionJumpToBottom:     "Jump to bottom of list",
		ActionHelp:             "Show this help menu",
		ActionYes:              "Answer yes",
		ActionNo:               "Answer no",
	}

	actionOrder := []KeybindingAction{
		ActionUp,
		ActionDown,
		ActionLeft,
		ActionRight,
		ActionSelect,
		ActionConfirm,
		ActionCancel,
		ActionJumpToTop,
		ActionJumpToBottom,
		ActionOpenGameSwitcher,
		ActionHelp,
		ActionQuit,
		ActionYes,
		ActionNo,
	}

	for _, action := range actionOrder {
		if description, exists := actionDescriptions[action]; exists {
			keys := GlobalKeybindingManager.GetKeysForAction(action)
			if len(keys) > 0 {
				keyStr := strings.Join(keys, ", ")
				line := m.s.BodyText.Render("  "+keyStr+"  ") +
					m.s.Label.Render(description)
				content.WriteString(line + "\n")
			}
		}
	}

	content.WriteString("\n" + m.s.Hint.Render("Press F1, ?, or Esc to close this menu"))

	boxedContent := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.s.Colors.Accent).
		Padding(1, 2).
		Width(int(float64(m.termW) * 0.6)).
		Render(content.String())

	return boxedContent
}

func (m *model) renderError() string {
	var content strings.Builder

	title := m.s.Title.Render("Something went wrong")
	content.WriteString(title + "\n\n")

	content.WriteString(m.s.BodyText.Render(m.err.Error()) + "\n")

	content.WriteString("\n" + m.s.Hint.Render("Press Esc to close this help menu"))

	boxedContent := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.s.Colors.Accent)).
		Padding(1, 2).
		Width(int(float64(m.termW) * 0.4)).
		Render(content.String())

	return boxedContent
}

func loadInitData(s *db.Store) tea.Cmd {
	return func() tea.Msg {
		r, err := s.GetActiveRun(context.Background())
		if err != nil && err != sql.ErrNoRows {
			return initDataMsg{nil, err}
		}
		arun := types.Run{
			ID:          int(r.ID),
			Name:        r.Name,
			Game:        sql.NullString{String: r.Game.String, Valid: true},
			Category:    sql.NullString{String: r.Category.String, Valid: true},
			Attempts:    int(r.Attempts.Int64),
			ActiveSplit: int(r.ActiveSplit.Int64),
		}
		return initDataMsg{&arun, nil}
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, loadInitData(m.db))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.showHelpMenu {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if MatchesHelp(msg) || MatchesCancel(msg) {
				m.showHelpMenu = false
				return m, nil
			}
		}
		return m, nil
	}

	if m.showError {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if MatchesQuit(msg) || MatchesCancel(msg) {
				m.showError = false
				m.err = nil
				return m, nil
			}
		}
		return m, nil
	}

	if m.showFuzzyPicker && m.picker != nil {
		var cmd tea.Cmd
		var sub tea.Model
		sub, cmd = m.picker.Update(msg)
		m.picker = sub.(*fuzzyFinder)

		switch mm := msg.(type) {
		case selectionMsg:
			if len(mm) > 0 {
				selected := mm[0]
				names, err := m.db.GetAllRunNames(context.Background())
				if err == nil {
					for _, n := range names {
						if n.Name == selected {
							if uerr := m.db.UpdateActiveRunByID(context.Background(), sql.NullInt64{Int64: int64(n.ID), Valid: true}); uerr != nil {
								m.err = uerr
								m.showError = true
							}
							break
						}
					}
				} else {
					m.err = err
					m.showError = true
				}
			}
			m.showFuzzyPicker = false
			m.picker = nil
			m.refreshAfterPicker()
			return m, nil
		case cancelPickerMsg:
			m.showFuzzyPicker = false
			m.picker = nil
			return m, nil
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termW = msg.Width
		m.termH = msg.Height
		m.vp.Width = msg.Width
		m.vp.Height = msg.Height - 2
		m.vp.Style = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(m.s.Colors.Accent).
			Padding(0)

		m.refreshContent()
		return m, nil
	case initDataMsg:
		if msg.err != nil {
			m.err = msg.err
			m.showError = true
			return m, nil
		}
		if msg.activeRun != nil {
			m.activeRun = msg.activeRun
			m.runName = msg.activeRun.Name
			m.runID = msg.activeRun.ID
		}
		m.refreshContent()
		if m.activeRun != nil {
			m.setCursorToActiveSplit()
		}
		return m, nil
	case tea.KeyMsg:
		if MatchesQuit(msg) {
			m.quitting = true
			return m, tea.Quit
		}

		if m.showSelection && MatchesCancel(msg) {
			m.showSelection = false
			return m, nil
		}
		if MatchesDown(msg) {
			m.cursor++
			m.clampCursor()
			m.ensureCursorVisible()
			m.updateSelectedSplit()
			m.rebuildList()

		} else if MatchesUp(msg) {
			m.cursor--
			m.clampCursor()
			m.ensureCursorVisible()
			m.updateSelectedSplit()
			m.rebuildList()
		} else if MatchesJumpToTop(msg) {
			m.cursor = 0
			m.ensureCursorVisible()
			m.updateSelectedSplit()
			m.rebuildList()
		} else if MatchesJumpToBottom(msg) {
			if len(m.rows) > 0 {
				m.cursor = len(m.rows) - 1
				m.ensureCursorVisible()
				m.updateSelectedSplit()
				m.rebuildList()
			}
		} else if MatchesOpenGameSwitcher(msg) {
			m.switchGame()
			return m, nil
		} else if MatchesHelp(msg) {
			m.showHelpMenu = true
			return m, nil
		}
		if m.activeRun == nil {
			if MatchesConfirm(msg) {
				if m.cursor < len(m.rows) && m.rows[m.cursor].Type == "run" {
					err := m.db.UpdateActiveRunByID(context.Background(), sql.NullInt64{Int64: int64(m.rows[m.cursor].ID), Valid: true})
					if err != nil {
						m.err = err
						m.showError = true
					} else {
						m.runID = m.rows[m.cursor].ID
						m.buildListWithSplits()
					}
				}
			}
		}
		return m, nil

	case errMsg:
		m.showError = true
		m.err = msg
		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m *model) clampCursor() {
	n := len(m.rows)
	if n == 0 {
		m.cursor = 0
		return
	}
	if m.cursor < 0 {
		m.cursor = 0
	} else if m.cursor >= n {
		m.cursor = n - 1
	}
}

func (m *model) ensureCursorVisible() {
	top := m.vp.YOffset
	bottom := top + m.vp.Height - 1
	if m.cursor < top {
		m.vp.YOffset = m.cursor
	} else if m.cursor > bottom {
		m.vp.YOffset = m.cursor - m.vp.Height + 1
	}
}

func (m *model) updateSelectedSplit() {
	if m.activeRun != nil && m.cursor < len(m.rows) {
		row := m.rows[m.cursor]
		if row.Type == "split" {
			m.splitID = row.ID
			err := m.db.Queries.UpdateActiveSplitByID(context.Background(), sqlc.UpdateActiveSplitByIDParams{ID: int64(m.splitID), ID_2: int64(m.runID)})
			if err != nil {
				m.err = err
				m.showError = true
			}
		}
	}
}

func (m *model) setCursorToActiveSplit() {
	if m.activeRun != nil {
		for i, row := range m.rows {
			if row.Type == "split" {
				split, err := m.db.GetSplitByID(context.Background(), int64(row.ID))
				if err != nil {
					m.err = err
					m.showError = true
					return
				}
				if split.Idx == int64(m.activeRun.ActiveSplit) {
					m.cursor = i
					m.splitID = row.ID
					break
				}
			}
		}
	}
}

func (m model) View() string {
	if m.quitting {
		return "\n"
	}

	if m.showSelection {
		selectionText := strings.Join(m.selection, ", ")
		centeredSelection := lipgloss.Place(
			m.termW,
			m.termH,
			lipgloss.Center,
			lipgloss.Center,
			m.s.Title.Render("Selected: "+selectionText),
		)
		return centeredSelection
	}

	leftTitle := m.s.Title.Render(m.runName)
	rightTitle := m.s.Title.Render(fmt.Sprintf("Attempt: %v", m.runAttempts))

	leftWidth := lipgloss.Width(leftTitle)
	rightWidth := lipgloss.Width(rightTitle)
	availableWidth := m.termW - leftWidth - rightWidth - 2

	spacer := ""
	if availableWidth > 0 {
		spacer = lipgloss.NewStyle().Width(availableWidth).Render("")
	}

	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftTitle,
		spacer,
		rightTitle,
	)
	headerRow := lipgloss.NewStyle().
		Width(m.termW).
		Padding(1, 1, 0, 1).
		Render(header)

	viewportContent := m.vp.View()

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		headerRow,
		viewportContent,
	)

	if m.showHelpMenu {
		helpContent := m.renderHelpMenu()
		helpWidth := lipgloss.Width(helpContent)
		helpHeight := lipgloss.Height(helpContent)
		x := (m.termW - helpWidth) / 2
		y := (m.termH - helpHeight) / 2
		return PlaceOverlay(x, y, helpContent, content)
	}

	if m.err != nil {
		modal := m.renderError()
		modalWidth := lipgloss.Width(modal)
		modalHeight := lipgloss.Height(modal)
		x := (m.termW - modalWidth) / 2
		y := (m.termH - modalHeight) / 2
		return PlaceOverlay(x, y, modal, content)
	}

	if m.showFuzzyPicker && m.picker != nil {
		modal := m.picker.ViewBox()
		modalWidth := lipgloss.Width(modal)
		modalHeight := lipgloss.Height(modal)
		x := (m.termW - modalWidth) / 2
		y := (m.termH - modalHeight) / 2
		return PlaceOverlay(x, y, modal, content)
	}
	return content
}

func (m *model) switchGame() {
	names, err := m.db.GetAllRunNames(context.Background())
	if err != nil {
		m.err = err
		m.showError = true
		return
	}

	runNames := make([]string, len(names))
	for i, name := range names {
		runNames[i] = name.Name
	}

	picker, err := newFuzzyFinderWithThemeAndFlag(
		runNames, false, m.s, "Select Run", true, m.db,
	)
	if err != nil {
		m.err = err
		m.showError = true
		return
	}
	picker.termW, picker.termH = m.termW, m.termH
	picker.boxW = int(float64(m.termW) * 0.5)
	picker.viewport.Width = picker.boxW
	picker.viewport.Height = min(12, m.termH/4)
	promptCells := lipgloss.Width(picker.input.Prompt)
	picker.input.Width = max(picker.boxW-promptCells-3, 0)
	picker.refreshContent()
	m.picker = picker
	m.showFuzzyPicker = true
}

func NewMainScreen(s styles.Styles, db *db.Store) {
	m := initialModel(s, db)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
