package models

import (
	"slices"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type KeybindingAction string

const (
	ActionUp               KeybindingAction = "up"
	ActionDown             KeybindingAction = "down"
	ActionLeft             KeybindingAction = "left"
	ActionRight            KeybindingAction = "right"
	ActionSelect           KeybindingAction = "select"
	ActionCancel           KeybindingAction = "cancel"
	ActionConfirm          KeybindingAction = "confirm"
	ActionQuit             KeybindingAction = "quit"
	ActionHelp             KeybindingAction = "help"
	ActionYes              KeybindingAction = "yes"
	ActionNo               KeybindingAction = "no"
	ActionOpenGameSwitcher KeybindingAction = "switch game"
	ActionJumpToTop        KeybindingAction = "jump to top"
	ActionJumpToBottom     KeybindingAction = "jump to bottom"
)

type KeybindingSet struct {
	Name        string
	Description string
	Bindings    map[KeybindingAction][]string
}

type KeybindingManager struct {
	availableSets map[string]KeybindingSet
	setOrder      []string
	currentSet    string
}

func init() {
	GlobalKeybindingManager.SetKeybindingOrder([]string{
		"arrow-vim",
		"default",
		"arrow-keys",
		"wasd",
	})
}

func NewKeybindingManager() *KeybindingManager {
	km := &KeybindingManager{
		availableSets: make(map[string]KeybindingSet),
		setOrder:      make([]string, 0, 8),
		currentSet:    "default",
	}

	km.availableSets["default"] = KeybindingSet{
		Name:        "default",
		Description: "Default keybindings with vim-style navigation",
		Bindings: map[KeybindingAction][]string{
			ActionUp:               {"up", "k"},
			ActionDown:             {"down", "j"},
			ActionLeft:             {"left", "h"},
			ActionRight:            {"right", "l"},
			ActionSelect:           {"tab", " "},
			ActionOpenGameSwitcher: {"s"},
			ActionJumpToTop:        {"g"},
			ActionJumpToBottom:     {"G"},
			ActionCancel:           {"esc"},
			ActionConfirm:          {"enter"},
			ActionQuit:             {"q", "esc", "ctrl+c"},
			ActionHelp:             {"?", "f1"},
			ActionYes:              {"y", "Y"},
			ActionNo:               {"n", "N"},
		},
	}
	km.setOrder = append(km.setOrder, "default")

	km.availableSets["arrow-keys"] = KeybindingSet{
		Name:        "arrow-keys",
		Description: "Arrow keys only, no vim-style bindings",
		Bindings: map[KeybindingAction][]string{
			ActionUp:               {"up"},
			ActionDown:             {"down"},
			ActionLeft:             {"left"},
			ActionRight:            {"right"},
			ActionSelect:           {"tab", " "},
			ActionOpenGameSwitcher: {"s"},
			ActionJumpToTop:        {"home"},
			ActionJumpToBottom:     {"end"},
			ActionCancel:           {"esc"},
			ActionConfirm:          {"enter"},
			ActionQuit:             {"q", "esc", "ctrl+c"},
			ActionHelp:             {"?", "f1"},
			ActionYes:              {"y", "Y"},
			ActionNo:               {"n", "N"},
		},
	}
	km.setOrder = append(km.setOrder, "arrow-keys")

	km.availableSets["wasd"] = KeybindingSet{
		Name:        "wasd",
		Description: "WASD gaming-style navigation",
		Bindings: map[KeybindingAction][]string{
			ActionUp:               {"up", "w", "W"},
			ActionDown:             {"down", "s", "S"},
			ActionLeft:             {"left", "a", "A"},
			ActionRight:            {"right", "d", "D"},
			ActionSelect:           {"tab", " "},
			ActionOpenGameSwitcher: {"g"},
			ActionJumpToTop:        {"home"},
			ActionJumpToBottom:     {"end"},
			ActionCancel:           {"esc"},
			ActionConfirm:          {"enter"},
			ActionQuit:             {"q", "esc", "ctrl+c"},
			ActionHelp:             {"?", "f1"},
			ActionYes:              {"y", "Y"},
			ActionNo:               {"n", "N"},
		},
	}
	km.setOrder = append(km.setOrder, "wasd")

	km.availableSets["arrow-vim"] = KeybindingSet{
		Name:        "arrow-vim",
		Description: "Arrow keys + vim-style navigation (best of both worlds)",
		Bindings: map[KeybindingAction][]string{
			ActionUp:               {"up", "k"},
			ActionDown:             {"down", "j"},
			ActionLeft:             {"left", "h"},
			ActionRight:            {"right", "l"},
			ActionSelect:           {"tab", " "},
			ActionOpenGameSwitcher: {"s"},
			ActionJumpToTop:        {"g"},
			ActionJumpToBottom:     {"G"},
			ActionCancel:           {"esc"},
			ActionConfirm:          {"enter"},
			ActionQuit:             {"q", "esc", "ctrl+c"},
			ActionHelp:             {"?", "f1"},
			ActionYes:              {"y", "Y"},
			ActionNo:               {"n", "N"},
		},
	}
	km.setOrder = append(km.setOrder, "arrow-vim")

	km.availableSets["vim-only"] = KeybindingSet{
		Name:        "vim-only",
		Description: "Pure vim-style navigation (hjkl only, no arrow keys)",
		Bindings: map[KeybindingAction][]string{
			ActionUp:               {"k"},
			ActionDown:             {"j"},
			ActionLeft:             {"h"},
			ActionRight:            {"l"},
			ActionSelect:           {"tab", " "},
			ActionOpenGameSwitcher: {"s"},
			ActionJumpToTop:        {"g"},
			ActionJumpToBottom:     {"G"},
			ActionCancel:           {"esc"},
			ActionConfirm:          {"enter"},
			ActionQuit:             {"q", "esc", "ctrl+c"},
			ActionHelp:             {"?", "f1"},
			ActionYes:              {"y", "Y"},
			ActionNo:               {"n", "N"},
		},
	}
	km.setOrder = append(km.setOrder, "vim-only")

	km.availableSets["fuzzy-picker"] = KeybindingSet{
		Name:        "fuzzy-picker",
		Description: "Arrow keys only for fuzzy picker components",
		Bindings: map[KeybindingAction][]string{
			ActionUp:      {"up"},
			ActionDown:    {"down"},
			ActionLeft:    {"left"},
			ActionRight:   {"right"},
			ActionSelect:  {"tab", " "},
			ActionCancel:  {"esc"},
			ActionConfirm: {"enter"},
			ActionQuit:    {"q", "esc", "ctrl+c"},
			ActionHelp:    {"?"},
			ActionYes:     {"y", "Y"},
			ActionNo:      {"n", "N"},
		},
	}
	km.setOrder = append(km.setOrder, "fuzzy-picker")

	return km
}

func (km *KeybindingManager) GetAvailableKeybindingSets() []string {
	out := make([]string, len(km.availableSets))
	copy(out, km.setOrder)
	return out
}

func (km *KeybindingManager) SetKeybindingOrder(order []string) {
	seen := make(map[string]bool, len(order))
	newOrder := make([]string, 0, len(km.availableSets))

	for _, name := range order {
		if _, ok := km.availableSets[name]; ok && !seen[name] {
			newOrder = append(newOrder, name)
			seen[name] = true
		}
	}

	for _, name := range km.setOrder {
		if !seen[name] {
			if _, ok := km.availableSets[name]; ok {
				newOrder = append(newOrder, name)
				seen[name] = true
			}
		}
	}

	km.setOrder = newOrder
}

func (km *KeybindingManager) AddKeybindingSet(set KeybindingSet) {
	_, existed := km.availableSets[set.Name]
	km.availableSets[set.Name] = set
	if !existed {
		km.setOrder = append(km.setOrder, set.Name)
	}
}

func (km *KeybindingManager) RemoveKeybindingSet(name string) bool {
	if name == km.currentSet {
		return false
	}
	if _, ok := km.availableSets[name]; !ok {
		return false
	}
	delete(km.availableSets, name)
	newOrder := make([]string, 0, len(km.setOrder))
	for _, n := range km.setOrder {
		if n != name {
			newOrder = append(newOrder, n)
		}
	}
	km.setOrder = newOrder
	return true
}

func (km *KeybindingManager) GetKeybindingSetDescription(setName string) string {
	if set, exists := km.availableSets[setName]; exists {
		return set.Description
	}
	return "Unknown keybinding set"
}

func (km *KeybindingManager) SetCurrentKeybindingSet(setName string) bool {
	if _, exists := km.availableSets[setName]; exists {
		km.currentSet = setName
		return true
	}
	return false
}

func (km *KeybindingManager) GetCurrentKeybindingSet() string {
	return km.currentSet
}

func (km *KeybindingManager) GetKeysForAction(action KeybindingAction) []string {
	if set, exists := km.availableSets[km.currentSet]; exists {
		if keys, actionExists := set.Bindings[action]; actionExists {
			return keys
		}
	}
	return []string{}
}

func (km *KeybindingManager) CreateKeyBinding(action KeybindingAction, help string) key.Binding {
	keys := km.GetKeysForAction(action)
	if len(keys) == 0 {
		return key.NewBinding()
	}

	return key.NewBinding(
		key.WithKeys(keys...),
		key.WithHelp(keys[0], help),
	)
}

func (km *KeybindingManager) MatchesAction(msg tea.KeyMsg, action KeybindingAction) bool {
	keys := km.GetKeysForAction(action)
	return slices.Contains(keys, msg.String())
}

var GlobalKeybindingManager = NewKeybindingManager()

func LoadKeymapFromConfig(keymapName string) bool {
	if keymapName == "" {
		keymapName = "arrow-vim"
	}
	return GlobalKeybindingManager.SetCurrentKeybindingSet(keymapName)
}

func GetCurrentKeymapName() string {
	return GlobalKeybindingManager.GetCurrentKeybindingSet()
}

func GetAvailableKeymaps() []string {
	allKeymaps := GlobalKeybindingManager.GetAvailableKeybindingSets()
	out := make([]string, 0, len(allKeymaps))
	for _, keymap := range allKeymaps {
		if keymap != "fuzzy-picker" {
			out = append(out, keymap)
		}
	}
	return out
}

func IsValidKeymap(keymapName string) bool {
	available := GetAvailableKeymaps()
	return slices.Contains(available, keymapName)
}

func GetQuitKeys() key.Binding {
	return GlobalKeybindingManager.CreateKeyBinding(ActionQuit, "quit")
}

func GetUpKeys() []string {
	return GlobalKeybindingManager.GetKeysForAction(ActionUp)
}

func GetDownKeys() []string {
	return GlobalKeybindingManager.GetKeysForAction(ActionDown)
}

func GetLeftKeys() []string {
	return GlobalKeybindingManager.GetKeysForAction(ActionLeft)
}

func GetRightKeys() []string {
	return GlobalKeybindingManager.GetKeysForAction(ActionRight)
}

func GetSelectKeys() []string {
	return GlobalKeybindingManager.GetKeysForAction(ActionSelect)
}

func GetCancelKeys() []string {
	return GlobalKeybindingManager.GetKeysForAction(ActionCancel)
}

func GetConfirmKeys() []string {
	return GlobalKeybindingManager.GetKeysForAction(ActionConfirm)
}

func GetYesKeys() []string {
	return GlobalKeybindingManager.GetKeysForAction(ActionYes)
}

func GetNoKeys() []string {
	return GlobalKeybindingManager.GetKeysForAction(ActionNo)
}

func MatchesQuit(msg tea.KeyMsg) bool {
	return GlobalKeybindingManager.MatchesAction(msg, ActionQuit)
}

func MatchesOpenGameSwitcher(msg tea.KeyMsg) bool {
	return GlobalKeybindingManager.MatchesAction(msg, ActionOpenGameSwitcher)
}

func MatchesUp(msg tea.KeyMsg) bool {
	return GlobalKeybindingManager.MatchesAction(msg, ActionUp)
}

func MatchesDown(msg tea.KeyMsg) bool {
	return GlobalKeybindingManager.MatchesAction(msg, ActionDown)
}

func MatchesLeft(msg tea.KeyMsg) bool {
	return GlobalKeybindingManager.MatchesAction(msg, ActionLeft)
}

func MatchesRight(msg tea.KeyMsg) bool {
	return GlobalKeybindingManager.MatchesAction(msg, ActionRight)
}

func MatchesSelect(msg tea.KeyMsg) bool {
	return GlobalKeybindingManager.MatchesAction(msg, ActionSelect)
}

func MatchesCancel(msg tea.KeyMsg) bool {
	return GlobalKeybindingManager.MatchesAction(msg, ActionCancel)
}

func MatchesConfirm(msg tea.KeyMsg) bool {
	return GlobalKeybindingManager.MatchesAction(msg, ActionConfirm)
}

func MatchesYes(msg tea.KeyMsg) bool {
	return GlobalKeybindingManager.MatchesAction(msg, ActionYes)
}

func MatchesNo(msg tea.KeyMsg) bool {
	return GlobalKeybindingManager.MatchesAction(msg, ActionNo)
}
func GetFuzzyUpKeys() []string {
	if set, exists := GlobalKeybindingManager.availableSets["fuzzy-picker"]; exists {
		if keys, actionExists := set.Bindings[ActionUp]; actionExists {
			return keys
		}
	}
	return []string{"up"}
}

func GetFuzzyDownKeys() []string {
	if set, exists := GlobalKeybindingManager.availableSets["fuzzy-picker"]; exists {
		if keys, actionExists := set.Bindings[ActionDown]; actionExists {
			return keys
		}
	}
	return []string{"down"}
}

func GetFuzzyLeftKeys() []string {
	if set, exists := GlobalKeybindingManager.availableSets["fuzzy-picker"]; exists {
		if keys, actionExists := set.Bindings[ActionLeft]; actionExists {
			return keys
		}
	}
	return []string{"left"}
}

func GetFuzzyRightKeys() []string {
	if set, exists := GlobalKeybindingManager.availableSets["fuzzy-picker"]; exists {
		if keys, actionExists := set.Bindings[ActionRight]; actionExists {
			return keys
		}
	}
	return []string{"right"}
}

func MatchesFuzzyUp(msg tea.KeyMsg) bool {
	keys := GetFuzzyUpKeys()
	return slices.Contains(keys, msg.String())
}

func MatchesFuzzyDown(msg tea.KeyMsg) bool {
	keys := GetFuzzyDownKeys()
	return slices.Contains(keys, msg.String())
}

func MatchesFuzzyLeft(msg tea.KeyMsg) bool {
	keys := GetFuzzyLeftKeys()
	return slices.Contains(keys, msg.String())
}

func MatchesFuzzyRight(msg tea.KeyMsg) bool {
	keys := GetFuzzyRightKeys()
	return slices.Contains(keys, msg.String())
}

func MatchesJumpToTop(msg tea.KeyMsg) bool {
	return GlobalKeybindingManager.MatchesAction(msg, ActionJumpToTop)
}

func MatchesJumpToBottom(msg tea.KeyMsg) bool {
	return GlobalKeybindingManager.MatchesAction(msg, ActionJumpToBottom)
}

func MatchesHelp(msg tea.KeyMsg) bool {
	return GlobalKeybindingManager.MatchesAction(msg, ActionHelp)
}
