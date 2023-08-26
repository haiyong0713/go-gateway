// Code generated from condition.g4 by ANTLR 4.8. DO NOT EDIT.

package parser // condition

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = reflect.Copy
var _ = strconv.Itoa

var parserATN = []uint16{
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 3, 24, 63, 4,
	2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 3, 2, 3, 2, 3,
	2, 3, 2, 3, 2, 3, 2, 3, 2, 3, 2, 3, 2, 3, 2, 5, 2, 23, 10, 2, 3, 2, 3,
	2, 3, 2, 3, 2, 3, 2, 3, 2, 7, 2, 31, 10, 2, 12, 2, 14, 2, 34, 11, 2, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 3, 4, 7, 4, 46, 10,
	4, 12, 4, 14, 4, 49, 11, 4, 3, 4, 3, 4, 3, 5, 3, 5, 3, 5, 3, 5, 3, 5, 3,
	5, 5, 5, 59, 10, 5, 3, 6, 3, 6, 3, 6, 2, 3, 2, 7, 2, 4, 6, 8, 10, 2, 4,
	3, 2, 12, 13, 3, 2, 3, 8, 2, 69, 2, 22, 3, 2, 2, 2, 4, 35, 3, 2, 2, 2,
	6, 39, 3, 2, 2, 2, 8, 58, 3, 2, 2, 2, 10, 60, 3, 2, 2, 2, 12, 13, 8, 2,
	1, 2, 13, 14, 7, 11, 2, 2, 14, 23, 5, 2, 2, 9, 15, 16, 7, 14, 2, 2, 16,
	17, 5, 2, 2, 2, 17, 18, 7, 15, 2, 2, 18, 23, 3, 2, 2, 2, 19, 23, 5, 4,
	3, 2, 20, 23, 5, 6, 4, 2, 21, 23, 5, 8, 5, 2, 22, 12, 3, 2, 2, 2, 22, 15,
	3, 2, 2, 2, 22, 19, 3, 2, 2, 2, 22, 20, 3, 2, 2, 2, 22, 21, 3, 2, 2, 2,
	23, 32, 3, 2, 2, 2, 24, 25, 12, 8, 2, 2, 25, 26, 7, 10, 2, 2, 26, 31, 5,
	2, 2, 9, 27, 28, 12, 7, 2, 2, 28, 29, 7, 9, 2, 2, 29, 31, 5, 2, 2, 8, 30,
	24, 3, 2, 2, 2, 30, 27, 3, 2, 2, 2, 31, 34, 3, 2, 2, 2, 32, 30, 3, 2, 2,
	2, 32, 33, 3, 2, 2, 2, 33, 3, 3, 2, 2, 2, 34, 32, 3, 2, 2, 2, 35, 36, 5,
	8, 5, 2, 36, 37, 5, 10, 6, 2, 37, 38, 5, 8, 5, 2, 38, 5, 3, 2, 2, 2, 39,
	40, 5, 8, 5, 2, 40, 41, 9, 2, 2, 2, 41, 42, 7, 14, 2, 2, 42, 47, 5, 8,
	5, 2, 43, 44, 7, 17, 2, 2, 44, 46, 5, 8, 5, 2, 45, 43, 3, 2, 2, 2, 46,
	49, 3, 2, 2, 2, 47, 45, 3, 2, 2, 2, 47, 48, 3, 2, 2, 2, 48, 50, 3, 2, 2,
	2, 49, 47, 3, 2, 2, 2, 50, 51, 7, 15, 2, 2, 51, 7, 3, 2, 2, 2, 52, 59,
	7, 18, 2, 2, 53, 59, 7, 19, 2, 2, 54, 59, 7, 20, 2, 2, 55, 59, 7, 21, 2,
	2, 56, 59, 7, 22, 2, 2, 57, 59, 7, 23, 2, 2, 58, 52, 3, 2, 2, 2, 58, 53,
	3, 2, 2, 2, 58, 54, 3, 2, 2, 2, 58, 55, 3, 2, 2, 2, 58, 56, 3, 2, 2, 2,
	58, 57, 3, 2, 2, 2, 59, 9, 3, 2, 2, 2, 60, 61, 9, 3, 2, 2, 61, 11, 3, 2,
	2, 2, 7, 22, 30, 32, 47, 58,
}
var deserializer = antlr.NewATNDeserializer(nil)
var deserializedATN = deserializer.DeserializeFromUInt16(parserATN)

