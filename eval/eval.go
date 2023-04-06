package eval

import (
	"dao/ast"
	"dao/meta"
	"fmt"
	"io"
	"os"
)

var (
	TRUE  = &meta.Bool{Value: true}
	FALSE = &meta.Bool{Value: false}
	NIL   = &meta.Nil{}
)

var builtins = map[string]*meta.Builtin{
	"len":  {Fn: Len},
	"echo": {Fn: Echo},
	"puts": {Fn: Puts},
}

func Eval(n ast.Node, e *meta.Env) meta.Meta {
	switch n := n.(type) {
	case *ast.Program:
		return program(n, e)
	case *ast.ExpressionStatement:
		return Eval(n.Expression, e)
	case *ast.IntegerLiteral:
		return &meta.Int{Value: n.Value}
	case *ast.BooleanLiteral:
		return nativeBool(n.Value)
	case *ast.PrefixExpression:
		right := Eval(n.Right, e)
		if isError(right) {
			return right
		}
		return prefixExp(n.Operator, right)
	case *ast.InfixExpression:
		left := Eval(n.Left, e)
		if isError(left) {
			return left
		}
		right := Eval(n.Right, e)
		if isError(right) {
			return right
		}
		return infixExp(n.Operator, left, right)
	case *ast.BlockStatement:
		return blockStatement(n, e)
	case *ast.IfExpression:
		cond := Eval(n.Condition, e)
		if isError(cond) {
			return cond
		}
		return ifExp(n, e)
	case *ast.ReturnStatement:
		val := Eval(n.ReturnValue, e)
		if isError(val) {
			return val
		}
		return &meta.ReturnValue{Value: val}
	case *ast.VarStatement:
		val := Eval(n.Value, e)
		if isError(val) {
			return val
		}
		e.Set(n.Name.Value, val)
	case *ast.AssignStatement:
		val := Eval(n.Value, e)
		if isError(val) {
			return val
		}
		assign(n, val, e)
	case *ast.ForStatement:
		return forStatement(n, e)
	case *ast.Identifier:
		return identifier(n, e)
	case *ast.FunctionLiteral:
		return function(n, e)
	case *ast.CallExpression:
		function := Eval(n.Function, e)
		if isError(function) {
			return function
		}

		args := expressions(n.Args, e)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(function, args)
	case *ast.StringLiteral:
		return &meta.String{Value: n.Value}
	}

	return NIL
}

func program(program *ast.Program, e *meta.Env) meta.Meta {
	var res meta.Meta

	for _, stmt := range program.Statements {
		res = Eval(stmt, e)

		switch res := res.(type) {
		case *meta.ReturnValue:
			return res.Value
		case *meta.Error:
			return res
		}
	}

	return res
}

func blockStatement(b *ast.BlockStatement, e *meta.Env) meta.Meta {
	var res meta.Meta

	for _, stmt := range b.Statements {
		res = Eval(stmt, e)

		if res != nil {
			rt := res.Type()
			if rt == meta.RETURN_VALUE || rt == meta.ERROR {
				return res
			}
		}
	}

	return res
}

func forStatement(f *ast.ForStatement, e *meta.Env) meta.Meta {
	var val meta.Meta

	if len(f.Condition) == 0 { // no condition
		for {
			val = blockStatement(f.Body, e)
			r, ok := val.(*meta.ReturnValue)

			if ok {
				return r.Value
			}
		}
	}

	for _, stmt := range f.Condition {
		_, ok := stmt.(*ast.VarStatement)
		if ok {
			Eval(stmt, e)
		}
	}

	for isTrue(Eval(f.Condition[len(f.Condition)-1], e)) {

		val = blockStatement(f.Body, e)
		r, ok := val.(*meta.ReturnValue)

		if ok {
			return r.Value
		}

		for _, stmt := range f.Condition {
			_, ok := stmt.(*ast.VarStatement)
			if !ok {
				Eval(stmt, e)
			}
		}
	}

	return val
}

func nativeBool(b bool) *meta.Bool {
	if b {
		return TRUE
	}
	return FALSE
}

func prefixExp(op string, right meta.Meta) meta.Meta {
	switch op {
	case "!":
		return bangOpExp(right)
	case "-":
		return minusOpExp(right)
	default:
		return newError("unknown operator: %s %s", op, right.Type())
	}
}

func bangOpExp(right meta.Meta) meta.Meta {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NIL:
		return TRUE
	default:
		return nil
	}
}

func minusOpExp(right meta.Meta) meta.Meta {
	if right.Type() != meta.INT {
		return newError("unknown operator: -%s", right.Type())
	}

	val := right.(*meta.Int).Value
	return &meta.Int{Value: -val}
}

