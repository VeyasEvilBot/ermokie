package setup

import (
	"codeberg.org/veya/ermokie/pkg/models/styles"
	"github.com/charmbracelet/lipgloss"
)

type ThemeManager struct {
	availableThemes map[string]Theme
	themeOrder      []string
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
		themeOrder:      make([]string, 0, 8),
	}

	tm.availableThemes["default"] = Theme{
		Name:        "default",
		Description: "Default terminal colors",
		Colors:      styles.DefaultColors(),
	}
	tm.themeOrder = append(tm.themeOrder, "default")

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
	tm.themeOrder = append(tm.themeOrder, "rose-pine-moon")

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
	tm.themeOrder = append(tm.themeOrder, "catppuccin-macchiato")

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
	tm.themeOrder = append(tm.themeOrder, "tokyo-night-dark")

	tm.SetThemeOrder([]string{
		"default",
		"catppuccin-macchiato",
		"rose-pine-moon",
		"tokyo-night-dark",
	})

	return tm
}

func (tm *ThemeManager) GetAvailableThemes() []string {
	out := make([]string, len(tm.themeOrder))
	copy(out, tm.themeOrder)
	return out
}

func (tm *ThemeManager) SetThemeOrder(order []string) {
	seen := make(map[string]bool, len(order))
	newOrder := make([]string, 0, len(tm.availableThemes))

	for _, name := range order {
		if _, ok := tm.availableThemes[name]; ok && !seen[name] {
			newOrder = append(newOrder, name)
			seen[name] = true
		}
	}

	for _, name := range tm.themeOrder {
		if !seen[name] {
			if _, ok := tm.availableThemes[name]; ok {
				newOrder = append(newOrder, name)
				seen[name] = true
			}
		}
	}

	tm.themeOrder = newOrder
}

func (tm *ThemeManager) AddTheme(t Theme) {
	_, existed := tm.availableThemes[t.Name]
	tm.availableThemes[t.Name] = t
	if !existed {
		tm.themeOrder = append(tm.themeOrder, t.Name)
	}
}

func (tm *ThemeManager) RemoveTheme(name string) bool {
	if name == tm.currentTheme {
		return false
	}
	if _, ok := tm.availableThemes[name]; !ok {
		return false
	}
	delete(tm.availableThemes, name)
	newOrder := new([]string)
	for _, n := range tm.themeOrder {
		if n != name {
			*newOrder = append(*newOrder, n)
		}
	}
	tm.themeOrder = *newOrder
	return true
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

func (tm *ThemeManager) ValidateTheme(name string) bool {
	for _, theme := range tm.availableThemes {
		if theme.Name == name {
			return true
		}
	}
	return false
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
