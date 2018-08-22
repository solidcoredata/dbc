package memrunner

import (
	"io"

	"github.com/solidcoredata/dbc/query"
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

func (r *MemoryStoreRunner) Run(s *query.Store, opt runner.Option) (query.StreamingResultSet, error) {
	return StreamingResultSet{}, nil
}

type MemoryStore struct {
	Version int64
}

func (ms *MemoryStore) Store() *query.Store {
	return nil
}

func (ms *MemoryStore) AddTable(t *query.StoreTable, data [][]interface{}) error {
	return nil
}

func (ms *MemoryStore) AddQuery(q *query.Query) error {
	return nil
}

type StreamingResultSet struct{}

func (StreamingResultSet) Next() (query.StreamItem, error) {
	return nil, io.EOF
}
