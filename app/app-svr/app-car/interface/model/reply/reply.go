package reply

import (
	"fmt"
	"strconv"
	"strings"

	"go-gateway/app/app-svr/app-car/interface/model"
)

const (
	_authorRex = `@%s`
)

type ReplyList struct {
	Cursor struct {
		AllCount int    `json:"all_count,omitempty"`
		IsBegin  bool   `json:"is_begin,omitempty"`
		Prev     int    `json:"prev,omitempty"`
		Next     int64  `json:"next,omitempty"`
		IsEnd    bool   `json:"is_end,omitempty"`
		Mode     int    `json:"mode,omitempty"`
		Name     string `json:"name,omitempty"`
		ShowType int    `json:"show_type,omitempty"`
	} `json:"cursor,omitempty"`
	Control struct {
		InputDisable   bool   `json:"input_disable,omitempty"`
		RootInputText  string `json:"root_input_text,omitempty"`
		ChildInputText string `json:"child_input_text,omitempty"`
		BgText         string `json:"bg_text,omitempty"`
	} `json:"control,omitempty"`
	Root        *ReplyInfo   `json:"root,omitempty"`
	Replies     []*ReplyInfo `json:"replies,omitempty"`
	TopReplies  []*ReplyInfo `json:"top_replies,omitempty"`
	HotsReplies []*ReplyInfo `json:"hots,omitempty"`
	Page        struct {
		Count int64 `json:"count"`
		Num   int64 `json:"num"`
		Size  int64 `json:"size"`
	} `json:"page,omitempty"`
}

type ReplyInfo struct {
	Rpid         int64         `json:"rpid,omitempty"`
	Oid          int64         `json:"oid,omitempty"`
	Type         int64         `json:"type,omitempty"`
	Mid          int64         `json:"mid,omitempty"`
	Root         int64         `json:"root,omitempty"`
	Parent       int64         `json:"parent,omitempty"`
	Dialog       int64         `json:"dialog,omitempty"`
	Count        int64         `json:"count,omitempty"`
	Ctime        int64         `json:"ctime,omitempty"`
	Like         int64         `json:"like,omitempty"`
	Replies      []*ReplyInfo  `json:"replies,omitempty"`
	Member       *Member       `json:"member,omitempty"`
	Content      *Content      `json:"content,omitempty"`
	ReplyControl *ReplyControl `json:"reply_control,omitempty"`
}

type ReplyControl struct {
	SubReplyEntryText string `json:"sub_reply_entry_text,omitempty"`
	SubReplyTitleText string `json:"sub_reply_title_text,omitempty"`
	// ext
	URI string `json:"uri,omitempty"`
}

type Content struct {
	Message string            `json:"message,omitempty"`
	Emote   map[string]*Emote `json:"emote,omitempty"`
	Members []*Member         `json:"members,omitempty"`
}

type Emote struct {
	Text   string `json:"text,omitempty"`
	Url    string `json:"url,omitempty"`
	GifUrl string `json:"gif_url,omitempty"`
	Meta   *struct {
		Size int `json:"size,omitempty"`
	} `json:"meta,omitempty"`
}

type ContentReply struct {
	Message string    `json:"message,omitempty"`
	Emote   []*Emote  `json:"emote,omitempty"`
	Members []*Member `json:"members,omitempty"`
}

type Member struct {
	Mid  string `json:"mid,omitempty"`
	Name string `json:"uname,omitempty"`
	Face string `json:"avatar,omitempty"`
}

type ReplyParam struct {
	model.DeviceInfo
	Oid        int64  `form:"oid"`
	Cid        int64  `form:"cid"`
	SeasonType int    `form:"season_type"`
	Otype      string `form:"otype"`
	Next       int64  `form:"next"`
	Mode       int64  `form:"mode"`
	Ps         int    `form:"ps"`
	Pn         int    `form:"pn"`
	Root       int64  `form:"root"`
	Jump       int64  `form:"rpid"`
}

type ReplyExtra struct {
	EpId       int64 `json:"epid"`
	SeasonId   int64 `json:"season_id"`
	SeasonType int   `json:"season_type"`
}

type ReplyShow struct {
	Items []*ReplyItem `json:"items,omitempty"`
	Page  *Page        `json:"page,omitempty"`
}

