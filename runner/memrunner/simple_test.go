package memrunner

import (
	"io"
	"testing"

	"github.com/solidcoredata/dbc/parser"
	"github.com/solidcoredata/dbc/runner"
)

func TestSimple(t *testing.T) {
	ms := &MemoryStore{}

	ms.AddTable(&parser.StoreTable{
		Name:    "Book",
		Alias:   "b",
		Display: "Library Books",
		Comment: "Contains all the available library books.",
		Tag:     []string{"soft-delete"},
		Column: []*parser.StoreColumn{
			{Name: "ID", Type: parser.TypeInteger, Key: true, Serial: true},
		},
		Read: []parser.Param{
			{
				Q:     "exists (select top 1 1 from Account a join AccountOrganization ao on a.ID = ao.Account where ao.Organization = b.Organization)",
				Input: []parser.Input{{Type: parser.TypeInteger, Name: "Account"}},
			},
		},
		Port: map[string]parser.StoreTablePort{
			"internal": parser.StoreTablePort{
				RoleAuthn: map[string]parser.Authn{
					"user": parser.AllowReturn | parser.AllowInsert,
				},
				DenyRead: parser.Param{},

				Column: []parser.StoreColumnPort{},
			},
		},
	}, [][]interface{}{
		{1},
		{2},
	})

	r := NewMemoryStoreRunner(ms)
	st := ms.Store()

	stream, err := r.Run(st, runner.Option{
		QueryName: "Book",

		Port: "internal",
		Role: []string{"accounting", "user"},
		Param: []runner.Param{
			runner.Param{Name: "Account", Value: 1},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for {
		item, err := stream.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}
		switch v := item.(type) {
		default:
			t.Fatal("unknown state")
		case parser.StreamItemResultSetSchema:
		case parser.StreamItemResult:
		case parser.StreamItemRow:
		case parser.StreamItemColumn:
		case parser.StreamItemEndOfResult:
		case parser.StreamItemEndOfSet:
		case parser.StreamItemError:
		}
	}
}
