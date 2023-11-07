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

type Visualizer struct{}

func NewVisualizer() *Visualizer {
	return &Visualizer{}
}

// renders timeline of history and performance analysis
func (v *Visualizer) Visualize(history []store.Operation) error {
	content := v.performance(history) + "\n" + v.timeline(history)
	today := time.Now().Format("01-02-2006_15-04-05")
	return utils.WriteStringToFile(content, fmt.Sprintf("test/results/single-client-correctness/%s", today))
}

func (v *Visualizer) timeline(history []store.Operation) string {
	events := makeEvents(history)

	build := strings.Builder{}
	build.WriteString("Event History:\n")
	for i := range events {
		build.WriteString(fmt.Sprintf("  %s\n", events[i].String()))
	}

	return build.String()
}

func (v *Visualizer) performance(history []store.Operation) string {
	reqTimes := []time.Duration{}
	for i := range history {
		latency := history[i].ReturnEvent.Sub(history[i].CallEvent)
		reqTimes = append(reqTimes, latency)
	}

	build := strings.Builder{}
	build.WriteString("Summary:\n")
	build.WriteString(fmt.Sprintf("  Total: %v\n", cumulative(reqTimes)))
	build.WriteString(fmt.Sprintf("  Slowest: %v\n", slowest(reqTimes)))
	build.WriteString(fmt.Sprintf("  Fastest: %v\n", fastest(reqTimes)))
	build.WriteString(fmt.Sprintf("  Average: %v\n", average(reqTimes)))
	build.WriteString("\n")

	build.WriteString("Latency Distribution:\n")
	build.WriteString(fmt.Sprintf("  p50: %v\n", calculateLatencyP(reqTimes, 0.50)))
	build.WriteString(fmt.Sprintf("  p75: %v\n", calculateLatencyP(reqTimes, 0.75)))
	build.WriteString(fmt.Sprintf("  p95: %v\n", calculateLatencyP(reqTimes, 0.95)))
	build.WriteString(fmt.Sprintf("  p99: %v\n", calculateLatencyP(reqTimes, 0.99)))

	return build.String()
}

func cumulative(latencies []time.Duration) time.Duration {
	var total time.Duration
	for _, l := range latencies {
		total += l
	}
	return total
}

func slowest(latencies []time.Duration) time.Duration {
	slow := latencies[0]
	for _, l := range latencies {
		if slow < l {
			slow = l
		}
	}
	return slow
}

func fastest(latencies []time.Duration) time.Duration {
	fast := latencies[0]
	for _, l := range latencies {
		if fast > l {
			fast = l
		}
	}
	return fast
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
