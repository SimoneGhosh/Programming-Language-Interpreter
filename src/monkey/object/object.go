package object

//represent every value we encounter when evaluating Monkey source code as
//an Object, an interface of our design. Every value will be wrapped inside a struct, which fulfills
//this Object interface.

import (
	"fmt"
)

type ObjectType string

const (
	INTEGER_OBJ = "INTEGER"
	BOOLEAN_OBJ = "BOOLEAN"
	NULL_OBJ    = "NULL"
)

// interface instead of struct
// every value needs a different internal representation and it’s easier to define
// two different struct types than trying to fit booleans and integers into the same struct field
type Object interface {
	Type() ObjectType
	Inspect() string
}

// INTEGER
type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }

// BOOLEAN
type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }

// NULL
type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }
