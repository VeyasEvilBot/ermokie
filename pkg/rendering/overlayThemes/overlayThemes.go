package overlaythemes

type OverlayThemes struct {
	Themes []OverlayTheme
}

type OverlayTheme struct {
	Name   string
	Colors OverlayThemeColors
}

type OverlayThemeColors struct {
	TextPrimary     string
	BorderPrimary   string
	BorderSecondary string
	BorderTertiary  string
	BgHeader        string
	BgRowEven       string
	BgRowOdd        string
	BgHighlighted   string
}

func (o OverlayThemes) GetThemeByName(name string) OverlayTheme {
	for _, theme := range o.Themes {
		if theme.Name == name {
			return theme
		}
	}
	var t OverlayTheme
	return t
}

func (o OverlayThemes) GetThemeNames() []string {
	var names []string
	for _, theme := range o.Themes {
		names = append(names, theme.Name)
	}
	return names
}

func NewOverlayThemes() OverlayThemes {
	return OverlayThemes{
		Themes: []OverlayTheme{
			{
				Name: "default",
				Colors: OverlayThemeColors{
					TextPrimary:     "#222222",
					BorderPrimary:   "#ffffff",
					BorderSecondary: "#c8c8c8",
					BorderTertiary:  "#bebebe",
					BgHeader:        "#ebebeb",
					BgRowEven:       "#fafafa",
					BgRowOdd:        "#f5f5f5",
					BgHighlighted:   "#ffeb3b",
				},
			},
			{
				Name: "ugly",
				Colors: OverlayThemeColors{
					TextPrimary:     "#ff00ff",
					BorderPrimary:   "#00ff00",
					BorderSecondary: "#ff6600",
					BorderTertiary:  "#00ffff",
					BgHeader:        "#0000ff",
					BgRowEven:       "#ffff00",
					BgRowOdd:        "#ff0000",
					BgHighlighted:   "#00ffff",
				},
			},
		},
	}
}
