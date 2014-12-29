package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/chrhlnd/cmdlang"
)

var grammers map[string]*Grammer

func evalGramFile(name string, file io.Reader) {
	hadErr := false

	if grammers == nil {
		grammers = make(map[string]*Grammer)
	}

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
		grammers[name] = g
	}

}

var gramFiles []string
var looseFiles []string

func main() {
	flag.Parse()

	modeGram := false

	for _, f := range flag.Args() {
		if f == "gram" {
			modeGram = true
			continue
		}

		if modeGram {
			gramFiles = append(gramFiles, f)
			modeGram = false
			continue
		}

		looseFiles = append(looseFiles, f)
	}

	for _, g := range gramFiles {
		file, err := os.Open(g)
		if err != nil {
			fmt.Print("GRAM ERR: ")
			fmt.Println(err)
			continue
		}

		evalGramFile(g, file)

		file.Close()
	}

	for _, f := range looseFiles {
		fmt.Println("-- ---->>>", f, " -----")

		file, err := os.Open(f)
		if err != nil {
			fmt.Print("ERR: ")
			fmt.Println(err)
		} else {
			for gname, g := range grammers {
				fmt.Println("-- --- using: ", gname, " ----- ")

				var buf bytes.Buffer
				err := g.Transform(file, &buf)
				if err != nil {
					fmt.Print("ERROR:")
					fmt.Println(err)
					fmt.Print("Parsed:\n", buf.String(), "\n")
				}
				fmt.Println(buf.String())
			}

			/*
				scanner := cmdlang.NewScanner(file)

				var tok cmdlang.TokInfo

				for tok = scanner.Scan(); tok.Token != cmdlang.TOK_EOF; tok = scanner.Scan() {
					fmt.Printf("%v\n", tok)
				}
				fmt.Printf("%v\n", tok)

			*/
			file.Close()
		}

		fmt.Println("-- ----<<< ", f, " -----")
	}
}
