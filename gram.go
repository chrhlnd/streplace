package main

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/chrhlnd/cmdlang"
)

type eval struct {
	items []interface{}
}

func (ev *eval) String() string {
	buf := bytes.Buffer{}
	for i, v := range ev.items {
		buf.WriteString(strconv.Itoa(i))
		buf.WriteString(") ")
		buf.WriteString(fmt.Sprintf("%v", v))
		buf.WriteString("\n")
	}
	return buf.String()
}

func (ev *eval) appendTok(tok cmdlang.TokInfo) {
	ev.items = append(ev.items, &tok)
}

func (ev *eval) appendSub(sub eval) {
	ev.items = append(ev.items, &sub)
}

func (ev *eval) getTok(idx int) *cmdlang.TokInfo {
	if idx >= len(ev.items) {
		return nil
	}

	if v, ok := ev.items[idx].(*cmdlang.TokInfo); ok {
		return v
	}

	return nil
}

func (ev *eval) getEval(idx int) *eval {
	if idx >= len(ev.items) {
		return nil
	}
	if v, ok := ev.items[idx].(*eval); ok {
		return v
	}
	return nil
}

type rule struct {
	id cmdlang.TokInfo
	ev eval
}

type emit struct {
	ev eval
}

type Grammer struct {
	rules []rule
	emits []emit
}

type ErrRpt func(cmdlang.TokInfo, error)

func (g *Grammer) addRule(r rule) {
	g.rules = append(g.rules, r)
}

func (g *Grammer) addEmit(e emit) {
	g.emits = append(g.emits, e)
}

var cEMIT = [...]byte{'!', 'e', 'm', 'i', 't'}

func (g *Grammer) handleEmit(ev eval, err ErrRpt) bool {
	if tok := ev.getTok(0); bytes.Compare(tok.Literal, cEMIT[0:]) != 0 {
		return false
	}

	g.addEmit(emit{ev: ev})
	return true
}

func (g *Grammer) handleRule(ev eval, err ErrRpt) bool {
	g.addRule(rule{id: *ev.getTok(0), ev: ev})
	return true
}

func (g *Grammer) String() string {
	var buf bytes.Buffer

	for _, r := range g.rules {
		buf.Write(r.id.Literal)
		buf.WriteString("\n")
	}

	return buf.String()
}

func NewGrammer(file io.Reader, errRpt ErrRpt) *Grammer {
	scanner := cmdlang.NewScanner(file)

	ret := &Grammer{}

	var scan func() eval

	eof := false

	scanId := 0

	verbose := false

	scan = func() eval {
		var tok cmdlang.TokInfo
		var capture eval
	SCAN:
		for tok = scanner.Scan(); tok.Token != cmdlang.TOK_EOF; tok = scanner.Scan() {
			if verbose {
				fmt.Print(scanId)
				fmt.Print(":")
				fmt.Println(tok)
			}
			switch tok.Token {
			case cmdlang.TOK_IDENT:
				capture.appendTok(tok)
			case cmdlang.TOK_WS:
			case cmdlang.TOK_EOC:
				break SCAN
			case cmdlang.TOK_BLOCK_START:
				capture.appendSub(scan())
			case cmdlang.TOK_BLOCK_END:
				break SCAN
			case cmdlang.TOK_COMMENT_BLOCK:
			case cmdlang.TOK_COMMENT_EOL:
			}
		}
		eof = tok.Token == cmdlang.TOK_EOF
		scanId++
		return capture
	}

	for !eof {
		ev := scan()

		if len(ev.items) < 1 {
			continue // skip empty lines
		}

		if ret.handleEmit(ev, errRpt) {
			continue
		}

		ret.handleRule(ev, errRpt)
	}

	//fmt.Printf("Grammer had %v Rules\n", len(ret.rules))
	//fmt.Printf("Grammer had %v Emits\n", len(ret.emits))

	//fmt.Println(ret)

	return ret
}
