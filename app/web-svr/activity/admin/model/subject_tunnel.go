package model

import (
	"encoding/json"
	"fmt"
	"strings"

	"go-gateway/app/web-svr/activity/admin/model/stime"
)

const (
	_tunnelInsertSQL = "INSERT INTO act_subject_tunnel(`sid`,`type`,`template_id`,`title`,`content`,`icon`,`link`,`sender_uid`) VALUES %s"
	_tunnelUpdateSQL = "UPDATE act_subject_tunnel SET template_id = CASE %s END,title = CASE %s END,content = CASE %s END,icon = CASE %s END,link = CASE %s END,sender_uid = CASE %s END WHERE id IN (%s)"
)

type SubjectTunnel struct {
	ID         int64          `json:"id"`
	Sid        int64          `json:"-"`
	Type       int64          `json:"type"`
	TemplateID int64          `json:"template_id"`
	Titles     *TunnelTitle   `json:"titles",omitempty`
	Contents   *TunnelContent `json:"contents",omitempty`
	Icon       string         `json:"icon"`
	Link       string         `json:"link"`
	SenderUid  int64          `json:"sender_uid"`
}

type ActSubjectTunnel struct {
	ID         int64  `json:"id"`
	Sid        int64  `json:"-"`
	Type       int64  `json:"type"`
	TemplateID int64  `json:"template_id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Icon       string `json:"icon"`
	Link       string `json:"link"`
	SenderUid  int64  `json:"sender_uid"`
}

type TunnelInfo struct {
	Sid       int64          `json:"sid"`
	IsPush    int64          `json:"is_push"`
	Index     *SubjectTunnel `json:"index"`
	Letter    *SubjectTunnel `json:"letter"`
	Dynamic   *SubjectTunnel `json:"dynamic"`
	PushStart stime.Time     `json:"push_start,omitempty" form:"-" time_format:"2006-01-02 15:04:05"`
	PushEnd   stime.Time     `json:"push_end,omitempty" form:"-" time_format:"2006-01-02 15:04:05"`
}

type TunnelTitle struct {
	Title  string `json:"title,omitempty"`
	Title1 string `json:"title1,omitempty"`
	Title2 string `json:"title2,omitempty"`
	Title3 string `json:"title3,omitempty"`
	Title4 string `json:"title4,omitempty"`
	Title5 string `json:"title5,omitempty"`
}

type TunnelContent struct {
	Content  string `json:"content,omitempty"`
	Content1 string `json:"content1,omitempty"`
	Content2 string `json:"content2,omitempty"`
	Content3 string `json:"content3,omitempty"`
	Content4 string `json:"content4,omitempty"`
	Content5 string `json:"content5,omitempty"`
}

type SubjectTunnelParam struct {
	Sid       int64  `json:"sid" form:"sid" validate:"min=1,required"`
	Index     string `json:"index" form:"index"`
	Letter    string `json:"letter" form:"letter"`
	Dynamic   string `json:"dynamic" form:"dynamic"`
	PushStart string `json:"push_start" form:"push_start" time_format:"2006-01-02 15:04:05" validate:"required"`
	PushEnd   string `json:"push_end" form:"push_end" time_format:"2006-01-02 15:04:05" validate:"required"`
}

type PushTemplate struct {
	TemplateID int64  `json:"template_id"`
	Titles     string `json:"titles"`
	Contents   string `json:"contents"`
}

// TableName ActSubjectTunnel def.
func (ActSubjectTunnel) TableName() string {
	return "act_subject_tunnel"
}

// TunnelBatchAddSQL .
func TunnelBatchAddSQL(tunnels []*SubjectTunnel) (sql string, param []interface{}) {
	if len(tunnels) == 0 {
		return "", []interface{}{}
	}
	var rowStrings []string
	for _, v := range tunnels {
		var title, content []byte
		rowStrings = append(rowStrings, "(?,?,?,?,?,?,?,?)")
		if v.Titles != nil {
			title, _ = json.Marshal(v.Titles)
		}
		if v.Contents != nil {
			content, _ = json.Marshal(v.Contents)
		}
		param = append(param, v.Sid, v.Type, v.TemplateID, string(title), string(content), v.Icon, v.Link, v.SenderUid)
	}
	return fmt.Sprintf(_tunnelInsertSQL, strings.Join(rowStrings, ",")), param
}

// TunnelBatchEditSQL .
func TunnelBatchEditSQL(tunnels []*SubjectTunnel) (sql string, param []interface{}) {
	if len(tunnels) == 0 {
		return "", []interface{}{}
	}
	var (
		templateStr, titleStr, contentStr, iconStr, linkStr, senderStr string
		ids                                                            []interface{}
		idSql                                                          []string
	)
	for _, tunnel := range tunnels {
		templateStr += " WHEN id = ? THEN ?"
		param = append(param, tunnel.ID, tunnel.TemplateID)
		idSql = append(idSql, "?")
		ids = append(ids, tunnel.ID)
	}
	for _, tunnel := range tunnels {
		var title []byte
		if tunnel.Titles != nil {
			title, _ = json.Marshal(tunnel.Titles)
		}
		titleStr += " WHEN id = ? THEN ?"
		param = append(param, tunnel.ID, string(title))
	}
	for _, tunnel := range tunnels {
		var content []byte
		if tunnel.Contents != nil {
			content, _ = json.Marshal(tunnel.Contents)
		}
		contentStr += " WHEN id = ? THEN ?"
		param = append(param, tunnel.ID, string(content))
	}
	for _, tunnel := range tunnels {
		iconStr += " WHEN id = ? THEN ?"
		param = append(param, tunnel.ID, tunnel.Icon)
	}
	for _, tunnel := range tunnels {
		linkStr += " WHEN id = ? THEN ?"
		param = append(param, tunnel.ID, tunnel.Link)
	}
	for _, tunnel := range tunnels {
		senderStr += " WHEN id = ? THEN ?"
		param = append(param, tunnel.ID, tunnel.SenderUid)
	}
	param = append(param, ids...)
	return fmt.Sprintf(_tunnelUpdateSQL, templateStr, titleStr, contentStr, iconStr, linkStr, senderStr, strings.Join(idSql, ",")), param
}
