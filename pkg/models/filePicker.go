package models

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"codeberg.org/veya/ermokie/pkg/models/styles"
	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type filePicker struct {
	filepicker   filepicker.Model
	selectedFile string
	quitting     bool
	err          error
	styles       struct {
		title    lipgloss.Style
		bodyText lipgloss.Style
		hint     lipgloss.Style
		error    lipgloss.Style
		selected lipgloss.Style
		border   lipgloss.Color
	}
	width  int
	height int
}

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func (f filePicker) Init() tea.Cmd {
	return f.filepicker.Init()
}

func (f filePicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.width = msg.Width
		f.height = msg.Height
		return f, nil
	case tea.KeyMsg:
		if MatchesQuit(msg) {
			f.quitting = true
			return f, tea.Quit
		}
	case clearErrorMsg:
		f.err = nil
	}

	var cmd tea.Cmd
	f.filepicker, cmd = f.filepicker.Update(msg)

	if didSelect, path := f.filepicker.DidSelectFile(msg); didSelect {
		f.selectedFile = path
		f.quitting = true
		return f, tea.Quit
	}

	if didSelect, path := f.filepicker.DidSelectDisabledFile(msg); didSelect {
		f.err = errors.New(path + " is not valid.")
		f.selectedFile = ""
		return f, tea.Batch(cmd, clearErrorAfter(2*time.Second))
	}

	return f, cmd
}

func (f filePicker) View() string {
	if f.quitting {
		return ""
	}

	if f.width <= 0 {
		f.width = 120
	}
	if f.height <= 0 {
		f.height = 30
	}

	contentWidth := int(float64(f.width) * 0.8)
	contentHeight := int(float64(f.height) * 0.8)

	var s strings.Builder

	title := f.styles.title.Render("File Picker")
	centeredTitle := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(title)
	s.WriteString(centeredTitle + "\n\n")

	var statusMsg string
	if f.err != nil {
		statusMsg = f.styles.error.Render("Error: " + f.err.Error())
	} else if f.selectedFile == "" {
		statusMsg = f.styles.bodyText.Render("Pick a file:")
	} else {
		statusMsg = f.styles.bodyText.Render("Selected file: ") + f.styles.selected.Render(f.selectedFile)
	}
	centeredStatus := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(statusMsg)
	s.WriteString(centeredStatus + "\n\n")

	filePickerContent := f.filepicker.View()
	s.WriteString(filePickerContent + "\n\n")

	upKeys := GetFuzzyUpKeys()
	downKeys := GetFuzzyDownKeys()
	selectKeys := GetSelectKeys()
	quitKeys := GlobalKeybindingManager.GetKeysForAction(ActionQuit)

	helpText := f.styles.hint.Render(fmt.Sprintf("%s/%s Navigate • %s Select • %s Quit",
		upKeys[0], downKeys[0], selectKeys[0], quitKeys[0]))
	centeredHelp := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(helpText)
	s.WriteString(centeredHelp)

	content := s.String()

	borderedContent := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(f.styles.border).
		Padding(1, 2).
		Width(contentWidth).
		Height(contentHeight).
		Render(content)

	if f.width > 0 && f.height > 0 {
		return lipgloss.Place(f.width, f.height, lipgloss.Center, lipgloss.Center, borderedContent)
	}
	return borderedContent
}

func (f *filePicker) SetStyles(title, bodyText, hint, error, selected lipgloss.Style) {
	f.styles.title = title
	f.styles.bodyText = bodyText
	f.styles.hint = hint
	f.styles.error = error
	f.styles.selected = selected
	if f.styles.border == "" {
		f.styles.border = lipgloss.Color("6")
	}
}

func (f *filePicker) SetStylesWithBorder(title, bodyText, hint, error, selected lipgloss.Style, border lipgloss.Color) {
	f.styles.title = title
	f.styles.bodyText = bodyText
	f.styles.hint = hint
	f.styles.error = error
	f.styles.selected = selected
	f.styles.border = border
}

func NewFilePicker() string {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".xml"}
	fp.CurrentDirectory, _ = os.UserHomeDir()

	width, height, _ := term.GetSize(int(os.Stdout.Fd()))

	contentHeight := int(float64(height) * 0.8)
	fp.Height = contentHeight - 6

	f := filePicker{
		filepicker: fp,
		width:      width,
		height:     height,
	}

	f.SetStylesWithBorder(
		lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true),
		lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true),
		lipgloss.NewStyle().Foreground(lipgloss.Color("1")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true),
		lipgloss.Color("6"),
	)

	tm, _ := tea.NewProgram(&f).Run()
	mm := tm.(filePicker)
	fmt.Println("\n  You selected: " + f.filepicker.Styles.Selected.Render(mm.selectedFile) + "\n")
	return mm.selectedFile
}

func NewFilePickerWithTheme(themeStyles styles.Styles) string {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".xml"}
	fp.CurrentDirectory, _ = os.UserHomeDir()

	width, height, _ := term.GetSize(int(os.Stdout.Fd()))

	contentHeight := int(float64(height) * 0.8)
	fp.Height = contentHeight - 6

	f := filePicker{
		filepicker: fp,
		width:      width,
		height:     height,
	}

	f.SetStylesWithBorder(themeStyles.Title, themeStyles.BodyText, themeStyles.Hint, themeStyles.Label, themeStyles.Title, themeStyles.Colors.Border)

	tm, _ := tea.NewProgram(&f).Run()
	mm := tm.(filePicker)
	fmt.Println("\n  You selected: " + f.filepicker.Styles.Selected.Render(mm.selectedFile) + "\n")
	return mm.selectedFile
}
