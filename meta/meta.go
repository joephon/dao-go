package meta

import (
	"bytes"
	"dao/ast"
	"fmt"
	"strings"
)

const (
	INT          = "INT"
	BOOL         = "BOOL"
	STRING       = "STRING"
	NIL          = "NIL"
	FUNC         = "FUNC"
	RETURN_VALUE = "RETURN_VALUE"
	ERROR        = "ERROR"
	BUILTIN      = "BUILTIN"
)

type MetaType string

type Meta interface {
	Type() MetaType
	Echo() string
}

type Int struct {
	Value int64
}

func (i *Int) Type() MetaType {
	return INT
}
func (i *Int) Echo() string {
	return fmt.Sprintf("%d", i.Value)
}

type Bool struct {
	Value bool
}

func (b *Bool) Type() MetaType {
	return BOOL
}
func (b *Bool) Echo() string {
	return fmt.Sprintf("%t", b.Value)
}

type String struct {
	Value string
}

func (s *String) Type() MetaType { return STRING }
func (s *String) Echo() string   { return s.Value }

type Nil struct{}

func (n *Nil) Type() MetaType {
	return NIL
}
func (n *Nil) Echo() string {
	return "nil"
}

type ReturnValue struct {
	Value Meta
}

func (rv *ReturnValue) Type() MetaType {
	return RETURN_VALUE
}
func (rv *ReturnValue) Echo() string {
	return rv.Value.Echo()
}

type Error struct {
	Msg string
}

func (e *Error) Type() MetaType {
	return ERROR
}
func (e *Error) Echo() string {
	return "syntax error: " + e.Msg
}

type Func struct {
	Name       *ast.Identifier
	ReturnType *ast.Identifier
	Args       []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Env
}

func (f *Func) Type() MetaType { return FUNC }
func (f *Func) Echo() string {
	var out bytes.Buffer

	args := []string{}
	for _, p := range f.Args {
		args = append(args, p.String()+" "+p.Type.Literal())
	}

	out.WriteString("func")
	if f.Name != nil {
		out.WriteString(" " + f.Name.Literal())
	}
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}

type Env struct {
	store map[string]Meta
	outer *Env
}

func (e *Env) Get(name string) (Meta, bool) {
	one, ok := e.store[name]
	if !ok && e.outer != nil {
		one, ok = e.outer.Get(name)
	}
	return one, ok
}

func (e *Env) GetWithEnv(name string) (Meta, *Env) {
	var outer *Env
	one, ok := e.store[name]
	if !ok && e.outer != nil {
		one, _ = e.outer.Get(name)
		outer = e.outer
	}
	return one, outer
}

func (e *Env) Set(name string, val Meta) Meta {
	e.store[name] = val
	return val
}

func NewEnv() *Env {
	s := make(map[string]Meta)
	return &Env{store: s, outer: nil}
}

func NewEnclosedEnv(outer *Env) *Env {
	env := NewEnv()
	env.outer = outer
	return env
}

type Builtin struct {
	Fn BuiltinFunc
}

type BuiltinFunc func(args ...Meta) Meta

func (b *Builtin) Type() MetaType { return BUILTIN }
func (b *Builtin) Echo() string   { return "builtin function" }