var literalNames = []string{
	"", "'=='", "'!='", "'>='", "'<='", "'>'", "'<'", "'||'", "'&&'", "'!'",
	"'in'", "'not in'", "'('", "')'", "'\"'", "','", "'true'", "'false'",
}
var symbolicNames = []string{
	"", "EQ", "NE", "GE", "LE", "GT", "LT", "OR", "AND", "NOT", "IN", "NIN",
	"LPAREN", "RPAREN", "QUOTE", "COMMA", "TRUE", "FALSE", "INT", "DOUBLE",
	"STRING", "IDENTIFIER", "WS",
}

var ruleNames = []string{
	"condition", "compare", "inOrNotIn", "variableDeclarator", "op",
}
var decisionToDFA = make([]*antlr.DFA, len(deserializedATN.DecisionToState))

func init() {
	for index, ds := range deserializedATN.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(ds, index)
	}
}

type conditionParser struct {
	*antlr.BaseParser
}

func NewconditionParser(input antlr.TokenStream) *conditionParser {
	this := new(conditionParser)

	this.BaseParser = antlr.NewBaseParser(input)

	this.Interpreter = antlr.NewParserATNSimulator(this, deserializedATN, decisionToDFA, antlr.NewPredictionContextCache())
	this.RuleNames = ruleNames
	this.LiteralNames = literalNames
	this.SymbolicNames = symbolicNames
	this.GrammarFileName = "condition.g4"

	return this
}

// conditionParser tokens.
const (
	conditionParserEOF        = antlr.TokenEOF
	conditionParserEQ         = 1
	conditionParserNE         = 2
	conditionParserGE         = 3
	conditionParserLE         = 4
	conditionParserGT         = 5
	conditionParserLT         = 6
	conditionParserOR         = 7
	conditionParserAND        = 8
	conditionParserNOT        = 9
	conditionParserIN         = 10
	conditionParserNIN        = 11
	conditionParserLPAREN     = 12
	conditionParserRPAREN     = 13
	conditionParserQUOTE      = 14
	conditionParserCOMMA      = 15
	conditionParserTRUE       = 16
	conditionParserFALSE      = 17
	conditionParserINT        = 18
	conditionParserDOUBLE     = 19
	conditionParserSTRING     = 20
	conditionParserIDENTIFIER = 21
	conditionParserWS         = 22
)

// conditionParser rules.
const (
	conditionParserRULE_condition          = 0
	conditionParserRULE_compare            = 1
	conditionParserRULE_inOrNotIn          = 2
	conditionParserRULE_variableDeclarator = 3
	conditionParserRULE_op                 = 4
)

// IConditionContext is an interface to support dynamic dispatch.
type IConditionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsConditionContext differentiates from other interfaces.
	IsConditionContext()
}

type ConditionContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyConditionContext() *ConditionContext {
	var p = new(ConditionContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = conditionParserRULE_condition
	return p
}

func (*ConditionContext) IsConditionContext() {}

func NewConditionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ConditionContext {
	var p = new(ConditionContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = conditionParserRULE_condition

	return p
}

func (s *ConditionContext) GetParser() antlr.Parser { return s.parser }

func (s *ConditionContext) CopyFrom(ctx *ConditionContext) {
	s.BaseParserRuleContext.CopyFrom(ctx.BaseParserRuleContext)
}

func (s *ConditionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConditionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type LogicalOp2Context struct {
	*ConditionContext
}

func NewLogicalOp2Context(parser antlr.Parser, ctx antlr.ParserRuleContext) *LogicalOp2Context {
	var p = new(LogicalOp2Context)

	p.ConditionContext = NewEmptyConditionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ConditionContext))

	return p
}

func (s *LogicalOp2Context) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LogicalOp2Context) Compare() ICompareContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*ICompareContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(ICompareContext)
}

func (s *LogicalOp2Context) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.EnterLogicalOp2(s)
	}
}

func (s *LogicalOp2Context) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.ExitLogicalOp2(s)
	}
}

type LogicalNotContext struct {
	*ConditionContext
}

func NewLogicalNotContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LogicalNotContext {
	var p = new(LogicalNotContext)

	p.ConditionContext = NewEmptyConditionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ConditionContext))

	return p
}

func (s *LogicalNotContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LogicalNotContext) NOT() antlr.TerminalNode {
	return s.GetToken(conditionParserNOT, 0)
}

