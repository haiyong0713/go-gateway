// Code generated from condition.g4 by ANTLR 4.8. DO NOT EDIT.

package parser // condition

import "github.com/antlr/antlr4/runtime/Go/antlr"

// conditionListener is a complete listener for a parse tree produced by conditionParser.
type conditionListener interface {
	antlr.ParseTreeListener

	// EnterLogicalOp2 is called when entering the logicalOp2 production.
	EnterLogicalOp2(c *LogicalOp2Context)

	// EnterLogicalNot is called when entering the logicalNot production.
	EnterLogicalNot(c *LogicalNotContext)

	// EnterParen is called when entering the paren production.
	EnterParen(c *ParenContext)

	// EnterCommomOp is called when entering the commomOp production.
	EnterCommomOp(c *CommomOpContext)

	// EnterLogicalAnd is called when entering the logicalAnd production.
	EnterLogicalAnd(c *LogicalAndContext)

	// EnterLogicalOr is called when entering the logicalOr production.
	EnterLogicalOr(c *LogicalOrContext)

	// EnterInOrNotInOp2 is called when entering the inOrNotInOp2 production.
	EnterInOrNotInOp2(c *InOrNotInOp2Context)

	// EnterLogicalOp is called when entering the logicalOp production.
	EnterLogicalOp(c *LogicalOpContext)

	// EnterInOrNotInOp is called when entering the inOrNotInOp production.
	EnterInOrNotInOp(c *InOrNotInOpContext)

	// EnterLogicalTrue is called when entering the logicalTrue production.
	EnterLogicalTrue(c *LogicalTrueContext)

	// EnterLogicalFalse is called when entering the logicalFalse production.
	EnterLogicalFalse(c *LogicalFalseContext)

	// EnterCommonInt is called when entering the commonInt production.
	EnterCommonInt(c *CommonIntContext)

	// EnterCommonDouble is called when entering the commonDouble production.
	EnterCommonDouble(c *CommonDoubleContext)

	// EnterCommonString is called when entering the commonString production.
	EnterCommonString(c *CommonStringContext)

	// EnterVariable is called when entering the variable production.
	EnterVariable(c *VariableContext)

	// EnterOp is called when entering the op production.
	EnterOp(c *OpContext)

	// ExitLogicalOp2 is called when exiting the logicalOp2 production.
	ExitLogicalOp2(c *LogicalOp2Context)

	// ExitLogicalNot is called when exiting the logicalNot production.
	ExitLogicalNot(c *LogicalNotContext)

	// ExitParen is called when exiting the paren production.
	ExitParen(c *ParenContext)

	// ExitCommomOp is called when exiting the commomOp production.
	ExitCommomOp(c *CommomOpContext)

	// ExitLogicalAnd is called when exiting the logicalAnd production.
	ExitLogicalAnd(c *LogicalAndContext)

	// ExitLogicalOr is called when exiting the logicalOr production.
	ExitLogicalOr(c *LogicalOrContext)

	// ExitInOrNotInOp2 is called when exiting the inOrNotInOp2 production.
	ExitInOrNotInOp2(c *InOrNotInOp2Context)

	// ExitLogicalOp is called when exiting the logicalOp production.
	ExitLogicalOp(c *LogicalOpContext)

	// ExitInOrNotInOp is called when exiting the inOrNotInOp production.
	ExitInOrNotInOp(c *InOrNotInOpContext)

	// ExitLogicalTrue is called when exiting the logicalTrue production.
	ExitLogicalTrue(c *LogicalTrueContext)

	// ExitLogicalFalse is called when exiting the logicalFalse production.
	ExitLogicalFalse(c *LogicalFalseContext)

	// ExitCommonInt is called when exiting the commonInt production.
	ExitCommonInt(c *CommonIntContext)

	// ExitCommonDouble is called when exiting the commonDouble production.
	ExitCommonDouble(c *CommonDoubleContext)

	// ExitCommonString is called when exiting the commonString production.
	ExitCommonString(c *CommonStringContext)

	// ExitVariable is called when exiting the variable production.
	ExitVariable(c *VariableContext)

	// ExitOp is called when exiting the op production.
	ExitOp(c *OpContext)
}
