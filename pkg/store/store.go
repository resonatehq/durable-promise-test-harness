package store

type Store struct {
	Done    chan struct{}
	history []Operation
}

func NewStore() *Store {
	return &Store{
		Done:    make(chan struct{}),
		history: make([]Operation, 0),
	}
}

func (s *Store) Run(results <-chan Operation) {
	for op := range results {
		s.history = append(s.history, op)
	}

	s.Done <- struct{}{}
}

func (s *Store) History() []Operation {
	return s.history
}
