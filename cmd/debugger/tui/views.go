package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/harshagw/viri/internal/code"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/vm"
)

func RenderUI(state *vm.VMState, current, total, width, height int, done bool, err error, scrollPos *int) string {
	// Calculate panel dimensions - 3 columns, 3 rows
	// Ensure minimum dimensions
	panelWidth := (width - 8) / 3
	if panelWidth < 20 {
		panelWidth = 20
	}
	panelHeight := (height - 8) / 3
	if panelHeight < 8 {
		panelHeight = 8
	}

	// Render title bar
	title := titleStyle.Render("Viri VM Debugger")
	position := positionStyle.Render(fmt.Sprintf("Step %d/%d", current, total))

	var statusMsg string
	if err != nil {
		statusMsg = errorStyle.Render(fmt.Sprintf("Error: %v", err))
	} else if done {
		statusMsg = doneStyle.Render("✓ Execution Complete")
	}

	titleBar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		strings.Repeat(" ", width-lipgloss.Width(title)-lipgloss.Width(position)-lipgloss.Width(statusMsg)-4),
		statusMsg,
		position,
	)

	// Render panels
	bytecodePanel := renderBytecodePanel(state, panelWidth, panelHeight, scrollPos)
	stackPanel := renderStackPanel(state, panelWidth, panelHeight)
	framesPanel := renderFramesPanel(state, panelWidth, panelHeight)
	localsPanel := renderLocalsPanel(state, panelWidth, panelHeight)
	globalsPanel := renderGlobalsPanel(state, panelWidth, panelHeight)
	closuresPanel := renderClosuresPanel(state, panelWidth, panelHeight)

	// Split bottom row between constants and output
	constantsWidth := (panelWidth*3 + 4) / 2
	outputWidth := (panelWidth*3 + 4) / 2
	constantsPanel := renderConstantsPanel(state, constantsWidth, panelHeight)
	outputPanel := renderOutputPanel(state, outputWidth, panelHeight)

	// Arrange panels in grid (3 columns, 3 rows)
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, bytecodePanel, stackPanel, framesPanel)
	middleRow := lipgloss.JoinHorizontal(lipgloss.Top, localsPanel, globalsPanel, closuresPanel)
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, constantsPanel, outputPanel)
	panels := lipgloss.JoinVertical(lipgloss.Left, topRow, middleRow, bottomRow)

	// Render help bar
	helpBar := renderHelpBar(width)

	// Combine all
	return lipgloss.JoinVertical(lipgloss.Left, titleBar, panels, helpBar)
}

