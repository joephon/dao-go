package eval

import (
	"dao/lexer"
	"dao/meta"
	"dao/parser"
	"testing"
)

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input       string
		expectedMsg string
	}{
		{"foobar", "identifier not found: foobar"},
		{
			"5 + true;",
			"type mismatch: INT + BOOL",
		},
		{
			"5 + true; 5;",
			"type mismatch: INT + BOOL",
		},
		{
			"-true",
			"unknown operator: -BOOL",
		},
		{
			"true + false;",
			"unknown operator: BOOL + BOOL",
		},
		{
			"5; true + false; 5",
			"unknown operator: BOOL + BOOL",
		},
		{
			"if (10 > 1) { true + false; }",
			"unknown operator: BOOL + BOOL",
		},
		{
			`
if 10 > 1 {
  if 10 < 1 {
    return true + false;
  } else if 10 < 2 * 3 {
	return !true
  } else {
	true + false
  }

  return 1;
}
`,
			"unknown operator: BOOL + BOOL",
		},
		{
			`"Hello" - "World"`,
			"unknown operator: STRING - STRING",
		},
		{
			`"hey" * true`, "type mismatch: STRING * BOOL",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		err, ok := evaluated.(*meta.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v)",
				evaluated, evaluated)
			continue
		}

		if err.Msg != tt.expectedMsg {
			t.Errorf("wrong error message. expected=%q, got=%q",
				tt.expectedMsg, err.Msg)
		}
	}
}
func TestEvalIntExp(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{input: "6", expected: 6},
		{input: "1", expected: 1},
		{input: "-7", expected: -7},
		{input: "-3", expected: -3},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tt := range tests {
		out := testEval(tt.input)
		testIntMeta(t, out, tt.expected)
	}
}

func TestEvalBoolExp(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"(1 > 2) == (2 > 3)", true},
		{"false == (1 > 2)", true},
	}

	for _, tt := range tests {
		out := testEval(tt.input)
		testBoolMeta(t, out, tt.expected)
	}
}

func TestBangOp(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!false", true},
		{"!true", false},
		{"!!false", false},
		{"!!true", true},
	}

	for _, tt := range tests {
		res := testEval(tt.input)
		testBoolMeta(t, res, tt.expected)
	}
}

func TestIfElseExp(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if true { 10 }", 10},
		{"if false { 10 }", nil},
		{"if 1 { 10 }", 10},
		{"if 1 < 2 { 10 }", 10},
		{"if 1 > 2 { 10 }", nil},
		{"if 1 > 2 { 10 } else { 20 }", 20},
		{"if 1 < 2 { 10 } else { 20 }", 10},
		{"if 1 > 2 { 10 } else if 3 < 4 { 20 }", 20},
		{"if 1 > 2 { 10 } else if 3 > 4 { 20 } else if 1 < 4 { 7 + 3 * 6 - 4 }", 21},
		{"if 1 > 2 - 2 { 21 } else if 3 > 4 { 21 } else if 1 > 4 { 21 } else if 1 < 4 { 7 + 3 * 6 - 4 }", 21},
	}

	for _, tt := range tests {
		res := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntMeta(t, res, int64(integer))
		} else {
			testNil(t, res)
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{
			`
if (10 > 1) {
  if (10 > 1) {
    return 10;
  }

  return 1;
}
`, 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntMeta(t, evaluated, tt.expected)
	}
}

func TestForStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`
 
for var x = 0; x < 8 {
	echo(x)
	x = x + 1
	if x > 6 {
		return x
	}
	
}


		`, 7},
		{`
 
for var x = 0; x = x + 1; x < 10 {
	echo(x)
	if x > 8 {
		return x
	}
	
}


		`, 9},
		{`
 
var x = 0
for {
	echo(x)
	x = x + 1
	if x > 8 {
		return x
	}
	
}


		`, 9},
		{`
var x = 0
for x < 10 {
	echo(x)
	x = x + 1
	if x > 8 {
		return x
	}

}

		`, 9},
		{`
var x = 0
for x < 10 {
	echo(x)
	x = x + 1
	return x

}

		`, 1},
	}

	for _, tt := range tests {
		testIntMeta(t, testEval(tt.input), tt.expected)
	}
}

func TestVarStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"var a int = 5; a", 5},
		{"var a = 5 * 5; a", 25},
		{"var a = 5; var b = a; b;", 5},
		{"var a = 5; var b = a; var c = a + b + 5; c;", 15},
		{"var a int = 13 % 6; a", 1},
	}

	for _, tt := range tests {
		testIntMeta(t, testEval(tt.input), tt.expected)
	}
}

func TestAssignStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"var a int = 5; a = a + 0; a", 5},
		{"var a = 5 * 5; a = a + 5; a", 30},
		{"var a = 5; var b = a; b = a + b; b", 10},
		{"var a = 5; var b = a; var c = a + b + 5; c = c + 5; c", 20},
		{"func a() { var b = 1; return func(x int) { b = b + x; return b;}}; var c = a(); c(1); c(1)", 3},
	}

	for _, tt := range tests {
		testIntMeta(t, testEval(tt.input), tt.expected)
	}
}

func TestFuncMeta(t *testing.T) {
	input := "func ooxx(x int) int { x + 2}"

	evaluated := testEval(input)
	fn, ok := evaluated.(*meta.Func)
	if !ok {
		t.Fatalf("target is not Function. got=%T (%+v)", evaluated, evaluated)
	}

	if len(fn.Args) != 1 {
		t.Fatalf("function has wrong parameters. Parameters=%+v",
			fn.Args)
	}

	if fn.Args[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Args[0])
	}

	expectedBody := "(x + 2)"

	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"var identity = func(x int) { x; }; identity(5);", 5},
		{"var identity = func add(x int) { return x; }; identity(5);", 5},
		{"var double = func(x int) { x * 2; }; double(5);", 10},
		{"var add = func(x int, y int) { x + y; }; add(5, 5);", 10},
		{"var add = func(x int, y int) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"func ooxx(x int) { x; }(5)", 5},
		{"func ooxx(x int) { x; }; ooxx(5)", 5},
	}

	for _, tt := range tests {
		testIntMeta(t, testEval(tt.input), tt.expected)
	}
}

func TestStringLiteral(t *testing.T) {
	input := `"Hello World!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*meta.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `"Hello" + " " + "World!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*meta.String)
	if !ok {
		t.Fatalf("target is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestStringComparation(t *testing.T) {
	tests := []struct {
		input    string
		expexted bool
	}{
		{`"Hello" > "ok"`, false},
		{`"Hello" < "ok"`, true},
		{`"Hello" != "ok"`, true},
		{`"Hello" == "ok"`, false},
		{`"Hello" == "Hello"`, true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		testBoolMeta(t, evaluated, tt.expexted)
	}
}

func TestStringPlusInt(t *testing.T) {
	tests := []struct {
		input    string
		expexted string
	}{
		{`"Hello" * 1`, "Hello"},
		{`"Hello " * 2`, "Hello Hello "},
		{`"Hello" * 3`, "HelloHelloHello"},
		{`1 * "Hello"`, "Hello"},
		{`2 * "Hello "`, "Hello Hello "},
		{`3 * "Hello"`, "HelloHelloHello"},
		{`"Hello" * 0`, ""},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		testStrMeta(t, evaluated, tt.expexted)
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, "argument to `len` not supported yet, got INT"},
		{`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
		{`echo("haha")`, NIL},
		{`puts("haha")`, NIL},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntMeta(t, evaluated, int64(expected))
		case string:
			errObj, ok := evaluated.(*meta.Error)
			if !ok {
				t.Errorf("target is not Error. got=%T (%+v)",
					evaluated, evaluated)
				continue
			}
			if errObj.Msg != expected {
				t.Errorf("wrong error message. expected=%q, got=%q",
					expected, errObj.Msg)
			}
		}
	}
}

func testNil(t *testing.T, m meta.Meta) bool {
	if m != NIL {
		t.Errorf("target is not NULL. got=%T (%+v)", m, m)
		return false
	}
	return true
}

func testEval(input string) meta.Meta {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.Parse()
	e := meta.NewEnv()

	return Eval(program, e)
}

func testIntMeta(t *testing.T, m meta.Meta, expected int64) bool {
	res, ok := m.(*meta.Int)
	if !ok {
		t.Errorf("target is not Int. got %T (%+v)", m, m)
		return false
	}

	if res.Value != expected {
		t.Errorf("target got wrong value. got %d. want %d", res.Value, expected)
		return false
	}

	return true
}

func testBoolMeta(t *testing.T, m meta.Meta, expected bool) bool {
	res, ok := m.(*meta.Bool)
	if !ok {
		t.Errorf("target is not Bool. got %T (%+v)", m, m)
		return false
	}

	if res.Value != expected {
		t.Errorf("target got wrong value. got %t. want %t", res.Value, expected)
		return false
	}

	return true
}

func testStrMeta(t *testing.T, m meta.Meta, expected string) bool {
	res, ok := m.(*meta.String)
	if !ok {
		t.Errorf("target is not string. got %T (%+v)", m, m)
		return false
	}

	if res.Value != expected {
		t.Errorf("target got wrong value. got %s. want %s", res.Value, expected)
		return false
	}

	return true
}
