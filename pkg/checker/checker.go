package checker

import (
	"fmt"

	"github.com/resonatehq/durable-promise-test-harness/pkg/store"
)

// Checker validates that a history is correct with respect to some model.
type Checker struct {
	visualizer
}

func New() *Checker {
	return &Checker{
		visualizer: newVisualizer(),
	}
}

// Check verifies the history is linearizable (for correctness).
func (c *Checker) Check(history []store.Operation) error {
	events := makeEvents(history)
	for _, e := range events {
		fmt.Println(e.value)
	}

	fmt.Println(c.visualize(events))

	model := NewDurablePromiseModel()

	return checkEvents(model, events)
}

func checkEvents(model *DurablePromiseModel, events []event) error {
	state := model.Init()
	eventIter := NewEventIterator(events)

	for {
		in, out, next := eventIter.Next()
		if !next {
			break // no more events
		}

		// TODO: show state -> newState, operation etc. better understandability when something fails
		newState, err := model.Step(state, in, out)
		if err != nil {
			return fmt.Errorf("not lineariable: received bad operation: input=%v output=%v: %v", in.value, out.value, err)
		}

		state = newState
	}

	return nil
}
