package checker

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/resonatehq/durable-promise-test-harness/pkg/openapi"
	"github.com/resonatehq/durable-promise-test-harness/pkg/store"
	"github.com/resonatehq/durable-promise-test-harness/pkg/utils"
)

// A Model is a sequential specification of the durable promise system.
type DurablePromiseModel struct {
	Verifiers map[store.API]StepVerifier
}

func NewDurablePromiseModel() *DurablePromiseModel {
	return &DurablePromiseModel{
		Verifiers: map[store.API]StepVerifier{
			store.Get:    newGetPromiseVerifier(),
			store.Create: newCreatePromiseVerifier(),
		},
	}
}

func (m *DurablePromiseModel) Init() interface{} {
	return make(State, 0)
}

func (m *DurablePromiseModel) Step(state interface{}, input interface{}, output interface{}) (interface{}, error) {
	st := state.(State)
	in := input.(event)
	out := output.(event)

	verif, ok := m.Verifiers[in.API]
	if !ok {
		return state, errors.New("unexpected operation")
	}
	return verif.Verify(st, in, out)
}

type StepVerifier interface {
	Verify(st State, in event, out event) (State, error)
}

type GetPromiseVerifier struct{}

func newGetPromiseVerifier() *GetPromiseVerifier {
	return &GetPromiseVerifier{}
}

// possible outcomes:
// [ invoke, ok, fail ]
// 1. get a promise that exists and it is correct one
// 2. get a promise that exists and it is not the correct one
// 3. get a promise that does not exist and get error (nil)
// 4. get a promise and server failure -- fail ( ... ) where to set status = fail
// TODO: include REJECTED_TIMEOUT
func (v *GetPromiseVerifier) Verify(state State, req, resp event) (State, error) {
	reqObj, ok := req.value.(string)
	if !ok {
		return state, errors.New("res.Value not of type string")
	}

	respObj, ok := resp.value.(*openapi.Promise)
	if !ok {
		return state, errors.New("res.Value not of type *openapi.Promise")
	}

	val, err := state.Get(reqObj)
	if err != nil {
		// does not exist, check if it should have existed
		// fix this in server: returns nil when getting a promise
		// that does not exist. return proper error message to check.
		// also can use status code. if goes in block it correctly
		// failed because it was a getting a promise that does not exist.
		if respObj.Id == nil && respObj.Param == nil && respObj.Timeout == nil {
			return state, nil
		}
		return state, err
	}

	if !reflect.DeepEqual(val.Id, respObj.Id) {
		return state, fmt.Errorf("expected '%s', got '%s'", *val.Id, *respObj.Id) // can be nil
	}
	if !reflect.DeepEqual(val.Param, respObj.Param) {
		return state, fmt.Errorf("expected '%v', got '%v'", *val.Param, *respObj.Param)
	}
	if !reflect.DeepEqual(val.Timeout, respObj.Timeout) {
		return state, fmt.Errorf("expected '%d', got '%d'", *val.Timeout, *respObj.Timeout)
	}

	// validate state

	return state, nil // state does not change
}

type CreatePromiseVerifier struct{}

func newCreatePromiseVerifier() *CreatePromiseVerifier {
	return &CreatePromiseVerifier{}
}

// possible outcomes:
// [ invoke, ok, fail ]
// 1. create a promise that does not exist, success
// 3. create a promise that does exist, error ( returns, object but bad status code )
// 4. create a promise that does exist w/ idempotency key, success
// 5. create a promise and server error
func (v *CreatePromiseVerifier) Verify(state State, req, resp event) (State, error) {
	_, ok := req.value.(*openapi.CreatePromiseRequest)
	if !ok {
		return state, errors.New("req.Value not of type *openapi.CreatePromiseRequest")
	}

	respObj, ok := resp.value.(*openapi.Promise)
	if !ok {
		return state, errors.New("resp.Value not of type *openapi.Promise")
	}

	// validate state -- check if existed, adn if it used idempotency to determine
	// if it got correct thing

	newState := utils.DeepCopy(state)
	newState.Set(*respObj.Id, respObj)

	return newState, nil
}

type State map[string]*openapi.Promise // TODO: more than promises too

func (s State) Set(key string, val *openapi.Promise) {
	s[key] = val
}

func (s State) Get(key string) (*openapi.Promise, error) {
	val, ok := s[key]
	if !ok {
		return nil, errors.New("promise not found")
	}

	return val, nil
}
