package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/harshagw/viri/internal/vm"
)

// Debugger interface to avoid circular dependency
type Debugger interface {
	Run()
	StepForward()
	StepBack()
	Continue()
	Reset()
	CurrentState() *vm.VMState
	Position() (current, total int)
	IsDone() bool
	Error() error
}

type Model struct {
	debugger          Debugger
	state             *vm.VMState
	width             int
	height            int
	ready             bool
	bytecodeScrollPos int // Track scroll position for bytecode panel
}

func NewModel(debugger Debugger) Model {
	return Model{
		debugger:          debugger,
		state:             nil,
		ready:             false,
		bytecodeScrollPos: 0,
	}
}

func (m Model) Init() tea.Cmd {
	// Start the debugger
	m.debugger.Run()
	return tea.Batch(
		waitForState(m.debugger),
		tea.EnterAltScreen,
	)
}

type stateMsg struct {
	state *vm.VMState
}

func waitForState(d Debugger) tea.Cmd {
	return func() tea.Msg {
		// Small delay to let state be captured
		state := d.CurrentState()
		return stateMsg{state: state}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case stateMsg:
		m.state = msg.state
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "n", " ", "enter":
			// Step forward
			m.debugger.StepForward()
			return m, waitForState(m.debugger)

		case "b", "p":
			// Step back
			m.debugger.StepBack()
			m.state = m.debugger.CurrentState()
			return m, nil

		case "c":
			// Continue to end
			go m.debugger.Continue()
			return m, waitForState(m.debugger)

		case "r":
			// Reset
			m.debugger.Reset()
			m.debugger.Run()
			m.bytecodeScrollPos = 0
			return m, waitForState(m.debugger)
		}
	}

	return m, nil
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	if m.state == nil {
		return "Waiting for state..."
	}

	// Get position info
	current, total := m.debugger.Position()

	// Render the UI with scroll position
	return RenderUI(m.state, current, total, m.width, m.height, m.debugger.IsDone(), m.debugger.Error(), &m.bytecodeScrollPos)
}
