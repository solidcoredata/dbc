package runner

import (
	"github.com/solidcoredata/dbc/query"
)

type Param struct {
	Name  string
	Value interface{}
}

type Option struct {
	QueryName string

	Port  string
	Role  []string
	Param []Param
}

type StoreRunner interface {
	Run(s *query.Store, opt Option) (query.StreamingResultSet, error)
}
