package setup

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stefanistkuhl/ermokie/pkg/models/styles"
)

type StepKind int

const (
	StepAsk StepKind = iota
	StepList
	StepInput
)

type StepSpec struct {
	ID          string
	Kind        StepKind
	Title       string
	Description string

	Question          string
	DefaultY          bool
	OnDone            func(any) tea.Cmd
	OnDoneWithStyles  func(any, styles.Styles) tea.Cmd
	OnDoneWithResults func(any, styles.Styles, map[string]any) tea.Cmd

	Options []string

	Placeholder string
	Validate    func(string) error
}

type StepResult struct {
	ID    string
	Value any
	Err   error
}

type ThemeChangeMsg struct {
	ThemeName string
	NewStyles styles.Styles
}

type StyleUpdateMsg struct {
	NewStyles styles.Styles
}
