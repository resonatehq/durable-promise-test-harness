package checker

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/resonatehq/durable-promise-test-harness/pkg/store"
	"github.com/resonatehq/durable-promise-test-harness/pkg/utils"
)

type visualizer struct{}

func newVisualizer() visualizer {
	return visualizer{}
}

// renders timeline of history and performance analysis
func (v *visualizer) Visualize(history []store.Operation) error {
	content := v.timeline(history) + "\n" + v.performance(history)
	today := time.Now().Format("01-02-2006_15-04-05")
	return utils.WriteStringToFile(content, fmt.Sprintf("results/single-client-correctness/%s", today))
}

func (v *visualizer) timeline(history []store.Operation) string {
	events := makeEvents(history)

	build := strings.Builder{}
	build.WriteString("EVENT HISTORY\n")
	build.WriteString("-------------\n")

	for i := range events {
		build.WriteString(fmt.Sprintf("(%d) %s\n", i+1, events[i].String()))
	}

	return build.String()
}

func (v *visualizer) performance(history []store.Operation) string {
	reqTimes := []time.Duration{}
	for i := range history {
		latency := history[i].ReturnEvent.Sub(history[i].CallEvent)
		reqTimes = append(reqTimes, latency)
	}

	build := strings.Builder{}
	build.WriteString("PERFORMANCE\n")
	build.WriteString("-------------\n")
	build.WriteString(fmt.Sprintf("cumulative: %v\n", cumulative(reqTimes)))
	build.WriteString(fmt.Sprintf("avg: %v\n", average(reqTimes)))
	build.WriteString(fmt.Sprintf("p50: %v\n", calculateLatencyP(reqTimes, 0.50)))
	build.WriteString(fmt.Sprintf("p75: %v\n", calculateLatencyP(reqTimes, 0.75)))
	build.WriteString(fmt.Sprintf("p95: %v\n", calculateLatencyP(reqTimes, 0.95)))
	build.WriteString(fmt.Sprintf("p99: %v\n", calculateLatencyP(reqTimes, 0.99)))

	return build.String()
}

func cumulative(latencies []time.Duration) time.Duration {
	var total time.Duration
	for _, l := range latencies {
		total += l
	}
	return total
}

func average(latencies []time.Duration) time.Duration {
	var total time.Duration
	for _, l := range latencies {
		total += l
	}
	return time.Duration(int64(total) / int64(len(latencies)))
}

func calculateLatencyP(latencies []time.Duration, percentile float64) time.Duration {
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})
	index := int(math.Ceil(float64(len(latencies)) * percentile))
	return latencies[index-1]
}
