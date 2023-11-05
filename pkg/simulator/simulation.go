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

// Simulation is a client
type Simulation struct {
	suite.Suite

	config *SimulationConfig
}

func New(config *SimulationConfig) *Simulation {
	// TODO: parse + validate
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

		time.Sleep(3 * time.Second)
	}

	if !ready {
		s.T().Fatal("server did not become ready in time.")
	}
}

func (s *Simulation) TearDownSuite() {
	// TODO: clean up resonate.db so doesn't affect stuff
	// TODO: delete everything created, leave server in original state
}

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
		hist := t.Store.History()
		t.Checker.Visualize(hist)
	}()

	ctx := context.Background()
	ops := t.Generator.Generate(10)
	for _, op := range ops {
		t.Store.Add(t.Client.Invoke(ctx, op))
	}

	return t.Checker.Check(t.Store.History())
}

func IsReady(Addr string) bool {
	// TODO: Addr define the format -- options -- remote local etc, tcp though
	serverAddr := strings.TrimSuffix(strings.TrimPrefix(Addr, "http://"), "/")
	conn, err := net.DialTimeout("tcp", serverAddr, 1*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
