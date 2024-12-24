package main

import (
	"fmt"
	"hash/fnv"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fireflycons/bubbles/messagebox"
	"github.com/fireflycons/bubbles/xtable"
)

type User struct {
	Name    string `xtable:"User Name"`
	Age     int
	private string // Unexported field, will be ignored
	Email   string `xtable:"E-mail Address"`
}

var users = []User{
	{Name: "Hank", Age: 59, Email: "Hank@example.com"},
	{Name: "Oscar", Age: 21, Email: "Oscar@sample.net"},
	{Name: "Quincy", Age: 59, Email: "Quincy@demo.co"},
	{Name: "Rita", Age: 62, Email: "Rita@example.com"},
	{Name: "Zane", Age: 31, Email: "Zane@test.org"},
	{Name: "Bob", Age: 36, Email: "Bob@example.com"},
	{Name: "Ivy", Age: 66, Email: "Ivy@sample.net"},
	{Name: "Jack", Age: 25, Email: "Jack@sample.net"},
	{Name: "Mona", Age: 23, Email: "Mona@example.com"},
	{Name: "Grace", Age: 50, Email: "Grace@test.org"},
	{Name: "Karen", Age: 60, Email: "Karen@sample.net"},
	{Name: "Nancy", Age: 51, Email: "Nancy@example.com"},
	{Name: "Bob", Age: 28, Email: "Bob@myemail.com"},
	{Name: "Charlie", Age: 36, Email: "Charlie@test.org"},
	{Name: "Diana", Age: 32, Email: "Diana@example.com"},
	{Name: "Quincy", Age: 37, Email: "Quincy@sample.net"},
	{Name: "Charlie", Age: 24, Email: "Charlie@sample.net"},
	{Name: "Nancy", Age: 33, Email: "Nancy@demo.co"},
	{Name: "Oscar", Age: 57, Email: "Oscar@example.com"},
	{Name: "Nancy", Age: 58, Email: "Nancy@myemail.com"},
	{Name: "Oscar", Age: 40, Email: "Oscar@myemail.com"},
	{Name: "Yara", Age: 59, Email: "Yara@test.org"},
	{Name: "Yara", Age: 27, Email: "Yara@sample.net"},
	{Name: "Zane", Age: 35, Email: "Zane@demo.co"},
	{Name: "Victor", Age: 61, Email: "Victor@test.org"},
	{Name: "Wendy", Age: 41, Email: "Wendy@sample.net"},
	{Name: "Frank", Age: 31, Email: "Frank@test.org"},
	{Name: "Eve", Age: 20, Email: "Eve@demo.co"},
	{Name: "Mona", Age: 51, Email: "Mona@demo.co"},
	{Name: "Wendy", Age: 22, Email: "Wendy@myemail.com"},
}

// actionReturn describes what should be done with the row upon which a user action (via messagebox) is performed
type actionReturn int

const (
	ACTION_NONE actionReturn = iota
	ACTION_DELETE
	ACTION_DELETE_ALL
)

// userAction describes a messagebox and subsequent action to be performed
// on the selected row when the key(s) identified by the key binding is pressed.
type userAction struct {
	// Key(s) to launch the action
	Launch key.Binding

	// Type of message box to show
	MessageBoxType messagebox.Type

	// Message to display
	Message string

	// Action to perform when message box is dismissed.
	// Interface argument contains row metadata
	Action func(messagebox.Button, interface{}) actionReturn
}

// Assert interface implementation
var _ xtable.Metadata = (*User)(nil)

func (u User) GetHashCode() uint64 {
	fnv_1 := fnv.New64a()
	fnv_1.Write([]byte(u.Email))
	return fnv_1.Sum64()
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table         xtable.Model
	msgBox        messagebox.Model
	help          help.Model
	actions       []userAction
	currentAction *userAction
	keymap        KeyMap
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.msgBox.IsActive() {
		// Send all messages to message box
		m1, cmd1 := m.msgBox.Update(msg)
		m.msgBox = m1.(messagebox.Model)
		return m, cmd1
	}

	switch msg := msg.(type) {

	case messagebox.Button:
		// Act on message box button
		if m.currentAction != nil && m.currentAction.Action != nil {
			switch m.currentAction.Action(msg, m.table.SelectedRow().Metadata) {
			case ACTION_DELETE_ALL:

				// There would be no data left to display, so
				return m, tea.Quit

			case ACTION_DELETE:

				if stillHaveRows := m.table.RemoveSelectedRow(); !stillHaveRows {
					// Deleted last row
					return m, tea.Quit
				}
			}
		}

	case tea.KeyMsg:

		// If key matches what's in the keymap, perform that action.
		for _, v := range m.keymap {
			if key.Matches(msg, v.binding) {

				m, cmd := v.action(m, msg)
				return m, cmd
			}
		}
	}
	return m, cmd
}

func (m model) View() string {
	sb := strings.Builder{}
	sb.WriteString(baseStyle.Render(m.table.View()) + "\n")
	sb.WriteString(m.help.View(m.keymap))

	// Overlay active messagebox if any
	return m.msgBox.Render(sb.String())
}

func main() {

	t := xtable.New(
		xtable.WithStructData(users), // Auto-populate table from slice of struct
		xtable.WithRowNumbers(),      // Add row number column
		xtable.WithFocused(true),
		xtable.WithKeyMap(defaultKeyMap.toTableMap()),
	)

	s := xtable.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	helpMdl := help.New()
	helpMdl.ShowAll = true
	helpMdl.Styles.FullKey = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))

	// Additional actions to process in model's Update method.
	actions := []userAction{
		{
			Launch:         key.NewBinding(key.WithKeys("delete"), key.WithHelp("DEL", "Delete")), // Key to launch action
			MessageBoxType: messagebox.YES_NO_ALL,                                                 // Type of message box to display (0 = no message box, just do it)
			Message:        "Delete selected?",                                                    // Message to display in box
			Action: func(b messagebox.Button, rowData interface{}) actionReturn {
				// Action to perform when message box is dismissed
				switch b {

				case messagebox.MB_YES:

					toDelete := rowData.(User)
					// Perform actions to delete user identified by rowData
					_ = toDelete

					return ACTION_DELETE // Remove row (handled by your model)

				case messagebox.MB_ALL:

					// Peform actions to delete all users that were in the table.
					// It is assumed this function has access to the []User that
					// was used to create the table.
					// Here it does, since var users is in this file.

					return ACTION_DELETE_ALL // Remove all rows and quit BubbleTea program (handled by your model)
				}

				return ACTION_NONE
			},
		},
	}

	m := model{
		table:   t,
		help:    helpMdl,
		actions: actions,
		keymap:  buildKeyMap(actions),
	}

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