func infixExp(op string, left meta.Meta, right meta.Meta) meta.Meta {
	switch {
	case left.Type() == meta.INT && right.Type() == meta.INT:
		return intInfixExp(op, left, right)
	case left.Type() == meta.STRING && right.Type() == meta.STRING:
		return strInfixExp(op, left, right)
	case left.Type() == meta.STRING && right.Type() == meta.INT || left.Type() == meta.INT && right.Type() == meta.STRING:
		return strPlusInfixExp(op, left, right)
	case op == "==":
		return nativeBool(left == right)
	case op == "!=":
		return nativeBool(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), op, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

func intInfixExp(op string, left meta.Meta, right meta.Meta) meta.Meta {
	lv := left.(*meta.Int).Value
	rv := right.(*meta.Int).Value
	switch op {
	case "+":
		return &meta.Int{Value: lv + rv}
	case "-":
		return &meta.Int{Value: lv - rv}
	case "*":
		return &meta.Int{Value: lv * rv}
	case "/":
		return &meta.Int{Value: lv / rv}
	case "%":
		return &meta.Int{Value: lv % rv}
	case ">":
		return nativeBool(lv > rv)
	case "<":
		return nativeBool(lv < rv)
	case "!=":
		return nativeBool(lv != rv)
	case "==":
		return nativeBool(lv == rv)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

func strInfixExp(op string, left meta.Meta, right meta.Meta) meta.Meta {
	lv := left.(*meta.String).Value
	rv := right.(*meta.String).Value
	switch op {
	case "+":
		return &meta.String{Value: lv + rv}
	case ">":
		return nativeBool(lv > rv)
	case "<":
		return nativeBool(lv < rv)
	case "!=":
		return nativeBool(lv != rv)
	case "==":
		return nativeBool(lv == rv)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

func strPlusInfixExp(op string, left meta.Meta, right meta.Meta) meta.Meta {
	switch op {
	case "*":
		if str, ok := left.(*meta.String); ok {
			if num, ok := right.(*meta.Int); ok {
				var res = ""
				for num.Value > 0 {
					res += str.Value
					num.Value -= 1
				}
				return &meta.String{Value: res}
			}
		}
		if num, ok := left.(*meta.Int); ok {
			if str, ok := right.(*meta.String); ok {
				var res = ""
				for num.Value > 0 {
					res += str.Value
					num.Value -= 1
				}
				return &meta.String{Value: res}
			}
		}
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), op, right.Type())
	}
}

func ifExp(m *ast.IfExpression, e *meta.Env) meta.Meta {
	cond := Eval(m.Condition, e)
	if isTrue(cond) {
		return Eval(m.Consequence, e)
	} else if len(m.Options) > 0 {
		for _, o := range m.Options {
			c := Eval(o.Condition, e)
			if isTrue(c) {
				return Eval(o.Consequence, e)
			}
		}

		if m.Alternative != nil {
			return Eval(m.Alternative, e)
		}

		return NIL

	} else if m.Alternative != nil {
		return Eval(m.Alternative, e)
	} else {
		return NIL
	}
}

func identifier(m *ast.Identifier, e *meta.Env) meta.Meta {

	if val, ok := e.Get(m.Value); ok {
		return val
	}

	if builtin, ok := builtins[m.Value]; ok {
		return builtin
	}

	return newError("identifier not found: " + m.Value)
}

func assign(m *ast.AssignStatement, val meta.Meta, e *meta.Env) {
	_, outer := e.GetWithEnv(m.Name.Value)
	e.Set(m.Name.Value, val)
	if outer != nil {
		outer.Set(m.Name.Value, val)
	}
}

func function(m *ast.FunctionLiteral, e *meta.Env) meta.Meta {
	args := m.Args
	body := m.Body
	name := m.Name

	funcMeta := &meta.Func{Args: args, Body: body, Env: e, Name: name}

	if name != nil { // for func name  eg. func add() {} call add()
		e.Set(name.Value, funcMeta)
	}

	return funcMeta
}

func expressions(exps []ast.Expression, e *meta.Env) []meta.Meta {
	var result []meta.Meta

	for _, exp := range exps {
		res := Eval(exp, e)
		if isError(res) {
			return []meta.Meta{res}
		}

		result = append(result, res)
	}
	return result
}

func applyFunction(fn meta.Meta, args []meta.Meta) meta.Meta {
	switch fn := fn.(type) {
	case *meta.Func:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *meta.Builtin:
		return fn.Fn(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *meta.Func, args []meta.Meta) *meta.Env {
	env := meta.NewEnclosedEnv(fn.Env)

	for i, arg := range fn.Args {
		env.Set(arg.Value, args[i])
	}

	return env
}

func unwrapReturnValue(m meta.Meta) meta.Meta {
	if returnValue, ok := m.(*meta.ReturnValue); ok {
		return returnValue.Value
	}

	return m
}

func isTrue(m meta.Meta) bool {
	switch m {
	case TRUE:
		return true
	case FALSE:
		return false
	case NIL:
		return false
	default:
		return true
	}
}

func isError(m meta.Meta) bool {
	if m != nil {
		return m.Type() == meta.ERROR
	}

	return false
}

func newError(format string, a ...interface{}) *meta.Error {
	return &meta.Error{Msg: fmt.Sprintf(format, a...)}
}

func Len(args ...meta.Meta) meta.Meta {
	if l := len(args); l != 1 {
		return newError("wrong number of arguments. got=%d, want=%d", l, 1)
	}

	switch arg := args[0].(type) {
	case *meta.String:
		return &meta.Int{Value: int64(len(arg.Value))}
	default:
		return newError("argument to `len` not supported yet, got %s", arg.Type())
	}
}

func Puts(args ...meta.Meta) meta.Meta {
	for _, arg := range args {
		io.WriteString(os.Stdout, arg.Echo())
		// if i < len(args)-1 {
		// 	io.WriteString(os.Stdout, ",")
		// }
	}
	io.WriteString(os.Stdout, "\n")

	return NIL
}

func Echo(args ...meta.Meta) meta.Meta {
	for _, arg := range args {
		io.WriteString(os.Stdout, arg.Echo())
		io.WriteString(os.Stdout, "\n")
	}

	return NIL
}
