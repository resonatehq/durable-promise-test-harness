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
			store.Cancel:  newCompletePromiseVerifier(),
			store.Resolve: newCompletePromiseVerifier(),
			store.Reject:  newCompletePromiseVerifier(),
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

type SearchPromiseVerifier struct{}

func newSearchPromiseVerifier() *SearchPromiseVerifier {
	return &SearchPromiseVerifier{}
}

func (v *SearchPromiseVerifier) Verify(state State, req, resp event) (State, error) {
	return state, nil
}

type GetPromiseVerifier struct{}

func newGetPromiseVerifier() *GetPromiseVerifier {
	return &GetPromiseVerifier{}
}

func (v *GetPromiseVerifier) Verify(state State, req, resp event) (State, error) {
	if !isValid(resp.status) {
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

	// TODO: verify state is previous write or timeout (strict check)

	return state, nil // state does not change
}

type CreatePromiseVerifier struct{}

func newCreatePromiseVerifier() *CreatePromiseVerifier {
	return &CreatePromiseVerifier{}
}

func (v *CreatePromiseVerifier) Verify(state State, req, resp event) (State, error) {
	if !isValid(resp.status) {
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
		if resp.code == http.StatusForbidden && state.Exists(*reqObj.Id) {
			return state, nil
		}
		return state, fmt.Errorf("got an unexpected failure status code '%d", resp.code)
	}

	if resp.code != http.StatusCreated && resp.code != http.StatusOK && *respObj.State == openapi.PENDING {
		return state, fmt.Errorf("go an unexpected ok status code '%d", resp.code)
	}

	newState := utils.DeepCopy(state)
	newState.Set(*respObj.Id, respObj)

	return newState, nil
}

type CompletePromiseVerifier struct{}

func newCompletePromiseVerifier() *CompletePromiseVerifier {
	return &CompletePromiseVerifier{}
}

func (v *CompletePromiseVerifier) Verify(state State, req, resp event) (State, error) {
	if !isValid(resp.status) {
		return state, fmt.Errorf("operation has unexpected status '%d'", resp.status)
	}

	reqObj, ok := req.value.(*openapi.CompletePromiseRequestWrapper)
	if !ok {
		return state, errors.New("req.Value not of type *simulator.CompletePromiseRequestWrapper")
	}
	respObj, ok := resp.value.(*openapi.Promise)
	if !ok {
		return state, errors.New("resp.Value not of type *openapi.Promise")
	}

	if resp.status == store.Fail {
		switch resp.code {
		case http.StatusForbidden:
			if state.Completed(*reqObj.Id) || isTimedOut(*respObj.State) {
				return state, nil
			}
			return state, fmt.Errorf("got an unexpected 403 status: promise not completed")
		case http.StatusNotFound:
			if !state.Exists(*reqObj.Id) {
				return state, nil
			}
			return state, fmt.Errorf("got an unexpected 404 status code: promise exists")
		default:
			return state, fmt.Errorf("got an unexpected failure status code '%d", resp.code)
		}
	}

	if resp.code != http.StatusCreated && resp.code != http.StatusOK && isCorrectCompleteState(resp.API, *respObj.State) {
		return state, fmt.Errorf("go an unexpected ok status code '%d", resp.code)
	}

	newState := utils.DeepCopy(state)
	newState.Set(*respObj.Id, respObj)

	return newState, nil
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

func (s State) Completed(key string) bool {
	val, ok := s[key]
	if !ok {
		return false
	}

	if val.State == nil {
		panic("got nil promise state")
	}

	switch *val.State {
	case openapi.RESOLVED, openapi.REJECTED, openapi.REJECTEDCANCELED, openapi.REJECTEDTIMEDOUT:
		return true
	default:
		return false
	}
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

func isValid(stat store.Status) bool {
	return stat == store.Ok || stat == store.Fail
}

func isTimedOut(state openapi.PromiseState) bool {
	return state == openapi.REJECTEDTIMEDOUT
}

func isCorrectCompleteState(api store.API, state openapi.PromiseState) bool {
	switch api {
	case store.Resolve:
		return state == openapi.RESOLVED
	case store.Reject:
		return state == openapi.REJECTED
	case store.Cancel:
		return state == openapi.REJECTEDCANCELED
	default:
		return false
	}
}
