package setup

import (
	"strings"

	"codeberg.org/veya/ermokie/pkg/models"
	"codeberg.org/veya/ermokie/pkg/models/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type helpScreen struct {
	s      styles.Styles
	width  int
	height int
	spec   StepSpec
}

func newScreenHelp(s styles.Styles, spec StepSpec) *helpScreen {
	return &helpScreen{
		s:    s,
		spec: spec,
	}
}

func (m *helpScreen) Init() tea.Cmd {
	return nil
}

func (m *helpScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case StyleUpdateMsg:
		m.s = msg.NewStyles
		return m, nil

	case tea.KeyMsg:
		if models.MatchesHelp(msg) || models.MatchesCancel(msg) || models.MatchesConfirm(msg) {
			// Close help and continue to next step
			return m, func() tea.Msg {
				return StepResult{
					ID:    m.spec.ID,
					Value: true, // Help was shown
					Err:   nil,
				}
			}
		}
	}
	return m, nil
}

func (m *helpScreen) View() string {
	// Get current keybinding set
	currentSet := models.GlobalKeybindingManager.GetCurrentKeybindingSet()
	setDescription := models.GlobalKeybindingManager.GetKeybindingSetDescription(currentSet)

	// Create help content
	var content strings.Builder

	// Title
	title := m.s.Title.Render("Help - Keybindings")
	content.WriteString(title + "\n\n")

	// Current keymap info
	keymapInfo := m.s.Heading.Render("Current Keymap: " + currentSet)
	content.WriteString(keymapInfo + "\n")
	content.WriteString(m.s.BodyText.Render(setDescription) + "\n\n")

	// Keybindings section
	content.WriteString(m.s.Heading.Render("Keybindings:") + "\n")

	// Define action descriptions
	actionDescriptions := map[models.KeybindingAction]string{
		models.ActionUp:               "Navigate up",
		models.ActionDown:             "Navigate down",
		models.ActionLeft:             "Navigate left",
		models.ActionRight:            "Navigate right",
		models.ActionSelect:           "Select item",
		models.ActionConfirm:          "Confirm selection",
		models.ActionCancel:           "Cancel/Go back",
		models.ActionQuit:             "Quit application",
		models.ActionOpenGameSwitcher: "Open game switcher",
		models.ActionJumpToTop:        "Jump to top of list",
		models.ActionJumpToBottom:     "Jump to bottom of list",
		models.ActionHelp:             "Show this help menu",
		models.ActionYes:              "Answer yes",
		models.ActionNo:               "Answer no",
	}

	// Define logical order for displaying keybindings
	actionOrder := []models.KeybindingAction{
		// Navigation
		models.ActionUp,
		models.ActionDown,
		models.ActionLeft,
		models.ActionRight,
		// Selection and confirmation
		models.ActionSelect,
		models.ActionConfirm,
		models.ActionCancel,
		// Navigation shortcuts
		models.ActionJumpToTop,
		models.ActionJumpToBottom,
		// Application actions
		models.ActionOpenGameSwitcher,
		models.ActionHelp,
		models.ActionQuit,
		// Yes/No responses
		models.ActionYes,
		models.ActionNo,
	}

	// Display keybindings in logical order
	for _, action := range actionOrder {
		if description, exists := actionDescriptions[action]; exists {
			keys := models.GlobalKeybindingManager.GetKeysForAction(action)
			if len(keys) > 0 {
				keyStr := strings.Join(keys, ", ")
				line := m.s.BodyText.Render("  "+keyStr+"  ") +
					m.s.Label.Render(description)
				content.WriteString(line + "\n")
			}
		}
	}

	// Footer
	content.WriteString("\n" + m.s.Hint.Render("Press F1, ?, Enter, or Esc to continue setup"))

	// Wrap in a border
	boxedContent := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.s.Colors.Accent).
		Padding(1, 2).
		Width(int(float64(m.width) * 0.6)).
		Render(content.String())

	// Center the help menu
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, boxedContent)
	}

	return boxedContent
}
