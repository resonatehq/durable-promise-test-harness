package checker

type visualizer struct{}

func newVisualizer() visualizer {
	return visualizer{}
}

// renders timeline of history and performance analysis
func (v *visualizer) visualize(events []event) string {
	return "TODO"
}
