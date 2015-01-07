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
var gARGS = []byte("!gArgs")

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

func setCapture(capt *capture, rtok cmdlang.TokInfo, ctok tConsumer, rs *rulestate) error {
	rs.inc()
	capt.Settings["!cap"] = rtok
	capt.Type = &rtok
	rs.Process = setName
	return nil
}

func setName(capt *capture, rtok cmdlang.TokInfo, ctok tConsumer, rs *rulestate) error {
	if bytes.Compare(rtok.Literal, gARG) == 0 {
		if tok, ok := ctok(); ok {
			rs.inc()
			capt.Settings["name"] = *tok
		}
	} else if bytes.Compare(rtok.Literal, gARGS) == 0 {
		rs.Process = setName

		// eat everything until the catpure ends
		if tok, ok := ctok(); ok {
			var lst []interface{}

			if v, ok := capt.Settings["name"]; ok {
				if lv, ok := v.([]interface{}); ok {
					lst = lv
				} else {
					lst = append(lst, v)
				}
			}

			lst = append(lst, *tok)
			capt.Settings["name"] = lst
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
		if bytes.Compare(rtok.Literal, gARG) == 0 {
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

		var setTok *cmdlang.TokInfo
		var didwork bool

		if bytes.Compare(rtok.Literal, gARG) == 0 {
			setTok, didwork = ctok()
			goto HANDLE_TOKEN
		}

		if bytes.Compare(rtok.Literal, gARGS) == 0 {
			setTok, didwork = ctok()
			goto HANDLE_TOKEN
		}

		//default case
		didwork = true
		setTok = &rtok

	HANDLE_TOKEN:

		if didwork {
			if lst != nil {
				lst = append(lst, *setTok)
				Settings[name] = lst
			} else {
				Settings[name] = *setTok
			}

			if bytes.Compare(rtok.Literal, gARGS) != 0 {
				rs.inc()
			}
			//} else {
			//	log.Print("in !gArgs capturing more")
			//}
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

		if current.RuleIdx == -1 { // no rule
			return newTErr(inTok, "No rule found")
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
				// log.Print("Processing block from ", inTok)
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

var dSpace = []byte(" ")

type EmitHandler func(data *capture, first cmdlang.TokInfo, ev *eval, out *outWriter) (bool, error)

type EmitDirective struct {
	Tag     []byte
	Handler EmitHandler
}

var _emitDirectives []EmitDirective

func getDirectiveTable() []EmitDirective {
	if _emitDirectives == nil {
		_emitDirectives = []EmitDirective{EmitDirective{[]byte("!pad"), handlePad},
			EmitDirective{[]byte("!delim"), handleDelim},
			EmitDirective{[]byte("!jclps"), handleJoinColapse},
			EmitDirective{[]byte("!emit"), handleEmit},
			EmitDirective{[]byte("!get"), handleGet},
			EmitDirective{[]byte("!pget"), handlePGet},
			EmitDirective{[]byte("!join"), handleJoin},
			EmitDirective{[]byte("!wrap"), handleWrap},
			EmitDirective{[]byte("!if"), handleIf},
			EmitDirective{[]byte("!ifn"), handleIfNot},
			EmitDirective{[]byte("!len"), handleLen},
		}
	}
	return _emitDirectives
}

func dumpEval(ev *eval, pad string) {
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
			dumpEval(child, pad+" ")
		}
	}
}

func logTok(v interface{}) string {
	if vv, ok := v.(*cmdlang.TokInfo); ok {
		return string(vv.Literal)
	}
	if vv, ok := v.(cmdlang.TokInfo); ok {
		return string(vv.Literal)
	}
	return fmt.Sprintf("%v", v)
}

func dumpData(data *capture, pad string) {
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

func dispatchEval(data *capture, first cmdlang.TokInfo, ev *eval, out *outWriter) (bool, error) {
	if first.Token != cmdlang.TOK_IDENT {
		return true, nil
	}

	for _, v := range getDirectiveTable() {
		if bytes.Compare(v.Tag, first.Literal) == 0 {
			//log.Printf("Calling %v", string(v.Tag))
			return v.Handler(data, first, ev, out)
		}
	}

	return false, newTErr(first, "Un handled command %v", string(first.Literal))
}

type wResult struct {
	Items [][]byte
}

func (wr *wResult) Reset() {
	wr.Items = nil
}

func (wr *wResult) Write(d []byte) {
	wr.Items = append(wr.Items, d)
}

func (wr *wResult) String() string {
	buf := bytes.Buffer{}
	for _, v := range wr.Items {
		buf.Write(v)
	}
	return buf.String()
}

func (wr *wResult) Bytes() []byte {
	buf := bytes.Buffer{}
	for _, v := range wr.Items {
		buf.Write(v)
	}
	return buf.Bytes()
}

func getEvalItem(data *capture, first cmdlang.TokInfo, ev *eval, pos int, result *wResult) error {
	if len(ev.items) <= pos {
		return newTErr(first, "Expected param {0} but wasn't found", pos)
	}

	if tok := ev.getTok(pos); tok != nil {
		result.Write(tok.Literal)
		return nil
	}

	if subev := ev.getEval(pos); subev != nil {
		buf := bytes.Buffer{}
		if ok, err := dispatchEval(data, *subev.getTok(0), subev, &outWriter{&buf, nil, 0}); ok {
			result.Write(buf.Bytes())
		} else {
			return err
		}
	}
	return nil
}

func handleDelim(data *capture, first cmdlang.TokInfo, ev *eval, out *outWriter) (bool, error) {
	for i := 1; i < len(ev.items); i++ {
		buf := wResult{}

		if err := getEvalItem(data, first, ev, i, &buf); err != nil {
			return true, err
		}

		for _, v := range buf.Items {
			//log.Println("Setting delim.. appending ", string(v))
			out.delim = append(out.delim, cmdlang.TokInfo{Literal: v})
		}
	}

	return true, nil
}

func handleJoinColapse(data *capture, first cmdlang.TokInfo, ev *eval, out *outWriter) (bool, error) {
	var err error

	buf := wResult{}

	if err = getEvalItem(data, first, ev, 1, &buf); err != nil {
		return true, err
	}

	setting := buf.String()

	buf.Reset()

	if err = getEvalItem(data, first, ev, 2, &buf); err != nil {
		return true, err
	}

	var groupno int

	if groupno, err = strconv.Atoi(buf.String()); err != nil {
		return true, err
	}

	buf.Reset()

	if err = getEvalItem(data, first, ev, 3, &buf); err != nil {
		return true, err
	}

	delim := buf.Bytes()

	buf = wResult{}

	buf.Reset()

	var lst []interface{}

	if lst = data.getSettingAsList(setting); lst == nil { // no data to operate on
		return true, nil
	}

	grouping := 0
	wrote := false

	writeItem := func(d []byte) {
		if !wrote {
			wrote = true
			if grouping > 0 {
				out.Write(delim)
			}
		}
		out.Write(d)
	}

	for i := 0; i+(groupno-1) < len(lst); i += groupno {
		if wrote {
			grouping++
			wrote = false
		}

		for q := 4; q < len(ev.items); q++ {
			buf.Reset()
			if err = getEvalItem(data, first, ev, q, &buf); err != nil {
				return true, err
			}

			tok := buf.Bytes()

			if tok[0] != '$' {
				writeItem(tok)
				continue
			}

			var n int
			if n, err = strconv.Atoi(string(tok[1:])); err != nil {
				return true, err
			}

			// have an element parameter
			switch tok := lst[i+(n-1)].(type) {
			case cmdlang.TokInfo:
				writeItem(tok.Literal)
			case *cmdlang.TokInfo:
				writeItem(tok.Literal)
			default:
				writeItem([]byte(fmt.Sprintf("%v", lst[i+(n-1)])))
			}
		}
	}
	return true, nil
}

func handleEmit(data *capture, first cmdlang.TokInfo, ev *eval, out *outWriter) (bool, error) {
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

					// walk params object emiting literals and interping sub commands
					for i := 1; i < len(sube.items); i++ {
						if subtok := sube.getTok(i); subtok != nil {
							out.Write(subtok.Literal)
							continue
						}

						if ssev := sube.getEval(i); ssev != nil {
							if ok, err := dispatchEval(ccap, *ssev.getTok(0), ssev, out); !ok || err != nil {
								return ok, err
							}
						}
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
	return true, nil
}

func handleGet(data *capture, first cmdlang.TokInfo, ev *eval, out *outWriter) (bool, error) {
	buf := wResult{}

	for i := 1; i < len(ev.items); i++ {
		buf.Reset()

		if err := getEvalItem(data, first, ev, i, &buf); err != nil {
			return true, err
		}

		setting := buf.String()

		var lst []interface{}
		if lst = data.getSettingAsList(setting); lst == nil {
			continue
		}

		for _, v := range lst {
			switch tv := v.(type) {
			case cmdlang.TokInfo:
				out.Write(tv.Literal)
			case *cmdlang.TokInfo:
				out.Write(tv.Literal)
			default:
				out.Write([]byte(fmt.Sprintf("%v", v)))
			}
		}

	}

	return true, nil
}

func handlePGet(data *capture, first cmdlang.TokInfo, ev *eval, out *outWriter) (bool, error) {
	if data.Parent == nil {
		return true, newTErr(first, "No Parent for context")
	}

	return handleGet(data.Parent, first, ev, out)
}

func handleJoin(data *capture, first cmdlang.TokInfo, ev *eval, out *outWriter) (bool, error) {
	buf := wResult{}

	if err := getEvalItem(data, first, ev, 1, &buf); err != nil {
		return true, err
	}

	delim := buf.Bytes()

	buf.Reset()

	emitCount := 0

	for i := 2; i < len(ev.items); i++ {
		if emitCount > 0 {
			out.Write(delim)
		}

		buf.Reset()

		if err := getEvalItem(data, first, ev, i, &buf); err != nil {
			return true, err
		}

		out.Write(buf.Bytes())

		emitCount++
	}
	return true, nil
}

func handleWrap(data *capture, first cmdlang.TokInfo, ev *eval, out *outWriter) (bool, error) {
	buf := wResult{}

	if err := getEvalItem(data, first, ev, 1, &buf); err != nil {
		return true, err
	}

	wrap := buf.Bytes()

	buf.Reset()

	out.Write(wrap)

	for i := 2; i < len(ev.items); i++ {
		buf.Reset()
		if err := getEvalItem(data, first, ev, i, &buf); err != nil {
			return true, err
		}
		out.Write(buf.Bytes())
	}

	out.Write(wrap)

	return true, nil
}

func handleIfNot(data *capture, first cmdlang.TokInfo, ev *eval, out *outWriter) (bool, error) {
	buf := wResult{}

	if err := getEvalItem(data, first, ev, 1, &buf); err != nil {
		return true, err
	}

	setting := buf.String()

	if _, ok := data.Settings[setting]; !ok {
		return false, nil // eval false
	}

	for i := 2; i < len(ev.items); i++ {
		buf.Reset()
		if err := getEvalItem(data, first, ev, i, &buf); err != nil {
			return true, err
		}
		for _, v := range buf.Items {
			out.Write(v)
		}
	}
	return true, nil
}

func handleIf(data *capture, first cmdlang.TokInfo, ev *eval, out *outWriter) (bool, error) {
	buf := wResult{}

	if err := getEvalItem(data, first, ev, 1, &buf); err != nil {
		return true, err
	}

	setting := buf.String()

	if _, ok := data.Settings[setting]; !ok {
		return true, nil // eval false
	}

	for i := 2; i < len(ev.items); i++ {
		buf.Reset()
		if err := getEvalItem(data, first, ev, i, &buf); err != nil {
			return true, err
		}
		for _, v := range buf.Items {
			out.Write(v)
		}
	}
	return true, nil
}

var chZERO = []byte("0")

func handleLen(data *capture, first cmdlang.TokInfo, ev *eval, out *outWriter) (bool, error) {
	buf := wResult{}

	if err := getEvalItem(data, first, ev, 1, &buf); err != nil {
		return true, err
	}

	setting := buf.String()

	if lst := data.getSettingAsList(setting); lst == nil {
		out.Write(chZERO)
		return true, nil
	} else {
		out.Write([]byte(strconv.Itoa(len(lst))))
	}

	return true, nil
}

func handlePad(data *capture, first cmdlang.TokInfo, ev *eval, out *outWriter) (bool, error) {
	buf := wResult{}

	if err := getEvalItem(data, first, ev, 1, &buf); err != nil {
		return true, err
	}

	val := buf.Bytes()

	switch val[0] {
	case '+':
		if v, err := strconv.Atoi(string(val[1:])); err != nil {
			return true, err
		} else {
			out.pad += v
		}
	case '-':
		if v, err := strconv.Atoi(string(val[1:])); err != nil {
			return true, err
		} else {
			out.pad -= v
		}
	default:
		if v, err := strconv.Atoi(string(val)); err != nil {
			return true, err
		} else {
			out.pad = v
		}
	}

	return true, nil
}

func (g *Grammer) runEmit(erule emit, current *capture, outw io.Writer) error {
	out := &outWriter{outw, nil, 0}

	//dumpData(current, "")

	first := erule.ev.getTok(0)

	routed, err := dispatchEval(current, *first, &erule.ev, out)
	if err != nil {
		return err
	}
	if !routed {
		return newTErr(*first, "Unable to handle command '%v'", string(first.Literal))
	}
	return nil

	/*
		isEmitTok := func(tok cmdlang.TokInfo) bool {
			if bytes.Compare(tok.Literal, tEMIT) == 0 {
				return true
			}
			return false
		}

				var printOut func(data *capture, ev *eval) error

				printOut = func(data *capture, ev *eval) error {
					// 1st token sets the type or is !emit which causes us to iterate on child captures

					first := ev.getTok(0)

					if isEmitTok(*first) {
						// walk params 1 at a time against child captures
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
	*/

}
