package memrunner

import (
	"github.com/solidcoredata/dbc/parser"
	"github.com/solidcoredata/dbc/runner"
)

var _ runner.StoreRunner = &MemoryStoreRunner{}

type MemoryStoreRunner struct {
	store *MemoryStore
}

func NewMemoryStoreRunner(st *MemoryStore) *MemoryStoreRunner {
	return &MemoryStoreRunner{
		store: st,
	}
}

func (r *MemoryStoreRunner) Run(s *parser.Store, opt runner.Option) (parser.StreamingResultSet, error) {
	return nil, nil
}

type MemoryStore struct{}

func (ms *MemoryStore) Store() *parser.Store {
	return nil
}

func (ms *MemoryStore) AddTable(t *parser.StoreTable, data [][]interface{}) error {
	return nil
}

func (ms *MemoryStore) AddQuery(t *parser.StoreQuery) error {
	return nil
}
