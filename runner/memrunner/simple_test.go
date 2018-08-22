package memrunner

import (
	"fmt"
	"io"
	"testing"

	"github.com/solidcoredata/dbc/internal/elist"
	"github.com/solidcoredata/dbc/query"
	"github.com/solidcoredata/dbc/runner"
)

func TestSimple(t *testing.T) {
	ms := &MemoryStore{
		Version: 1,
	}

	err := ms.AddTable(&query.StoreTable{
		Name:    "Book",
		Alias:   "b",
		Display: "Library Books",
		Comment: "Contains all the available library books.",
		Tag:     []string{"soft-delete"},
		Column: []*query.StoreColumn{
			{Name: "ID", Type: query.TypeInteger, Key: true, Serial: true},
			{Name: "Name", Type: query.TypeString, Display: "Book Name"},
		},
		Read: []query.Param{
			{
				Q:     "exists (select top 1 1 from Account a join AccountOrganization ao on a.ID = ao.Account where ao.Organization = b.Organization)",
				Input: []query.Input{{Type: query.TypeInteger, Name: "Account"}},
			},
		},
		Port: map[string]query.StoreTablePort{
			"internal": query.StoreTablePort{
				RoleAuthn: map[string]query.Authn{
					"user": query.AllowReturn | query.AllowInsert,
				},
				DenyRead: query.Param{},

				Column: []query.StoreColumnPort{},
			},
		},
	}, [][]interface{}{
		{1, "Never a Dull Moment"},
		{2, "To Kill a Bird"},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = ms.AddQuery(&query.Query{ /*
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
					Column: []query.StoreQueryColumn{
						{Table: "Book", StoreName: "ID", QueryName: "ID", UIBindName: "", Display: "", ReadOnly: true},
						{Table: "Book", StoreName: "Name", QueryName: "Name", UIBindName: "Name", Display: "Book Name", ReadOnly: false},
					},
		*/})
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

func bufferFromStream(stream query.StreamingResultSet) (*query.ResultSetBuffer, error) {
	set := &query.ResultSetBuffer{}

	var el elist.EList
	var result *query.ResultBuffer
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
		case query.StreamItemResultSetSchema:
			set.Schema = v.Schema
			set.Set = make([]query.ResultBuffer, len(v.Schema.Set))
		case query.StreamItemResult:
			result = &set.Set[v.SchemaIndex]
			result.Schema = set.Schema.Set[v.SchemaIndex]
		case query.StreamItemRow:
			// result.Row = append(result.Row, v.Row)
		case query.StreamItemEndOfResult:
			result = nil
		case query.StreamItemEndOfSet:
			break
		case query.StreamItemError:
			el.Add(v.Error)
		}
	}
	return set, el.ErrNil()
}
