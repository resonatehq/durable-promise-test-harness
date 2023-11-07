package simulator

import (
	"context"
	"errors"
	"log"
	"math/rand"
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

func (s *Simulation) Load() error {
	return nil
}

func (s *Simulation) Multiple() error {
	return nil
}

func (s *Simulation) Single() error {
	if err := s.setupSuite(); err != nil {
		return err
	}

	return s.testSingleClientCorrectness()
}

func (s *Simulation) setupSuite() error {
	log.Printf("testing server readiness at %s\n", s.config.Addr)
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

func (s *Simulation) tearDownSuite() error {
	return nil
}

func (s *Simulation) testSingleClientCorrectness() error {
	defer func() {
		if r := recover(); r != nil {
			s.tearDownSuite()
			panic(r)
		}
	}()

	localStore := store.NewStore()

	client, err := NewClient(s.config.Addr)
	if err != nil {
		panic(err)
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

type LoadTestCase struct{}

func NewLoadTestCase() TestCase { return nil }

type MultipleTestCase struct{}

func NewMultipleTestCase() TestCase { return nil }

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
	for _, op := range ops {
		t.Store.Add(t.Client.Invoke(ctx, op))
	}

	return t.Checker.Check(t.Store.History())
}

//
// utils
//
