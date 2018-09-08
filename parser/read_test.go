// Copyright 2018 solidcoredata authors.

package parser

import (
	"context"
	"testing"
	"time"
)

func TestLex(t *testing.T) {
	list := []struct {
		src string
	}{
		{
			src: `module foo/fee/v1

require (
	s/v1/time v1.0.1
)
`,
		},
		{
			src: `package foo

table "account"  {
	id int64 serial -- This is a comment.
	name text default 'Hello World''s' /* This is
a multiline comment
*/
}
`,
		},
		{
			src: `package foo

query "dancing" {
	join   book b
	join   account a and b.Account = a.ID
	and    b.Deleted = false
	and    b.Name = 'Robert'
	and    a.ID = in.account
	select b.ID, b.Name
	select b.store
	insert temp.bns tb (id bigint, name text, store text)
	select tb.id
}
`,
		},
	}

	bg := context.Background()

	for _, item := range list {
		ctx, cancel := context.WithTimeout(bg, time.Second*1)
		err := Lex(ctx, item.src, func(e Token) {
			t.Logf("%v", e)
		})
		cancel()
		if err != nil {
			t.Fatal(err)
		}
	}
}
