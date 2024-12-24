package messagebox

// Package MessageBox implements a modal messsage box for bubbletea.
//
// Activate by calling the MessageBox function from the Update method of the owning control.
// When a button is pressed, and the message box is dismissed, a value of type Button is returned
// wrapped in a tea.Cmd so that it can ben handled in the next call to the owning control's Update method.
//
// The control ownning the message box should call messageBox.Render as the last step in that control's View method
// to overlay the message box.

import (
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

type options struct {
	xpos  int
	ypos  int
	width int
	style *Styles
}

type optionFunc func(*options)

// Button represents a button in the message box
type Button int

const (
	MB_OK Button = 1 << (iota + 1)
	MB_CANCEL
	MB_YES
	MB_NO
	MB_ALL
)

// Type defines the type of message box to display
type Type int

// Available message box types
var (
	OK         = Type(MB_OK)
	OK_CANCEL  = Type(MB_OK | MB_CANCEL)
	YES_NO     = Type(MB_YES | MB_NO)
	YES_NO_ALL = Type(MB_YES | MB_NO | MB_ALL)
)

// Coloring for buttons
var (
	buttonFg     = lipgloss.Color("0")
	buttonBg     = lipgloss.Color("244")
	buttonSelBg  = lipgloss.Color("7")
	buttonHotkey = lipgloss.Color("196")
	border       = lipgloss.Color("63")
)

const defaultViewPortWidth = 40

var buttonText = map[Button]string{
	MB_OK:     "&Ok",
	MB_CANCEL: "&Cancel",
	MB_YES:    "&Yes",
	MB_NO:     "&No",
	MB_ALL:    "&All",
}

// highlightChar gets the char to highlight as a hotkey (preceded by &)
func (b Button) highlightChar() string {
	text := buttonText[b]
	highlight := strings.Index(text, "&")
	return string(text[highlight+1])
}

// keyBinding generates a key.Binding for this button.
func (b Button) keyBinding() key.Binding {

	highlight := b.highlightChar()

	binding := key.NewBinding(key.WithKeys(func() []string {
		keys := []string{strings.ToLower(highlight)}

		if b&(MB_CANCEL|MB_NO) != 0 {
			keys = append(keys, "esc")
		}

		return keys
	}()...))

	return binding
}

// render renders the button
func (b Button) render(style Styles, selected bool) string {

	text := buttonText[b]
	highlight := strings.Index(text, "&")
	highlightedChar := text[highlight+1]
	pre := text[:highlight]
	post := text[highlight+2:]

	buttonStyle := func() lipgloss.Style {
		if selected {
			return style.SelectedButton
		}

		return style.Button
	}()

	return buttonStyle.Render(" "+pre) +
		buttonStyle.Underline(true).
			Foreground(buttonHotkey).
			Render(string(highlightedChar)) +
		buttonStyle.Render(post+" ")
}

// Styles contains style definitions for this list component. By default, these
// values are generated by DefaultStyles.
type Styles struct {
	Border         lipgloss.Style
	Button         lipgloss.Style
	SelectedButton lipgloss.Style
	HotKey         lipgloss.Color // Text color of hotkey. Hotkey will also be undelined
}

// DefaultStyles returns a set of default style definitions for this table.
func DefaultStyles() Styles {
	return Styles{
		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(border)),
		Button: lipgloss.NewStyle().
			Foreground(lipgloss.Color(buttonFg)).
			Background(lipgloss.Color(buttonBg)),
		SelectedButton: lipgloss.NewStyle().
			Foreground(lipgloss.Color(buttonFg)).
			Background(lipgloss.Color(buttonSelBg)),
		HotKey: lipgloss.Color(buttonHotkey),
	}
}

// box manages an active message box
type box struct {
	message        string
	buttons        []Button
	selectedButton int
}

// Model is the bubbletea model for message box.
type Model struct {
	// Viewport with which to render the box
	viewport viewport.Model

	// Active message box, or nil when no message box showing
	box *box

	// X position in cursor coords
	xpos int

	// Y position in cursor coords
	ypos int

	// Width of box
	width int

	// Message box styling
	styles Styles
}

// WithPosition sets the position of the top left of the messagebox in
// columns from the left (x), and rows from the top (y).
func WithPosition(x, y int) optionFunc {
	return func(o *options) {
		o.xpos = x
		o.ypos = y
	}
}

// WithWidth sets the width of the message box. This will not be narrower than the space required to render the buttons.
// Height is computed from the message text.
func WithWidth(w int) optionFunc {
	return func(o *options) {
		o.width = w
	}
}

// WithStyle overrides the default style for the message box
func WithStyle(s Styles) optionFunc {
	return func(o *options) {
		o.style = &s
	}
}

