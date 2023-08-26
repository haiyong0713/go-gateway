package dynHandler

import (
	"fmt"
	"strings"

	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
)

const (
	_titleSearchWordFormat = "<em class=\"keyword\">%s</em>"
)

func (schema *CardSchema) getTitle(title string, dynCtx *dynmdlV2.DynamicContext) string {
	// 搜索词飘红
	if dynCtx.SearchWordRed {
		titleArr := schema.titleSearchWordProc(title, dynCtx)
		var result string
		for _, cardTitle := range titleArr {
			result = result + cardTitle
		}
		return result
	} else {
		return title
	}
}

func (schema *CardSchema) titleSearchWordProc(title string, dynCtx *dynmdlV2.DynamicContext) []string {
	wordLen, index := 0, -1
	for _, searchWord := range dynCtx.SearchWords {
		index = strings.Index(title, searchWord)
		if index != -1 {
			wordLen = len(searchWord)
			break
		}
	}
	var res []string
	if index == -1 {
		tmp := title
		res = append(res, tmp)
		return res
	}
	end := index + wordLen
	pre := title[:index]
	top := title[index:end]
	aft := title[end:]
	if pre != "" {
		tmp := schema.titleSearchWordProc(pre, dynCtx)
		res = append(res, tmp...)
	}
	tmp := fmt.Sprintf(_titleSearchWordFormat, top)
	res = append(res, tmp)
	if aft != "" {
		tmp := schema.titleSearchWordProc(aft, dynCtx)
		res = append(res, tmp...)
	}
	return res
}
