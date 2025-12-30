package main

import (
	"sync"

	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/vm"
)

type Debugger struct {
	vm       *vm.VM
	program  *objects.CompiledProgram
	history  []*vm.VMState
	position int

	stepChan chan struct{} // signals VM to continue
	done     bool
	err      error
	mu       sync.Mutex
}

func NewDebugger(program *objects.CompiledProgram) *Debugger {
	machine := vm.New(program)

	d := &Debugger{
		vm:       machine,
		program:  program,
		history:  make([]*vm.VMState, 0),
		position: -1,
		stepChan: make(chan struct{}),
	}

	machine.SetOnStep(func() {
		d.mu.Lock()
		// Capture state BEFORE execution
		state := machine.GetState()
		d.history = append(d.history, state)
		d.position = len(d.history) - 1
		d.mu.Unlock()

		// Block until user steps forward
		<-d.stepChan
	})

	return d
}

// Run starts VM in goroutine
func (d *Debugger) Run() {
	go func() {
		d.err = d.vm.RunProgram()
		d.mu.Lock()
		d.done = true
		// Capture final state
		d.history = append(d.history, d.vm.GetState())
		d.position = len(d.history) - 1
		d.mu.Unlock()
	}()
}

// StepForward moves to next state
func (d *Debugger) StepForward() {
	d.mu.Lock()
	if d.position < len(d.history)-1 {
		// Viewing history, just move forward
		d.position++
		d.mu.Unlock()
	} else if !d.done {
		// At latest, execute next opcode
		d.mu.Unlock()
		d.stepChan <- struct{}{}
	} else {
		d.mu.Unlock()
	}
}

// StepBack moves to previous state
func (d *Debugger) StepBack() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.position > 0 {
		d.position--
	}
}

// Continue runs to completion
func (d *Debugger) Continue() {
	for {
		d.mu.Lock()
		done := d.done
		d.mu.Unlock()
		if done {
			break
		}
		d.stepChan <- struct{}{}
	}
}

// Reset creates a fresh VM and clears history
func (d *Debugger) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Create new VM
	machine := vm.New(d.program)
	d.vm = machine
	d.history = make([]*vm.VMState, 0)
	d.position = -1
	d.done = false
	d.err = nil
	d.stepChan = make(chan struct{})

	// Set up the callback again
	machine.SetOnStep(func() {
		d.mu.Lock()
		state := machine.GetState()
		d.history = append(d.history, state)
		d.position = len(d.history) - 1
		d.mu.Unlock()
		<-d.stepChan
	})
}

// CurrentState returns state at current position
func (d *Debugger) CurrentState() *vm.VMState {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.position >= 0 && d.position < len(d.history) {
		return d.history[d.position]
	}
	return nil
}

// Position info for UI
func (d *Debugger) Position() (current, total int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.position + 1, len(d.history)
}

func (d *Debugger) IsDone() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.done
}

func (d *Debugger) Error() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.err
}
