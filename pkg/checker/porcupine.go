package checker

import (
	"fmt"
	"reflect"

	"github.com/anishathalye/porcupine"
	"github.com/resonatehq/durable-promise-test-harness/pkg/store"
)

// newPorcupineModel is being used as a wrapper around the model for its functionality.
func newPorcupineModel() porcupine.Model {

	return porcupine.Model{
		Init: func() interface{} {
			model := newDurablePromiseModel()
			return model.Init()
		},
		Step: func(state, input, output interface{}) (bool, interface{}) {
			s := state.(State)
			in := input.(event)
			out := output.(event)

			model := newDurablePromiseModel()

			newState, err := model.Step(s, in, out)
			if err != nil {
				panic(fmt.Sprintf(
					"bad op: %v: in=%s out=%s from %s",
					err,
					in.String(),
					out.String(),
					s.String(),
				))
			}

			return err == nil, newState
		},
		Equal: func(state1, state2 interface{}) bool {
			s1 := state1.(State)
			s2 := state2.(State)
			return reflect.DeepEqual(s1, s2)
		},
	}

}

func makePorcupineEvents(ops []store.Operation) []porcupine.Event {
	porcupineEvents, events := make([]porcupine.Event, 0), makeEvents(ops)

	for _, event := range events {
		porcupineEvents = append(porcupineEvents, porcupine.Event{
			Id:       event.id,
			ClientId: event.clientId,
			Kind:     porcupine.EventKind(event.kind),
			Value:    event,
		})
	}
	return porcupineEvents
}