func (s *LogicalNotContext) Condition() IConditionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IConditionContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IConditionContext)
}

func (s *LogicalNotContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.EnterLogicalNot(s)
	}
}

func (s *LogicalNotContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.ExitLogicalNot(s)
	}
}

type ParenContext struct {
	*ConditionContext
}

func NewParenContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ParenContext {
	var p = new(ParenContext)

	p.ConditionContext = NewEmptyConditionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ConditionContext))

	return p
}

func (s *ParenContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ParenContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(conditionParserLPAREN, 0)
}

func (s *ParenContext) Condition() IConditionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IConditionContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IConditionContext)
}

func (s *ParenContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(conditionParserRPAREN, 0)
}

func (s *ParenContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.EnterParen(s)
	}
}

func (s *ParenContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.ExitParen(s)
	}
}

type CommomOpContext struct {
	*ConditionContext
}

func NewCommomOpContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *CommomOpContext {
	var p = new(CommomOpContext)

	p.ConditionContext = NewEmptyConditionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ConditionContext))

	return p
}

func (s *CommomOpContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CommomOpContext) VariableDeclarator() IVariableDeclaratorContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IVariableDeclaratorContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IVariableDeclaratorContext)
}

func (s *CommomOpContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.EnterCommomOp(s)
	}
}

func (s *CommomOpContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.ExitCommomOp(s)
	}
}

type LogicalAndContext struct {
	*ConditionContext
}

func NewLogicalAndContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LogicalAndContext {
	var p = new(LogicalAndContext)

	p.ConditionContext = NewEmptyConditionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ConditionContext))

	return p
}

func (s *LogicalAndContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LogicalAndContext) AllCondition() []IConditionContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IConditionContext)(nil)).Elem())
	var tst = make([]IConditionContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IConditionContext)
		}
	}

	return tst
}

func (s *LogicalAndContext) Condition(i int) IConditionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IConditionContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IConditionContext)
}

func (s *LogicalAndContext) AND() antlr.TerminalNode {
	return s.GetToken(conditionParserAND, 0)
}

func (s *LogicalAndContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.EnterLogicalAnd(s)
	}
}

func (s *LogicalAndContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.ExitLogicalAnd(s)
	}
}

type LogicalOrContext struct {
	*ConditionContext
}

func NewLogicalOrContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LogicalOrContext {
	var p = new(LogicalOrContext)

	p.ConditionContext = NewEmptyConditionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ConditionContext))

	return p
}

func (s *LogicalOrContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LogicalOrContext) AllCondition() []IConditionContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IConditionContext)(nil)).Elem())
	var tst = make([]IConditionContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IConditionContext)
		}
	}

	return tst
}

func (s *LogicalOrContext) Condition(i int) IConditionContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IConditionContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IConditionContext)
}

func (s *LogicalOrContext) OR() antlr.TerminalNode {
	return s.GetToken(conditionParserOR, 0)
}

func (s *LogicalOrContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.EnterLogicalOr(s)
	}
}

func (s *LogicalOrContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.ExitLogicalOr(s)
	}
}

type InOrNotInOp2Context struct {
	*ConditionContext
}

func NewInOrNotInOp2Context(parser antlr.Parser, ctx antlr.ParserRuleContext) *InOrNotInOp2Context {
	var p = new(InOrNotInOp2Context)

	p.ConditionContext = NewEmptyConditionContext()
	p.parser = parser
	p.CopyFrom(ctx.(*ConditionContext))

	return p
}

func (s *InOrNotInOp2Context) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InOrNotInOp2Context) InOrNotIn() IInOrNotInContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IInOrNotInContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IInOrNotInContext)
}

func (s *InOrNotInOp2Context) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.EnterInOrNotInOp2(s)
	}
}

func (s *InOrNotInOp2Context) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.ExitInOrNotInOp2(s)
	}
}

func (p *conditionParser) Condition() (localctx IConditionContext) {
	return p.condition(0)
}

