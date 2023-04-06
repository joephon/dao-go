package ast

import (
	"bytes"
	"dao/token"
	"strings"
)

type Node interface {
	Literal() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) Literal() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].Literal()
	} else {
		return ""
	}

	// for _, stmt := range p.Statements {
	// 	return stmt.Literal()
	// }
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

type VarStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (vs *VarStatement) statementNode() {}
func (vs *VarStatement) Literal() string {
	return vs.Token.Literal
}
func (vs *VarStatement) String() string {
	var out bytes.Buffer

	out.WriteString(vs.Literal() + " ")
	out.WriteString(vs.Name.String())
	out.WriteString(" = ")

	if vs.Value != nil {
		out.WriteString(vs.Value.String())
	}

	out.WriteString(";\n")

	return out.String()
}

type AssignStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (as *AssignStatement) statementNode() {}
func (as *AssignStatement) Literal() string {
	return as.Token.Literal
}
func (as *AssignStatement) String() string {
	var out bytes.Buffer

	out.WriteString(as.Name.String())
	out.WriteString(" = ")

	if as.Value != nil {
		out.WriteString(as.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()  {}
func (rs *ReturnStatement) Literal() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.Literal() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}

type Identifier struct {
	Token token.Token
	Value string
	Type  *Identifier
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) Literal() string {
	return i.Token.Literal
}
func (i *Identifier) String() string {
	return i.Value
}

type ExpressionStatement struct {
	Token      token.Token // 该表达式中的第一个词法单元
	Expression Expression
}

func (es *ExpressionStatement) statementNode()  {}
func (es *ExpressionStatement) Literal() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode() {}
func (il *IntegerLiteral) Literal() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string  { return il.Token.Literal }

type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode() {}
func (bl *BooleanLiteral) Literal() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string  { return bl.Token.Literal }

type NilLiteral struct {
	Token token.Token
	Value string
}

func (bl *NilLiteral) expressionNode() {}
func (bl *NilLiteral) Literal() string { return bl.Token.Literal }
func (bl *NilLiteral) String() string  { return bl.Token.Literal }

type PrefixExpression struct {
	Token    token.Token // 前缀词法单元，如!
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode() {}
func (pe *PrefixExpression) Literal() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

type InfixExpression struct {
	Token    token.Token
	Operator string
	Left     Expression
	Right    Expression
}

func (ie *InfixExpression) expressionNode() {}
func (ie *InfixExpression) Literal() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" ")
	out.WriteString(ie.Operator)
	out.WriteString(" ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

type ForStatement struct {
	Token     token.Token
	Condition []Statement
	Body      *BlockStatement
}

func (fs *ForStatement) statementNode()  {}
func (fs *ForStatement) Literal() string { return fs.Token.Literal }
func (fs *ForStatement) String() string {
	var out bytes.Buffer

	out.WriteString("for")
	for _, s := range fs.Condition {
		out.WriteString(s.String())
	}
	// out.WriteString(fs.String())
	out.WriteString(" ")

	return out.String()
}

type IfExpression struct {
	Token       token.Token // 'if'词法单元
	Condition   Expression
	Options     []*IfExpression // elif
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode() {}
func (ie *IfExpression) Literal() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())

	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}

	return out.String()
}

type BlockStatement struct {
	Token      token.Token // '{'词法单元
	Statements []Statement
}

func (bs *BlockStatement) statementNode()  {}
func (bs *BlockStatement) Literal() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

type FunctionLiteral struct {
	Token      token.Token // 'func'词法单元
	Name       *Identifier
	ReturnType *Identifier
	Args       []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode() {}
func (fl *FunctionLiteral) Literal() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Args {
		params = append(params, p.String())
	}

	out.WriteString(fl.Literal())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())

	return out.String()
}

type CallExpression struct {
	Token    token.Token // '('词法单元
	Function Expression  // 标识符或函数字面量
	Args     []Expression
}

func (ce *CallExpression) expressionNode() {}
func (ce *CallExpression) Literal() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ce.Args {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode() {}
func (sl *StringLiteral) Literal() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string  { return sl.Token.Literal }
