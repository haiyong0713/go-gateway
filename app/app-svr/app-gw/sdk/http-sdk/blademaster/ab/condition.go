package ab

import (
	"fmt"
	"strconv"
	"strings"

	"go-common/library/log"
	parsing "go-gateway/app/app-svr/app-gw/sdk/http-sdk/blademaster/ab/internal/parsing"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/pkg/errors"
)

var (
	// ErrEnvVarNotExist is an error if env variable does not exist.
	ErrEnvVarNotExist = errors.New("ab: env variable not exist")
)

// UserDefinedPredicate is a predicate which can be tested in condition
type UserDefinedPredicate func(t *T) bool

type Condition interface {
	Matches(t *T) bool
}

type falseCondition struct{}

func (fc *falseCondition) Matches(t *T) bool {
	return false
}

var FALSE = &falseCondition{}

type trueCondition struct{}

func (tc *trueCondition) Matches(t *T) bool {
	return true
}

var TRUE = &trueCondition{}

type andCondition struct {
	children []Condition
}

func (ac *andCondition) add(c Condition) {
	ac.children = append(ac.children, c)
}

func (ac *andCondition) Matches(t *T) bool {
	if len(ac.children) <= 0 {
		return true
	}
	for _, cond := range ac.children {
		if !cond.Matches(t) {
			return false
		}
	}
	return true
}

type orCondition struct {
	children []Condition
}

func (oc *orCondition) add(c Condition) {
	oc.children = append(oc.children, c)
}

func (oc *orCondition) Matches(t *T) bool {
	if len(oc.children) <= 0 {
		return true
	}
	for _, cond := range oc.children {
		if cond.Matches(t) {
			return true
		}
	}
	return false
}

type notCondition struct {
	child Condition
}

func (nc *notCondition) Matches(t *T) bool {
	return !nc.child.Matches(t)
}

type didCondition struct {
	groupID int64
}

func (dc *didCondition) Matches(t *T) bool {
	return t.did(dc.groupID)
}

type udpCondition struct {
	udp UserDefinedPredicate
}

func (uc *udpCondition) Matches(t *T) bool {
	return uc.udp(t)
}

type compareCondition struct {
	left  uint // slice index of env kv
	right KV   // type & value of condition literal
}

func (cc *compareCondition) Left(t *T) *KV {
	return t.env[cc.left]
}

func (cc *compareCondition) Right(t *T) (kv *KV) {
	return &cc.right
}

func newCompareCondition(k string, v string) (cc compareCondition, err error) {
	var (
		e     envVar
		ok    bool
		value KV
	)

	e, ok = parseConditionVar(k)
	if !ok {
		err = fmt.Errorf("ab: fail to parse env variable (%s)", k)
		return
	}

	value, err = parseKV(e.kv, strings.Trim(v, "\""))
	if err != nil {
		return
	}

	cc = compareCondition{
		left:  e.index,
		right: value,
	}
	return
}

type eqCondition struct {
	compareCondition
}

func newEqCondition(k string, val string) Condition {
	cc, err := newCompareCondition(k, val)
	if err != nil {
		return FALSE
	}
	return &eqCondition{cc}
}

func (ec *eqCondition) Matches(t *T) (b bool) {
	left, right := ec.Left(t), ec.Right(t)
	if left == nil && right == nil {
		b = true
	} else if left == nil || right == nil {
		b = false
	} else if left.Type != right.Type {
		b = false
	} else {
		switch left.Type {
		case typeString:
			b = left.String == right.String
		case typeInt64:
			b = left.Int64 == right.Int64
		case typeFloat64:
			b = left.Float64 == right.Float64
		case typeBool:
			b = left.Bool == right.Bool
		case typeVersion:
			b = left.Version.eq(right.Version)
		default:
			b = false
		}
	}
	return
}

type geCondition struct {
	compareCondition
}

func newGeCondition(k string, val string) Condition {
	cc, err := newCompareCondition(k, val)
	if err != nil {
		return FALSE
	}

	return &geCondition{cc}
}

func (gc *geCondition) Matches(t *T) (b bool) {
	left, right := gc.Left(t), gc.Right(t)
	if left == nil && right == nil {
		b = true
	} else if left == nil || right == nil {
		b = false
	} else if left.Type != right.Type {
		b = false
	} else {
		switch left.Type {
		case typeInt64:
			b = left.Int64 >= right.Int64
		case typeFloat64:
			b = left.Float64 >= right.Float64
		case typeVersion:
			b = left.Version.ge(right.Version)
		default:
			b = false
		}
	}
	return
}

type gtCondition struct {
	compareCondition
}

func newGtCondition(k string, val string) Condition {
	cc, err := newCompareCondition(k, val)
	if err != nil {
		return FALSE
	}

	return &gtCondition{cc}
}

func (gc *gtCondition) Matches(t *T) (b bool) {
	left, right := gc.Left(t), gc.Right(t)
	if left == nil && right == nil {
		b = true
	} else if left == nil || right == nil {
		b = false
	} else if left.Type != right.Type {
		b = false
	} else {
		switch left.Type {
		case typeInt64:
			b = left.Int64 > right.Int64
		case typeFloat64:
			b = left.Float64 > right.Float64
		default:
			b = false
		}
	}
	return
}

type leCondition struct {
	notCondition
}

