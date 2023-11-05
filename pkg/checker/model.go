package checker

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"

	"github.com/resonatehq/durable-promise-test-harness/pkg/openapi"
	"github.com/resonatehq/durable-promise-test-harness/pkg/store"
	"github.com/resonatehq/durable-promise-test-harness/pkg/utils"
)

// A Model is a sequential specification of the durable promise system.
type DurablePromiseModel struct {
	SequentialSpec map[store.API]StepVerifier
}

func newDurablePromiseModel() *DurablePromiseModel {
	return &DurablePromiseModel{
		SequentialSpec: map[store.API]StepVerifier{
			store.Search:  newSearchPromiseVerifier(),
			store.Get:     newGetPromiseVerifier(),
			store.Create:  newCreatePromiseVerifier(),
			store.Cancel:  newCancelPromiseVerifier(),
			store.Resolve: newResolvePromiseVerifier(),
			store.Reject:  newRejectPromiseVerifier(),
		},
	}
}

func (m *DurablePromiseModel) Init() State {
	return make(State, 0)
}

func (m *DurablePromiseModel) Step(state State, input, output event) (State, error) {
	verif, ok := m.SequentialSpec[input.API]
	if !ok {
		return state, fmt.Errorf("unexpected operation '%d'", input.API)
	}
	return verif.Verify(state, input, output)
}

type StepVerifier interface {
	Verify(st State, in event, out event) (State, error)
}

// possible outcomes:
// give me or not -- are all promise given, found locally and no more
type SearchPromiseVerifier struct{}

func newSearchPromiseVerifier() *SearchPromiseVerifier {
	return &SearchPromiseVerifier{}
}

func (v *SearchPromiseVerifier) Verify(state State, req, resp event) (State, error) {
	return state, nil
}

// possible outcomes:
// [ invoke, ok, fail ]
// 1. get a promise that exists and it is correct one - 200
// 2. get a promise that exists and it is not the correct one - 200, check here its correct
// 3. get a promise that does not exist and get error (returns nil) -- fix create issue -- 404
type GetPromiseVerifier struct{}

func newGetPromiseVerifier() *GetPromiseVerifier {
	return &GetPromiseVerifier{}
}

func (v *GetPromiseVerifier) Verify(state State, req, resp event) (State, error) {
	if !isCompleted(resp.status) {
		return state, fmt.Errorf("operation has unexpected status '%d'", resp.status)
	}

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
		if resp.status == store.Fail && resp.code == http.StatusNotFound {
			return state, nil
		}
		return state, err
	}

	if !reflect.DeepEqual(store.Ok, resp.status) {
		return state, fmt.Errorf("expected '%d', got '%d'", store.Ok, resp.status)
	}
	if !reflect.DeepEqual(http.StatusOK, resp.code) {
		return state, fmt.Errorf("expected '%d', got '%d'", http.StatusOK, resp.code)
	}
	if !reflect.DeepEqual(val.Id, respObj.Id) {
		return state, fmt.Errorf("expected '%s', got '%s'", utils.SafeDereference(val.Id), utils.SafeDereference(respObj.Id))
	}
	if !reflect.DeepEqual(val.Param, respObj.Param) {
		return state, fmt.Errorf("expected '%v', got '%v'", utils.SafeDereference(val.Param), utils.SafeDereference(respObj.Param))
	}
	if !reflect.DeepEqual(val.Timeout, respObj.Timeout) {
		return state, fmt.Errorf("expected '%d', got '%d'", utils.SafeDereference(val.Timeout), utils.SafeDereference(respObj.Timeout))
	}

	// TODO: validate promise STATE, what can it be, this has a few options
	// if no reject or resolve were created than should be, PENDING or TIMEDOUt ?
	// once completed it can't be change so others only affect once if not PENDING

	return state, nil // state does not change
}

// possible outcomes:
// [ invoke, ok, fail ]
// [ ok ]
// 1. create a promise that does not exist, success - 201
// 2. create a promise that does exist w/ idempotency key, success (first gets 201, then 200 -- should be the same though, no? - for put in both  -- if not documented for sure)
// [ fail ]
// 1. create a promise that does exist NO Idempotency, error ( returns, object (weird fix) but bad status code ) -- 403, should be 409 [ fix ]
type CreatePromiseVerifier struct{}

func newCreatePromiseVerifier() *CreatePromiseVerifier {
	return &CreatePromiseVerifier{}
}

func (v *CreatePromiseVerifier) Verify(state State, req, resp event) (State, error) {
	if !isCompleted(resp.status) {
		return state, fmt.Errorf("operation has unexpected status '%d'", resp.status)
	}

	reqObj, ok := req.value.(*openapi.CreatePromiseRequest)
	if !ok {
		return state, errors.New("req.Value not of type *openapi.CreatePromiseRequest")
	}
	respObj, ok := resp.value.(*openapi.Promise)
	if !ok {
		return state, errors.New("resp.Value not of type *openapi.Promise")
	}

	if resp.status == store.Fail {
		// the client correctly got a forbidden status code since the promise
		// already had been created by the client.
		if resp.code == http.StatusForbidden && state.Exists(*reqObj.Id) {
			return state, nil
		}

		// failed even though promise doesn't exist and/or got unexpected status code
		return state, fmt.Errorf("got an unexpected failure status code '%d", resp.code)
	}

	// TODO: be strict that if it is 200 it must have an idempotency key
	if resp.code != http.StatusCreated && resp.code != http.StatusOK {
		return state, fmt.Errorf("go an unexpected ok status code '%d", resp.code)
	}

	// validate promise state, only PENDING ?? -- idempotency key ??? might me something else
	// separate those two ??
	// if !reflect.DeepEqual() {
	// 	return state, fmt.Errorf("got ")
	// }

	newState := utils.DeepCopy(state)
	newState.Set(*respObj.Id, respObj)

	return newState, nil
}

// possible outcomes:
// if completed [ resolve, rejected, or canceled ] don't update state
type CancelPromiseVerifier struct{}

func newCancelPromiseVerifier() *CancelPromiseVerifier {
	return &CancelPromiseVerifier{}
}

func (v *CancelPromiseVerifier) Verify(state State, req, resp event) (State, error) {
	return state, nil
}

// possible outcomes:
// if completed [ resolve, rejected, or canceled ] don't update state
type ResolvePromiseVerifier struct{}

func newResolvePromiseVerifier() *ResolvePromiseVerifier {
	return &ResolvePromiseVerifier{}
}

func (v *ResolvePromiseVerifier) Verify(state State, req, resp event) (State, error) {
	return state, nil
}

// possible outcomes:
// if completed [ resolve, rejected, or canceled ] don't update state
type RejectPromiseVerifier struct{}

func newRejectPromiseVerifier() *RejectPromiseVerifier {
	return &RejectPromiseVerifier{}
}

func (v *RejectPromiseVerifier) Verify(state State, req, resp event) (State, error) {
	return state, nil
}

// State holds the expectation of the client
type State map[string]*openapi.Promise

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

func (s State) Exists(key string) bool {
	_, ok := s[key]
	return ok
}

func (s State) String() string {
	// sorts key for consistent output
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	build := strings.Builder{}
	build.WriteString("STATE\n")
	build.WriteString("-----\n")
	for _, k := range keys {
		build.WriteString(fmt.Sprintf(
			"promise(Id=%v, state=%v)\n",
			*s[k].Id,
			*s[k].State,
		))
	}

	return build.String()
}

//
// utils
//

func isCompleted(stat store.Status) bool {
	return stat == store.Ok || stat == store.Fail
}
