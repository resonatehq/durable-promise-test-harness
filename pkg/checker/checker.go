package checker

import (
	"fmt"

	"github.com/resonatehq/durable-promise-test-harness/pkg/store"
)

// Checker validates that a history is correct with respect to some model.
type Checker struct {
	visualizer
}

// Creates a new Checker with reasonable defaults.
func NewChecker() *Checker {
	return &Checker{
		visualizer: newVisualizer(),
	}
}

// Check verifies the history is linearizable (for correctness).
func (c *Checker) Check(history []store.Operation) error {
	model, events := newDurablePromiseModel(), makeEvents(history)
	return checkEvents(model, events)
}

// checkEvents is loop that actually goes through all the steps.
func checkEvents(model *DurablePromiseModel, events []event) error {
	state := model.Init()
	eventIter := newEventIterator(events)

	for {
		in, out, next := eventIter.Next()
		if !next {
			break
		}

		newState, err := model.Step(state, in, out)
		if err != nil {
			return fmt.Errorf("bad op: %v: in=%s out=%s from %s", err, in.String(), out.String(), state.String())
		}

		state = newState
	}

	return nil
}
