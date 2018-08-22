package memrunner

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/solidcoredata/dbc/internal/elist"
	"github.com/solidcoredata/dbc/parser"
	"github.com/solidcoredata/dbc/runner"
)

func TestSimple(t *testing.T) {
	ms := &MemoryStore{
		Version: 1,
	}

	err := ms.AddTable(&parser.StoreTable{
		Name:    "Book",
		Alias:   "b",
		Display: "Library Books",
		Comment: "Contains all the available library books.",
		Tag:     []string{"soft-delete"},
		Column: []*parser.StoreColumn{
			{Name: "ID", Type: parser.TypeInteger, Key: true, Serial: true},
			{Name: "Name", Type: parser.TypeString, Display: "Book Name"},
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
		{1, "Never a Dull Moment"},
		{2, "To Kill a Bird"},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = ms.AddQuery(&parser.StoreQuery{
		Type: "jsonnet",
		Query: `
local t = import("table");
local f = import("func");
{
	local b = t.Book,
	from: join([b, t.Account, f.join(t.AccountBook, f.eq(t.Account.ID, t.AccountBook.Book))],
	select: [b.ID, b.Name],
	where: [f.like(b.Name, "%bob%")],
}

[
	join(b, t.Account, and
]

join   book b
join   account a and b.Account = a.ID
and    b.Deleted = false
and    b.Name = 'Robert'
and    a.ID = in.account
select b.ID, b.Name
select b.store
insert temp.bns (id bigint, name text, store text)
;

join temp.bns t
select t.*
;


join (
	book b,
	account a and b.Account = a.ID,
)
and (
	b.Deleted = false,
	b.Name = 'Robert',
)
select (
	b.ID,
	b.Name,
	b.store
);

join (
	book b,
	account a and (
		b.Account = a.ID,
	),
)
and (
	b.Deleted = false,
	b.Name = 'Robert',
)
select (
	b.ID,
	b.Name,
	b.store,
);

q([
	local b = t.book,
	local a = t.account,
	local tmp = t.temp("bns"),
	join(b),
	join(a, eq(a.book, b.id)),
	eq(b.Deleted, false),
	eq(b.Name, 'Robert'),
	select(b.id, b.name),
	select(b.store),
	insert(tmp),
]) + q([

])
`,
		Column: []parser.StoreQueryColumn{
			{Table: "Book", StoreName: "ID", QueryName: "ID", UIBindName: "", Display: "", ReadOnly: true},
			{Table: "Book", StoreName: "Name", QueryName: "Name", UIBindName: "Name", Display: "Book Name", ReadOnly: false},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

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

	set, err := bufferFromStream(stream)
	if err != nil {
		t.Fatal(err)
	}
	_ = set
}

func bufferFromStream(stream parser.StreamingResultSet) (*parser.ResultSetBuffer, error) {
	set := &parser.ResultSetBuffer{}

	var el elist.EList
	var result *parser.ResultBuffer
	for {
		item, err := stream.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		switch v := item.(type) {
		default:
			return nil, fmt.Errorf("unknown state: %v", v.StreamState())
		case parser.StreamItemResultSetSchema:
			set.Schema = v.Schema
			set.Set = make([]parser.ResultBuffer, len(v.Schema.Set))
		case parser.StreamItemResult:
			result = &set.Set[v.SchemaIndex]
			result.Schema = set.Schema.Set[v.SchemaIndex]
		case parser.StreamItemRow:
			// result.Row = append(result.Row, v.Row)
		case parser.StreamItemColumn:
			return set, errors.New("column store not implemented")
		case parser.StreamItemEndOfResult:
			result = nil
		case parser.StreamItemEndOfSet:
			break
		case parser.StreamItemError:
			el.Add(v.Error)
		}
	}
	return set, el.ErrNil()
}