func (p *conditionParser) condition(_p int) (localctx IConditionContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()
	_parentState := p.GetState()
	localctx = NewConditionContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IConditionContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 0
	p.EnterRecursionRule(localctx, 0, conditionParserRULE_condition, _p)

	defer func() {
		p.UnrollRecursionContexts(_parentctx)
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(20)
	p.GetErrorHandler().Sync(p)
	switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 0, p.GetParserRuleContext()) {
	case 1:
		localctx = NewLogicalNotContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(11)
			p.Match(conditionParserNOT)
		}
		{
			p.SetState(12)
			p.condition(7)
		}

	case 2:
		localctx = NewParenContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(13)
			p.Match(conditionParserLPAREN)
		}
		{
			p.SetState(14)
			p.condition(0)
		}
		{
			p.SetState(15)
			p.Match(conditionParserRPAREN)
		}

	case 3:
		localctx = NewLogicalOp2Context(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(17)
			p.Compare()
		}

	case 4:
		localctx = NewInOrNotInOp2Context(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(18)
			p.InOrNotIn()
		}

	case 5:
		localctx = NewCommomOpContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(19)
			p.VariableDeclarator()
		}

	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(30)
	p.GetErrorHandler().Sync(p)
	_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 2, p.GetParserRuleContext())

	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(28)
			p.GetErrorHandler().Sync(p)
			switch p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 1, p.GetParserRuleContext()) {
			case 1:
				localctx = NewLogicalAndContext(p, NewConditionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, conditionParserRULE_condition)
				p.SetState(22)

				if !(p.Precpred(p.GetParserRuleContext(), 6)) {
					panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 6)", ""))
				}
				{
					p.SetState(23)
					p.Match(conditionParserAND)
				}
				{
					p.SetState(24)
					p.condition(7)
				}

			case 2:
				localctx = NewLogicalOrContext(p, NewConditionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, conditionParserRULE_condition)
				p.SetState(25)

				if !(p.Precpred(p.GetParserRuleContext(), 5)) {
					panic(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 5)", ""))
				}
				{
					p.SetState(26)
					p.Match(conditionParserOR)
				}
				{
					p.SetState(27)
					p.condition(6)
				}

			}

		}
		p.SetState(32)
		p.GetErrorHandler().Sync(p)
		_alt = p.GetInterpreter().AdaptivePredict(p.GetTokenStream(), 2, p.GetParserRuleContext())
	}

	return localctx
}

// ICompareContext is an interface to support dynamic dispatch.
type ICompareContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsCompareContext differentiates from other interfaces.
	IsCompareContext()
}

type CompareContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCompareContext() *CompareContext {
	var p = new(CompareContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = conditionParserRULE_compare
	return p
}

func (*CompareContext) IsCompareContext() {}

func NewCompareContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CompareContext {
	var p = new(CompareContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = conditionParserRULE_compare

	return p
}

func (s *CompareContext) GetParser() antlr.Parser { return s.parser }

func (s *CompareContext) CopyFrom(ctx *CompareContext) {
	s.BaseParserRuleContext.CopyFrom(ctx.BaseParserRuleContext)
}

func (s *CompareContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CompareContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type LogicalOpContext struct {
	*CompareContext
}

func NewLogicalOpContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LogicalOpContext {
	var p = new(LogicalOpContext)

	p.CompareContext = NewEmptyCompareContext()
	p.parser = parser
	p.CopyFrom(ctx.(*CompareContext))

	return p
}

func (s *LogicalOpContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LogicalOpContext) AllVariableDeclarator() []IVariableDeclaratorContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IVariableDeclaratorContext)(nil)).Elem())
	var tst = make([]IVariableDeclaratorContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IVariableDeclaratorContext)
		}
	}

	return tst
}

func (s *LogicalOpContext) VariableDeclarator(i int) IVariableDeclaratorContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IVariableDeclaratorContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IVariableDeclaratorContext)
}

func (s *LogicalOpContext) Op() IOpContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IOpContext)(nil)).Elem(), 0)

	if t == nil {
		return nil
	}

	return t.(IOpContext)
}

func (s *LogicalOpContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.EnterLogicalOp(s)
	}
}

func (s *LogicalOpContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.ExitLogicalOp(s)
	}
}

func (p *conditionParser) Compare() (localctx ICompareContext) {
	localctx = NewCompareContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, conditionParserRULE_compare)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	localctx = NewLogicalOpContext(p, localctx)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(33)
		p.VariableDeclarator()
	}
	{
		p.SetState(34)
		p.Op()
	}
	{
		p.SetState(35)
		p.VariableDeclarator()
	}

	return localctx
}

