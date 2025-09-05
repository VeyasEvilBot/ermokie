package setup

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stefanistkuhl/ermokie/pkg/models"
	"github.com/stefanistkuhl/ermokie/pkg/models/styles"
)

type screenInput struct {
	s           styles.Styles
	title       string
	desc        string
	placeholder string
	validate    func(string) error
	ti          textinput.Model
	errText     string
	w, h        int
}

func newScreenInput(s styles.Styles, spec StepSpec) *screenInput {
	ti := textinput.New()
	ti.Placeholder = spec.Placeholder
	ti.Prompt = "> "
	ti.CharLimit = 256
	ti.Focus()

	// Apply global styling to the text input
	ti.PromptStyle = ti.PromptStyle.Foreground(s.Colors.Accent)
	ti.TextStyle = ti.TextStyle.Foreground(s.Colors.Fg)
	ti.PlaceholderStyle = ti.PlaceholderStyle.Foreground(s.Colors.Muted)

	return &screenInput{
		s:           s,
		title:       spec.Title,
		desc:        spec.Description,
		placeholder: spec.Placeholder,
		validate:    spec.Validate,
		ti:          ti,
	}
}

func (m *screenInput) Init() tea.Cmd { return nil }

func (m *screenInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch k := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = k.Width, k.Height
		return m, nil
	case tea.KeyMsg:
		if models.MatchesConfirm(k) {
			val := m.ti.Value()
			if m.validate != nil {
				if err := m.validate(val); err != nil {
					m.errText = err.Error()
					return m, nil
				}
			}
			return m, func() tea.Msg { return StepResult{Value: val} }
		}

		if models.MatchesCancel(k) {
			return m, func() tea.Msg { return StepResult{Value: ""} }
		}

		if models.MatchesQuit(k) {
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.ti, cmd = m.ti.Update(msg)
	return m, cmd
}

func (m *screenInput) View() string {
	body := lipgloss.JoinVertical(lipgloss.Left,
		m.s.Heading.Render(m.title),
		m.s.BodyText.Render(m.desc),
		"",
		m.ti.View(),
		func() string {
			if m.errText == "" {
				return m.s.Hint.Render(fmt.Sprintf("(%s to confirm, %s to cancel)",
					models.GetConfirmKeys()[0], models.GetCancelKeys()[0]))
			}
			return m.s.Label.Render(fmt.Sprintf("Error: %s", m.errText))
		}(),
	)

	content := m.s.Window.Render(body)

	if m.w > 0 && m.h > 0 {
		return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, content)
	}

	return content
}
