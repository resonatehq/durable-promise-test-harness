package checker

import (
	"errors"
	"os"
	"time"

	"github.com/anishathalye/porcupine"
	"github.com/resonatehq/durable-promise-test-harness/pkg/store"
)

// Checker validates that a history is correct with respect to some model.
type Checker struct {
	*Visualizer
}

// Creates a new Checker with reasonable defaults.
func NewChecker() *Checker {
	return &Checker{
		Visualizer: NewVisualizer(),
	}
}

// Check verifies the history is linearizable (for correctness).
func (c *Checker) Check(history []store.Operation) error {
	model, events := newPorcupineModel(), makePorcupineEvents(history)

	res, info := porcupine.CheckEventsVerbose(model, events, 1*time.Hour)
	if res == porcupine.Illegal {
		return errors.New("failed linearizability check")
	}

	file, err := os.CreateTemp("test/results/", "*.html")
	if err != nil {
		return err
	}
	err = porcupine.Visualize(model, info, file)
	if err != nil {
		return err
	}

	return nil
}
