package memrunner

import (
	"github.com/solidcoredata/dbc/parser"
	"github.com/solidcoredata/dbc/runner"
)

var _ runner.StoreRunner = &MemoryStoreRunner{}

type MemoryStoreRunner struct{}

func NewMemoryStoreRunner(st *parser.Store) *MemoryStoreRunner {
	return nil
}
