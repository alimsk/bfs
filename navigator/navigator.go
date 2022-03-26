// flutter-like navigation
package navigator

import tea "github.com/charmbracelet/bubbletea"

type navAction int

const (
	actionPush navAction = iota
	actionPushReplacement
	actionPop
	actionPushAndRemoveUntil
)

type navMsg struct {
	model     tea.Model
	predicate func(int, tea.Model) bool
	msg       tea.Msg
	action    navAction
}

func (n navMsg) cmd() tea.Cmd {
	if n.msg != nil {
		return func() tea.Msg { return n.msg }
	}
	return nil
}

func Push(m tea.Model) tea.Cmd {
	return func() tea.Msg {
		return navMsg{
			model:  m,
			action: actionPush,
		}
	}
}

func PushReplacement(m tea.Model) tea.Cmd {
	return func() tea.Msg {
		return navMsg{
			model:  m,
			action: actionPushReplacement,
		}
	}
}

func Pop() tea.Cmd {
	return func() tea.Msg {
		return navMsg{
			action: actionPop,
		}
	}
}

func PopWithResult[T any](v T) tea.Cmd {
	return func() tea.Msg {
		return navMsg{
			action: actionPop,
			msg:    ResultMsg[T]{v},
		}
	}
}

// pop until predicate returns true.
//
// current model will not be popped when it returns true.
func PushAndRemoveUntil(m tea.Model, predicate func(int, tea.Model) bool) tea.Cmd {
	return func() tea.Msg {
		return navMsg{
			model:     m,
			predicate: predicate,
			action:    actionPushAndRemoveUntil,
		}
	}
}

type ResultMsg[T any] struct{ Value T }

type Navigator struct {
	winsize tea.WindowSizeMsg
	models  []tea.Model
}

func New(initialModel tea.Model) Navigator {
	return Navigator{
		models: []tea.Model{initialModel},
	}
}

func (m Navigator) Init() tea.Cmd { return m.models[0].Init() }

func (m Navigator) View() string {
	if len(m.models) == 0 {
		return ""
	}
	return m.models[len(m.models)-1].View()
}

func (m Navigator) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.winsize = msg
	case navMsg:
		switch msg.action {
		case actionPush:
			m.models = append(m.models, msg.model)
			return m, tea.Batch(msg.model.Init(), m.winsizeCmd)
		case actionPushReplacement:
			m.models[len(m.models)-1] = msg.model
			return m, tea.Batch(msg.model.Init(), m.winsizeCmd)
		case actionPop:
			if len(m.models) == 1 {
				return m, tea.Quit
			}
			m.models = m.models[:len(m.models)-1]
			return m, tea.Batch(m.winsizeCmd, msg.cmd())
		case actionPushAndRemoveUntil:
			for i, model := range m.models {
				if msg.predicate(i, model) {
					break
				}
				m.models = m.models[:len(m.models)-1]
			}
			m.models = append(m.models, msg.model)
			return m, tea.Batch(msg.model.Init(), m.winsizeCmd)
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.models[len(m.models)-1], cmd = m.models[len(m.models)-1].Update(msg)
	return m, cmd
}

func (m Navigator) winsizeCmd() tea.Msg { return m.winsize }
