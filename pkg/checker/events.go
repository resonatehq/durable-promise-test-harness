package checker

import (
	"sort"
	"time"

	"github.com/resonatehq/durable-promise-test-harness/pkg/store"
)

type eventKind bool

const (
	callEvent   eventKind = false
	returnEvent eventKind = true
)

type event struct {
	id       int
	clientId int
	kind     eventKind
	API      store.API
	value    interface{}
	time     time.Time
	status   store.Status
}

func makeEvents(history []store.Operation) []event {
	events := make([]event, 0)

	for _, op := range history {
		// request
		events = append(events, event{
			id:       op.ID,
			clientId: op.ClientID,
			kind:     callEvent,
			API:      op.API,
			value:    op.Input,
			time:     op.CallEvent,
			status:   op.Status,
		})

		// response
		events = append(events, event{
			id:       op.ID,
			clientId: op.ClientID,
			kind:     returnEvent,
			API:      op.API,
			value:    op.Output,
			time:     op.ReturnEvent,
			status:   op.Status,
		})
	}

	sort.Sort(timeSortedEvent(events))

	return events
}

type timeSortedEvent []event

func (t timeSortedEvent) Len() int {
	return len(t)
}

func (t timeSortedEvent) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t timeSortedEvent) Less(i, j int) bool {
	if t[i].time != t[j].time {
		return t[i].time.Before(t[j].time)
	}
	// if the timestamps are the same, we need to make sure we order calls before returns
	return t[i].kind == callEvent && t[j].kind == returnEvent
}

type eventIterator struct {
	index  int
	events []event
}

func NewEventIterator(events []event) *eventIterator {
	return &eventIterator{index: 0, events: events}
}

func (ei *eventIterator) Next() (event, event, bool) {
	if ei.index+1 < len(ei.events) {
		in, out := ei.events[ei.index], ei.events[ei.index+1]
		ei.index += 2
		return in, out, true
	}
	return event{}, event{}, false
}
