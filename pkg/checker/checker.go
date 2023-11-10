package checker

import (
	"fmt"
	"time"

	"github.com/anishathalye/porcupine"
	"github.com/resonatehq/durable-promise-test-harness/pkg/store"
	"github.com/resonatehq/durable-promise-test-harness/pkg/utils"
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

// Check verifies the history is linearizably consistent with respect to the model.
func (c *Checker) Check(history []store.Operation) error {
	model, events := newPorcupineModel(), makePorcupineEvents(history)

	var pass bool

	res, info := porcupine.CheckEventsVerbose(model, events, 1*time.Hour)
	if res != porcupine.Illegal {
		pass = true
	}

	today := time.Now().Format("01-02-2006_15-04-05")

	filePath := fmt.Sprintf("test/results/%s/visualization.html", today)
	err := utils.WriteStringToFile("", filePath)
	if err != nil {
		return err
	}

	err = porcupine.VisualizePath(model, info, filePath)
	if err != nil {
		return err
	}

	c.Summary(pass, today, history)

	return nil
}