func newLeCondition(k string, val string) Condition {
	return &leCondition{notCondition{newGtCondition(k, val)}}
}

type ltCondition struct {
	notCondition
}

func newLtCondition(k string, val string) Condition {
	return &ltCondition{notCondition{newGeCondition(k, val)}}
}

type neCondition struct {
	notCondition
}

func newNeCondition(k string, val string) Condition {
	return &neCondition{notCondition{newEqCondition(k, val)}}
}

//nolint:deadcode,unused
type inCondition struct {
	orCondition
}

func newInCondition(k string, val ...string) (c Condition) {
	if len(val) <= 0 {
		c = FALSE
	} else {
		o := &orCondition{}
		for _, v := range val {
			o.add(newEqCondition(k, v))
		}
		c = o
	}
	return
}

//nolint:deadcode,unused
type notInCondition struct {
	notCondition
}

func newNotInCondition(k string, val ...string) Condition {
	return &notCondition{newInCondition(k, val...)}
}

func parseConditionVar(k string) (e envVar, ok bool) {
	e, ok = Registry.loadEnv()[k]
	if !ok {
		log.Warn("ab: failed to find variable in condition(%s)", k)
	}
	return
}

func ParseCondition(condStr string) Condition {
	return parseCondition(condStr)
}

func ParseConditionWithError(condStr string) (Condition, error) {
	return parseConditionWithError(condStr)
}

type errorReporter struct {
	antlr.DefaultErrorListener
	condStr       string
	errorMessages []string
}

func (er *errorReporter) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	er.errorMessages = append(er.errorMessages, "line "+strconv.Itoa(line)+":"+strconv.Itoa(column)+" "+msg)
}
func (er *errorReporter) ChainedError() error {
	if len(er.errorMessages) <= 0 {
		return nil
	}
	return errors.Errorf("Failed to parsing condition: `%s`: %s", er.condStr, strings.Join(er.errorMessages, "\n"))
}

func parseCondition(condStr string) Condition {
	cond, err := parseConditionWithError(condStr)
	if err != nil {
		return nil
	}
	return cond
}

func parseConditionWithError(condStr string) (Condition, error) {
	trimStr := strings.TrimSpace(condStr)
	if len(trimStr) <= 0 {
		return TRUE, nil
	}
	lexer := parsing.NewconditionLexer(antlr.NewInputStream(trimStr))
	parser := parsing.NewconditionParser(antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel))
	er := &errorReporter{condStr: condStr}
	lexer.AddErrorListener(er)
	cond := parseCond(parser.Condition())
	if err := er.ChainedError(); err != nil {
		return nil, err
	}
	return cond, nil
}

func parseCond(ic parsing.IConditionContext) (c Condition) {
	c = FALSE
	switch x := ic.(type) {
	case *parsing.LogicalNotContext:
		child := parseCond(ic.GetChild(1).(parsing.IConditionContext))
		c = &notCondition{child}
	case *parsing.LogicalAndContext:
		left := parseCond(ic.GetChild(0).(parsing.IConditionContext))
		right := parseCond(ic.GetChild(2).(parsing.IConditionContext))
		ac := &andCondition{}
		ac.add(left)
		ac.add(right)
		c = ac
	case *parsing.LogicalOrContext:
		left := parseCond(ic.GetChild(0).(parsing.IConditionContext))
		right := parseCond(ic.GetChild(2).(parsing.IConditionContext))
		ac := &orCondition{}
		ac.add(left)
		ac.add(right)
		c = ac
	case *parsing.ParenContext:
		c = parseCond(ic.GetChild(1).(parsing.IConditionContext))
	case *parsing.InOrNotInOp2Context:
		tc := x.InOrNotIn()
		var vals []string
		for i := 3; i < tc.GetChildCount()-1; i = i + 2 {
			vals = append(vals, tc.GetChild(i).(antlr.ParseTree).GetText())
		}
		if tc.GetChild(1).(antlr.ParseTree).GetText() == "in" {
			c = newInCondition(tc.GetChild(0).(antlr.ParseTree).GetText(), vals...)
		} else {
			c = newNotInCondition(tc.GetChild(0).(antlr.ParseTree).GetText(), vals...)
		}
	case *parsing.LogicalOp2Context:
		tc := x.Compare()
		op := tc.GetChild(1)
		opText := op.(antlr.ParseTree).GetText()
		varName := tc.GetChild(0).(antlr.ParseTree).GetText()
		val := tc.GetChild(2).(antlr.ParseTree).GetText()
		switch opText {
		case "==":
			return newEqCondition(varName, val)
		case ">=":
			return newGeCondition(varName, val)
		case "<=":
			return newLeCondition(varName, val)
		case ">":
			return newGtCondition(varName, val)
		case "<":
			return newLtCondition(varName, val)
		case "!=":
			return newNeCondition(varName, val)
		default:
		}
	case *parsing.CommomOpContext:
		text := ic.GetText()
		if strings.HasPrefix(text, "did_") {
			gid, err := strconv.Atoi(strings.Split(text, "_")[1])
			if err != nil {
				log.Error("ab: fail to parse group id(%s)", text)
				return
			}
			c = &didCondition{int64(gid)}
		} else if strings.HasPrefix(text, "USER_") {
			c = &udpCondition{Registry.udp[strings.Split(text, "_")[1]]}
		}
	default:
	}
	return
}