func renderBytecodePanel(state *vm.VMState, width, height int, scrollPos *int) string {
	title := panelTitleStyle.Render("BYTECODE")

	var lines []string
	lines = append(lines, title)
	lines = append(lines, strings.Repeat("─", width-4))

	// Disassemble all instructions first
	ins := state.Instructions
	currentIP := state.IP

	type instruction struct {
		offset int
		line   string
	}

	var allInstructions []instruction
	i := 0
	for i < len(ins) {
		def, err := code.Lookup(ins[i])
		if err != nil {
			i++
			continue
		}

		operands, read := code.ReadOperands(def, ins[i+1:])

		// Format instruction
		var operandStr string
		if len(operands) > 0 {
			parts := make([]string, len(operands))
			for j, op := range operands {
				parts[j] = fmt.Sprintf("%d", op)
			}
			operandStr = " " + operandStyle.Render(strings.Join(parts, " "))
		}

		instrLine := fmt.Sprintf("%04d %s%s", i, def.Name, operandStr)

		if i == currentIP {
			instrLine = currentInstructionStyle.Render(instrLine)
			instrLine = instructionPointerStyle.Render("▶ ") + instrLine
		} else {
			instrLine = instructionStyle.Render("  " + instrLine)
		}

		allInstructions = append(allInstructions, instruction{offset: i, line: instrLine})
		i += 1 + read
	}

	// Find current instruction index
	currentIdx := -1
	for idx, instr := range allInstructions {
		if instr.offset == currentIP {
			currentIdx = idx
			break
		}
	}

	// Calculate visible range with stateful scrolling
	maxLines := height - 4
	if maxLines < 1 {
		maxLines = 1
	}

	start := *scrollPos

	// Adjust scroll position only if current instruction is outside visible range
	if currentIdx >= 0 {
		// If current instruction is above visible area, scroll up
		if currentIdx < start {
			start = currentIdx
		}
		// If current instruction is below visible area, scroll down
		if currentIdx >= start+maxLines {
			start = currentIdx - maxLines + 1
		}

		// Clamp to valid range
		if start < 0 {
			start = 0
		}
		if start+maxLines > len(allInstructions) {
			start = len(allInstructions) - maxLines
			if start < 0 {
				start = 0
			}
		}

		// Update scroll position
		*scrollPos = start
	}

	// Add visible instructions
	end := start + maxLines
	if end > len(allInstructions) {
		end = len(allInstructions)
	}

	for i := start; i < end && i < len(allInstructions); i++ {
		lines = append(lines, allInstructions[i].line)
	}

	// Fill remaining space
	for len(lines) < height-2 {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	return panelStyle.Width(width).Height(height).Render(content)
}

func renderStackPanel(state *vm.VMState, width, height int) string {
	title := panelTitleStyle.Render(fmt.Sprintf("STACK (SP: %d)", state.SP))

	var lines []string
	lines = append(lines, title)
	lines = append(lines, strings.Repeat("─", width-4))

	if len(state.Stack) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(mutedColor).Render("  (empty)"))
	} else {
		// Show stack from top to bottom
		maxLines := height - 4
		if maxLines < 1 {
			maxLines = 1
		}
		start := 0
		if len(state.Stack) > maxLines {
			start = len(state.Stack) - maxLines
		}

		for i := len(state.Stack) - 1; i >= start; i-- {
			obj := state.Stack[i]
			index := stackIndexStyle.Render(fmt.Sprintf("[%d]", i))

			var value string
			var objType string
			if obj == nil {
				value = "nil"
				objType = stackTypeStyle.Render("(NIL)")
			} else {
				value = truncate(objects.Stringify(obj), width-10)
				objType = stackTypeStyle.Render(fmt.Sprintf("(%s)", obj.Type()))
			}

			line := fmt.Sprintf("%s %s %s", index, stackValueStyle.Render(value), objType)
			lines = append(lines, line)
		}
	}

	// Fill remaining space
	for len(lines) < height-2 {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	return panelStyle.Width(width).Height(height).Render(content)
}

func renderFramesPanel(state *vm.VMState, width, height int) string {
	title := panelTitleStyle.Render(fmt.Sprintf("FRAMES (%d active)", state.FrameIndex))

	var lines []string
	lines = append(lines, title)
	lines = append(lines, strings.Repeat("─", width-4))

	if len(state.Frames) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(mutedColor).Render("  (no frames)"))
	} else {
		// Show frames from current to bottom
		for i := len(state.Frames) - 1; i >= 0; i-- {
			frame := state.Frames[i]

			marker := "  "
			style := inactiveFrameStyle
			if i == len(state.Frames)-1 {
				marker = frameMarkerStyle.Render("▶ ")
				style = activeFrameStyle
			}

			frameName := "main"
			if i > 0 {
				frameName = "closure"
			}

			line := fmt.Sprintf("%s[%d] IP:%d/%d BP:%d locals:%d (%s)",
				marker, i, frame.IP, frame.InstructionSize, frame.BasePointer, frame.ClosureInfo.NumLocals, frameName)
			lines = append(lines, style.Render(line))
		}
	}

	// Fill remaining space
	for len(lines) < height-2 {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	return panelStyle.Width(width).Height(height).Render(content)
}

