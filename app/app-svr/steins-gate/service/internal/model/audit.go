package model

import (
	"encoding/json"
	"unicode/utf8"
)

// AuditParam def.
type AuditParam struct {
	ID           int64  `form:"id" validate:"required,min=1"`
	RejectID     int    `form:"reject_id"`
	RejectTitle  string `form:"reject_title"`
	RejectReason string `form:"reject_reason"`
	WithNotify   int    `form:"with_notify"`
	State        int    `form:"state" validate:"required"`
}

// AegisMetaData .
type AegisMetaData struct {
	DiffMsg  string `json:"diff_msg"`
	VarsName string `json:"vars_name"`
}

// BuildMsg 如果标题未勾选只返回理由，如果标题勾选了则进行打*处理后一起返回
func (v *AuditParam) BuildMsg() (result string) {
	if v.RejectTitle == "" {
		return v.RejectReason
	}
	a := []rune(v.RejectTitle)
	var b []rune
	for i := 0; i < len(a); i++ {
		if i%3 != 0 {
			b = append(b, rune('*'))
		} else {
			b = append(b, a[i])
		}
	}
	return string(b) + v.RejectReason
}

func AegisMetaLen(msg, varsName string) (res int) {
	var metaData []byte
	meta := &AegisMetaData{DiffMsg: msg, VarsName: varsName}
	metaData, _ = json.Marshal(meta)
	res = utf8.RuneCountInString(string(metaData))
	return
}

type GraphAuditDB struct {
	GraphDB
	ResultGID int64
}

func (v *GraphAuditDB) HasNoDiffVarsName(in []*RegionalVal) (res bool) {
	var (
		old    []*RegionalVal
		oldMap = make(map[string]struct{})
	)
	if err := json.Unmarshal([]byte(v.RegionalVars), &old); err != nil {
		res = false
		return
	}
	for _, item := range old {
		if item.IsShow == 1 {
			oldMap[item.Name] = struct{}{}
		}
	}
	for _, item := range in {
		if item.IsShow == 1 {
			if _, ok := oldMap[item.Name]; !ok {
				res = false
				return
			}
		}
	}
	res = true
	return

}
