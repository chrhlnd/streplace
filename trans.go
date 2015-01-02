package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/chrhlnd/cmdlang"
)

type TransformError struct {
	Msg   string
	Token cmdlang.TokInfo
}

func newTErr(tok cmdlang.TokInfo, frmt string, args ...interface{}) TransformError {
	ret := TransformError{}
	ret.Msg = fmt.Sprintf(frmt, args...)
	ret.Token = tok
	return ret
}

func (te TransformError) Error() string {
	return fmt.Sprintf("Error '%v' at %v", te.Msg, te.Token)
}

type tConsumer func() (*cmdlang.TokInfo, bool)
type Processor func(cap *capture, op cmdlang.TokInfo, consumeTok tConsumer, rs *rulestate) error

type rulestate struct {
	Value   int
	Process Processor
	Child   *rulestate
	Parent  *rulestate
}

func (t *rulestate) get() int {
	return t.Value
}

func (t *rulestate) inc() {
	t.Value++
}

func (t *rulestate) set(p int) {
	t.Value = p
}

func (t *rulestate) push() *rulestate {
	t.Child = &rulestate{Parent: t}
	return t.Child
}

func (t *rulestate) pop() *rulestate {
	t.Parent.Child = nil
	return t.Parent
}

type capture struct {
	Type     *cmdlang.TokInfo
	Settings map[string]interface{}
	Vars     map[string]interface{}
	Parent   *capture

	Children []*capture

	RuleIdx int

	Rs rulestate
}

func (c *capture) getSettingAsList(name string) []interface{} {
	var v interface{}
	var ok bool

	if v, ok = c.Settings[name]; !ok {
		if c.Parent != nil {
			return c.Parent.getSettingAsList(name)
		}
	}

	if vv, ok := v.([]interface{}); ok {
		return vv
	}

	if v != nil {
		ret := make([]interface{}, 1)
		ret[0] = v
		return ret
	}
	return nil
}

func (c *capture) getSettingToken(name string) *cmdlang.TokInfo {
	if v, ok := c.Settings[name]; ok {
		if ret, ok := v.(cmdlang.TokInfo); ok {
			return &ret
		}
	}

	if c.Parent != nil {
		return c.Parent.getSettingToken(name)
	}

	return nil
}

func (c *capture) getVarStr(name string) string {
	if v, ok := c.Vars[name]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	} else {
		if c.Parent != nil {
			return c.Parent.getVarStr(name)
		}
	}
	return ""
}

func (c *capture) setVar(name string, val interface{}) {
	c.Vars[name] = val
}

var gARG = []byte("!gArg")

func newCapture(p *capture) *capture {
	ret := &capture{}
	ret.Parent = p
	ret.RuleIdx = -1
	ret.Settings = make(map[string]interface{})
	ret.Vars = make(map[string]interface{})
	return ret
}

/*
	restoreState

	evalRules
		if want tok, use if there else return
*/

func restoreRuleState(current *capture, evRule rule) (bool, *eval, *rulestate, int) {
	depth := 0
	ruleEv := &evRule.ev
	ruleState := &current.Rs

	for {
		// walk existing child states
		for ruleState.Child != nil {
			child := ruleEv.getEval(ruleState.get())
			ruleEv = child
			ruleState = ruleState.Child
			depth++
		}

		// walk new child states
		ruleEval := ruleEv.getEval(ruleState.get())
		if ruleEval != nil {
			ruleState = ruleState.push()
			ruleEv = ruleEval
			depth++
			continue
		}

		// if we've advanced past the number of rule items
		if ruleState.get() >= len(ruleEv.items) {
			if depth > 0 {
				parentState := ruleState.pop()
				parentState.inc()
				ruleEv = &evRule.ev
				ruleState = &current.Rs
				depth--
				continue
			}

			// we've advanced all the way to the top
			current.Rs.set(0)
			return true, nil, nil, 0
		}
		break
	}

	return false, ruleEv, ruleState, depth
}

func isGArg(data []byte) bool {
	return bytes.Compare(data, gARG) == 0
}

func setCapture(capt *capture, rtok cmdlang.TokInfo, ctok tConsumer, rs *rulestate) error {
	rs.inc()
	capt.Settings["!cap"] = rtok
	capt.Type = &rtok
	rs.Process = setName
	return nil
}

