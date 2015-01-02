package main

import (
	"fmt"
	"io"
	//	"log"
	"os"

	"github.com/chrhlnd/cmdlang"
)

func evalGramFile(name string, file io.Reader) (*Grammer, bool) {
	hadErr := false

	errorRpt := func(tok cmdlang.TokInfo, err error) {
		fmt.Print(name)
		fmt.Print(" - ")
		fmt.Print("ERR: ")
		fmt.Println(err)
		fmt.Print(" @ tok: ")
		fmt.Println(tok)
		hadErr = true
	}

	g := NewGrammer(file, errorRpt)
	if !hadErr {
		return g, true
	}
	return nil, false
}

var gramFiles []string
var looseFiles []string

func main() {
	arg0 := os.Args[0]
	args := os.Args[1:]

	var gramName string
	var gram *Grammer
	var file *os.File
	var err error
	var out io.Writer

	out = os.Stdout

	cmtPfx := "--"

	printHelp := func() {
		fmt.Println("Usage: ", arg0, "[cmt <string>] <gram file> [files ...] ... [<gram file> [files...]]")
	}

	if len(args) == 0 {
		printHelp()
		os.Exit(1)
	}

	exitNo := 0

WORK:
	for i := 0; i < len(args); i++ {
		option := args[i]

		//log.Println("Handling arg ", i, " ", option)

		switch option {
		case "cmt":
			i++
			if i > len(args) {
				fmt.Println("Error: ", arg0, " cmt <str>, expected string argument")
				printHelp()
				exitNo = 1
				break WORK
			}
			cmtPfx = args[i]
			//log.Println("Set comment to ", cmtPfx)
		case "gram":
			i++
			if i > len(args) {
				fmt.Println("Error: ", arg0, " gram <grammer file>, expected grammer file")
				printHelp()
				exitNo = 1
				break WORK
			}

			gramName = args[i]

			file, err = os.Open(gramName)
			if err != nil {
				fmt.Println("Error: '", gramName, "':  ", err)
				exitNo = 1
				break WORK
			}

			if g, ok := evalGramFile(gramName, file); !ok {
				exitNo = 1
				break WORK
			} else {
				gram = g
				//log.Println("Set gram ", gramName, " to ", gram)
			}

			file.Close()
			file = nil
		default:
			if gram == nil {
				fmt.Println("Error: no grammer file to process '", option, "' exiting")
				printHelp()
				exitNo = 1
				break WORK
			}
			//log.Println("Processing ", option)
			fmt.Fprintln(out, cmtPfx, " begin ", option)
			fmt.Fprintln(out, cmtPfx, " applying ", gramName)

			file, err = os.Open(option)
			if err != nil {
				fmt.Println("Error: '", option, "':  ", err)
				exitNo = 1
				break WORK
			}

			err = gram.Transform(file, out)
			if err != nil {
				fmt.Println("Error: '", option, "': ", err)
				exitNo = 1
				break WORK
			}

			fmt.Fprintln(out, cmtPfx, " end ", option)
			file.Close()
			file = nil
		}
	}
	os.Exit(exitNo)
}
