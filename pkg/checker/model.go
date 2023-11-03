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
			store.Get:    &GetPromiseVerifier{},
			store.Create: &CreatePromiseVerifier{},
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

// TODO: edge cases (status, nil stuff everywhere)
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

	return state, nil // state does not change
}

type CreatePromiseVerifier struct{}

// TODO: edge cases (status, duplicates with no idempotency key, nil stuff everywhere)
func (v *CreatePromiseVerifier) Verify(state State, req, resp event) (State, error) {
	_, ok := req.value.(*openapi.CreatePromiseRequest)
	if !ok {
		panic("something went wrong-1")
	}

	respObj, ok := resp.value.(*openapi.Promise)
	if !ok {
		panic(resp.value)
	}

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
		return nil, errors.New("promise not found ")
	}

	return val, nil
}
