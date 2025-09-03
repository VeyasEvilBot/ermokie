package setup

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stefanistkuhl/ermokie/pkg/models/styles"
)

type screenAsk struct {
	s        styles.Styles
	title    string
	desc     string
	question string
	defaultY bool
	focusIdx int
	w, h     int
}

func newScreenAsk(s styles.Styles, spec StepSpec) *screenAsk {
	f := 1
	if spec.DefaultY {
		f = 0
	}
	return &screenAsk{
		s:        s,
		title:    spec.Title,
		desc:     spec.Description,
		question: spec.Question,
		defaultY: spec.DefaultY,
		focusIdx: f,
	}
}

func (m *screenAsk) Init() tea.Cmd { return nil }

func (m *screenAsk) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch k := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = k.Width, k.Height
		return m, nil
	case tea.KeyMsg:
		switch k.String() {
		case "left", "h", "up", "k":
			if m.focusIdx > 0 {
				m.focusIdx--
			}
			return m, nil
		case "right", "l", "down", "j":
			if m.focusIdx < 1 {
				m.focusIdx++
			}
			return m, nil

		case "y", "Y":
			return m, func() tea.Msg { return StepResult{Value: true} }
		case "n", "N":
			return m, func() tea.Msg { return StepResult{Value: false} }

		case "enter", " ":
			val := (m.focusIdx == 0)
			return m, func() tea.Msg { return StepResult{Value: val} }

		case "esc":
			return m, func() tea.Msg { return StepResult{Value: false} }
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *screenAsk) View() string {
	btn := func(label string, focused bool) string {
		base := m.s.BodyText
		if focused {
			base = m.s.Title
		}
		first := label[:1]
		rest := ""
		if len(label) > 1 {
			rest = label[1:]
		}
		emph := base.Bold(true).Underline(true)
		return base.Render("[ ") + emph.Render(first) + base.Render(rest+" ]")
	}

	row := lipgloss.JoinHorizontal(lipgloss.Center,
		btn("Yes", m.focusIdx == 0), "  ", btn("No", m.focusIdx == 1),
	)

	body := lipgloss.JoinVertical(
		lipgloss.Center,
		m.s.Heading.Render(m.title),
		m.s.BodyText.Render(m.desc),
		"",
		m.s.Label.Render(m.question),
		"",
		row,
		m.s.Hint.Render("←/→ h/j/k/l to move, Enter to confirm; y/n also work"),
	)

	styledBody := m.s.Window.Render(body)

	if m.w > 0 && m.h > 0 {
		return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, styledBody)
	}

	return styledBody
}
