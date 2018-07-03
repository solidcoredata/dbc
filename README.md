# Solid Core Data: DB Compiler

 * Statically verify database schemas, schema changes, queries.
 * Write queries that can run in many systems.
 * Authorize data use for all queries.
 * Custom queries, ORM, CRUD.

## Motivation

Databases are often initially designed, then slowly change over time as business
needs change or are discovered. By manifesting the database schema at compile time
and parsing or generating the queries against the database an entire class of bugs
can be eliminated. For while it is easy to get a query correct when initially
writting it against a database, maintaining it over time becomes a challenge.

By snapshotting the database state over time and storing them side-by-side,
database migrations are simple to declare and do so without being redundant
keeping a final schema up-to-date.

CRUD queries and ORM functionality come almost for free once the database
schema is known up front.

Because all queries are run through a single translation pipeline, developers
can declare that certain tables must restrict access to users with sufficent
authorization. It also allows CRUD queries to dynamically append search criteria
for fully featured search-list-detail screens.

## Design

There are two primary components: a developer tool that verifies and compiles
the database schema and queries into a single compact representation, and a
runtime server that loads the schema, schema migrations, and queries for
active use.

### Example Use

First the database and queries must be defined on the filesystem:
```
database/
	ar/
		accounts.scd
		payments.scd

--- accounts.scd ---

package ar

create account table {
	ID int64 serial key
	name text
	number int64 null default null

	xname index [cluster] [unique] [concurrent] (Name) include (number) [using <name> [string params]] [where <filter>]
}


func doit(part float64) table {
	create temp.Foo table {
		ID int64 serial key
		name text
		number int64 null default null

		xname index [cluster] [unique] [concurrent] (Name) include (number) [using <name> [string params]] [where <filter>]
	}
	drop temp.Foo

	type ti interface {
		Name text
		Deleted bool
	}
	func notDeleted(t ti) {
		from
			t
		and
			t.Deleted = false
	}

	func cte1(name text) table {
		from
			Table3 t3
		and
			t3.Type = 'dance'
			and t3.Name = name
		select
			t3.ID, t3.Table2,
	}

	// Update table with join to other tables.
	from
		Table1 t1
		join Table t2 and t1.ID = t2.ID
		join ct3(t1.Name) t3 and t3.Table2 = t2.ID
	and
		t2.Part = part
		#where1
		or (
			t3.ID = 0
			t1.Name = 'Nothing'
		)
	update t1
		Name = 'Hello',
	
	// Insert row by value into table.
	from
		Table1 t1
	insert t1
		Name = 'Hello',
	
	// Insert into table, output inserted row (OUTPUT or RETURNING).
	from
		Table1 t1
	insert t1
		Name = 'Hello',
	select
		t1.ID,
	
	// Insert from other existing data which is joined to inserting table.
	from
		Table1 t1
		join Table t2 and t1.ID = t2.ID
		join ct3(t1.Name) t3 and t3.Table2 = t2.ID
	and
		t2.Part = part
		#where1
		or (
			t3.ID = 0
			t1.Name = 'Nothing'
		)
	insert t1
		Name = 'Hello',
	
	// Insert from other existing data which is not joined to inserting table.
	from
		Table t2
		join ct3(t1.Name) t3 and t3.Table2 = t2.ID
	from
		Table1 t1
	and
		t2.ID = 4
	insert t1
		Name = 'Hello',
	
	from
		Table1 t1
		join Table t2 and t1.ID = t2.ID
		join ct3(t1.Name) t3 and t3.Table2 = t2.ID
	and
		t2.Part = part
		#where1
		or (
			t3.ID = 0
			t1.Name = 'Nothing'
		)
	select
		t1.Name, NamePart = t2.Part,
	order
		t1.Name asc
	limit 50 offset 10
	
	// For insert, update, and select, "name = t.name," is the same thing as
	// writting "t.name,".
}

--- payments.scd ---

package ar

create payment table {
	ID int64 serial key
	account fk<account.ID> null
	amount decimal default 0
}

mixin (pay payment) IsPositive(IsPositive bool) {
	if IsPositive {
		and
			pay.amount > 0
	}
}
```

Now the user runs `dbc build databae/ar` which parses the defined
schema and queries and if there are no errors, outputs two binary files,
one for the schema and one for the queries.

When a migration is ready to be defined, the developer runs `dbc schema commit`
which takes the current and previous schema definitions, and optionally an
migration instruction file, to produce an alter definitions.

The developer can package the schema, schema migration, and queries and hand them
to the operations team. If the developer is the "operations team", they can
deploy it themselves or package it along with their application.

## Implementation

Both the developer tools and the runtime server will be written in Go. However
there is not restriction on what languages or runtimes they can be used with.
For ORM uses, code generators are pluggable and simple to implement.

### Keywords

select, insert, update, and, or, order, limit, offset,
func, create, table, index

### Security

To report a security bug plase email [security@solidcoredata.org](mailto:security@solidcoredata.org).
Your email will be acknowledged within 72 hours.

### License

Source code files should not list authors names directly.
Each file should have a standard header:
```
// Copyright 2018 The Solid Core Data Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
```

At the moment copyright is not assigned. However before external contributions
are accepted or the framework used in production, a business must be formed
and copyright assigned directly to the project. The Solid Core Data project will
be made distinct from the business name.
