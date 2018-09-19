# Solid Core Data: DB Compiler

 * Statically verify database schemas, schema changes, queries.
 * Standard method for declaring and executing schema alters.
 * Write queries that can run in many systems.
 * Authorize data use for all queries.
 * Setup a standard package system for queries, views, conditions (expressions) and schemas.
 * Make it easy to write custom queries and request ORM style records.

## Developing

 1. Declare schema, use for alters. Working MVP.
 2. Improve alter tooling, ensure alters are step-wise compatible.
 3. Create a separate product that reads the declared schema and connection string
    and provides a record based query system, results return a row delta.
	Ignore permissions for the most part. Working MVP.
 4. Expand the above service to allow custom queries, declare required
    meta-information alogn with the database specific query text.
	Limited parameter expansion for now.
 5. Allow manupulating the schema tables based on Tags and Named with query
    expressions.
 6. Allow creating unit tests of queries, views, and conditions. On each expression,
    extract the schema that is used, then substitute test vectors in for each
	referenced table. This allows test vectors to be narrower (only referenced
	columns need to be populated).
 7. Explore the topic of actually parsing a query to automatically extract and
    manipulate the query.

## Motivation

Databases are often initially designed, then slowly change over time as business
needs change or are discovered. By manifesting the database schema at compile time
and parsing or generating the queries against the database an entire class of bugs
can be eliminated. For while it is easy to get a query correct when initially
writting it against a database, maintaining it over time becomes a challenge.

By snapshotting the database state over time and storing them side-by-side,
database migrations are simple to declare and do so without being redundant
keeping a final schema up-to-date.

Make it easy to unit test SQL expressions (queries, views, and conditions) by
reducing the expression to referenced columns and allow substituting tables
with test vectors.

CRUD queries and ORM functionality come almost for free once the database
schema is known up front.

All queries are run through a single translation pipeline so developers
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

import (
	coredata.biz/app1/role	
)

{type: "varblock", options: [
	{type: "keyword", id: "table"},
	{type: "identifier"},
	{type: "varblock", options: [
		{type: "property},
		{type: "line", parts: [
			{type: "identifier"},
			{type: "identifier"},
			{type: "varblock", options: [
				{type: "property"},
			]},
		},
	]},
]}

// account holds a name and account number for use in the general ledger.
account table {
	alias: a
	display: Personal Account
	
	id int64 serial key
	name text
	number int64 {null:, default: null}
	number int64 {
		null:
		concurrent:
		unique:
		default: null
	}

	xname index [cluster] [unique] [concurrent] (Name) include (number) [using <name> [string params]] [where <filter>]
	
	name_number query {
		type: text
		or (
			name_number = a.name
			and (
				name_number:?int64 = true
				name_number::int64 = a.number
			)
			exists (
				from ledger l
				from account_ledger al and (l.id = al.ledger)
				and (
					al.account = a.id
					name_number = l.name
				)
			)
			exists (
				from (
					ledger l
					account_ledger al and (l.id = al.ledger)
				)
				and (
					al.account = a.id
					name_number = l.name
				)
			)
		)
	}
}

param (a account) name_number text {
	or (
		name_number = a.name
		and (
			name_number:?int64 = true
			name_number::int64 = a.number
		)
	)
}

table ledger {
	id int64 serial key
	name text
	balance decimal {
		default: 0
	}
}

table account_ledger {
	id int64 serial key {
		comment: used for primary key
		tag: xyz
		tag: abc
		display: ID of Join
	}
	account *account.id
	ledger *ledger.id
}

ckone query {
	param: aid *account.id
	from account a
	from account_ledger al and(a.id = al.account)
	from ledger l and(l.id = al.ledger)
	and a.id = aid
	select a.name "Account Name", l.name "Ledger", l.balance bal
}

```

```

func doit(part float64) table {
	var foo table {
		ID int64 serial key
		name text
		number int64 null default null

		xname index [cluster] [unique] [concurrent] (Name) include (number) [using <name> [string params]] [where <filter>]
	}

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

## Query Testing

A given query / view / condition can be analysized for the tables and columns it
uses. An implicit schema interface can be derived for testing. Then smaller
data vectors can be used for unit tests, and isolating the test data from other
un-related schema changes.

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
