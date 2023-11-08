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

	var err error
	switch s.config.Mode {
	case Load:
		err = s.TestLoad()
	case Linearizability:
		err = s.TestLinearizability()
	default:
		return errors.New("received unknown mode")
	}

	if err := s.TearDownSuite(); err != nil {
		return fmt.Errorf("error tearing down suite: %v", err)
	}

	return err
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

func (s *Simulation) TestLoad() error {
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

	test := NewLoadTestCase(
		localStore,
		clients,
		generator,
	)

	if err := test.Run(); err != nil {
		return err
	}

	return nil
}

func (s *Simulation) TestLinearizability() error {
	defer func() {
		if r := recover(); r != nil {
			s.TearDownSuite()
			panic(r)
		}
	}()

	localStore := store.NewStore()

	client, err := NewClient(s.config.Addr)
	if err != nil {
		return err
	}

	generator := NewGenerator(&GeneratorConfig{
		r:           rand.New(rand.NewSource(0)),
		numRequests: s.config.NumRequests,
		Ids:         100,
		Data:        100,
	})

	checker := checker.NewChecker()

	test := NewSingleTestCase(
		localStore,
		client,
		generator,
		checker,
	)

	if err := test.Run(); err != nil {
		return err
	}

	return nil
}

type TestCase interface {
	Run() error
}

type LoadTestCase struct {
	Store     *store.Store
	Clients   []*Client
	Generator *Generator
}

func NewLoadTestCase(s *store.Store, cs []*Client, g *Generator) TestCase {
	return &LoadTestCase{
		Store:     s,
		Clients:   cs,
		Generator: g,
	}
}

func (t *LoadTestCase) Run() error {
	defer func() {
		checker.NewVisualizer().Visualize(t.Store.History())
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

	return nil
}

type SingleTestCase struct {
	Store     *store.Store
	Client    *Client
	Generator *Generator
	Checker   *checker.Checker
}

func NewSingleTestCase(s *store.Store, c *Client, g *Generator, ch *checker.Checker) TestCase {
	return &SingleTestCase{
		Store:     s,
		Client:    c,
		Generator: g,
		Checker:   ch,
	}
}

func (t *SingleTestCase) Run() error {
	defer func() {
		t.Checker.Visualize(t.Store.History())
	}()

	ctx := context.Background()
	ops := t.Generator.Generate()
	results := make(chan store.Operation, len(ops))

	go func() {
		t.Store.Run(results)
	}()

	for _, op := range ops {
		results <- t.Client.Invoke(ctx, op)
	}

	close(results)
	<-t.Store.Done

	return t.Checker.Check(t.Store.History())
}