// New creates a new modal message box with the given options.
// You would normally do this in the parent control's Update method in response to a key message.
//
// While the message box in in an active state, you should direct all UI messages to its update method.
func (m Model) New(message string, boxType Type, opts ...optionFunc) Model {

	o := &options{}

	for _, opt := range opts {
		opt(o)
	}

	m.xpos = o.xpos
	m.ypos = o.ypos

	if o.style == nil {
		m.styles = DefaultStyles()
	} else {
		m.styles = *o.style
	}

	buttons := []Button{}
	var selectedButton int

	for _, b := range []Button{MB_OK, MB_YES, MB_NO, MB_ALL, MB_CANCEL} {
		if Button(boxType)&b != 0 {
			buttons = append(buttons, b)
		}
	}

	if len(buttons) == 1 {
		selectedButton = 0
	} else {
		selectedButton = slices.Index(buttons, MB_CANCEL)

		if selectedButton == -1 {
			selectedButton = slices.Index(buttons, MB_NO)
		}

		if selectedButton == -1 {
			selectedButton = 0
		}
	}

	m.box = &box{
		buttons:        buttons,
		selectedButton: selectedButton,
	}

	m.width = defaultViewPortWidth

	// Size the viewport. Has to be wide enough for button bar.
	buttonBar := m.renderButtons()
	buttonsWidth := runewidth.StringWidth(buttonBar) + 2

	if o.width != 0 {
		// User requested width
		m.width = max(buttonsWidth, o.width)
	}

	m.box.message = runewidth.Wrap(strings.TrimSpace(message), m.width-2)
	m.viewport = viewport.New(m.width, strings.Count(m.box.message, "\n")+3)

	return m
}

// Init satisfies the BubbleTea Model interface.
// Does nothing here.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update satisfies the BubbleTea Model interface.
// Processes key messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.box == nil {
		return m, nil
	}

	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.Type {

		case tea.KeyEsc:

			// Return the button most suited to "take no action"
			buttonToReturn := func() Button {
				switch {
				case slices.Contains(m.box.buttons, MB_CANCEL):
					return MB_CANCEL

				case slices.Contains(m.box.buttons, MB_NO):
					return MB_NO

				default:
					return MB_CANCEL
				}
			}

			// Dismiss message box
			m.box = nil

			return m, func() tea.Cmd {
				return func() tea.Msg {
					// Return pressed button as message for caller's model update
					return buttonToReturn
				}
			}()

		case tea.KeyCtrlI, tea.KeyRight:

			// Forward tab between buttons
			m.box.selectedButton = (m.box.selectedButton + 1) % len(m.box.buttons)
			return m, nil

		case tea.KeyShiftTab, tea.KeyLeft:

			// Reverse tab between buttons
			m.box.selectedButton = (len(m.box.buttons) + m.box.selectedButton - 1) % len(m.box.buttons)
			return m, nil

		case tea.KeySpace, tea.KeyEnter:

			// Get selected button before dismissal
			selectedButton := m.box.buttons[m.box.selectedButton]

			// Dismiss message box
			m.box = nil

			return m, func() tea.Cmd {
				return func() tea.Msg {
					// Return pressed button as message for caller's model update
					return selectedButton
				}
			}()

		default:
			// If a bound key is pressed, return that key's button and dismiss message box
			for _, b := range m.box.buttons {
				if key.Matches(msg, b.keyBinding()) {
					// Dismiss message box
					m.box = nil
					return m, func() tea.Cmd {
						return func() tea.Msg {
							// Return pressed button as message for caller's model update
							return b
						}
					}()
				}
			}
		}
	}

	return m, nil
}

// View doesn't do anything, and it should never be called directly
// Implemented as part of BubbleTea Model interface
func (m Model) View() string {
	return ""
}

// Render takes in the main view content and overlays the model's active message box.
// This function expects you to build the entirety of your view's content before calling
// this function. It's recommended for this to be the final call of your model's View().
// Returns a string representation of the content with overlayed message box.
func (m Model) Render(content string) string {
	if m.box == nil {
		return content
	}

	center := lipgloss.NewStyle().Width(m.width - 2).Align(lipgloss.Center)

	m.viewport.SetContent(
		center.Render(m.box.message) + "\n\n" + center.Render(m.renderButtons()),
	)

	return PlaceOverlay(m.xpos, m.ypos, m.styles.Border.Render(m.viewport.View()), content)
}

// IsActive returns true if a message box is currently being displayed
func (m Model) IsActive() bool {
	return m.box != nil
}

// Renders the buttons
func (m Model) renderButtons() string {
	bs := []string{}

	for i, b := range m.box.buttons {
		bs = append(bs, b.render(m.styles, i == m.box.selectedButton))
	}

	return strings.Join(bs, " ")
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}