// IInOrNotInContext is an interface to support dynamic dispatch.
type IInOrNotInContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsInOrNotInContext differentiates from other interfaces.
	IsInOrNotInContext()
}

type InOrNotInContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyInOrNotInContext() *InOrNotInContext {
	var p = new(InOrNotInContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = conditionParserRULE_inOrNotIn
	return p
}

func (*InOrNotInContext) IsInOrNotInContext() {}

func NewInOrNotInContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *InOrNotInContext {
	var p = new(InOrNotInContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = conditionParserRULE_inOrNotIn

	return p
}

func (s *InOrNotInContext) GetParser() antlr.Parser { return s.parser }

func (s *InOrNotInContext) CopyFrom(ctx *InOrNotInContext) {
	s.BaseParserRuleContext.CopyFrom(ctx.BaseParserRuleContext)
}

func (s *InOrNotInContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InOrNotInContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type InOrNotInOpContext struct {
	*InOrNotInContext
}

func NewInOrNotInOpContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *InOrNotInOpContext {
	var p = new(InOrNotInOpContext)

	p.InOrNotInContext = NewEmptyInOrNotInContext()
	p.parser = parser
	p.CopyFrom(ctx.(*InOrNotInContext))

	return p
}

func (s *InOrNotInOpContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *InOrNotInOpContext) AllVariableDeclarator() []IVariableDeclaratorContext {
	var ts = s.GetTypedRuleContexts(reflect.TypeOf((*IVariableDeclaratorContext)(nil)).Elem())
	var tst = make([]IVariableDeclaratorContext, len(ts))

	for i, t := range ts {
		if t != nil {
			tst[i] = t.(IVariableDeclaratorContext)
		}
	}

	return tst
}

func (s *InOrNotInOpContext) VariableDeclarator(i int) IVariableDeclaratorContext {
	var t = s.GetTypedRuleContext(reflect.TypeOf((*IVariableDeclaratorContext)(nil)).Elem(), i)

	if t == nil {
		return nil
	}

	return t.(IVariableDeclaratorContext)
}

func (s *InOrNotInOpContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(conditionParserLPAREN, 0)
}

func (s *InOrNotInOpContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(conditionParserRPAREN, 0)
}

func (s *InOrNotInOpContext) IN() antlr.TerminalNode {
	return s.GetToken(conditionParserIN, 0)
}

func (s *InOrNotInOpContext) NIN() antlr.TerminalNode {
	return s.GetToken(conditionParserNIN, 0)
}

func (s *InOrNotInOpContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(conditionParserCOMMA)
}

func (s *InOrNotInOpContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(conditionParserCOMMA, i)
}

func (s *InOrNotInOpContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.EnterInOrNotInOp(s)
	}
}

func (s *InOrNotInOpContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.ExitInOrNotInOp(s)
	}
}

func (p *conditionParser) InOrNotIn() (localctx IInOrNotInContext) {
	localctx = NewInOrNotInContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, conditionParserRULE_inOrNotIn)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	localctx = NewInOrNotInOpContext(p, localctx)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(37)
		p.VariableDeclarator()
	}
	{
		p.SetState(38)
		_la = p.GetTokenStream().LA(1)

		if !(_la == conditionParserIN || _la == conditionParserNIN) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}
	{
		p.SetState(39)
		p.Match(conditionParserLPAREN)
	}
	{
		p.SetState(40)
		p.VariableDeclarator()
	}
	p.SetState(45)
	p.GetErrorHandler().Sync(p)
	_la = p.GetTokenStream().LA(1)

	for _la == conditionParserCOMMA {
		{
			p.SetState(41)
			p.Match(conditionParserCOMMA)
		}
		{
			p.SetState(42)
			p.VariableDeclarator()
		}

		p.SetState(47)
		p.GetErrorHandler().Sync(p)
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(48)
		p.Match(conditionParserRPAREN)
	}

	return localctx
}

// IVariableDeclaratorContext is an interface to support dynamic dispatch.
type IVariableDeclaratorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsVariableDeclaratorContext differentiates from other interfaces.
	IsVariableDeclaratorContext()
}

type VariableDeclaratorContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyVariableDeclaratorContext() *VariableDeclaratorContext {
	var p = new(VariableDeclaratorContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = conditionParserRULE_variableDeclarator
	return p
}

func (*VariableDeclaratorContext) IsVariableDeclaratorContext() {}

func NewVariableDeclaratorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *VariableDeclaratorContext {
	var p = new(VariableDeclaratorContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = conditionParserRULE_variableDeclarator

	return p
}

func (s *VariableDeclaratorContext) GetParser() antlr.Parser { return s.parser }

func (s *VariableDeclaratorContext) CopyFrom(ctx *VariableDeclaratorContext) {
	s.BaseParserRuleContext.CopyFrom(ctx.BaseParserRuleContext)
}

func (s *VariableDeclaratorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VariableDeclaratorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type CommonStringContext struct {
	*VariableDeclaratorContext
}

func NewCommonStringContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *CommonStringContext {
	var p = new(CommonStringContext)

	p.VariableDeclaratorContext = NewEmptyVariableDeclaratorContext()
	p.parser = parser
	p.CopyFrom(ctx.(*VariableDeclaratorContext))

	return p
}

func (s *CommonStringContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CommonStringContext) STRING() antlr.TerminalNode {
	return s.GetToken(conditionParserSTRING, 0)
}

func (s *CommonStringContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.EnterCommonString(s)
	}
}

func (s *CommonStringContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.ExitCommonString(s)
	}
}

type CommonIntContext struct {
	*VariableDeclaratorContext
}

func NewCommonIntContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *CommonIntContext {
	var p = new(CommonIntContext)

	p.VariableDeclaratorContext = NewEmptyVariableDeclaratorContext()
	p.parser = parser
	p.CopyFrom(ctx.(*VariableDeclaratorContext))

	return p
}

func (s *CommonIntContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CommonIntContext) INT() antlr.TerminalNode {
	return s.GetToken(conditionParserINT, 0)
}

func (s *CommonIntContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.EnterCommonInt(s)
	}
}

func (s *CommonIntContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.ExitCommonInt(s)
	}
}

type LogicalTrueContext struct {
	*VariableDeclaratorContext
}

func NewLogicalTrueContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LogicalTrueContext {
	var p = new(LogicalTrueContext)

	p.VariableDeclaratorContext = NewEmptyVariableDeclaratorContext()
	p.parser = parser
	p.CopyFrom(ctx.(*VariableDeclaratorContext))

	return p
}

func (s *LogicalTrueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LogicalTrueContext) TRUE() antlr.TerminalNode {
	return s.GetToken(conditionParserTRUE, 0)
}

func (s *LogicalTrueContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.EnterLogicalTrue(s)
	}
}

func (s *LogicalTrueContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.ExitLogicalTrue(s)
	}
}

type LogicalFalseContext struct {
	*VariableDeclaratorContext
}

func NewLogicalFalseContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *LogicalFalseContext {
	var p = new(LogicalFalseContext)

	p.VariableDeclaratorContext = NewEmptyVariableDeclaratorContext()
	p.parser = parser
	p.CopyFrom(ctx.(*VariableDeclaratorContext))

	return p
}

func (s *LogicalFalseContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LogicalFalseContext) FALSE() antlr.TerminalNode {
	return s.GetToken(conditionParserFALSE, 0)
}

func (s *LogicalFalseContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.EnterLogicalFalse(s)
	}
}

func (s *LogicalFalseContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.ExitLogicalFalse(s)
	}
}

type VariableContext struct {
	*VariableDeclaratorContext
}

func NewVariableContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *VariableContext {
	var p = new(VariableContext)

	p.VariableDeclaratorContext = NewEmptyVariableDeclaratorContext()
	p.parser = parser
	p.CopyFrom(ctx.(*VariableDeclaratorContext))

	return p
}

func (s *VariableContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VariableContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(conditionParserIDENTIFIER, 0)
}

func (s *VariableContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.EnterVariable(s)
	}
}

func (s *VariableContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.ExitVariable(s)
	}
}

type CommonDoubleContext struct {
	*VariableDeclaratorContext
}

func NewCommonDoubleContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *CommonDoubleContext {
	var p = new(CommonDoubleContext)

	p.VariableDeclaratorContext = NewEmptyVariableDeclaratorContext()
	p.parser = parser
	p.CopyFrom(ctx.(*VariableDeclaratorContext))

	return p
}

func (s *CommonDoubleContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CommonDoubleContext) DOUBLE() antlr.TerminalNode {
	return s.GetToken(conditionParserDOUBLE, 0)
}

