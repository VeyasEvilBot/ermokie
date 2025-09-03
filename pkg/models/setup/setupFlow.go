package setup

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stefanistkuhl/ermokie/pkg/models/styles"
)

type flowModel struct {
	s            styles.Styles
	width        int
	height       int
	menuIdx      int
	inChild      tea.Model
	stepSpecs    []StepSpec
	stepIdx      int
	results      map[string]any
	quitAfter    bool
	loadDefaults func() error
	themeManager *ThemeManager
	config       interface{}
}

func NewSetupFlow(loadDefaults func() error, steps []StepSpec, config interface{}) tea.Model {
	return &flowModel{
		s:            styles.DefaultStyles(),
		stepSpecs:    steps,
		results:      make(map[string]any),
		themeManager: NewThemeManager(),
		loadDefaults: loadDefaults,
		config:       config,
	}
}

func (m *flowModel) Init() tea.Cmd { return nil }

func (m *flowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.inChild != nil {
		var cmd tea.Cmd
		m.inChild, cmd = m.inChild.Update(msg)
		switch r := msg.(type) {
		case StepResult:
			id := m.stepSpecs[m.stepIdx].ID
			m.results[id] = r.Value
			spec := m.stepSpecs[m.stepIdx]

			if id == "choose-theme" && r.Value != nil {
				if themeName, ok := r.Value.(string); ok {
					m.s = m.themeManager.ApplyTheme(themeName, m.s)
					if m.inChild != nil {
						m.inChild.Update(StyleUpdateMsg{NewStyles: m.s})
					}
				}
			}

			if spec.OnDoneWithResults != nil {
				cmd = tea.Batch(cmd, spec.OnDoneWithResults(r.Value, m.s, m.results))
			} else if spec.OnDoneWithStyles != nil {
				cmd = tea.Batch(cmd, spec.OnDoneWithStyles(r.Value, m.s))
			} else if spec.OnDone != nil {
				cmd = tea.Batch(cmd, spec.OnDone(r.Value))
			}

			m.stepIdx++
			if m.stepIdx >= len(m.stepSpecs) {
				// All steps completed - save config and exit
				if m.loadDefaults != nil {
					if err := m.loadDefaults(); err != nil {
						// add showing and error message in the ui
						return m, nil
					}
				}
				m.inChild = nil
				m.quitAfter = true
				return m, tea.Quit
			}
			return m, tea.Batch(cmd, m.startChildFor(m.stepSpecs[m.stepIdx]))
		}
		return m, cmd
	}

	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = v.Width, v.Height
		if m.inChild != nil {
			m.inChild.Update(msg)
		}
		return m, nil

	case tea.KeyMsg:
		switch v.String() {
		case "left", "h", "up", "k":
			if m.menuIdx > 0 {
				m.menuIdx--
			}
			return m, nil

		case "right", "l", "down", "j":
			if m.menuIdx < 1 {
				m.menuIdx++
			}
			return m, nil

		case "enter":
			if m.menuIdx == 0 {
				// Quickstart - save config with defaults and exit
				if m.loadDefaults != nil {
					if err := m.loadDefaults(); err != nil {
						// add showing and error message in the ui
						return m, nil
					}
				}
				return m, tea.Quit
			}
			if len(m.stepSpecs) == 0 {
				return m, tea.Quit
			}
			return m, m.startChildFor(m.stepSpecs[0])

		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *flowModel) View() string {
	if m.inChild != nil {
		return m.inChild.View()
	}
	opt := func(i int, label string) string {
		if i == m.menuIdx {
			return m.s.Title.Render(styles.CursorGlyph + label)
		}
		return m.s.BodyText.Render("  " + label)
	}
	menu := m.s.Window.Render(
		opt(0, "Quick Start") + "\n" +
			m.s.Hint.Render("Loads defaults and exits") + "\n\n" +
			opt(1, "Custom Setup") + "\n" +
			m.s.Hint.Render("Run interactive steps"),
	)

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, menu)
	}

	return menu
}

func (m *flowModel) startChildFor(spec StepSpec) tea.Cmd {
	m.stepIdx = indexOfSpec(m.stepSpecs, spec.ID)
	switch spec.Kind {
	case StepAsk:
		child := newScreenAsk(m.s, spec)
		if m.width > 0 && m.height > 0 {
			child.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		}
		m.inChild = child
	case StepList:
		m.inChild = newScreenList(m.s, spec, m.width, m.height)
	case StepInput:
		child := newScreenInput(m.s, spec)
		if m.width > 0 && m.height > 0 {
			child.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		}
		m.inChild = child
	default:
		panic("unknown step kind")
	}
	return nil
}

func (m *flowModel) updateChildStyles() {
	if m.inChild != nil {
		if m.width > 0 && m.height > 0 {
			m.inChild.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		}
	}
}

func indexOfSpec(steps []StepSpec, id string) int {
	for i, s := range steps {
		if s.ID == id {
			return i
		}
	}
	return 0
}
