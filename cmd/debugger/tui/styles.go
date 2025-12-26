package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("#7D56F4")
	secondaryColor = lipgloss.Color("#00D9FF")
	successColor   = lipgloss.Color("#04B575")
	errorColor     = lipgloss.Color("#FF0000")
	warningColor   = lipgloss.Color("#FFA500")
	mutedColor     = lipgloss.Color("#626262")
	borderColor    = lipgloss.Color("#383838")

	// Base styles
	baseStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Title bar
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 1)

	positionStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Padding(0, 1)

	// Panel styles
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(0, 1)

	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor)

	// Bytecode styles
	instructionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF"))

	currentInstructionStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(successColor).
				Background(lipgloss.Color("#1A1A1A"))

	instructionPointerStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(successColor)

	operandStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	// Stack styles
	stackIndexStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	stackValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	stackTypeStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Italic(true)

	// Frame styles
	activeFrameStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(successColor)

	inactiveFrameStyle = lipgloss.NewStyle().
				Foreground(mutedColor)

	frameMarkerStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(successColor)

	// Closure styles
	closureKeyStyle = lipgloss.NewStyle().
			Foreground(primaryColor)

	closureValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF"))

	cellStyle = lipgloss.NewStyle().
			Foreground(warningColor)

	// Help bar
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Padding(0, 1)

	keyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(secondaryColor)

	// Error/Done styles
	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(errorColor).
			Padding(0, 1)

	doneStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(successColor).
			Padding(0, 1)
)