func setName(capt *capture, rtok cmdlang.TokInfo, ctok tConsumer, rs *rulestate) error {
	if isGArg(rtok.Literal) {
		if tok, ok := ctok(); ok {
			rs.inc()
			capt.Settings["name"] = *tok
		}
	} else {
		rs.inc()
		capt.Settings["name"] = rtok
	}
	return nil
}

func checkCapture(capt *capture, rtok cmdlang.TokInfo, ctok tConsumer, rs *rulestate) error {
	rs.inc()
	if cc := capt.getSettingToken("!cap"); cc != nil {

		if bytes.Index(rtok.Literal, []byte("|")) > -1 {
			if bytes.Index(rtok.Literal, cc.Literal) == -1 {
				return newTErr(rtok, "Expected context state of [%v] was [%v]", string(rtok.Literal), string(cc.Literal))
			}
		} else {
			if bytes.Compare(cc.Literal, rtok.Literal) != 0 {
				return newTErr(rtok, "Expected context state of [%v] was [%v]", string(rtok.Literal), string(cc.Literal))
			}
		}
	} else {
		return newTErr(rtok, "Invalid context state, expected capture context but value was nil")
	}
	rs.Process = checkCapture

	return nil
}

func makeRunList(funcs ...Processor) Processor {
	pos := 0

	var runner Processor

	runner = func(capt *capture, rtok cmdlang.TokInfo, ctok tConsumer, rs *rulestate) error {
		fn := funcs[pos]

		cctok := func() (*cmdlang.TokInfo, bool) {
			tok, ok := ctok()
			if !ok {
				pos-- // need to re run this one
			}
			return tok, ok
		}

		err := fn(capt, rtok, cctok, rs)
		pos++
		if pos < len(funcs) {
			rs.Process = runner
		}
		return err
	}

	return runner
}

func makeVarCapture(vname string) Processor {
	return func(capt *capture, rtok cmdlang.TokInfo, ctok tConsumer, rs *rulestate) error {
		if isGArg(rtok.Literal) {
			if tok, ok := ctok(); ok {
				capt.setVar(vname, string(tok.Literal))
				rs.inc()
			}
		} else {
			capt.setVar(vname, string(rtok.Literal))
			rs.inc()
		}
		return nil
	}
}

func makeSetNamedByVar(vname string) Processor {
	var setFn Processor

	setFn = func(capt *capture, rtok cmdlang.TokInfo, ctok tConsumer, rs *rulestate) error {
		rs.Process = setFn

		name := capt.getVarStr(vname)

		var lst []interface{}
		var ok bool

		Settings := capt.Settings

		if capt.Type == nil { // we don't have a capture type, we set into the parent
			Settings = capt.Parent.Settings
		}

		curv := Settings[name]
		if curv != nil {
			if lst, ok = curv.([]interface{}); !ok {
				lst = append(lst, curv)
			}
		}

		if isGArg(rtok.Literal) {
			if tok, ok := ctok(); ok {
				if lst != nil {
					lst = append(lst, *tok)
					Settings[name] = lst
				} else {
					Settings[name] = *tok
				}
				rs.inc()
			}
		} else {
			if lst != nil {
				lst = append(lst, rtok)
				Settings[name] = lst
			} else {
				Settings[name] = rtok
			}
			rs.inc()
		}
		return nil
	}

	return setFn
}

func setNameCapture(capt *capture, rtok cmdlang.TokInfo, ctok tConsumer, rs *rulestate) error {
	rs.Process = makeRunList(makeVarCapture("name"), makeSetNamedByVar("name"))
	return nil
}

