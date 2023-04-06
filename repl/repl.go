package repl

import (
	"bufio"
	"dao/eval"
	"dao/lexer"
	"dao/meta"
	"dao/parser"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

const PROMPT = "|☰☷☳☶☱☴☵☲|"

func Run(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	e := meta.NewEnv()

	for {
		fmt.Fprint(out, PROMPT)
		scaned := scanner.Scan()
		if !scaned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)
		program := p.Parse()

		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		res := eval.Eval(program, e)
		if res != nil {
			io.WriteString(out, res.Echo())
			io.WriteString(out, "\n")
		}
	}
}

func Eat(path string) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println("can not open file ", path)
		defer f.Close()
		return
	}

	in, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println("read file failed!")
		return
	}

	l := lexer.New(string(in))
	p := parser.New(l)
	program := p.Parse()
	e := meta.NewEnv()

	if len(p.Errors()) != 0 {
		printParserErrors(os.Stdout, p.Errors())
	}

	res := eval.Eval(program, e)
	if res != nil {
		io.WriteString(os.Stdout, res.Echo())
		io.WriteString(os.Stdout, "\n")
	}
}

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
