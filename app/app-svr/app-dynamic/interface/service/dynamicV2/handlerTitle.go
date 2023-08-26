package dynamicV2

import (
	"fmt"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	"strings"
)

const _titleSearchWordFormat = "<em class=\"keyword\">%s</em>"

func (s *Service) getTitle(title string, dynCtx *mdlv2.DynamicContext) string {
	// 搜索词飘红
	if dynCtx.SearchWordRed {
		titleArr := s.titleSearchWordProc(title, dynCtx)
		var result string
		for _, cardTitle := range titleArr {
			result = result + cardTitle
		}
		return result
	} else {
		return title
	}
}

func (s *Service) titleSearchWordProc(title string, dynCtx *mdlv2.DynamicContext) []string {
	index := -1
	wordLen := 0
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
		tmp := s.titleSearchWordProc(pre, dynCtx)
		res = append(res, tmp...)
	}
	tmp := fmt.Sprintf(_titleSearchWordFormat, top)
	res = append(res, tmp)
	if aft != "" {
		tmp := s.titleSearchWordProc(aft, dynCtx)
		res = append(res, tmp...)
	}
	return res
}