func renderLocalsPanel(state *vm.VMState, width, height int) string {
	title := panelTitleStyle.Render("LOCALS")

	var lines []string
	lines = append(lines, title)
	lines = append(lines, strings.Repeat("─", width-4))

	if len(state.Frames) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(mutedColor).Render("  (no frames)"))
	} else {
		currentFrame := state.Frames[len(state.Frames)-1]
		numLocals := currentFrame.ClosureInfo.NumLocals
		bp := currentFrame.BasePointer

		if numLocals == 0 {
			lines = append(lines, lipgloss.NewStyle().Foreground(mutedColor).Render("  (no locals)"))
		} else {
			maxLines := height - 4
			if maxLines < 1 {
				maxLines = 1
			}
			displayCount := numLocals
			if displayCount > maxLines {
				displayCount = maxLines
			}

			for i := 0; i < displayCount; i++ {
				stackIdx := bp + i
				var obj objects.Object
				if stackIdx < len(state.Stack) {
					obj = state.Stack[stackIdx]
				}

				index := stackIndexStyle.Render(fmt.Sprintf("[%d]", i))

				var value string
				var objType string
				if obj == nil {
					value = "nil"
					objType = stackTypeStyle.Render("(NIL)")
				} else {
					value = truncate(objects.Stringify(obj), width-10)
					objType = stackTypeStyle.Render(fmt.Sprintf("(%s)", obj.Type()))
				}

				line := fmt.Sprintf("%s %s %s", index, stackValueStyle.Render(value), objType)
				lines = append(lines, line)
			}
		}
	}

	// Fill remaining space
	for len(lines) < height-2 {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	return panelStyle.Width(width).Height(height).Render(content)
}

func renderGlobalsPanel(state *vm.VMState, width, height int) string {
	title := panelTitleStyle.Render("GLOBALS")

	var lines []string
	lines = append(lines, title)
	lines = append(lines, strings.Repeat("─", width-4))

	// Count non-nil globals
	nonNilGlobals := []int{}
	for i, obj := range state.Globals {
		if obj != nil {
			nonNilGlobals = append(nonNilGlobals, i)
		}
	}

	if len(nonNilGlobals) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(mutedColor).Render("  (no globals)"))
	} else {
		maxLines := height - 4
		if maxLines < 1 {
			maxLines = 1
		}
		displayCount := len(nonNilGlobals)
		if displayCount > maxLines {
			displayCount = maxLines
		}

		for i := 0; i < displayCount; i++ {
			globalIdx := nonNilGlobals[i]
			obj := state.Globals[globalIdx]

			index := stackIndexStyle.Render(fmt.Sprintf("[%d]", globalIdx))
			value := truncate(objects.Stringify(obj), width-10)
			objType := stackTypeStyle.Render(fmt.Sprintf("(%s)", obj.Type()))

			line := fmt.Sprintf("%s %s %s", index, stackValueStyle.Render(value), objType)
			lines = append(lines, line)
		}
	}

	// Fill remaining space
	for len(lines) < height-2 {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	return panelStyle.Width(width).Height(height).Render(content)
}

func renderClosuresPanel(state *vm.VMState, width, height int) string {
	title := panelTitleStyle.Render("CELLS")

	var lines []string
	lines = append(lines, title)
	lines = append(lines, strings.Repeat("─", width-4))

	if len(state.Frames) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(mutedColor).Render("  (no frames)"))
	} else {
		// Show free variables for current frame
		currentFrame := state.Frames[len(state.Frames)-1]

		if currentFrame.ClosureInfo.NumFree == 0 {
			lines = append(lines, lipgloss.NewStyle().Foreground(mutedColor).Render("  (no free variables)"))
		} else {
			lines = append(lines, closureKeyStyle.Render(fmt.Sprintf("  Frame[%d] Free Variables:", len(state.Frames)-1)))

			for i, freeVar := range currentFrame.ClosureInfo.FreeVars {
				var value string
				var objType string
				if freeVar == nil {
					value = "nil"
					objType = stackTypeStyle.Render("(NIL)")
				} else {
					value = truncate(objects.Stringify(freeVar), width-12)
					objType = stackTypeStyle.Render(fmt.Sprintf("(%s)", freeVar.Type()))
				}

				line := fmt.Sprintf("    [%d] %s → %s %s",
					i,
					cellStyle.Render("Cell"),
					closureValueStyle.Render(value),
					objType)
				lines = append(lines, line)
			}
		}
	}

	// Fill remaining space
	for len(lines) < height-2 {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	return panelStyle.Width(width).Height(height).Render(content)
}

