package reply

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-car/interface/conf"
	replydao "go-gateway/app/app-svr/app-car/interface/dao/reply"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/reply"
)

const (
	_psMax = 20
)

type Service struct {
	c   *conf.Config
	dao *replydao.Dao
}

func New(c *conf.Config) *Service {
	s := &Service{
		c:   c,
		dao: replydao.New(c),
	}
	return s
}

func (s *Service) Replys(c context.Context, mid int64, param *reply.ReplyParam) (*reply.ReplyShow, error) {
	var paramExt *reply.ReplyExtra
	if param.Otype == model.GotoPGC {
		paramExt = &reply.ReplyExtra{
			EpId:       param.Cid,
			SeasonId:   param.Oid,
			SeasonType: param.SeasonType,
		}
	}
	replys, err := s.dao.ReplyMain(c, _psMax, param.Oid, 1, param.Mode, param.Next, mid, paramExt)
	if err != nil {
		return nil, err
	}
	// 普通评论
	var (
		items     []*reply.ReplyItem
		hotsItems []*reply.ReplyItem
		topItems  []*reply.ReplyItem
	)
	// 热门
	for _, v := range replys.HotsReplies {
		item := &reply.ReplyItem{}
		if ok := item.FromReplyInfo(0, v, false, false, param.Build); !ok {
			continue
		}
		hotsItems = append(hotsItems, item)
	}
	// 置顶
	for _, v := range replys.TopReplies {
		item := &reply.ReplyItem{}
		if ok := item.FromReplyInfo(0, v, true, false, param.Build); !ok {
			continue
		}
		topItems = append(topItems, item)
	}
	for _, v := range replys.Replies {
		// 去重复
		for _, hot := range hotsItems {
			if hot.Rpid == v.Rpid {
				continue
			}
		}
		for _, top := range topItems {
			if top.Rpid == v.Rpid {
				continue
			}
		}
		item := &reply.ReplyItem{}
		if ok := item.FromReplyInfo(0, v, false, false, param.Build); !ok {
			continue
		}
		items = append(items, item)
	}
	if param.Next == 0 {
		hotsItems = append(hotsItems, items...)
		topItems = append(topItems, hotsItems...)
	} else {
		topItems = items
	}
	res := &reply.ReplyShow{
		Items: topItems,
		Page: &reply.Page{
			Next:  replys.Cursor.Next,
			IsEnd: replys.Cursor.IsEnd,
			Mode:  replys.Cursor.Mode,
		},
	}
	return res, nil
}

func (s *Service) ReplyChild(c context.Context, mid int64, param *reply.ReplyParam) (*reply.ReplyChild, error) {
	// 兜底
	if param.Pn == 0 && param.Jump == 0 {
		param.Pn = 1
	}
	// 后续翻页不依赖jump
	if param.Pn > 0 {
		param.Jump = 0
	}
	replys, err := s.dao.ReplyChild(c, param.Pn, _psMax, param.Oid, 1, param.Root, param.Jump, param.Mode, param.Next, mid)
	if err != nil {
		return nil, err
	}
	// 普通评论
	var (
		items []*reply.ReplyItem
	)
	for _, v := range replys.Replies {
		item := &reply.ReplyItem{}
		if ok := item.FromReplyInfo(0, v, false, true, param.Build); !ok {
			continue
		}
		items = append(items, item)
	}
	res := &reply.ReplyChild{
		Related: &reply.ReplyRcmd{
			Items: items,
			Title: fmt.Sprintf("相关回复共%d条", replys.Page.Count),
		},
		Page: &reply.PageChild{
			Pn: replys.Page.Num,
		},
	}
	top := &reply.ReplyItem{}
	if ok := top.FromReplyInfo(0, replys.Root, true, true, param.Build); ok {
		res.TopItem = top
	}
	// 最后
	if (replys.Page.Num+1)*replys.Page.Size >= replys.Page.Count {
		res.Page.IsEnd = true
	}
	return res, nil
}