// evalApplyRule consume the inTok in some way or error, returns  done, consumed, err
func (g *Grammer) evalApplyRule(current *capture, inTok cmdlang.TokInfo, evRule rule) (bool, bool, error) {
	consumed := false
	consumeInput := func() (*cmdlang.TokInfo, bool) {
		if !consumed {
			consumed = true
			return &inTok, true
		}
		return nil, false
	}

	for {
		done, ruleEv, ruleState, _ := restoreRuleState(current, evRule)
		if done {
			return true, consumed, nil
		}

		ruleTok := ruleEv.getTok(ruleState.get())

		if ruleState.Parent == nil && bytes.Compare(inTok.Literal, ruleTok.Literal) == 0 {
			var ok bool

			if _, ok = consumeInput(); !ok {
				return false, consumed, nil // ask for more input
			}

			//current.Type = ruleTok
			ruleState.inc()
			continue
		}

		if ruleState.Process != nil {
			fnProcess := ruleState.Process
			ruleState.Process = nil

			wanted := false
			got := false

			cInput := func() (*cmdlang.TokInfo, bool) {
				wanted = true
				if tok, ok := consumeInput(); ok {
					got = true
					return tok, ok
				}
				return nil, false
			}

			err := fnProcess(current, *ruleTok, cInput, ruleState)

			if err != nil {
				return false, consumed, err
			}

			if wanted && !got {
				ruleState.Process = fnProcess // try again after input
				return false, consumed, nil   // ask for more input
			}

			continue
		}

		// effectivly default processors
		cmd := string(ruleTok.Literal)
		switch cmd {
		case "!cap":
			ruleState.inc()
			ruleState.Process = setCapture
		case "!rcap":
			ruleState.inc()
			ruleState.Process = checkCapture
		case "!set":
			ruleState.inc()
			ruleState.Process = setNameCapture
		default:
			return false, consumed, newTErr(inTok, "Unknown rule token %v", ruleTok)
		}
	}

	return false, consumed, nil
}

func (g *Grammer) evalIdent(current *capture, inTok cmdlang.TokInfo) error {
	for {
		if current.RuleIdx == -1 {
			for i, r := range g.rules {
				if bytes.Compare(r.id.Literal, inTok.Literal) == 0 {
					current.RuleIdx = i
					break
				}
			}
		}

		done, consumed, err := g.evalApplyRule(current, inTok, g.rules[current.RuleIdx])

		if err != nil {
			return err
		}

		if done {
			current.RuleIdx = -1
			if consumed {
				break
			}
			continue
		}

		if consumed {
			break
		}
	}
	return nil
}

func (g *Grammer) Transform(in io.Reader, out io.Writer) error {
	scanner := cmdlang.NewScanner(in)

	var parse func(curcap *capture, depth int) error

	topCapture := newCapture(nil)

	addTop := func(tcap *capture) {
		topCapture.Children = append(topCapture.Children, tcap)
	}

	parse = func(curcap *capture, depth int) error {
		var inTok cmdlang.TokInfo

	PARSE:
		for inTok = scanner.Scan(); inTok.Token != cmdlang.TOK_EOF; inTok = scanner.Scan() {
			switch inTok.Token {
			case cmdlang.TOK_IDENT:
				if err := g.evalIdent(curcap, inTok); err != nil {
					return err
				}
			case cmdlang.TOK_BLOCK_START:
				newCap := newCapture(curcap)
				log.Print("Processing block from ", inTok)
				if err := parse(newCap, depth+1); err != nil {
					return err
				}
				if newCap.Type != nil {
					curcap.Children = append(curcap.Children, newCap)
				}
			case cmdlang.TOK_BLOCK_END:
				break PARSE
			case cmdlang.TOK_EOC:
				if depth == 0 && curcap.Type != nil {
					addTop(curcap)
					curcap = newCapture(nil)
				} // else this is just a blank top level line
				if depth > 0 {
					break PARSE
				}
			}
		}

		if depth == 0 && curcap.Type != nil {
			addTop(curcap)
			curcap = newCapture(nil)
		} // else this is just a blank top level line
		return nil
	}

	if err := parse(newCapture(nil), 0); err != nil {
		return err
	}

	for _, e := range g.emits {
		if err := g.runEmit(e, topCapture, out); err != nil {
			return err
		}
	}

	return nil

}

var tPAD = []byte("!pad")
var tDELIM = []byte("!delim")
var tJCLPS = []byte("!jclps")
var tEMIT = []byte("!emit")
var tGET = []byte("!get")
var tPGET = []byte("!pget")
var tJOIN = []byte("!join")
var tWRAP = []byte("!wrap")
var tIF = []byte("!if")
var tEOL = []byte("!eol")
var vEOL = []byte(fmt.Sprintf("\n"))

type outWriter struct {
	out   io.Writer
	delim []cmdlang.TokInfo
	pad   int
}

