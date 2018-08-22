// dbc is the database compiler. It reads a schema, detects changs to it,
// and suggests alters.
package main

import (
	"context"
	"log"
	"os"

	"github.com/kardianos/task"
)

func main() {
	err := run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	cmd := &task.Command{
		Commands: []*task.Command{
			{
				Name:   "build",
				Usage:  "Build the current schema, but do not perform a release",
				Action: build{},
			},
			{
				Name:   "release",
				Usage:  "Release the current schema, increment the version number",
				Action: build{release: true},
			},
		},
	}

	return cmd.Exec(os.Args[1:]).Run(ctx, task.DefaultState(), nil)
}

type build struct {
	release bool
}

func (b build) Run(ctx context.Context, st *task.State, sc task.Script) error {
	return nil
}
