package main

import (
	"fmt"
	"io"
	//	"log"
	"os"
	"strings"

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

func showErrorToken(file io.Reader, tok cmdlang.TokInfo) {
	scanner := cmdlang.NewScanner(file)

	context := make([]cmdlang.TokInfo, 0, 10)

	var inTok cmdlang.TokInfo

	tokAt := 0

	for inTok = scanner.Scan(); inTok.Token != cmdlang.TOK_EOF; inTok = scanner.Scan() {
		context = append(context, inTok)

		if inTok.Cstart == tok.Cstart && inTok.Cend == tok.Cend {
			tokAt = len(context) - 1
		}

		if inTok.Lend > tok.Lend {
			break
		}
	}

	startAt := tokAt
	ident := 0
	for i := tokAt - 1; i > -1 && ident < 10; i-- {
		if context[i].Token == cmdlang.TOK_IDENT {
			ident++
		}
		startAt--
	}

	for i := startAt - 1; i > -1; i-- {
		if context[i].Lstart != context[startAt].Lstart {
			break
		}
	}

	padStart := tokAt
	for q := tokAt - 1; q > -1 && context[q].Lend == context[tokAt].Lstart; q-- {
		padStart = q
	}

	doPad := func() {
		for q := padStart; q < tokAt; q++ {
			fmt.Print(strings.Replace(string(context[q].Literal), "\n", "", 1))
		}
	}

	for i := startAt; i < len(context); i++ {
		v := context[i]
		if i == tokAt {
			fmt.Println("\n Error @ line: ", v.Lstart, " -----------")

			errv := string(v.Literal)

			doPad()
			fmt.Println(errv)
			doPad()
			for z := 0; z < len(errv); z++ {
				fmt.Print("^")
			}
			fmt.Println("")
			doPad()
			for z := 0; z < len(errv); z++ {
				fmt.Print(" ")
			}

		} else {
			fmt.Print(string(v.Literal))
		}
	}
}

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
		fmt.Println("Usage: ", arg0, "[cmt <string>] [pfx <spec>] <gram file> [files ...] ... [<gram file> [files...]]")
		fmt.Println("       <spec> = <id>:<string> - causes (!pfx <id>) to emit string in emission rules. ")
	}

	if len(args) == 0 {
		printHelp()
		os.Exit(1)
	}

	exitNo := 0

	pfxKeys := make(map[string]string)

WORK:
	for i := 0; i < len(args); i++ {
		option := args[i]

		//log.Println("Handling arg ", i, " ", option)

		switch option {
		case "pfx":
			i++
			if i > len(args) {
				fmt.Println("Error: ", arg0, " pfx missing specifier <id>:stmt")
				printHelp()
				exitNo = 1
				break WORK
			}
			if !strings.Contains(args[i], ":") {
				fmt.Println("Error: ", arg0, " pfx <spec> - invalid spec settings `", args[i], "` must be <id>:<string><option>")
				printHelp()
				exitNo = 1
				break WORK

			}
			parts := strings.Split(args[i], ":")
			pfxKeys[parts[0]] = parts[1]
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

			err = gram.Transform(file, out, pfxKeys)
			if err != nil {
				fmt.Println("Error: '", option, "': ", err)

				if terr, ok := err.(TransformError); ok {
					if _, serr := file.Seek(0, 0); serr != nil {
						fmt.Println(serr)
					} else {
						showErrorToken(file, terr.Token)
					}
				}

				exitNo = 1
				break WORK
			}

			fmt.Fprintln(out, cmtPfx, " end ", option)
			file.Close()
			file = nil
		}
	}

	if file != nil {
		file.Close()
		file = nil
	}
	os.Exit(exitNo)
}
