package evaluator

import (
	"github.com/noculture/plug/lexer"
	"github.com/noculture/plug/object"
	"github.com/noculture/plug/parser"
	"testing"
)

type TestCase struct {
	input    string
	expected interface{}
}

type IntegerTestCase struct {
	input    string
	expected int64
}

type BooleanTestCase struct {
	input    string
	expected bool
}

func TestFunctionObject(t *testing.T) {
	input := "func(x) {x + 2;}"
	evaluated := testEval(input)
	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not a function, got %T (%+v)", evaluated, evaluated)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters, got %+v", fn.Parameters)
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not x, got %q", fn.Parameters[0])
	}

	expectedBody := "(x + 2)"
	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q, got %q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	testCases := []IntegerTestCase{
		{"let identity = func(x) { x; }; identity(5);", 5},
		{"let identity = func(x) { return x; }; identity(5);", 5},
		{"let double = func(x) { x * 2; }; double(5);", 10},
		{"let add = func(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = func(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"func(x) { x; }(5)", 5},
	}

	testIntegerCases(testCases, t)
}

func TestIfElseExpressions(t *testing.T) {
	testCases := []TestCase{
		{"if (true) { return 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}

	for _, testCase := range testCases {
		evaluated := testEval(testCase.input)
		integer, ok := testCase.expected.(int)
		if ok {
			testIntegerObject(t, int64(integer), evaluated)
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestLetStatements(t *testing.T) {
	testCases := []IntegerTestCase{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = b + a + 5; c;", 15},
	}

	testIntegerCases(testCases, t)
}

func TestForStatements(t *testing.T) {
	testCases := []IntegerTestCase{
		{"for i = range(5) { i }", 4},
		{"for i = range(5) { return i; }", 0},
	}

	testIntegerCases(testCases, t)
}

func TestReturnStatements(t *testing.T) {
	testCases := []IntegerTestCase{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 10; 9;", 10},
		{`if (10 > 1) {
					if (10 > 1) {
						return 10;
					}
					return 1;
				}`, 10},
	}

	testIntegerCases(testCases, t)
}

func TestEvalIntegerExpression(t *testing.T) {
	testCases := []IntegerTestCase{
		{"5", 5},
		{"7", 7},
		{"-7", -7},
		{"-7", -7},
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

	testIntegerCases(testCases, t)
}

func TestStringLiteral(t *testing.T) {
	input := `"Hello World!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String, got %T (%+v)", evaluated, evaluated)
	}
	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value, got %q", str.Value)
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `"Hello" + " " + "World!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String, got %T (%+v)", evaluated, evaluated)
	}
	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value, got %q", str.Value)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	testCases := []BooleanTestCase{
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
	}

	for _, testCase := range testCases {
		evaluated := testEval(testCase.input)
		testBoolObject(t, testCase.expected, evaluated)
	}
}

func TestBangOperator(t *testing.T) {
	testCases := []BooleanTestCase{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, testCase := range testCases {
		evaluated := testEval(testCase.input)
		testBoolObject(t, testCase.expected, evaluated)
	}
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 3, 4 + 5]"

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object is not an array. got %T (%+v)", evaluated, evaluated)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("length of elements is not 3, got %d", len(result.Elements))
	}

	testIntegerObject(t, 1, result.Elements[0])
	testIntegerObject(t, 6, result.Elements[1])
	testIntegerObject(t, 9, result.Elements[2])
}

func TestArrayIndexExpressions(t *testing.T) {
	testCases := []TestCase{
		{"[1, 2, 3][0]", 1},
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][2]", 3},
		{"let i = 0; [1][i];", 1},
		{"[1, 2, 3][1 + 1];", 3},
		{"let myArray = [1, 2, 3]; myArray[2];", 3},
		{"let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];", 6},
		{"let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i]", 2},
		{"[1, 2, 3][3]", nil},
		{"[1, 2, 3][-1]", nil},
	}
	for _, testCase := range testCases {
		evaluated := testEval(testCase.input)
		integer, ok := testCase.expected.(int)
		if ok {
			testIntegerObject(t, int64(integer), evaluated)
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestErrorHandling(t *testing.T) {
	testCases := []struct {
		input           string
		expectedMessage string
	}{
		{"5 + true;", "type mismatch: INTEGER + BOOLEAN"},
		{"5 + true; 5", "type mismatch: INTEGER + BOOLEAN"},
		{"-true;", "unknown operator: -BOOLEAN"},
		{"false + true;", "unknown operator: BOOLEAN + BOOLEAN"},
		{"5; false + true; 5", "unknown operator: BOOLEAN + BOOLEAN"},
		{"if (10 > 1) {true + false}", "unknown operator: BOOLEAN + BOOLEAN"},
		{`if (10 > 1) {
					if (10 > 1) {
						return true + false;
					}
					return 1;
				}`, "unknown operator: BOOLEAN + BOOLEAN"},
		{"foobar", "identifier not found: foobar"},
		{`"Hello" - "World"`, "unknown operator: STRING - STRING"},
	}

	for _, testCase := range testCases {
		evaluated := testEval(testCase.input)
		errorObject, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("No error object returned, got %T (%+v)", evaluated, evaluated)
			continue
		}

		if errorObject.Message != testCase.expectedMessage {
			t.Errorf("Wrong error message, expected %q, got %q", testCase.expectedMessage, errorObject.Message)
		}
	}
}

func TestBuiltinFunctions(t *testing.T) {
	testCases := []TestCase{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len([])`, 0},
		{`len([3, 9, 5])`, 3},
		{`len(1)`, "argument to `len` not supported, got INTEGER"},
		{`len("one", "train")`, "invalid number of arguments to `len`, expected 1, got 2"},
		{`first([])`, nil},
		{`first([3, 9, 5])`, 3},
		{`last([])`, nil},
		{`last([3, 9, 5])`, 5},
		{`rest([3, 9, 5])`, nil},
		{`rest([3, 9, 5])`, []int{9, 5}},
		{`push([3, 9, 5], 6)`, []int{3, 9, 5, 6}},
		{`push([], 1)`, []int{1}},
		{`push("b", "a")`, "first argument to `push` not supported, expected ARRAY, got STRING"},
	}

	for _, testCase := range testCases {
		evaluated := testEval(testCase.input)
		switch expected := testCase.expected.(type) {
		case int:
			testIntegerObject(t, int64(expected), evaluated)
		case string:
			error, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("object is not an error, got %T (%+v)", evaluated, evaluated)
				continue
			}
			if error.Message != expected {
				t.Errorf("wrong error mesage, expected %q, got %q", expected, error.Message)
			}
		case []int:
			array, ok := evaluated.(*object.Array)
			if !ok {
				t.Errorf("object is not an array, got %T (%+v)", evaluated, evaluated)
				continue
			}
			for index, element := range expected {
				testIntegerObject(t, int64(element), array.Elements[index])
			}
		}
	}
}

func testEval(input string) object.Object {
	lex := lexer.New(input)
	p := parser.New(lex)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	return Eval(program, env)
}

func testIntegerCases(testCases []IntegerTestCase, t *testing.T) {
	for _, testCase := range testCases {
		testIntegerObject(t, testCase.expected, testEval(testCase.input))
	}
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("object is not null, got %T (%+v)", obj, obj)
		return false
	}
	return true
}

func testIntegerObject(t *testing.T, expected int64, evaluated object.Object) bool {
	result, ok := evaluated.(*object.Integer)
	if !ok {
		t.Errorf("evaluated isn't a plug integer, got %T (%v),", evaluated, evaluated)
		return false
	}
	if result.Value != expected {
		t.Errorf("got wrong integer value, expected %d, got %d", expected, result.Value)
	}

	return true
}

func testBoolObject(t *testing.T, expected bool, evaluated object.Object) bool {
	result, ok := evaluated.(*object.Boolean)
	if !ok {
		t.Errorf("evaluated isn't a plug boolean, got %T (%v),", evaluated, evaluated)
		return false
	}
	if result.Value != expected {
		t.Errorf("got wrong boolean value, expected %t, got %t", expected, result.Value)
	}

	return true
}
