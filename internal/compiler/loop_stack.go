package compiler

// loopContext tracks the addresses needed for break/continue statements within a loop
type loopContext struct {
	continuePos   int   // address to jump to for continue (-1 if not yet known)
	breakJumps    []int // positions of break jumps to patch
	continueJumps []int // positions of continue jumps to patch (for for-loops)
}

// LoopStack manages nested loop contexts for break/continue compilation
type LoopStack struct {
	stack []*loopContext
}

// NewLoopStack creates a new empty loop stack
func NewLoopStack() *LoopStack {
	return &LoopStack{
		stack: []*loopContext{},
	}
}

// Push enters a new loop context with the given continue position
func (ls *LoopStack) Push(continuePos int) {
	ls.stack = append(ls.stack, &loopContext{
		continuePos:   continuePos,
		breakJumps:    []int{},
		continueJumps: []int{},
	})
}

// Pop exits the current loop context
func (ls *LoopStack) Pop() {
	if len(ls.stack) > 0 {
		ls.stack = ls.stack[:len(ls.stack)-1]
	}
}

// IsInLoop returns true if currently inside a loop
func (ls *LoopStack) IsInLoop() bool {
	return len(ls.stack) > 0
}

// current returns the innermost loop context, or nil if not in a loop
func (ls *LoopStack) current() *loopContext {
	if len(ls.stack) == 0 {
		return nil
	}
	return ls.stack[len(ls.stack)-1]
}

// AddBreakJump records a break jump position to be patched later
func (ls *LoopStack) AddBreakJump(pos int) {
	if loop := ls.current(); loop != nil {
		loop.breakJumps = append(loop.breakJumps, pos)
	}
}

// AddContinueJump records a continue jump position to be patched later (for for-loops)
func (ls *LoopStack) AddContinueJump(pos int) {
	if loop := ls.current(); loop != nil {
		loop.continueJumps = append(loop.continueJumps, pos)
	}
}

// ContinuePos returns the continue target address, or -1 if not yet known
func (ls *LoopStack) ContinuePos() int {
	if loop := ls.current(); loop != nil {
		return loop.continuePos
	}
	return -1
}

// SetContinuePos sets the continue target address for the current loop
func (ls *LoopStack) SetContinuePos(pos int) {
	if loop := ls.current(); loop != nil {
		loop.continuePos = pos
	}
}

// BreakJumps returns all break jump positions that need patching
func (ls *LoopStack) BreakJumps() []int {
	if loop := ls.current(); loop != nil {
		return loop.breakJumps
	}
	return nil
}

// ContinueJumps returns all continue jump positions that need patching
func (ls *LoopStack) ContinueJumps() []int {
	if loop := ls.current(); loop != nil {
		return loop.continueJumps
	}
	return nil
}
