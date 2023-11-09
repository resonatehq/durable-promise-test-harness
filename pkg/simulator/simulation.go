package simulator

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/resonatehq/durable-promise-test-harness/pkg/checker"
	"github.com/resonatehq/durable-promise-test-harness/pkg/store"
	"github.com/resonatehq/durable-promise-test-harness/pkg/utils"
)

type Simulation struct {
	config *SimulationConfig
}

func NewSimulation(config *SimulationConfig) *Simulation {
	return &Simulation{
		config: config,
	}
}

func (s *Simulation) Run() error {
	if err := s.SetupSuite(); err != nil {
		return fmt.Errorf("error setting up suite: %v", err)
	}

	if err := s.Verify(); err != nil {
		return fmt.Errorf("error running test: %v", err)
	}

	if err := s.TearDownSuite(); err != nil {
		return fmt.Errorf("error tearing down suite: %v", err)
	}

	return nil
}

func (s *Simulation) SetupSuite() error {
	var ready bool
	for i := 0; i < 10; i++ {
		if utils.IsReady(s.config.Addr) {
			ready = true
			break
		}

		time.Sleep(1 * time.Second)
	}

	if !ready {
		return errors.New("server did not become ready in time")
	}

	return nil
}

func (s *Simulation) TearDownSuite() error {
	return nil
}

func (s *Simulation) Verify() error {
	defer func() {
		if r := recover(); r != nil {
			s.TearDownSuite()
			panic(r)
		}
	}()

	localStore := store.NewStore()

	clients := make([]*Client, 0)
	for i := 0; i < s.config.NumClients; i++ {
		client, err := NewClient(s.config.Addr)
		if err != nil {
			return err
		}
		clients = append(clients, client)
	}

	generator := NewGenerator(&GeneratorConfig{
		r:           rand.New(rand.NewSource(0)),
		numRequests: s.config.NumRequests,
		Ids:         100,
		Data:        100,
	})

	checker := checker.NewChecker()

	test := NewTestCase(
		localStore,
		clients,
		generator,
		checker,
	)

	if err := test.Run(); err != nil {
		return err
	}

	return nil
}

type TestCase struct {
	Store     *store.Store
	Clients   []*Client
	Generator *Generator
	Checker   *checker.Checker
}

func NewTestCase(s *store.Store, cs []*Client, g *Generator, ch *checker.Checker) *TestCase {
	return &TestCase{
		Store:     s,
		Clients:   cs,
		Generator: g,
		Checker:   ch,
	}
}

func (t *TestCase) Run() error {
	defer func() {
		checker.NewVisualizer().Visualize(t.Store.History())
		// checker.PorcupineVisualize(t.Store.History())
		// TODO: fix this to output this too
	}()

	ctx := context.Background()
	results := make(chan store.Operation, len(t.Clients))

	go func() {
		t.Store.Run(results)
	}()

	var wg sync.WaitGroup
	wg.Add(len(t.Clients))
	for _, c := range t.Clients {
		go func(client *Client) {
			defer wg.Done()
			ops := t.Generator.Generate()
			for _, op := range ops {
				results <- client.Invoke(ctx, op)
			}
		}(c)
	}
	wg.Wait()

	close(results)
	<-t.Store.Done

	return t.Checker.Check(t.Store.History())
}
