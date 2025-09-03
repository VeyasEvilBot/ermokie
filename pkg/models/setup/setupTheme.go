package setup

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/stefanistkuhl/ermokie/pkg/models/styles"
)

type ThemeManager struct {
	availableThemes map[string]Theme
	currentTheme    string
}

type Theme struct {
	Name        string
	Description string
	Colors      styles.SetupColors
}

func NewThemeManager() *ThemeManager {
	tm := &ThemeManager{
		availableThemes: make(map[string]Theme),
		currentTheme:    "rose-pine-moon",
	}

	tm.availableThemes["default"] = Theme{
		Name:        "default",
		Description: "Default terminal colors",
		Colors:      styles.DefaultColors(),
	}

	tm.availableThemes["rose-pine-moon"] = Theme{
		Name:        "rose-pine-moon",
		Description: "Elegant dark theme with rose accents",
		Colors: styles.SetupColors{
			Bg:           lipgloss.Color("235"),
			Fg:           lipgloss.Color("189"),
			Muted:        lipgloss.Color("60"),
			Accent:       lipgloss.Color("168"),
			Border:       lipgloss.Color("67"),
			Title:        lipgloss.Color("179"),
			SubTitle:     lipgloss.Color("152"),
			Highlight:    lipgloss.Color("183"),
			SelectedText: lipgloss.Color("233"),
		},
	}

	tm.availableThemes["catppuccin-macchiato"] = Theme{
		Name:        "catppuccin-macchiato",
		Description: "Warm and cozy dark theme",
		Colors: styles.SetupColors{
			Bg:           lipgloss.Color("235"),
			Fg:           lipgloss.Color("188"),
			Muted:        lipgloss.Color("60"),
			Accent:       lipgloss.Color("209"),
			Border:       lipgloss.Color("67"),
			Title:        lipgloss.Color("218"),
			SubTitle:     lipgloss.Color("150"),
			Highlight:    lipgloss.Color("183"),
			SelectedText: lipgloss.Color("233"),
		},
	}

	tm.availableThemes["tokyo-night-dark"] = Theme{
		Name:        "tokyo-night-dark",
		Description: "Deep blue theme with vibrant accents",
		Colors: styles.SetupColors{
			Bg:           lipgloss.Color("16"),
			Fg:           lipgloss.Color("188"),
			Muted:        lipgloss.Color("60"),
			Accent:       lipgloss.Color("67"),
			Border:       lipgloss.Color("87"),
			Title:        lipgloss.Color("168"),
			SubTitle:     lipgloss.Color("150"),
			Highlight:    lipgloss.Color("183"),
			SelectedText: lipgloss.Color("17"),
		},
	}

	return tm
}

func (tm *ThemeManager) GetAvailableThemes() []string {
	themes := make([]string, 0, len(tm.availableThemes))
	for name := range tm.availableThemes {
		themes = append(themes, name)
	}
	return themes
}

func (tm *ThemeManager) GetThemeDescription(themeName string) string {
	if theme, exists := tm.availableThemes[themeName]; exists {
		return theme.Description
	}
	return "Unknown theme"
}

func (tm *ThemeManager) ApplyTheme(themeName string, s styles.Styles) styles.Styles {
	if theme, exists := tm.availableThemes[themeName]; exists {
		tm.currentTheme = themeName
		newStyles := s
		newStyles.Colors = theme.Colors

		c := theme.Colors
		newStyles.Window = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(c.Border).
			Foreground(c.Fg).
			Padding(1, 2).
			Align(lipgloss.Center).
			AlignHorizontal(lipgloss.Center)
		newStyles.Title = lipgloss.NewStyle().Foreground(c.Title).Bold(true)
		newStyles.Heading = lipgloss.NewStyle().
			Foreground(c.Title).Bold(true)
		newStyles.BodyText = lipgloss.NewStyle().Foreground(c.Fg)
		newStyles.Label = lipgloss.NewStyle().Foreground(c.SubTitle).Bold(true)
		newStyles.Hint = lipgloss.NewStyle().Foreground(c.Muted).Italic(true)

		return newStyles
	}
	return s
}

func (tm *ThemeManager) GetCurrentTheme() string {
	return tm.currentTheme
}

func (tm *ThemeManager) GetThemeColors(themeName string) (styles.SetupColors, bool) {
	if theme, exists := tm.availableThemes[themeName]; exists {
		return theme.Colors, true
	}
	return styles.SetupColors{}, false
}

func (tm *ThemeManager) PreviewTheme(themeName string, text string) string {
	if theme, exists := tm.availableThemes[themeName]; exists {
		style := lipgloss.NewStyle().
			Foreground(theme.Colors.Fg).
			Background(theme.Colors.SelectedText).
			Padding(0, 1)
		return style.Render(text)
	}
	return text
}
