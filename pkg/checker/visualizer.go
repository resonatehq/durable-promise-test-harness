package checker

import (
	"encoding/json"
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
func (v *Visualizer) Summary(pass bool, today string, history []store.Operation) error {
	summary := v.summary(pass)
	performance := v.performance(history)
	timeline := v.timeline(history)

	content := summary + "\n" + performance + "\n" + timeline
	err := utils.WriteStringToFile(content, fmt.Sprintf("test/results/%s/summary.txt", today))
	if err != nil {
		return err
	}

	fmt.Println(summary + "\n" + performance)

	return nil
}

func (v *Visualizer) summary(pass bool) string {
	out := "PASS"
	if !pass {
		out = "FAIL"
	}
	build := strings.Builder{}
	build.WriteString("Summary\n")
	build.WriteString("=====================\n")
	build.WriteString(fmt.Sprintf("Linearizability Check: %s\n", out))
	return build.String()
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

	// Requests
	build.WriteString("Requests:\n")
	build.WriteString(fmt.Sprintf("  Total: %v\n", cumulative(history)))
	build.WriteString(fmt.Sprintf("  Slowest: %v\n", slowest(reqTimes)))
	build.WriteString(fmt.Sprintf("  Fastest: %v\n", fastest(reqTimes)))
	build.WriteString(fmt.Sprintf("  Average: %v\n", average(reqTimes)))
	build.WriteString(fmt.Sprintf("  Requests/Sec: %.2f\n", calculateThroughputRPS(history)))
	build.WriteString("\n")

	// Data
	build.WriteString("Data:\n")
	build.WriteString(fmt.Sprintf("  Total Data: %.4f MB\n", totalDataSize(history)))
	build.WriteString(fmt.Sprintf("  Size/Sec: %.4f MB\n", dataSizePerSecond(history)))
	build.WriteString("\n")

	// Latency
	build.WriteString("Latency Distribution:\n")
	build.WriteString(fmt.Sprintf("  p50: %v\n", calculateLatencyP(reqTimes, 0.50)))
	build.WriteString(fmt.Sprintf("  p75: %v\n", calculateLatencyP(reqTimes, 0.75)))
	build.WriteString(fmt.Sprintf("  p95: %v\n", calculateLatencyP(reqTimes, 0.95)))
	build.WriteString(fmt.Sprintf("  p99: %v\n", calculateLatencyP(reqTimes, 0.99)))
	build.WriteString("\n")

	// Status codes
	build.WriteString("Status Code Distribution:\n")
	statusCodes := calculateStatusCodeDistribution(history)
	for code := range statusCodes {
		build.WriteString(fmt.Sprintf("  %d: %d responses\n", code, statusCodes[code]))
	}

	return build.String()
}

// per second stuff
func cumulative(history []store.Operation) time.Duration {
	sort.Slice(history, func(i, j int) bool {
		return history[i].CallEvent.Before(history[j].CallEvent)
	})
	firstOp, lastOp := history[0], history[len(history)-1]
	return lastOp.CallEvent.Sub(firstOp.CallEvent)
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

func calculateThroughputRPS(history []store.Operation) float64 {
	sort.Slice(history, func(i, j int) bool {
		return history[i].CallEvent.Before(history[j].CallEvent)
	})
	firstOp, lastOp := history[0], history[len(history)-1]
	duration := lastOp.CallEvent.Sub(firstOp.CallEvent)
	return float64(len(history)) / float64(duration.Seconds()) // TODO:
}

func totalDataSize(history []store.Operation) float64 {
	var total float64
	for i := range history {
		jsonData, _ := json.Marshal(history[i].Input)
		sizeBytes := len(jsonData)
		sizeMB := float64(sizeBytes) / (1000 * 1000)
		total += sizeMB
	}
	return total
}

func dataSizePerSecond(history []store.Operation) float64 {
	sort.Slice(history, func(i, j int) bool {
		return history[i].CallEvent.Before(history[j].CallEvent)
	})
	firstOp, lastOp := history[0], history[len(history)-1]
	duration := lastOp.CallEvent.Sub(firstOp.CallEvent)
	return totalDataSize(history) / float64(duration.Seconds())
}

func calculateLatencyP(latencies []time.Duration, percentile float64) time.Duration {
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})
	index := int(math.Ceil(float64(len(latencies)) * percentile))
	return latencies[index-1]
}

func calculateStatusCodeDistribution(history []store.Operation) map[int]int {
	statusCodes := map[int]int{}
	for i := range history {
		if history[i].Status == store.Invoke {
			continue
		}
		statusCodes[history[i].Code]++
	}
	return statusCodes
}
