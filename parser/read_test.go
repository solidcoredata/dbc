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
			src: `package foo

"account" table {
	id int64 serial -- This is a comment.
	name text default 'Hello World''s' /* This is
a multiline comment
*/
}
`,
		},
	}

	bg := context.Background()

	for _, item := range list {
		ctx, cancel := context.WithTimeout(bg, time.Second*1)
		err := run(ctx, item.src, func(token Token) {
			t.Logf("%#v", token)
		})
		cancel()
		if err != nil {
			t.Fatal(err)
		}
	}
}
