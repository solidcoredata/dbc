package parser

// Begin Schema

//go:generate stringer -type=Authn

type Authn uint8

const (
	AllowRead Authn = 1 << iota
	AllowInsert
	AllowUpdate
	AllowDelete

	AllowNone Authn = 0
	AllowRW   Authn = AllowRead | AllowInsert | AllowUpdate | AllowDelete
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

type StoreTable struct {
	Name      string // Table name.
	Alias     string // Suggested alias for queries.
	Comment   string
	Tag       []string
	RoleAuthn map[string]Authn
	Column    []*StoreColumn
	Read      []Param

	// Per-named interface, per row checks to deny an operation.
	DenyRead   map[string]Param
	DenyInsert map[string]Param
	DenyUpdate map[string]Param
	DenyDelete map[string]Param
}

type Input struct {
	DataType DataType
	Name     string
}

type Param struct {
	Q     string
	Input []Input
}

type StoreColumn struct {
	Name      string
	Comment   string
	Tag       []string
	RoleAuthn map[string]Authn
	Display   string // Suggested default name to display for the column.

	// Per-named interface, per column, per row checks to deny an operation.
	DenyRead   map[string]Param
	DenyUpdate map[string]Param

	// Properties used for normal tables.
	Key          bool
	Serial       bool
	Nullable     bool
	UpdateLock   bool // True if column should be compared prior to update, only allow if same.
	DeleteLock   bool // True if the column should be compared prior to delete, only allow if same.
	Length       int32
	DataType     DataType
	Default      interface{}
	LinkToTable  string
	LinkToColumn string
}

// Begin Result Schema

type ResultSetSchema struct {
	// Set of result schemas.
	Set []*ResultSchema
}

type ResultSchema struct {
	Role       string
	NotPrimary bool // Not present in primary data set, may be interleaved.
	Column     []*ColumnSchema
}

type ResultTableSchema struct {
	Name           string
	Allow          Authn // Client should be advised server will enforce these restrictions.
	IsArity        bool
	EncodeByColumn bool
}
type ColumnSchema struct {
	Table      *ResultTableSchema
	StoreName  string // Name of the column in the database.
	QueryName  string // Name of the column in the query.
	UIBindName string // Name of the UI field to bind to.
	Display    string // Suggested default name to display for the column.
	ReadOnly   bool   // Column may not be updated. Often the case for computed columns.

	// Properties used for normal tables.
	Key          bool
	Serial       bool
	Nullable     bool
	UpdateLock   bool // True if column should be compared prior to update, only allow if same.
	DeleteLock   bool // True if the column should be compared prior to delete, only allow if same.
	Length       int32
	DataType     DataType
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
	StreamColumn                      // Value is an array of values forming a column of data.
	StreamEndOfResult                 // No value.
	StreamEndOfSet                    // No value.
	StreamError                       // Value is an error, signalling termination of the stream.
)

type StreamingResultSet interface {
	Next() (StreamState, []byte)
}

// Begin Buffer

type ResultSetBuffer struct {
	Schema *ResultSetSchema
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
