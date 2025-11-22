package setup

import (
	"strings"

	"codeberg.org/veya/ermokie/pkg/models"
	"codeberg.org/veya/ermokie/pkg/models/styles"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type screenList struct {
	viewport          viewport.Model
	s                 styles.Styles
	title             string
	desc              string
	options           []string
	cursor            int
	w, h              int
	listWidth         int
	listHeight        int
	isThemeSelection  bool
	isKeymapSelection bool
	themeManager      *ThemeManager
	flowStyles        *styles.Styles
}

func NewScreenList(s styles.Styles, spec StepSpec, w, h int) *screenList {
	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle()

	m := &screenList{
		viewport: vp,
		s:        s,
		title:    spec.Title,
		desc:     spec.Description,
		options:  spec.Options,
		w:        w,
		h:        h,
	}

	if spec.ID == "choose-theme" {
		m.isThemeSelection = true
		m.themeManager = NewThemeManager()
		if len(spec.Options) > 0 {
			m.s = m.themeManager.ApplyTheme(spec.Options[0], m.s)
		}
		m.title = "Theme Selector"
	}

	if spec.ID == "choose-keymap" {
		m.isKeymapSelection = true
		m.title = "Keymap Selector"
	}

	m.applySize(w, h)
	m.rebuildList()
	return m
}

func (m *screenList) applySize(w, h int) {
	if w <= 0 {
		w = 80
	}
	if h <= 0 {
		h = 24
	}
	m.w, m.h = w, h
	m.listWidth = int(float64(w) * 0.4)
	maxRows := 12
	m.listHeight = min(maxRows, max(3, h-8))
	m.viewport.Width = m.listWidth
	m.viewport.Height = m.listHeight
}

func (m *screenList) Init() tea.Cmd { return nil }

func (m *screenList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.applySize(msg.Width, msg.Height)
		m.rebuildList()
		return m, nil
	case StyleUpdateMsg:
		m.s = msg.NewStyles
		m.refreshUI()
		return m, nil

	case tea.KeyMsg:
		if models.MatchesUp(msg) {
			if m.cursor > 0 {
				m.cursor--
				m.ensureCursorVisible()
				// Apply theme immediately when moving cursor
				if m.isThemeSelection && m.themeManager != nil && len(m.options) > 0 {
					themeName := m.options[m.cursor]
					m.s = m.themeManager.ApplyTheme(themeName, m.s)
					m.refreshUI() // Force complete UI refresh
				} else if m.isKeymapSelection && len(m.options) > 0 {
					// Apply keymap immediately when moving cursor
					keymapName := m.options[m.cursor]
					models.LoadKeymapFromConfig(keymapName)
					m.rebuildList()
				} else {
					m.rebuildList()
				}
			}
		}

		if models.MatchesDown(msg) {
			if m.cursor < max(0, len(m.options)-1) {
				m.cursor++
				m.ensureCursorVisible()
				if m.isThemeSelection && m.themeManager != nil && len(m.options) > 0 {
					themeName := m.options[m.cursor]
					m.s = m.themeManager.ApplyTheme(themeName, m.s)
					m.refreshUI()
				} else if m.isKeymapSelection && len(m.options) > 0 {
					keymapName := m.options[m.cursor]
					models.LoadKeymapFromConfig(keymapName)
					m.rebuildList()
				} else {
					m.rebuildList()
				}
			}
		}

		if models.MatchesConfirm(msg) {
			if len(m.options) > 0 {
				val := m.options[m.cursor]
				if m.isThemeSelection && m.themeManager != nil {
					m.s = m.themeManager.ApplyTheme(val, m.s)
					m.refreshUI()
				} else if m.isKeymapSelection {
					models.LoadKeymapFromConfig(val)
				}
				return m, func() tea.Msg { return StepResult{Value: val} }
			}
		}

		if models.MatchesCancel(msg) {
			return m, func() tea.Msg { return StepResult{Value: nil} }
		}

		if models.MatchesQuit(msg) {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *screenList) rebuildList() {
	inner := max(m.viewport.Width, 0)
	wrap := lipgloss.NewStyle().Width(inner)

	var lines []string
	for i, option := range m.options {
		var txt string
		if i == m.cursor {
			highlightColor := m.s.Colors.Highlight
			txt = lipgloss.NewStyle().
				Foreground(highlightColor).
				Bold(true).
				Padding(0, 1).
				Render(styles.CursorGlyph + option)
		} else {
			txt = "  " + option
		}

		if m.isThemeSelection && m.themeManager != nil {
			desc := m.themeManager.GetThemeDescription(option)
			if desc != "" {
				txt += " - " + desc
			}
		} else if m.isKeymapSelection {
			desc := models.GlobalKeybindingManager.GetKeybindingSetDescription(option)
			if desc != "" {
				txt += " - " + desc
			}
		}

		lines = append(lines, wrap.Render(txt))
	}
	m.viewport.SetContent(strings.Join(lines, "\n"))
}

func (m *screenList) refreshUI() {
	m.viewport.SetContent("")
	m.rebuildList()
	if m.w > 0 && m.h > 0 {
		m.applySize(m.w, m.h)
	}
}

func (m *screenList) updateStyles(newStyles styles.Styles) {
	m.s = newStyles
	m.refreshUI()
}

func (m *screenList) ensureCursorVisible() {
	top := m.viewport.YOffset
	bottom := top + m.viewport.Height - 1
	if m.cursor < top {
		m.viewport.YOffset = m.cursor
	} else if m.cursor > bottom {
		m.viewport.YOffset = m.cursor - m.viewport.Height + 1
	}
}

func (m *screenList) View() string {
	header := []string{m.s.Heading.Render(m.title)}
	desc := m.s.BodyText.Render(m.desc)
	if m.listWidth > 0 {
		desc = lipgloss.NewStyle().Width(m.listWidth).Render(desc)
	}
	header = append(header, desc, "")

	listWithBar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.viewport.View(),
		" ",
	)

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		strings.Join(header, "\n"),
		listWithBar,
	)

	styled := m.s.Window.Width(m.listWidth).Render(body)
	if m.w > 0 && m.h > 0 {
		return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, styled)
	}
	return styled
}
