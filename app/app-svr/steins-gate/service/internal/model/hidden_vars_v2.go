package model

import (
	"encoding/json"
	"regexp"

	"go-common/library/log"
)

const regExpression = "^v-" // 去掉固定前缀v-

// hvarIDTransform 去掉老var_id的特殊字符
func hvarIDTransform(originID string) (displayID string) {
	reg, err := regexp.Compile(regExpression)
	if err != nil {
		log.Error("HvarIDTransform originID %s, err %v", originID, err)
		return
	}
	return reg.ReplaceAllString(originID, "")
}

// IsExpr 返回是否为新的表达式
func (v *EdgeAttribute) IsExpr() bool {
	return v.ActionType == attributeIsExpr
}

// ApplyAttrs 支持针对多个edge的选择进行结算 支持表达式（不含特殊符号）和Syntax树（含v-）
func (out *HiddenVarsRecord) ApplyAttrs(edgesAttrStrs []string) (err error) {
	var edgeAttrsStr string
	for _, edgeAttrStrs := range edgesAttrStrs {
		if edgeAttrStrs == "" { // 当前edge的attributes
			continue
		}
		var edgeAttrs []*EdgeAttribute
		if err = json.Unmarshal([]byte(edgeAttrStrs), &edgeAttrs); err != nil {
			continue
		}
		for _, edgeAttr := range edgeAttrs {
			if edgeAttr.IsExpr() { // 如果是表达式 针对表达式进行解析
				edgeAttrsStr += edgeAttr.Action + split
				continue
			}
			if _, ok := out.Vars[edgeAttr.VarID]; !ok { // 如果attribute中含有graph中未声明的变量，不处理
				continue
			}
			out.Vars[edgeAttr.VarID].MergeAttribute(edgeAttr) // 合入当前节点的action
		}
	}
	if edgeAttrsStr != "" {
		//nolint:errcheck
		out.ExpressionEval(edgeAttrsStr)
	}
	return

}
