package main

import (
	"sync"
	"time"

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
	d := &Debugger{
		program:  program,
		history:  make([]*vm.VMState, 0),
		position: -1,
		stepChan: make(chan struct{}),
	}
	d.attachVM(vm.New(program))
	return d
}

// attachVM wires a VM to this debugger session. The onStep callback checks
// that its VM is still the active one, so a superseded session (after Reset)
// cannot corrupt the new session's history.
func (d *Debugger) attachVM(machine *vm.VM) {
	d.vm = machine
	machine.SetOnStep(func() {
		d.mu.Lock()
		if d.vm != machine {
			// Stale session after Reset: run to completion without recording
			d.mu.Unlock()
			return
		}
		// Capture state BEFORE execution
		state := machine.GetState()
		d.history = append(d.history, state)
		d.position = len(d.history) - 1
		ch := d.stepChan
		d.mu.Unlock()

		// Block until user steps forward
		<-ch
	})
}

// Run starts VM in goroutine
func (d *Debugger) Run() {
	d.mu.Lock()
	machine := d.vm
	d.mu.Unlock()

	go func() {
		err := machine.RunProgram()
		d.mu.Lock()
		defer d.mu.Unlock()
		if d.vm != machine {
			return // superseded by Reset
		}
		d.err = err
		d.done = true
		// Capture final state
		d.history = append(d.history, machine.GetState())
		d.position = len(d.history) - 1
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
		// At latest, execute next opcode. Non-blocking: if the VM is not
		// currently paused (e.g. it just finished), the step is dropped
		// instead of deadlocking this goroutine.
		ch := d.stepChan
		d.mu.Unlock()
		select {
		case ch <- struct{}{}:
		default:
		}
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
		ch := d.stepChan
		d.mu.Unlock()
		if done {
			break
		}
		// Timeout guards the race where the VM finishes between the check
		// above and this send (no receiver would ever arrive).
		select {
		case ch <- struct{}{}:
		case <-time.After(50 * time.Millisecond):
		}
	}
}

// Reset creates a fresh VM and clears history
func (d *Debugger) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Release a VM goroutine that may still be blocked on the old channel;
	// its onStep callback sees the swapped d.vm and stops recording.
	close(d.stepChan)

	d.history = make([]*vm.VMState, 0)
	d.position = -1
	d.done = false
	d.err = nil
	d.stepChan = make(chan struct{})

	d.attachVM(vm.New(d.program))
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