func (ow *outWriter) Write(data []byte) (int, error) {
	if ow.delim != nil {
		for _, v := range ow.delim {
			if amt, err := ow.dowrite(v.Literal); err != nil {
				return amt, err
			}
		}
		ow.delim = nil
	}
	return ow.dowrite(data)
}

var dSpace = []byte(" ")

func (ow *outWriter) dowrite(data []byte) (int, error) {
	if bytes.Compare(data, tEOL) == 0 {
		if n, err := ow.out.Write(vEOL); err == nil {
			if ow.pad > 0 {
				np := 0
				for i := 0; i < ow.pad; i++ {
					if x, perr := ow.out.Write(dSpace); perr != nil {
						return n + np + x, perr
					} else {
						np += x
					}
				}
				return n + np, nil
			}
			return n, err
		} else {
			return n, err
		}
	} else {
		return ow.out.Write(data)
	}
}

func (g *Grammer) runEmit(erule emit, current *capture, outw io.Writer) error {
	var dump func(ev *eval, pad string)

	out := &outWriter{outw, nil, 0}

	dump = func(ev *eval, pad string) {
		for i := 0; i < len(ev.items); i++ {
			tok := ev.getTok(i)
			if tok != nil {
				buf := bytes.Buffer{}
				buf.WriteString(pad)
				buf.WriteString(strconv.Itoa(i))
				buf.WriteString(") ")
				buf.Write(tok.Literal)

				log.Print(buf.String())
			}

			child := ev.getEval(i)
			if child != nil {
				log.Print(pad, strconv.Itoa(i), ")")

				dump(child, pad+" ")
			}
		}
	}
	dump(&erule.ev, "")

	logTok := func(v interface{}) string {
		if vv, ok := v.(*cmdlang.TokInfo); ok {
			return string(vv.Literal)
		}
		if vv, ok := v.(cmdlang.TokInfo); ok {
			return string(vv.Literal)
		}
		return fmt.Sprintf("%v", v)
	}

	var dumpData func(data *capture, pad string)
	dumpData = func(data *capture, pad string) {
		if data.Type != nil {
			log.Print(pad, "type: ", string(data.Type.Literal), ":")
		} else {
			log.Print(pad, "Empty type:")
		}
		pad += " "
		for k, v := range data.Settings {
			if lst, ok := v.([]interface{}); ok {
				log.Print(pad, "setting: ", k)

				for i, vv := range lst {
					log.Print(pad+" ", strconv.Itoa(i), ")")
					log.Print(pad+" ", "->", logTok(vv))
				}
			} else {
				log.Print(pad, "setting: ", k, " = ", logTok(v))
			}
		}
		for _, v := range data.Children {
			dumpData(v, pad+"  ")
		}
	}

	dumpData(current, "")

	isEmitTok := func(tok cmdlang.TokInfo) bool {
		if bytes.Compare(tok.Literal, tEMIT) == 0 {
			return true
		}
		return false
	}

	var handleAsCommand func(data *capture, first cmdlang.TokInfo, ev *eval, out *outWriter) (bool, error)

	handleAsCommand = func(data *capture, first cmdlang.TokInfo, ev *eval, out *outWriter) (bool, error) {
		if bytes.Compare(first.Literal, tPAD) == 0 {
			if pad := ev.getTok(1); pad != nil {
				parse := pad.Literal

				pos := false
				neg := false

				if parse[0] == '+' {
					parse = parse[1:]
					pos = true
				} else if parse[0] == '-' {
					parse = parse[1:]
					neg = true
				}

				if v, err := strconv.Atoi(string(parse)); err == nil {
					if pos {
						out.pad += v
					} else if neg {
						out.pad -= v
					} else {
						out.pad = v
					}
				}
			}
			return true, nil
		}
		//var tJCLPS = []byte("!jclps")

		// format !jclps 'setting' <group #> <delim> [formats...] $1 = $X
		if bytes.Compare(first.Literal, tJCLPS) == 0 {
			if len(ev.items) < 4 {
				return true, newTErr(first, "Invalid jclps command expected at least 4 args")
			}
			setting := ev.getTok(1)
			if setting == nil {
				return true, newTErr(first, "Invalid param expecting token")
			}
			var groupno int
			if tok := ev.getTok(2); tok != nil {
				var err error
				if groupno, err = strconv.Atoi(string(tok.Literal)); err != nil {
					return true, newTErr(first, "Failed parsing group size %v", err)
				}
			} else {
				return true, newTErr(first, "Invalid param expecting token")
			}
			delim := ev.getTok(3)
			if delim == nil {
				return true, newTErr(first, "Invalid param expecting token")
			}
			lst := data.getSettingAsList(string(setting.Literal))
			if lst != nil {
				wcnt := 0
				for i := 0; i+(groupno-1) < len(lst); i += groupno {
					wrote := false

					wtt := func(d []byte) {
						if !wrote {
							wrote = true
							if wcnt > 0 {
								out.Write(delim.Literal)
							}
						}
						out.Write(d)
					}

					for q := 4; q < len(ev.items); q++ {
						if tok := ev.getTok(q); tok != nil {
							if tok.Literal[0] == '$' {
								if n, err := strconv.Atoi(string(tok.Literal[1:])); err != nil {
									return true, newTErr(*tok, "Invalid capture index expected number %v", err)
								} else {
									if tok, ok := lst[i+(n-1)].(cmdlang.TokInfo); ok {
										wtt(tok.Literal)
										continue
									}
									if tok, ok := lst[i+(n-1)].(*cmdlang.TokInfo); ok {
										wtt(tok.Literal)
										continue
									}
									wtt([]byte(fmt.Sprintf("%v", lst[i+(n-1)])))
								}
							} else {
								wtt(tok.Literal)
							}
						}
					}

					if wrote {
						wcnt++
					}

				}
			}
			return true, nil
		}

		if bytes.Compare(first.Literal, tDELIM) == 0 {
			for i := 1; i < len(ev.items); i++ {
				if tok := ev.getTok(i); tok != nil {
					out.delim = append(out.delim, *tok)
					continue
				}

				if subev := ev.getEval(i); subev != nil {
					buf := bytes.Buffer{}
					if ok, err := handleAsCommand(data, *subev.getTok(0), subev, &outWriter{&buf, nil, 0}); ok {
						if err != nil {
							return true, err
						}
					} else {
						return true, newTErr(*subev.getTok(0), "Unhandled token type for !delim command")
					}
					out.delim = append(out.delim, cmdlang.TokInfo{Literal: buf.Bytes()})
				}
			}
			return true, nil
		}

		if bytes.Compare(first.Literal, tIF) == 0 {
			if test := ev.getTok(1); test != nil {
				if _, ok := data.Settings[string(test.Literal)]; ok {
					for i := 2; i < len(ev.items); i++ {
						if tok := ev.getTok(i); tok != nil {
							out.Write(tok.Literal)
							continue
						}

						if subev := ev.getEval(i); subev != nil {
							if ok, err := handleAsCommand(data, *subev.getTok(0), subev, out); ok {
								if err != nil {
									return true, err
								}
							} else {
								return true, newTErr(*subev.getTok(0), "Expected sub command in if true clause")
							}
						}
					}
				}
				return true, nil
			} else {
				return true, newTErr(first, "Expected param to be token for var test")
			}
		}

		runGet := func(data *capture, ev *eval, start int, end int, out *outWriter) (bool, error) {
			for i := start; i < end; i++ {
				if tok := ev.getTok(i); tok != nil {
					if stok := data.getSettingAsList(string(tok.Literal)); stok != nil {
						for _, v := range stok {
							if vv, ok := v.(cmdlang.TokInfo); ok {
								out.Write(vv.Literal)
								continue
							}
							if vv, ok := v.(*cmdlang.TokInfo); ok {
								out.Write(vv.Literal)
								continue
							}
							out.Write([]byte(fmt.Sprintf("%v", v)))
						}
						continue
					}
				}

				if subev := ev.getEval(i); subev != nil {
					if ok, err := handleAsCommand(data, *subev.getTok(0), subev, out); ok {
						if err != nil {
							return true, err
						}
					} else {
						return true, newTErr(*subev.getTok(0), "Unhandled token type for !get command")
					}
				}
			}
			return true, nil
		}

		if bytes.Compare(first.Literal, tGET) == 0 {
			return runGet(data, ev, 1, len(ev.items), out)
		}

		if bytes.Compare(first.Literal, tPGET) == 0 {
			return runGet(data.Parent, ev, 1, len(ev.items), out)
		}

		if bytes.Compare(first.Literal, tJOIN) == 0 {
			var delim []byte

			if wtok := ev.getTok(1); wtok != nil {
				delim = wtok.Literal
			}
			if wtoke := ev.getEval(1); wtoke != nil {
				buf := bytes.Buffer{}
				if ok, err := handleAsCommand(data, *wtoke.getTok(0), wtoke, &outWriter{&buf, nil, 0}); ok {
					if err != nil {
						return true, err
					}
					delim = buf.Bytes()
				} else {
					return true, newTErr(*wtoke.getTok(0), "Unable to resolve join delimieter")
				}
			}

			ecnt := 0
			for i := 2; i < len(ev.items); i++ {
				if ecnt > 0 {
					out.Write(delim)
				}

				if tok := ev.getTok(i); tok != nil {
					if stok := data.getSettingToken(string(tok.Literal)); stok != nil {
						out.Write(stok.Literal)
						continue
					}
					out.Write(tok.Literal)
				}

				if subev := ev.getEval(i); subev != nil {
					if ok, err := handleAsCommand(data, *subev.getTok(0), subev, out); ok {
						if err != nil {
							return true, err
						}
					} else {
						return true, newTErr(*subev.getTok(0), "Unhandled token type for !join command")
					}

				}
				ecnt++
			}
			return true, nil
		}

		if bytes.Compare(first.Literal, tWRAP) == 0 {
			var wrap []byte

			if wtok := ev.getTok(1); wtok != nil {
				wrap = wtok.Literal
			}
			if wtoke := ev.getEval(1); wtoke != nil {
				buf := bytes.Buffer{}
				if ok, err := handleAsCommand(data, *wtoke.getTok(0), wtoke, &outWriter{&buf, nil, 0}); ok {
					if err != nil {
						return true, err
					}
					wrap = buf.Bytes()
				} else {
					return true, newTErr(*wtoke.getTok(0), "Unable to resolve wrap value")
				}
			}

			for i := 2; i < len(ev.items); i++ {
				out.Write(wrap)

				if tok := ev.getTok(i); tok != nil {
					out.Write(tok.Literal)
				}

				if toke := ev.getEval(i); toke != nil {
					if ok, err := handleAsCommand(data, *toke.getTok(0), toke, out); ok {
						if err != nil {
							return true, err
						}
					} else {
						return true, newTErr(*toke.getTok(0), "Unable to resolve wrap data")
					}
				}

				out.Write(wrap)
			}
		}
		return false, nil
	}

	var printOut func(data *capture, ev *eval) error

	printOut = func(data *capture, ev *eval) error {
		// 1st token sets the type or is !emit which causes us to iterate on child captures

		first := ev.getTok(0)

		if isEmitTok(*first) {
			// walk params 1 at a time against child captures
			for i := 1; i < len(ev.items); i++ {
				if tok := ev.getTok(i); tok != nil {
					out.Write(tok.Literal)
					continue
				}

				if sube := ev.getEval(i); sube != nil {
					ctx := sube.getTok(0).Literal

					var lastName *cmdlang.TokInfo

					for _, ccap := range data.Children {
						if bytes.Compare(ccap.Type.Literal, ctx) == 0 {

							if lastName != nil {
								ccap.Settings["sibname"] = *lastName
							}

							if err := printOut(ccap, sube); err != nil {
								return err
							}

							delete(ccap.Settings, "sibname")
							if ln, ok := ccap.Settings["name"].(cmdlang.TokInfo); ok {
								lastName = &ln
							}
						}
					}
				}
			}
			out.delim = nil
			return nil
		}

		if ok, err := handleAsCommand(data, *first, ev, out); ok {
			return err
		}

		if bytes.Compare(data.Type.Literal, first.Literal) != 0 {
			return newTErr(*first, "Capture type was [%v] eval had [%v], eval was %v", string(data.Type.Literal), string(first.Literal), ev)
		}

		for i := 1; i < len(ev.items); i++ {
			if tok := ev.getTok(i); tok != nil {
				out.Write(tok.Literal)
				continue
			}

			child := ev.getEval(i)
			if child != nil {
				if err := printOut(data, child); err != nil {
					return err
				}
			}
		}
		return nil
	}

	return printOut(current, &erule.ev)
}