type ReplyChild struct {
	TopItem *ReplyItem `json:"top_item,omitempty"`
	Related *ReplyRcmd `json:"related,omitempty"`
	Page    *PageChild `json:"page,omitempty"`
}

type ReplyRcmd struct {
	Items []*ReplyItem `json:"items,omitempty"`
	Title string       `json:"title,omitempty"`
}

type ReplyItem struct {
	Owner        *Owner        `json:"owner,omitempty"`
	Member       *Member       `json:"member,omitempty"`
	Rpid         int64         `json:"rpid,omitempty"`
	Root         int64         `json:"root,omitempty"`
	Desc1        string        `json:"desc_1,omitempty"`
	Desc2        string        `json:"desc_2,omitempty"`
	ContentIcon  model.Icon    `json:"content_icon,omitempty"`
	Content      *ContentReply `json:"content,omitempty"`
	Items        []*ReplyItem  `json:"items,omitempty"`
	ReplyControl *ReplyControl `json:"reply_control,omitempty"`
}

type Owner struct {
	Mid  int64  `json:"mid,omitempty"`
	Name string `json:"name,omitempty"`
	Face string `json:"face,omitempty"`
	URI  string `json:"uri,omitempty"`
}

type Page struct {
	Next  int64 `json:"next"`
	IsEnd bool  `json:"is_end"`
	Mode  int   `json:"mode"`
}

type PageChild struct {
	Pn    int64 `json:"pn"`
	IsEnd bool  `json:"is_end"`
}

func (s *ReplyItem) FromReplyInfo(upmid int64, rep *ReplyInfo, isTop, replyChild bool, build int) bool {
	const (
		_replyRep   = "%s %s"
		_replyNoRep = "%s：%s"
	)
	if rep.Member == nil || rep.Content == nil || rep.Content.Message == "" {
		return false
	}
	userMid, _ := strconv.ParseInt(rep.Member.Mid, 10, 64)
	s.Content = &ContentReply{}
	s.Content.FromReplyContent(rep.Content)
	s.Rpid = rep.Rpid
	s.Root = rep.Root
	if len(rep.Replies) >= 3 || (isTop && replyChild) {
		s.ReplyControl = rep.ReplyControl
	}
	if len(rep.Content.Members) > 0 && build < 1100000 {
		// @用户名
		s.Content.Message = authorProc(s.Content.Message, rep.Content.Members[0].Name)
	}
	s.Member = rep.Member
	if rep.Root == 0 || replyChild {
		s.Owner = &Owner{
			Mid:  userMid,
			Name: rep.Member.Name,
			Face: rep.Member.Face,
			URI:  model.FillURI(model.GotoSpace, 0, 0, rep.Member.Mid, nil),
		}
		s.Desc1 = rep.Member.Name
		s.Desc2 = model.ReplyDataString(rep.Ctime)
		s.Content.Message = strings.Replace(s.Content.Message, ":回复", "：回复 ", -1)
	} else {
		if upmid != userMid {
			text := _replyNoRep
			if strings.Contains(s.Content.Message, "回复") {
				text = _replyRep
			}
			// nolint:gomnd
			if build >= 1100000 {
				s.Content.Message = fmt.Sprintf(text, fmt.Sprintf(_authorRex, rep.Member.Name), s.Content.Message)
			} else {
				s.Content.Message = fmt.Sprintf(text, fmt.Sprintf("<font color=\"#178BCF\">%s</font>", rep.Member.Name), s.Content.Message)
			}
		}
	}
	if isTop && !replyChild {
		s.ContentIcon = model.IconTop
	}
	for _, v := range rep.Replies {
		if v.Member == nil || v.Content == nil || v.Content.Message == "" {
			continue
		}
		item := &ReplyItem{}
		if ok := item.FromReplyInfo(userMid, v, false, false, build); !ok {
			continue
		}
		s.Items = append(s.Items, item)
	}
	return true
}

func (s *ContentReply) FromReplyContent(c *Content) {
	s.Members = c.Members
	s.Message = c.Message
	for _, v := range c.Emote {
		s.Emote = append(s.Emote, v)
	}
}

func authorProc(desc, name string) string {
	if name == "" || desc == "" {
		return desc
	}
	userName := fmt.Sprintf(_authorRex, name)
	return strings.Replace(desc, userName, fmt.Sprintf("<font color=\"#178BCF\">%s</font>", userName), -1)
}
