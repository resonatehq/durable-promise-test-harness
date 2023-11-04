package checker

import (
	"fmt"
	"strings"

	"github.com/resonatehq/durable-promise-test-harness/pkg/store"
)

type visualizer struct{}

func newVisualizer() visualizer {
	return visualizer{}
}

// renders timeline of history and performance analysis
func (v *visualizer) Timeline(history []store.Operation) string {
	events := makeEvents(history)

	build := strings.Builder{}
	build.WriteString("EVENT HISTORY\n")
	build.WriteString("-------------\n")

	for i := range events {
		build.WriteString(fmt.Sprintf("(%d) %s\n", i+1, events[i].String()))
	}

	return build.String()
}
