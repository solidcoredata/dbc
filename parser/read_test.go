// Copyright 2018 solidcoredata authors.

package parser

import (
	"context"
	"testing"
	"time"
)

func TestLex(t *testing.T) {
	list := []struct {
		name string
		src  string
	}{
		{
			name: "module",
			src: `module foo/fee/v1

require (
	s/v1/time v1.0.1
)
`,
		},
		{
			name: "table",
			src: `package foo

define "account" table {
	id int64 serial -- This is a comment.
	name text default 'Hello World''s' /* This is
a multiline comment
*/
}
`,
		},
		{
			name: "query",
			src: `package foo

define "dancing" query {
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
		t.Run(item.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(bg, time.Second*1)
			tc := make(chan Token, 100)
			done := make(chan bool)
			go func() {
				defer close(done)
				for {
					select {
					case tok, ok := <-tc:
						if !ok {
							return
						}
						t.Logf("%v", tok)
					}
				}
			}()
			err := Lex1(ctx, item.src, tc)
			close(tc)
			cancel()
			if err != nil {
				t.Fatal(err)
			}
			<-done
		})
	}
}
