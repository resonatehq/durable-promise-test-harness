package checker

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/resonatehq/durable-promise-test-harness/pkg/store"
)

type eventKind bool

const (
	callEvent   eventKind = false
	returnEvent eventKind = true
)

func (e eventKind) String() string {
	if e {
		return "RETURN"
	}
	return "CALL"
}

type event struct {
	id       int
	clientId int
	kind     eventKind
	API      store.API
	value    interface{}
	time     time.Time
	status   store.Status
	code     int
}

func (e event) String() string {
	v, _ := json.Marshal(e.value)

	return fmt.Sprintf(
		"Event(id=%d, clientId=%d, kind=%v, api=%v, value=%s, time=%v, status=%v, code=%d)",
		e.id,
		e.clientId,
		e.kind.String(),
		e.API.String(),
		string(v),
		e.time,
		e.status.String(),
		e.code,
	)
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
			status:   store.Invoke, // status is invoking
			code:     -1,           // code is unknown
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
			code:     op.Code,
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

func newEventIterator(events []event) *eventIterator {
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
