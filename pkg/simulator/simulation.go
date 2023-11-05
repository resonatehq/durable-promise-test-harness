package simulator

import (
	"context"
	"math/rand"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/resonatehq/durable-promise-test-harness/pkg/checker"
	"github.com/resonatehq/durable-promise-test-harness/pkg/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type Simulation struct {
	suite.Suite

	config *SimulationConfig
}

func NewSimulation(config *SimulationConfig) *Simulation {
	return &Simulation{
		config: config,
	}
}

func (s *Simulation) SetupSuite() {
	s.T().Logf("testing server readiness at %s\n", s.config.Addr)

	var ready bool
	for i := 0; i < 10; i++ {
		if IsReady(s.config.Addr) {
			ready = true
			break
		}

		time.Sleep(1 * time.Second)
	}

	if !ready {
		s.T().Fatal("server did not become ready in time.")
	}
}

func (s *Simulation) TearDownSuite() {}

func (s *Simulation) TestSingleClientCorrectness() {
	defer func() {
		if r := recover(); r != nil {
			s.TearDownSuite()
			panic(r)
		}
	}()

	s.T().Run("single client correctness", func(t *testing.T) {
		localStore := store.NewStore()

		client, err := NewClient(s.config.Addr)
		if err != nil {
			panic(err)
		}

		generator := NewGenerator(&GeneratorConfig{
			r:    rand.New(rand.NewSource(0)),
			Ids:  100,
			Data: 100,
		})

		checker := checker.NewChecker()

		test := NewTestCase(
			localStore,
			client,
			generator,
			checker,
		)

		assert.Nil(t, test.Run())
	})
}

type TestCase struct {
	Store     *store.Store
	Client    *Client
	Generator *Generator
	Checker   *checker.Checker
}

func NewTestCase(s *store.Store, c *Client, g *Generator, ch *checker.Checker) *TestCase {
	return &TestCase{
		Store:     s,
		Client:    c,
		Generator: g,
		Checker:   ch,
	}
}

func (t *TestCase) Run() error {
	defer func() {
		t.Checker.Visualize(t.Store.History())
	}()

	ctx := context.Background()
	ops := t.Generator.Generate(.1 * 1000)
	for _, op := range ops {
		t.Store.Add(t.Client.Invoke(ctx, op))
	}

	return t.Checker.Check(t.Store.History())
}

func IsReady(Addr string) bool {
	serverAddr := strings.TrimSuffix(strings.TrimPrefix(Addr, "http://"), "/")
	conn, err := net.DialTimeout("tcp", serverAddr, 1*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
