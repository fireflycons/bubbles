package messagebox

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

type options struct {
	ypos int
}

type optionFunc func(*options)

type Button int

const (
	MB_OK Button = 1 << (iota + 1)
	MB_CANCEL
	MB_YES
	MB_NO
	MB_ALL
)

type Type int

var (
	OK_CANCEL  = Type(MB_OK | MB_CANCEL)
	YES_NO     = Type(MB_YES | MB_NO)
	YES_NO_ALL = Type(MB_YES | MB_NO | MB_ALL)
)

const (
	buttonFg     = "0"
	buttonBg     = "7"
	buttonHotkey = "1"
)

const viewPortWidth = 40

var buttonText = map[Button]string{
	MB_OK:     "&Ok",
	MB_CANCEL: "&Cancel",
	MB_YES:    "&Yes",
	MB_NO:     "&No",
	MB_ALL:    "&All",
}

func (b Button) highlightChar() string {
	text := buttonText[b]
	highlight := strings.Index(text, "&")
	return string(text[highlight+1])

}
func (b Button) keyBinding() key.Binding {

	highlight := b.highlightChar()

	binding := key.NewBinding(key.WithKeys(func() []string {
		keys := []string{strings.ToLower(highlight)}

		if b&(MB_OK|MB_YES) != 0 {
			keys = append(keys, "enter")
		}

		if b&(MB_CANCEL|MB_NO) != 0 {
			keys = append(keys, "esc")
		}

		return keys
	}()...))

	return binding
}

var buttonStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color(buttonFg)).
	Background(lipgloss.Color(buttonBg))

func (b Button) render() string {

	text := buttonText[b]
	highlight := strings.Index(text, "&")
	highlightedChar := text[highlight+1]
	pre := text[:highlight]
	post := text[highlight+2:]

	return buttonStyle.Render(" ") +
		buttonStyle.Render(pre) +
		buttonStyle.
			Underline(true).
			Foreground(lipgloss.Color(buttonHotkey)).
			Background(lipgloss.Color(buttonBg)).
			Render(string(highlightedChar)) +
		buttonStyle.Render(post+" ")
}

var viewportStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63"))

type Box struct {
	message string
	buttons Button
}

type Model struct {
	viewport viewport.Model
	box      *Box
	ypos     int
}

func New() Model {
	return Model{
		viewport: viewport.New(viewPortWidth, 1),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.box == nil {
		return m, nil
	}

	switch msg := msg.(type) {

	case tea.KeyMsg:

		for b := range buttonText {
			if m.box.buttons&b != 0 && key.Matches(msg, b.keyBinding()) {
				// Dismiss message box
				m.box = nil
				return m, func() tea.Cmd {
					return func() tea.Msg {
						// Return pressed button as message for result_viewer Update method
						return b
					}
				}()
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

func WithYPosition(ypos int) optionFunc {
	return func(o *options) {
		o.ypos = ypos
	}
}

func (m Model) MessageBox(message string, boxType Type, opts ...optionFunc) Model {

	o := &options{}

	for _, opt := range opts {
		opt(o)
	}

	m.ypos = o.ypos
	msg := runewidth.Wrap(strings.TrimSpace(message), viewPortWidth-2)

	m.viewport.Height = strings.Count(message, "\n") + 3
	// m.windowH = windowH
	// m.windowW = windowW
	m.box = &Box{
		message: msg,
		buttons: Button(boxType),
	}

	return m
}

// Render takes in the main view content and overlays the model's active message box.
// This function expects you to build the entirety of your view's content before calling
// this function. It's recommended for this to be the final call of your model's View().
// Returns a string representation of the content with overlayed message box.
func (m Model) Render(content string) string {
	if m.box == nil {
		return content
	}

	buttons := func() string {

		bs := []string{}

		for _, b := range []Button{MB_OK, MB_YES, MB_NO, MB_CANCEL} {
			if m.box.buttons&b != 0 {
				bs = append(bs, b.render())
			}
		}

		return lipgloss.NewStyle().Width(viewPortWidth - 2).Align(lipgloss.Center).Render(strings.Join(bs, " "))
	}()

	m.viewport.SetContent(lipgloss.NewStyle().Width(viewPortWidth-2).Align(lipgloss.Center).Render(m.box.message) + "\n\n" + buttons)
	return PlaceOverlay(3, m.ypos, viewportStyle.Render(m.viewport.View()), content)
}

func (m Model) IsActive() bool {
	return m.box != nil
}
