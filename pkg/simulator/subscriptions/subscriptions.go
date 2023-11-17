package subscriptions

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/resonatehq/durable-promise-test-harness/pkg/openapi"
	"github.com/resonatehq/durable-promise-test-harness/pkg/simulator"
	"github.com/resonatehq/durable-promise-test-harness/pkg/store"
	"github.com/resonatehq/durable-promise-test-harness/pkg/utils"
)

type registry struct {
	mu   sync.Mutex
	data map[string]bool
}

func NewRegistry() *registry {
	return &registry{
		data: make(map[string]bool),
	}
}

func (r *registry) register(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[id] = true
}

func (r *registry) exists(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.data[id]
}

// experimental
func Run() {
	logCh := make(chan string, 10)

	go func() {
		notificationRegistry := NewRegistry()
		http.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				logCh <- "Method not allowed"
				return
			}

			var promise openapi.Promise
			err := json.NewDecoder(r.Body).Decode(&promise)
			if err != nil {
				logCh <- fmt.Sprintf("Error parsing request body: %s", err.Error())
				return
			}

			if !notificationRegistry.exists(*promise.Id) {
				notificationRegistry.register(*promise.Id)
				logCh <- fmt.Sprintf("received notification from '%s'", *promise.Id)
			}

			w.WriteHeader(http.StatusOK)
		})

		http.ListenAndServe(":8080", nil)
	}()

	log.Println("waiting for server to be ready")
	time.Sleep(3 * time.Second)

	log.Println("starting simulation")
	simulate()

	log.Println("waiting for all notifications")
	for i := 0; i < 10; i++ {
		log.Println(<-logCh)
	}

	log.Println("all notifications received")
}

func simulate() {
	client, err := simulator.NewClient(0, "http://0.0.0.0:8001/")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// create promise
	for i := 0; i < 10; i++ {
		client.Invoke(ctx, store.Operation{
			ID:       int(uuid.New().ID()),
			ClientID: client.ID,
			API:      store.Create,
			Input: &openapi.CreatePromiseRequest{
				Id: utils.ToPointer(fmt.Sprintf("promise-%d", i)),
				Param: &openapi.Value{
					Data: utils.ToPointer(base64.StdEncoding.EncodeToString([]byte(`Created Durable Promise`))),
				},
				Timeout: utils.ToPointer(2524608000000),
			},
		})
	}

	// create subscription
	for i := 0; i < 10; i++ {
		data := fmt.Sprintf(`{"promiseId":"promise-%d","url":"http://localhost:8080/subscribe","retryPolicy":{"delay":3,"attempts":5}}`, i)
		resp, err := http.Post(fmt.Sprintf("http://0.0.0.0:8001/subscriptions/sub-%d/create", i), "application/json", bytes.NewBuffer([]byte(data)))
		if err != nil {
			panic(err)
		}
		resp.Body.Close()
	}

	// complete promise
	for i := 0; i < 10; i++ {
		client.Invoke(ctx, store.Operation{
			ID:       int(uuid.New().ID()),
			ClientID: client.ID,
			API:      store.Resolve,
			Input: &openapi.CompletePromiseRequestWrapper{
				Id: utils.ToPointer(fmt.Sprintf("promise-%d", i)),
				Request: &openapi.ResolvePromiseRequest{
					Value: &openapi.Value{
						Data: utils.ToPointer(base64.StdEncoding.EncodeToString([]byte(`Created Durable Promise`))),
					},
				},
			},
		})
	}
}
