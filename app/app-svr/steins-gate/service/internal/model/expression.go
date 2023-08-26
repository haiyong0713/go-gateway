package model

import (
	"fmt"
	"strings"

	"go-gateway/app/app-svr/steins-gate/service/internal/model/compiler"
)

const (
	split              = ";"
	conditionSeparator = " && "
	param              = "$"

	// 	editorPrefix      = "v-" // 编辑器生成变量的前缀
	attributeIsExpr   = 1
	attributeIsSyntax = 0
)

// ExpressionEval 表达式求值后将结果赋入隐藏变量存档中
func (out *HiddenVarsRecord) ExpressionEval(expr string) (err error) {
	var (
		values = make(map[string]float64, len(out.Vars))
		varMap = make(map[string]string, len(out.Vars))
	)
	for _, v := range out.Vars {
		values[hvarIDToExpr(v.ID)] = v.Value
		varMap[hvarIDToExpr(v.ID)] = v.ID // 存下映射关系，便于转换回来
	}
	calc := new(compiler.Calculator)
	//nolint:errcheck
	calc.InitAndEval(expr, values)
	for k, v := range calc.Values {
		realVarID, ok := varMap[k] // 找回原来的varID
		if !ok {
			continue
		}
		if _, ok := out.Vars[realVarID]; !ok { // 计算后的结果赋值 只存下来graph中的隐藏变量的值，不关心局部变量
			continue
		}
		out.Vars[realVarID].Value = v
	}
	return
}

// ExpressionCondition 判断是不是满足当前条件
func (out *HiddenVarsRecord) ExpressionCondition(expr string) (res bool, err error) {
	var (
		values = make(map[string]float64, len(out.Vars))
		result float64
	)
	for _, v := range out.Vars {
		values[hvarIDToExpr(v.ID)] = v.Value
	}
	calc := new(compiler.Calculator)
	if result, err = calc.InitAndEval(expr, values); err == nil && result == 1 {
		return true, nil
	}
	return
}

// hvarIDToExpr v-a2134@#33=>$a1234__33 处理所有非法字符
func hvarIDToExpr(varID string) string {
	return param + compiler.TreatVarName(hvarIDTransform(varID))
}

// FromSyntax 用一个语法树的edge的attribute来初始化一个表达式的edge的attribute => $a1=$a1+34.5
func (v *EdgeAttribute) FromSyntax(syntaxEdge *EdgeAttribute) {
	newVarID := hvarIDToExpr(syntaxEdge.VarID)
	v.ActionType = attributeIsExpr
	switch syntaxEdge.Action {
	case EdgeAttrActionSub:
		v.Action += fmt.Sprintf("%s=%s-%0.2f;", newVarID, newVarID, float64(syntaxEdge.Value))
	case EdgeAttrActionAdd:
		v.Action += fmt.Sprintf("%s=%s+%0.2f;", newVarID, newVarID, float64(syntaxEdge.Value))
	case EdgeAttrActionAssign:
		v.Action += fmt.Sprintf("%s=%0.2f;", newVarID, float64(syntaxEdge.Value))
	}
	v.Action = strings.TrimRight(v.Action, split)
	v.VarID = syntaxEdge.VarID
	v.ActionType = attributeIsExpr
}

// FromSyntax 用一个语法树的edge的condition来初始化一个表达式的edge的condition
func (v *EdgeCondition) FromSyntax(syntaxCondition *EdgeCondition) {
	newVarID := hvarIDToExpr(syntaxCondition.VarID) // 过滤老隐藏变量ID中的非法字符
	switch syntaxCondition.Condition {
	case EdgeConditionTypeGe:
		v.Condition = fmt.Sprintf("%s>=%0.2f", newVarID, float64(syntaxCondition.Value))
	case EdgeConditionTypeGt:
		v.Condition = fmt.Sprintf("%s>%0.2f", newVarID, float64(syntaxCondition.Value))
	case EdgeConditionTypeEq:
		v.Condition = fmt.Sprintf("%s==%0.2f", newVarID, float64(syntaxCondition.Value))
	case EdgeConditionTypeLt:
		v.Condition = fmt.Sprintf("%s<%0.2f", newVarID, float64(syntaxCondition.Value))
	case EdgeConditionTypeLe:
		v.Condition = fmt.Sprintf("%s<=%0.2f", newVarID, float64(syntaxCondition.Value))
	}
	v.VarID = syntaxCondition.VarID
	v.ActionType = attributeIsExpr

}
