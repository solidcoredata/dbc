package query

import (
	"context"
)

// Begin Schema

//go:generate stringer -type=Authn

type Authn uint8

const (
	AllowRead   Authn = 1             // Read allows fields to be used in where clauses.
	AllowReturn Authn = AllowRead + 2 // Return allows fields to be used in select, update, or insert clauses. Implies AllowRead.
	AllowInsert Authn = 4             // Insert allows fields to be set to none default values per columns, or allows rows to be inserted for tables.
	AllowUpdate Authn = 8             // Update allows fields to be updated per columns, or allows rows to be updated for tables.
	AllowDelete Authn = 16            // Delete allows fields to be reset to their default values per column, or allows rows to be deleted for tables.

	AllowNone Authn = 0                                                     // Allow no reads and no writes.
	AllowFull Authn = AllowReturn | AllowInsert | AllowUpdate | AllowDelete // Allow any operation to the field, column, or table.
)

type DataType int32

const (
	TypeUnknown    DataType = iota
	TypeString              // Unicode text
	TypeBinary              // []byte
	TypeBoolean             // bool
	TypeInteger             // int64
	TypeFloat               // float64
	TypeDecimal             // Arbitrary precision decimal
	TypeRational            // Arbitrary precision rational
	TypeTime                // Time of day
	TypeDate                // Date only
	TypeDatez               // Date + Timezone
	TypeTimestamp           // Time + Date
	TypeTimestampZ          // Time + Date + Timezone
	TypeUUID                // UUID
	TypeJSON                // JSON compatible store.
	TypeArray               // Array
)

// TODO(kardianos): Okay, I have a decent model for encoding table schemas
// and permissions (to an initial reasonable extent). But recording queries
// and conditions are lacking. This representation should already be parsed.
// The table schema is already parsed, the queries should be too. Furthermore,
// all structures should be designed to be easily serializable.
//
// Queries need to understand:
//  * Know what tables and columns are used. (For permissions)
//  * Know what tables and columns are returned. (For permissions and UI static checking and static UI building)
//  * Be able to insert additional conditions to statements. (To insert dynamic parameters and authroization checks)

type Store struct {
	Table []*StoreTable
	Query []Query
}

type RunParam struct {
	Name  string
	Value interface{}
}

type RunOption struct {
	Port  string
	Role  []string
	Param []RunParam
}

type ExpList interface {
	AddCondition(exp Exp)
}

type Exp struct {
	// Body contains the actual expression, like:
	//
	//    and (a.Foo = b.Foo, a.Deleted = 0)
	Body string
}

type Stmt struct {
	ExpList ExpList

	Read   []*ColumnSchema
	Return []*ColumnSchema
	Insert []*ColumnSchema
	Update []*ColumnSchema
	Delete []*ResultTableSchema
}

type Query struct {
	Stmt []Stmt
}

type StoreRunner interface {
	Run(ctx context.Context, q Query, opt RunOption) (StreamingResultSet, error)
}

type StoreQueryColumn struct {
	Table      string
	StoreName  string // Name of the column in the database.
	QueryName  string // Name of the column in the query.
	UIBindName string // Name of the UI field to bind to.
	Display    string // Suggested default name to display for the column.
	ReadOnly   bool   // Column may not be updated. Often the case for computed columns.
}

type StoreTable struct {
	Name    string // Table name.
	Alias   string // Suggested alias for queries.
	Display string // Suggested Display for the table.
	Comment string
	Tag     []string
	Column  []*StoreColumn
	Read    []Param

	Port map[string]StoreTablePort
}

// StoreTablePort defines a view of the database based on how it is accessed.
// For instance,
type StoreTablePort struct {
	RoleAuthn map[string]Authn

	// Per-named interface, per row checks to deny an operation.
	DenyRead   Param
	DenyInsert Param
	DenyUpdate Param
	DenyDelete Param

	Column []StoreColumnPort
}

type StoreColumnPort struct {
	RoleAuthn map[string]Authn

	// Per-named interface, per column, per row checks to deny an operation.
	DenyRead   Param
	DenyUpdate Param
}

type Input struct {
	Type DataType
	Name string
}

type Param struct {
	Q     string
	Input []Input
}

type StoreColumn struct {
	Name    string
	Comment string
	Tag     []string
	Display string // Suggested default name to display for the column.

	// Properties used for normal tables.
	Key          bool
	Serial       bool
	Nullable     bool
	Length       int32
	Type         DataType
	Default      interface{}
	LinkToTable  string
	LinkToColumn string

	UpdateLock bool // True if column should be compared prior to update, only allow if same.
	DeleteLock bool // True if the column should be compared prior to delete, only allow if same.
}

// Begin Result Schema

type ResultSetSchema struct {
	// Set of result schemas.
	Set []*ResultSchema
}

type ResultSchema struct {
	Role   string
	Column []*ColumnSchema
}

type ResultTableSchema struct {
	Name    string
	Alias   string // Table alias when used in a query.
	IsArity bool   // Table is the query arity.
}
type ColumnSchema struct {
	Table      *ResultTableSchema
	StoreName  string // Name of the column in the database.
	QueryName  string // Name of the column in the query.
	UIBindName string // Name of the UI field to bind to.
	Display    string // Suggested default name to display for the column.
	Allow      Authn  // Client should be advised server will enforce these restrictions.

	// Properties used for normal tables.
	Key          bool
	Serial       bool
	Nullable     bool
	UpdateLock   bool // True if column should be compared prior to update, only allow if same.
	DeleteLock   bool // True if the column should be compared prior to delete, only allow if same.
	Length       int32
	Type         DataType
	Default      interface{}
	LinkToTable  string
	LinkToColumn string
}

// Begin Stream

type StreamState byte

const (
	StreamUnknown         StreamState = iota
	StreamResultSetSchema             // Schema of the expected result.
	StreamResult                      // Value is the result set schema index.
	StreamRow                         // Value is an array of values forming a row of data.
	StreamEndOfResult                 // No value.
	StreamEndOfSet                    // No value.
	StreamError                       // Value is an error, signalling termination of the stream.
)

type StreamItem interface {
	StreamState() StreamState
}

type StreamField []byte

type StreamItemResultSetSchema struct{ Schema ResultSetSchema }
type StreamItemResult struct{ SchemaIndex int64 }
type StreamItemRow struct{ Row []StreamField }
type StreamItemEndOfResult struct{}
type StreamItemEndOfSet struct{}
type StreamItemError struct{ Error error }

func (StreamItemResultSetSchema) StreamState() StreamState { return StreamResultSetSchema }
func (StreamItemResult) StreamState() StreamState          { return StreamResult }
func (StreamItemRow) StreamState() StreamState             { return StreamRow }
func (StreamItemEndOfResult) StreamState() StreamState     { return StreamEndOfResult }
func (StreamItemEndOfSet) StreamState() StreamState        { return StreamEndOfSet }
func (StreamItemError) StreamState() StreamState           { return StreamError }

type StreamingResultSet interface {
	// returns io.EOF when done.
	Next() (StreamItem, error)
}

// Begin Buffer

type ResultSetBuffer struct {
	Schema ResultSetSchema
	Set    []ResultBuffer
}

type ResultBuffer struct {
	Schema *ResultSchema
	Row    []RowBuffer
}

type RowBuffer struct {
	Schema     *ResultSchema
	Column     []ValueBuffer
	Interleave []ResultBuffer
}

type ValueBuffer struct {
	Schema *ColumnSchema
	Allow  Authn       // Some columns may be read only (computed) or be denied read or update per row.
	Value  interface{} // Value is nil if NULL.
}
