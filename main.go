// main.go

package main

import (
	"dao/repl"
	"fmt"
	"os"
	"os/user"
)

const VERSION = "0.0.1"

func main() {
	cmdLen := len(os.Args)
	switch {
	case cmdLen == 1:
		runRepl()
	case cmdLen == 2:
		if os.Args[1] == "-V" || os.Args[1] == "-v" || os.Args[1] == "version" {
			fmt.Println("v0.0.1")
		} else if os.Args[1] == "-h" || os.Args[1] == "help" {
			help()
		} else {
			repl.Eat(os.Args[1])
		}
	default:
		runRepl()
	}

}

func runRepl() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hey %s! welcome to the Dao programming language!\n",
		user.Username)
	fmt.Printf("current version: v%s\n", VERSION)
	help()
	repl.Run(os.Stdin, os.Stdout)
}

func help() {
	fmt.Println("Dao compiler usage:")
	fmt.Println(`
dao -h:        help list;
dao:           run the interpreter;
dao <source file>: eval the source file   
	`)
}
