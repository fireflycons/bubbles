package main

import (
	"maps"
	"slices"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fireflycons/bubbles/messagebox"
	"github.com/fireflycons/bubbles/xtable"
)

// orderedKeyBinding describes the actions performed for given key presses
type orderedKeyBinding struct {
	// Key binding
	binding key.Binding

	// Order for help display and determining
	// which UI component the action applies to
	order int

	// Action to perform in model's Update method
	action func(m model, msg tea.Msg) (model, tea.Cmd)
}

type KeyMap map[string]orderedKeyBinding

var _ help.KeyMap = (*KeyMap)(nil)

// xtableAction is the action to send a key to the xtable component for processing.
func xtableAction(m model, msg tea.Msg) (model, tea.Cmd) {
	if m.table.Focused() {

		mdl, cmd := m.table.Update(msg)
		m.table = mdl
		return m, cmd
	}

	return m, nil
}

// defaultKeyMap is the bindings map for all built-in actions
var defaultKeyMap = KeyMap{
	// Global keys (focus independent)
	"Quit": orderedKeyBinding{
		binding: key.NewBinding(
			key.WithKeys("esc", "ctrl+c"),
			key.WithHelp("ESC", "quit "),
		),
		order: 0,
		action: func(m model, _ tea.Msg) (model, tea.Cmd) {
			return m, tea.Quit
		},
	},

	// xtable keys (passed to xtable model)
	"LineUp": orderedKeyBinding{
		binding: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "up "),
		),
		order:  10,
		action: xtableAction,
	},

	"LineDown": orderedKeyBinding{
		binding: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "down "),
		),
		order:  11,
		action: xtableAction,
	},
	"PageUp": orderedKeyBinding{
		binding: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up "),
		),
		order:  12,
		action: xtableAction,
	},
	"PageDown": orderedKeyBinding{
		binding: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdn", "page down "),
		),
		order:  13,
		action: xtableAction,
	},
	"HalfPageUp": orderedKeyBinding{
		binding: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("^u", "½ page up "),
		),
		order:  14,
		action: xtableAction,
	},
	"HalfPageDown": orderedKeyBinding{
		binding: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("^d", "½ page down "),
		),
		order:  15,
		action: xtableAction,
	},
	"GotoTop": orderedKeyBinding{
		binding: key.NewBinding(
			key.WithKeys("home"),
			key.WithHelp("home", "go to start "),
		),
		order:  16,
		action: xtableAction,
	},
	"GotoBottom": orderedKeyBinding{
		binding: key.NewBinding(
			key.WithKeys("end"),
			key.WithHelp("end", "go to end "),
		),
		order:  17,
		action: xtableAction,
	},

	// Additional commands for example table
	"Sort": orderedKeyBinding{
		binding: key.NewBinding(
			key.WithKeys("0", "1", "2", "3", "4", "5", "6", "7", "8", "9"),
			key.WithHelp("1..0", "sort col "),
		),
		order: 20,
		action: func(m model, msg tea.Msg) (model, tea.Cmd) {
			// Do sort on numeric keys
			if m.table.Focused() {
				ascii := msg.(tea.KeyMsg).Runes[0]
				m.table.SortBy(sortKeyToIndex(ascii), xtable.SortAscending, "")
			}
			return m, nil
		},
	},
	"SortDesc": orderedKeyBinding{
		binding: key.NewBinding(
			key.WithKeys("alt+0", "alt+1", "alt+2", "alt+3", "alt+4", "alt+5", "alt+6", "alt+7", "alt+8", "alt+9"),
			key.WithHelp("M-1..0", "sort col desc "),
		),
		order: 21,
		action: func(m model, msg tea.Msg) (model, tea.Cmd) {
			// Do sort on numeric keys
			if m.table.Focused() {
				ascii := msg.(tea.KeyMsg).Runes[0]
				m.table.SortBy(sortKeyToIndex(ascii), xtable.SortDescending, "")
			}
			return m, nil
		},
	},
}

// toTableMap returns key bindings to pass to xtable component
// in the format it expects.
func (km KeyMap) toTableMap() xtable.KeyMap {
	return xtable.KeyMap{
		LineUp:       km["LineUp"].binding,
		LineDown:     km["LineDown"].binding,
		PageUp:       km["PageUp"].binding,
		PageDown:     km["PageDown"].binding,
		HalfPageUp:   km["HalfPageUp"].binding,
		HalfPageDown: km["HalfPageDown"].binding,
		GotoTop:      km["GotoTop"].binding,
		GotoBottom:   km["GotoBottom"].binding,
	}
}

// ShortHelp implements the KeyMap interface.
func (km KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{km["LineUp"].binding, km["LineDown"].binding}
}

// FullHelp implements the KeyMap interface.
func (km KeyMap) FullHelp() [][]key.Binding {

	helpBindings := [][]key.Binding{}

	// Go 1.23 - orderedBindings := slices.SortedFunc(maps.Values(m), func(a, b orderedKeyBinding) int {...})
	orderedBindings := mapValues(km)
	slices.SortFunc(orderedBindings, func(a, b orderedKeyBinding) int {
		if a.order > b.order {
			return 1
		}

		if a.order < b.order {
			return -1
		}

		return 0
	})

	// Map binding into 2D array.
	// 2 rows and as many columns as needed to print 2 lines of help keys at the bottom.
	for i := 0; i < len(orderedBindings); i += 2 {
		end := i + 2
		if end > len(orderedBindings) {
			end = len(orderedBindings)
		}
		col := []key.Binding{}
		for _, b := range orderedBindings[i:end] {
			col = append(col, b.binding)
		}
		helpBindings = append(helpBindings, col)
	}

	return helpBindings
}

var messageBoxStyle = func() messagebox.Styles {
	s := messagebox.DefaultStyles()
	s.Border = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("12"))
	return s
}()

// buildKeyMap creates map of key bindings to action function for user actions passed in
// appending them to the default actions
func buildKeyMap(actions []userAction) KeyMap {

	keymap := maps.Clone(defaultKeyMap)

	for i, a := range actions {
		keymap[a.Message] = orderedKeyBinding{
			binding: a.Launch,
			order:   30 + i,
			action: func(m model, msg tea.Msg) (model, tea.Cmd) {

				m.currentAction = &a

				if a.MessageBoxType > 0 {
					// Action has an accociated message box for confirmation
					y := m.table.SelectedRowYOffset()
					m.msgBox = m.msgBox.New(a.Message, a.MessageBoxType, messagebox.WithPosition(3, y+6), messagebox.WithStyle(messageBoxStyle))
					return m, nil

				} else {

					// Direct action without message box
					return m, func() tea.Msg {
						// Simulate messagebox raised and OK pressed
						return messagebox.MB_OK
					}
				}
			},
		}
	}

	return keymap
}

func isdigit(ascii rune) bool {
	return ascii >= 48 && ascii <= 57
}

func sortKeyToIndex(key rune) int {
	if !isdigit(key) {
		return 0
	}

	ind := int(key) - 48

	if ind == 0 {
		ind = 10
	}
	return ind
}

// mapValues gets values of a map as a slice, without requiring maps.Values in newer golang versions
func mapValues[K comparable, V any](m map[K]V) []V {
	result := make([]V, 0, len(m))

	for _, v := range m {
		result = append(result, v)
	}

	return result
}
