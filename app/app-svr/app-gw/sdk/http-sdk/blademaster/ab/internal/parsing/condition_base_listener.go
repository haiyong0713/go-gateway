// Code generated from condition.g4 by ANTLR 4.8. DO NOT EDIT.

package parser // condition

import "github.com/antlr/antlr4/runtime/Go/antlr"

// BaseconditionListener is a complete listener for a parse tree produced by conditionParser.
type BaseconditionListener struct{}

var _ conditionListener = &BaseconditionListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseconditionListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseconditionListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseconditionListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseconditionListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterLogicalOp2 is called when production logicalOp2 is entered.
func (s *BaseconditionListener) EnterLogicalOp2(ctx *LogicalOp2Context) {}

// ExitLogicalOp2 is called when production logicalOp2 is exited.
func (s *BaseconditionListener) ExitLogicalOp2(ctx *LogicalOp2Context) {}

// EnterLogicalNot is called when production logicalNot is entered.
func (s *BaseconditionListener) EnterLogicalNot(ctx *LogicalNotContext) {}

// ExitLogicalNot is called when production logicalNot is exited.
func (s *BaseconditionListener) ExitLogicalNot(ctx *LogicalNotContext) {}

// EnterParen is called when production paren is entered.
func (s *BaseconditionListener) EnterParen(ctx *ParenContext) {}

// ExitParen is called when production paren is exited.
func (s *BaseconditionListener) ExitParen(ctx *ParenContext) {}

// EnterCommomOp is called when production commomOp is entered.
func (s *BaseconditionListener) EnterCommomOp(ctx *CommomOpContext) {}

// ExitCommomOp is called when production commomOp is exited.
func (s *BaseconditionListener) ExitCommomOp(ctx *CommomOpContext) {}

// EnterLogicalAnd is called when production logicalAnd is entered.
func (s *BaseconditionListener) EnterLogicalAnd(ctx *LogicalAndContext) {}

// ExitLogicalAnd is called when production logicalAnd is exited.
func (s *BaseconditionListener) ExitLogicalAnd(ctx *LogicalAndContext) {}

// EnterLogicalOr is called when production logicalOr is entered.
func (s *BaseconditionListener) EnterLogicalOr(ctx *LogicalOrContext) {}

// ExitLogicalOr is called when production logicalOr is exited.
func (s *BaseconditionListener) ExitLogicalOr(ctx *LogicalOrContext) {}

// EnterInOrNotInOp2 is called when production inOrNotInOp2 is entered.
func (s *BaseconditionListener) EnterInOrNotInOp2(ctx *InOrNotInOp2Context) {}

// ExitInOrNotInOp2 is called when production inOrNotInOp2 is exited.
func (s *BaseconditionListener) ExitInOrNotInOp2(ctx *InOrNotInOp2Context) {}

// EnterLogicalOp is called when production logicalOp is entered.
func (s *BaseconditionListener) EnterLogicalOp(ctx *LogicalOpContext) {}

// ExitLogicalOp is called when production logicalOp is exited.
func (s *BaseconditionListener) ExitLogicalOp(ctx *LogicalOpContext) {}

// EnterInOrNotInOp is called when production inOrNotInOp is entered.
func (s *BaseconditionListener) EnterInOrNotInOp(ctx *InOrNotInOpContext) {}

// ExitInOrNotInOp is called when production inOrNotInOp is exited.
func (s *BaseconditionListener) ExitInOrNotInOp(ctx *InOrNotInOpContext) {}

// EnterLogicalTrue is called when production logicalTrue is entered.
func (s *BaseconditionListener) EnterLogicalTrue(ctx *LogicalTrueContext) {}

// ExitLogicalTrue is called when production logicalTrue is exited.
func (s *BaseconditionListener) ExitLogicalTrue(ctx *LogicalTrueContext) {}

// EnterLogicalFalse is called when production logicalFalse is entered.
func (s *BaseconditionListener) EnterLogicalFalse(ctx *LogicalFalseContext) {}

// ExitLogicalFalse is called when production logicalFalse is exited.
func (s *BaseconditionListener) ExitLogicalFalse(ctx *LogicalFalseContext) {}

// EnterCommonInt is called when production commonInt is entered.
func (s *BaseconditionListener) EnterCommonInt(ctx *CommonIntContext) {}

// ExitCommonInt is called when production commonInt is exited.
func (s *BaseconditionListener) ExitCommonInt(ctx *CommonIntContext) {}

// EnterCommonDouble is called when production commonDouble is entered.
func (s *BaseconditionListener) EnterCommonDouble(ctx *CommonDoubleContext) {}

// ExitCommonDouble is called when production commonDouble is exited.
func (s *BaseconditionListener) ExitCommonDouble(ctx *CommonDoubleContext) {}

// EnterCommonString is called when production commonString is entered.
func (s *BaseconditionListener) EnterCommonString(ctx *CommonStringContext) {}

// ExitCommonString is called when production commonString is exited.
func (s *BaseconditionListener) ExitCommonString(ctx *CommonStringContext) {}

// EnterVariable is called when production variable is entered.
func (s *BaseconditionListener) EnterVariable(ctx *VariableContext) {}

// ExitVariable is called when production variable is exited.
func (s *BaseconditionListener) ExitVariable(ctx *VariableContext) {}

// EnterOp is called when production op is entered.
func (s *BaseconditionListener) EnterOp(ctx *OpContext) {}

// ExitOp is called when production op is exited.
func (s *BaseconditionListener) ExitOp(ctx *OpContext) {}
