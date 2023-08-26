package service

import (
	"context"

	acccli "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model"
	"go-gateway/app/web-svr/activity/admin/model/up"
)

func (s *Service) UpActList(c context.Context, uid, state, pn, ps int64) (res []*up.Item, count int64, err error) {
	var list *up.ListReply
	if list, err = s.dao.SearchUpList(c, uid, state, pn, ps); err != nil {
		log.Error("UpActList s.dao.SearchUpList(%d,%d,%d,%d) error(%v)", uid, state, pn, ps, err)
		return
	}
	if list == nil || len(list.List) == 0 {
		return
	}
	count = list.Count
	mids := make([]int64, 0)
	for _, v := range list.List {
		if v.Mid <= 0 {
			continue
		}
		mids = append(mids, v.Mid)
	}
	if len(mids) == 0 {
		return
	}
	var account *acccli.InfosReply
	if account, err = s.accClient.Infos3(c, &acccli.MidsReq{Mids: mids}); err != nil {
		log.Error(" s.accClient.Infos3(%v) error(%v)", mids, err)
		return
	}
	for _, v := range list.List {
		tmp := &up.Item{UpAct: v}
		if acc, ok := account.Infos[v.Mid]; ok {
			tmp.Name = acc.Name
		}
		res = append(res, tmp)
	}
	return
}

func countSuffix(id int64) int64 {
	return id % 10
}

func (s *Service) UpActEdit(c context.Context, id, uid, state int64, isBig int) (err error) {
	var suffix int64
	if isBig == 0 {
		suffix = countSuffix(id)
	} else {
		suffix = uid
	}
	if err = s.dao.UpActEdit(c, id, state, suffix); err != nil {
		log.Error("UpActEdit s.dao.UpActEdit(%d,%d) error(%v)", id, state, err)
		return
	}
	var content string
	if state == 1 {
		content = s.c.Up.PassContent
	} else if state == 2 {
		content = s.c.Up.UnPassContent
	}
	if uid == 0 {
		return
	}
	// 发私信
	l := &model.LetterParam{
		RecverIDs: []uint64{uint64(uid)},
		SenderUID: s.c.Up.SenderUid,
		MsgType:   int32(1),
		Content:   content,
	}
	if _, err = s.dao.SendLetter(c, l); err != nil {
		log.Error("UpActEdit s.dao.SendLetter error(%v)", err)
	}
	return
}

func (s *Service) UpActOffline(c context.Context, id, offline int64) (err error) {
	if err = s.dao.UpActOffline(c, id, offline); err != nil {
		log.Error("UpActEdit s.dao.UpActOffline(%d,%d) error(%v)", id, offline, err)
	}
	return
}
