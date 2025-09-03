package styles

import "github.com/charmbracelet/lipgloss"

type SetupColors struct {
	Bg           lipgloss.Color
	Fg           lipgloss.Color
	Muted        lipgloss.Color
	Accent       lipgloss.Color
	Border       lipgloss.Color
	Title        lipgloss.Color
	SubTitle     lipgloss.Color
	Highlight    lipgloss.Color
	SelectedText lipgloss.Color
}

type Styles struct {
	Colors   SetupColors
	Window   lipgloss.Style
	Title    lipgloss.Style
	Heading  lipgloss.Style
	BodyText lipgloss.Style
	Label    lipgloss.Style
	Hint     lipgloss.Style
}

func DefaultColors() SetupColors {
	return SetupColors{
		Bg:           lipgloss.Color("0"),
		Fg:           lipgloss.Color("7"),
		Muted:        lipgloss.Color("8"),
		Accent:       lipgloss.Color("6"),
		Border:       lipgloss.Color("8"),
		Title:        lipgloss.Color("14"),
		SubTitle:     lipgloss.Color("12"),
		Highlight:    lipgloss.Color("5"),
		SelectedText: lipgloss.Color("0"),
	}
}

func DefaultStyles() Styles {
	c := DefaultColors()
	return Styles{
		Colors: c,
		Window: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(c.Accent).
			Foreground(c.Fg).
			Padding(1, 2).
			Align(lipgloss.Center).
			AlignHorizontal(lipgloss.Center),
		Title: lipgloss.NewStyle().Foreground(c.Title).Bold(true),
		Heading: lipgloss.NewStyle().
			Foreground(c.Accent).Bold(true),
		BodyText: lipgloss.NewStyle().Foreground(c.Fg),
		Label:    lipgloss.NewStyle().Foreground(c.SubTitle).Bold(true),
		Hint:     lipgloss.NewStyle().Foreground(c.Muted).Italic(true),
	}
}

const CursorGlyph = "› "
