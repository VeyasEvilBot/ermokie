package models

import (
	"reflect"
	"testing"

	"codeberg.org/veya/ermokie/pkg/models/styles"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestFinder(t *testing.T, rows []string, multi, embedded bool) *fuzzyFinder {
	t.Helper()
	finder, err := newFuzzyFinderWithThemeAndFlag(rows, multi, styles.DefaultStyles(), "test", embedded, nil)
	if err != nil {
		t.Fatal(err)
	}
	model, _ := finder.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	updated, ok := model.(*fuzzyFinder)
	if !ok {
		t.Fatalf("window update model = %T, want *fuzzyFinder", model)
	}
	return updated
}

func TestFuzzyFinderDeduplicatesRowsAndNavigatesWithConfiguredKeys(t *testing.T) {
	old := GlobalKeybindingManager
	GlobalKeybindingManager = NewKeybindingManager()
	defer func() { GlobalKeybindingManager = old }()
	if err := GlobalKeybindingManager.ApplyOverrides(map[string][]string{"down": {"n"}}); err != nil {
		t.Fatal(err)
	}

	finder := newTestFinder(t, []string{"alpha", "alpha", "beta"}, false, true)
	if !reflect.DeepEqual(finder.rows, []string{"alpha", "beta"}) {
		t.Fatalf("rows = %#v", finder.rows)
	}
	model, _ := finder.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	updated, ok := model.(*fuzzyFinder)
	if !ok {
		t.Fatalf("key update model = %T, want *fuzzyFinder", model)
	}
	if got := updated.cursor; got != 1 {
		t.Fatalf("cursor = %d, want 1", got)
	}
}

func TestFuzzyFinderCancelQuitsStandaloneAndNotifiesEmbeddedParent(t *testing.T) {
	standalone := newTestFinder(t, []string{"alpha"}, false, false)
	_, cmd := standalone.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("standalone cancel returned no command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("standalone cancel message = %T, want tea.QuitMsg", cmd())
	}

	embedded := newTestFinder(t, []string{"alpha"}, false, true)
	_, cmd = embedded.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("embedded cancel returned no command")
	}
	if _, ok := cmd().(cancelPickerMsg); !ok {
		t.Fatalf("embedded cancel message = %T, want cancelPickerMsg", cmd())
	}
}

func TestFuzzyFinderDoesNotConfirmWhenFilterHasNoMatches(t *testing.T) {
	finder := newTestFinder(t, []string{"alpha"}, false, true)
	finder.input.SetValue("zzz")
	finder.liveValue = "zzz"
	finder.refreshContent()

	model, cmd := finder.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatalf("empty result unexpectedly emitted %T", cmd())
	}
	updated, ok := model.(*fuzzyFinder)
	if !ok {
		t.Fatalf("confirm model = %T, want *fuzzyFinder", model)
	}
	if got := updated.selection; len(got) != 0 {
		t.Fatalf("selection = %#v", got)
	}
}

func TestFuzzyFinderSelectsFilteredResult(t *testing.T) {
	finder := newTestFinder(t, []string{"alpha", "beta", "gamma"}, false, true)
	finder.input.SetValue("gm")
	finder.liveValue = "gm"
	finder.refreshContent()

	_, cmd := finder.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("confirm returned no command")
	}
	selection, ok := cmd().(selectionMsg)
	if !ok || !reflect.DeepEqual([]string(selection), []string{"gamma"}) {
		t.Fatalf("selection message = %#v", selection)
	}
}
