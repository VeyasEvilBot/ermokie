package models

import (
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAvailableKeybindingSetsAreOrderedWithoutEmptyEntries(t *testing.T) {
	km := NewKeybindingManager()
	km.SetKeybindingOrder([]string{"arrow-vim", "default", "arrow-keys", "wasd"})

	got := km.GetAvailableKeybindingSets()
	want := []string{"arrow-vim", "default", "arrow-keys", "wasd", "vim-only", "fuzzy-picker"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetAvailableKeybindingSets() = %#v, want %#v", got, want)
	}
}

func TestApplyOverridesChangesOnlyRequestedActions(t *testing.T) {
	km := NewKeybindingManager()

	err := km.ApplyOverrides(map[string][]string{
		"up":   {"w"},
		"quit": {"x", "ctrl+c"},
	})
	if err != nil {
		t.Fatalf("ApplyOverrides() error = %v", err)
	}

	if got := km.GetKeysForAction(ActionUp); !reflect.DeepEqual(got, []string{"w"}) {
		t.Fatalf("up keys = %#v", got)
	}
	if got := km.GetKeysForAction(ActionDown); !reflect.DeepEqual(got, []string{"down", "j"}) {
		t.Fatalf("down keys unexpectedly changed: %#v", got)
	}
	if !km.MatchesAction(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, ActionQuit) {
		t.Fatal("custom quit key does not match")
	}
}

func TestApplyOverridesRejectsUnknownAndEmptyBindings(t *testing.T) {
	km := NewKeybindingManager()
	if err := km.ApplyOverrides(map[string][]string{"explode": {"x"}}); err == nil {
		t.Fatal("unknown action was accepted")
	}
	if err := km.ApplyOverrides(map[string][]string{"up": {}}); err == nil {
		t.Fatal("empty binding was accepted")
	}
}
