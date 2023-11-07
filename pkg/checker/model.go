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
	if !isValid(resp.status) {
		return state, fmt.Errorf("operation has unexpected status '%d'", resp.status)
	}

	reqObj, ok := req.value.(*openapi.SearchPromisesParams)
	if !ok {
		return state, errors.New("res.Value not of type *openapi.SearchPromiseResponse")
	}
	respObj, ok := resp.value.(*openapi.SearchPromiseResponse)
	if !ok {
		return state, errors.New("res.Value not of type *openapi.Promise")
	}

	// search never fails from user's point of view - ok
	if !reflect.DeepEqual(store.Ok, resp.status) {
		return state, fmt.Errorf("expected '%d', got '%d'", store.Ok, resp.status)
	}
	if !reflect.DeepEqual(http.StatusOK, resp.code) {
		return state, fmt.Errorf("expected '%d', got '%d'", http.StatusOK, resp.code)
	}

	localSearchResults, serverSearchResults := state.Search(*reqObj.State, req.time.UnixMilli()), *respObj.Promises

	sort.Slice(localSearchResults, func(i, j int) bool {
		return *localSearchResults[i].Id < *localSearchResults[j].Id
	})

	sort.Slice(serverSearchResults, func(i, j int) bool {
		return *serverSearchResults[i].Id < *serverSearchResults[j].Id
	})

	err := deepEqualPromiseList(localSearchResults, serverSearchResults)
	if err != nil {
		return state, fmt.Errorf("got mistmatched promises search results: %v", err)
	}

	return state, nil // state does not change
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

	local, err := state.Get(reqObj, req.time.UnixMilli())
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
	err = deepEqualPromise(local, respObj)
	if err != nil {
		return state, fmt.Errorf("got incorrect promise result: %v", err)
	}

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
		if resp.code == http.StatusForbidden && state.Exists(*reqObj.Id, req.time.UnixMilli()) {
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
			if state.Completed(*reqObj.Id, req.time.UnixMilli()) || isTimedOut(*respObj.State) {
				return state, nil
			}
			return state, fmt.Errorf("got an unexpected 403 status: promise not completed")
		case http.StatusNotFound:
			if !state.Exists(*reqObj.Id, req.time.UnixMilli()) {
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

func (s State) Search(stateParam string, callEvent int64) []openapi.Promise {
	filter := make([]openapi.Promise, 0)
	for key, promise := range s {
		if promise == nil && promise.State == nil {
			continue
		}
		// before every read update for timeout since its implicit
		p := s.SetImplicitTimeout(key, promise, callEvent)
		if strings.EqualFold(stateParam, string(openapi.REJECTED)) && isRejectedState(*p.State) {
			filter = append(filter, *p)
			continue
		}
		if strings.EqualFold(stateParam, string(*p.State)) {
			filter = append(filter, *p)
		}
	}
	return filter
}

func (s State) Get(key string, callEvent int64) (*openapi.Promise, error) {
	val, ok := s[key]
	if !ok {
		return nil, errors.New("promise not found")
	}
	// before every read update for timeout since its implicit
	return s.SetImplicitTimeout(key, val, callEvent), nil
}

func (s State) SetImplicitTimeout(key string, val *openapi.Promise, callEvent int64) *openapi.Promise {
	var timeout int64
	if val.Timeout == nil {
		timeout = int64(0)
	} else {
		timeout = int64(*val.Timeout)
	}
	// based on what to expect
	if int64(timeout) <= callEvent {
		val.State = utils.ToPointer(openapi.REJECTEDTIMEDOUT)
		s[key] = val
	}

	return val
}

func (s State) Exists(key string, callEvent int64) bool {
	_, err := s.Get(key, callEvent)
	return err == nil
}

func (s State) Completed(key string, callEvent int64) bool {
	val, err := s.Get(key, callEvent)
	if err != nil {
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
		promise := s[k]
		build.WriteString(fmt.Sprintf(
			"Promise(Id=%v, state=%v)\n",
			*promise.Id,
			*promise.State,
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

func isRejectedState(state openapi.PromiseState) bool {
	switch state {
	case openapi.REJECTED, openapi.REJECTEDCANCELED, openapi.REJECTEDTIMEDOUT:
		return true
	default:
		return false
	}
}

func deepEqualPromiseList(local, external []openapi.Promise) error {
	if len(local) != len(external) {
		return fmt.Errorf("expected '%v' promises, got '%v'instead", len(local), len(external))
	}
	for i := range local {
		err := deepEqualPromise(&local[i], &external[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func deepEqualPromise(local, external *openapi.Promise) error {
	// intentionally ignore createdOn and completedOn fields
	if !reflect.DeepEqual(local.CreatedOn, external.CreatedOn) {
		return fmt.Errorf("expected 'CreatedOn' %v, got %v", local.CreatedOn, external.CreatedOn)
	}
	if !reflect.DeepEqual(local.Id, external.Id) {
		return fmt.Errorf("expected 'Id' %v, got %v", local.Id, external.Id)
	}
	if !reflect.DeepEqual(local.Param, external.Param) {
		return fmt.Errorf("expected 'Param' %v, got %v", local.Param, external.Param)
	}
	if !reflect.DeepEqual(local.State, external.State) {
		return fmt.Errorf("expected'State' %v, got %v", local.State, external.State)
	}
	if !reflect.DeepEqual(local.Tags, external.Tags) {
		return fmt.Errorf("expected 'Tags' %v, got %v", local.Tags, external.Tags)
	}
	if !reflect.DeepEqual(local.Timeout, external.Timeout) {
		return fmt.Errorf("expected 'Timeout' %v, got %v", local.Timeout, external.Timeout)
	}
	if !reflect.DeepEqual(local.Value, external.Value) {
		return fmt.Errorf("expected 'Value' %v, got %v", local.Value, external.Value)
	}
	return nil
}
