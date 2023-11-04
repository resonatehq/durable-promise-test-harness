package checker

import (
	"fmt"

	"github.com/resonatehq/durable-promise-test-harness/pkg/store"
)

// Checker validates that a history is correct with respect to some model.
type Checker struct {
	visualizer
}

func NewChecker() *Checker {
	return &Checker{
		visualizer: newVisualizer(),
	}
}

// Check verifies the history is linearizable (for correctness).
func (c *Checker) Check(history []store.Operation) error {
	events := makeEvents(history)
	model := NewDurablePromiseModel()
	return checkEvents(model, events)
}

func checkEvents(model *DurablePromiseModel, events []event) error {
	state := model.Init()
	eventIter := NewEventIterator(events)

	for {
		in, out, next := eventIter.Next()
		if !next {
			break
		}

		// TODO: show state -> newState, operation etc. better understandability when something fails
		newState, err := model.Step(state, in, out)
		if err != nil {
			return fmt.Errorf("error: received bad operation: input=%s output=%s: %v", in.String(), out.String(), err)
		}

		state = newState
	}

	return nil
}