func renderHelpBar(width int) string {
	keys := []string{
		keyStyle.Render("[n/Space/Enter]") + " Step",
		keyStyle.Render("[b]") + " Back",
		keyStyle.Render("[c]") + " Run to End",
		keyStyle.Render("[r]") + " Restart",
		keyStyle.Render("[q]") + " Quit",
	}

	helpText := strings.Join(keys, "   ")
	padding := width - lipgloss.Width(helpText) - 2
	if padding < 0 {
		padding = 0
	}

	return helpStyle.Render(helpText + strings.Repeat(" ", padding))
}

func renderConstantsPanel(state *vm.VMState, width, height int) string {
	title := panelTitleStyle.Render(fmt.Sprintf("CONSTANTS (%d total)", len(state.Constants)))

	var lines []string
	lines = append(lines, title)
	lines = append(lines, strings.Repeat("─", width-4))

	if len(state.Constants) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(mutedColor).Render("  (no constants)"))
	} else {
		maxLines := height - 4
		if maxLines < 1 {
			maxLines = 1
		}
		displayCount := len(state.Constants)
		if displayCount > maxLines {
			displayCount = maxLines
		}

		for i := 0; i < displayCount; i++ {
			obj := state.Constants[i]
			index := stackIndexStyle.Render(fmt.Sprintf("[%d]", i))

			var value string
			var objType string
			if obj == nil {
				value = "nil"
				objType = stackTypeStyle.Render("(NIL)")
			} else {
				// For constants, show more detail
				switch obj.Type() {
				case objects.TypeCompiledFunction:
					cf := obj.(*objects.CompiledFunction)
					value = fmt.Sprintf("<fn: %d instructions, %d params, %d locals>",
						len(cf.Instructions), cf.NumParameters, cf.NumLocals)
					objType = stackTypeStyle.Render("(COMPILED_FUNCTION)")
				default:
					value = truncate(objects.Stringify(obj), width-20)
					objType = stackTypeStyle.Render(fmt.Sprintf("(%s)", obj.Type()))
				}
			}

			line := fmt.Sprintf("%s %s %s", index, stackValueStyle.Render(value), objType)
			lines = append(lines, line)
		}

		// Show if there are more constants
		if len(state.Constants) > displayCount {
			remaining := len(state.Constants) - displayCount
			lines = append(lines, lipgloss.NewStyle().Foreground(mutedColor).Render(
				fmt.Sprintf("  ... and %d more", remaining)))
		}
	}

	// Fill remaining space
	for len(lines) < height-2 {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	return panelStyle.Width(width).Height(height).Render(content)
}

func renderOutputPanel(state *vm.VMState, width, height int) string {
	title := panelTitleStyle.Render(fmt.Sprintf("OUTPUT (%d lines)", len(state.Output)))

	var lines []string
	lines = append(lines, title)
	lines = append(lines, strings.Repeat("─", width-4))

	if len(state.Output) == 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(mutedColor).Render("  (no output)"))
	} else {
		maxLines := height - 4
		if maxLines < 1 {
			maxLines = 1
		}

		// Show most recent output (from the end)
		start := 0
		if len(state.Output) > maxLines {
			start = len(state.Output) - maxLines
		}

		for i := start; i < len(state.Output); i++ {
			output := state.Output[i]
			// Truncate long lines
			if len(output) > width-6 {
				output = output[:width-9] + "..."
			}
			line := stackValueStyle.Render("  " + output)
			lines = append(lines, line)
		}

		// Show if there's more output
		if start > 0 {
			lines = append(lines, lipgloss.NewStyle().Foreground(mutedColor).Render(
				fmt.Sprintf("  ... %d more lines above", start)))
		}
	}

	// Fill remaining space
	for len(lines) < height-2 {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")
	return panelStyle.Width(width).Height(height).Render(content)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