func (s *CommonDoubleContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.EnterCommonDouble(s)
	}
}

func (s *CommonDoubleContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.ExitCommonDouble(s)
	}
}

func (p *conditionParser) VariableDeclarator() (localctx IVariableDeclaratorContext) {
	localctx = NewVariableDeclaratorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, conditionParserRULE_variableDeclarator)

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.SetState(56)
	p.GetErrorHandler().Sync(p)

	switch p.GetTokenStream().LA(1) {
	case conditionParserTRUE:
		localctx = NewLogicalTrueContext(p, localctx)
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(50)
			p.Match(conditionParserTRUE)
		}

	case conditionParserFALSE:
		localctx = NewLogicalFalseContext(p, localctx)
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(51)
			p.Match(conditionParserFALSE)
		}

	case conditionParserINT:
		localctx = NewCommonIntContext(p, localctx)
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(52)
			p.Match(conditionParserINT)
		}

	case conditionParserDOUBLE:
		localctx = NewCommonDoubleContext(p, localctx)
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(53)
			p.Match(conditionParserDOUBLE)
		}

	case conditionParserSTRING:
		localctx = NewCommonStringContext(p, localctx)
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(54)
			p.Match(conditionParserSTRING)
		}

	case conditionParserIDENTIFIER:
		localctx = NewVariableContext(p, localctx)
		p.EnterOuterAlt(localctx, 6)
		{
			p.SetState(55)
			p.Match(conditionParserIDENTIFIER)
		}

	default:
		panic(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
	}

	return localctx
}

// IOpContext is an interface to support dynamic dispatch.
type IOpContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// IsOpContext differentiates from other interfaces.
	IsOpContext()
}

type OpContext struct {
	*antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOpContext() *OpContext {
	var p = new(OpContext)
	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(nil, -1)
	p.RuleIndex = conditionParserRULE_op
	return p
}

func (*OpContext) IsOpContext() {}

func NewOpContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *OpContext {
	var p = new(OpContext)

	p.BaseParserRuleContext = antlr.NewBaseParserRuleContext(parent, invokingState)

	p.parser = parser
	p.RuleIndex = conditionParserRULE_op

	return p
}

func (s *OpContext) GetParser() antlr.Parser { return s.parser }

func (s *OpContext) GE() antlr.TerminalNode {
	return s.GetToken(conditionParserGE, 0)
}

func (s *OpContext) LE() antlr.TerminalNode {
	return s.GetToken(conditionParserLE, 0)
}

func (s *OpContext) EQ() antlr.TerminalNode {
	return s.GetToken(conditionParserEQ, 0)
}

func (s *OpContext) NE() antlr.TerminalNode {
	return s.GetToken(conditionParserNE, 0)
}

func (s *OpContext) GT() antlr.TerminalNode {
	return s.GetToken(conditionParserGT, 0)
}

func (s *OpContext) LT() antlr.TerminalNode {
	return s.GetToken(conditionParserLT, 0)
}

func (s *OpContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OpContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *OpContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.EnterOp(s)
	}
}

func (s *OpContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(conditionListener); ok {
		listenerT.ExitOp(s)
	}
}

func (p *conditionParser) Op() (localctx IOpContext) {
	localctx = NewOpContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, conditionParserRULE_op)
	var _la int

	defer func() {
		p.ExitRule()
	}()

	defer func() {
		if err := recover(); err != nil {
			if v, ok := err.(antlr.RecognitionException); ok {
				localctx.SetException(v)
				p.GetErrorHandler().ReportError(p, v)
				p.GetErrorHandler().Recover(p, v)
			} else {
				panic(err)
			}
		}
	}()

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(58)
		_la = p.GetTokenStream().LA(1)

		if !(((_la)&-(0x1f+1)) == 0 && ((1<<uint(_la))&((1<<conditionParserEQ)|(1<<conditionParserNE)|(1<<conditionParserGE)|(1<<conditionParserLE)|(1<<conditionParserGT)|(1<<conditionParserLT))) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

	return localctx
}

func (p *conditionParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 0:
		var t *ConditionContext = nil
		if localctx != nil {
			t = localctx.(*ConditionContext)
		}
		return p.Condition_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *conditionParser) Condition_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 6)

	case 1:
		return p.Precpred(p.GetParserRuleContext(), 5)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
