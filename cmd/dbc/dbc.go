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
	flags := []*task.Flag{
		{Name: "alter", Usage: "alters output directory", Default: "alter"},
		{Name: "schema", Usage: "schema definition directory", Default: "schema"},
	}
	cmd := &task.Command{
		Commands: []*task.Command{
			{
				Name:   "build",
				Usage:  "Build the current schema, but do not perform a release",
				Flags:  flags,
				Action: build{},
			},
			{
				Name:   "release",
				Usage:  "Release the current schema, increment the version number",
				Flags:  flags,
				Action: build{release: true},
			},
		},
	}

	return cmd.Exec(os.Args[1:]).Run(ctx, task.DefaultState(), nil)
}

type build struct {
	// A release does the same thing as a build, plus extra steps at the end,
	// if the build worked, to record the next release.
	release bool
}

func (b build) Run(ctx context.Context, st *task.State, sc task.Script) error {
	// There are two things we need for a build:
	//  1. The directory where the table schema can be found.
	//  2. The directory where the alters can be found.
	//
	// I expect both of the above data to be checked in to source control, so
	// keep them as plain text. Actually, the alters could probably be stored
	// as a binary blob, even within a zip file. It needs to be checked in
	// and not altered. Also, the full schema will still show the diff
	// in source control.

	alterPath := st.Filepath(st.Get("alter"))
	schemaPath := st.Filepath(st.Get("schema"))

	// 1. Read current schema files from schema directory.
	// 2. Lex and parse the schema files. On error, fail and display errors.
	// 3. Verify the schema is valid and consistent.
	// 4. Read the most recent alter version.
	// 5. Verify the new schema is compatible with the previous version.
	//     The schema may introduce a field or table, or remove an unused field
	//     or table. The exact rules for compatible incremental changes need to
	//     be defined.
	// 6. If not a release or on error, exit.
	// 7. Update the schema version and write a new alter version.
	//     Each alter version needs to record the full schema as it stands
	//     at that version.

	return nil
}
