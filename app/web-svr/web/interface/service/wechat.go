package service

import (
	"context"
	"errors"
	"go-common/library/log"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web/interface/model"
)

var (
	_emptyWxArc     = make([]*model.WxArchive, 0)
	_emptyWxArcTag  = make([]*model.WxArcTag, 0)
	_teenageModeYes = 1
)

// WxHot get wx hot archives.
func (s *Service) WxHot(c context.Context, pn, ps int, platform string, teenage_mode int, mid int64, buvid string) (res []*model.WxArchive, count int, err error) {
	var (
		pageCards []*model.WxArchiveCard
		pageAids  []int64
		arcs      map[int64]*model.WxArchive
		_av       = "av"
	)
	if teenage_mode == _teenageModeYes {
		// 青少年模式走AI推荐
		data, _, aiErr := func() (res []*model.WXTeenageRcmdItem, code int, err error) {
			data, resCode, aiErr := s.dao.GetTeenageRcmdCards(c, mid, buvid, 0, ps)
			if aiErr != nil {
				log.Error("日志报警 【WxHot】 GetTeenageRcmdCards mid:%d buvid:%s ps:%d error(%v)", mid, buvid, ps, err)
				return nil, resCode, err
			}
			if len(data) == 0 {
				return nil, 0, errors.New("日志报警 【WxHot】 rcmd aids len is 0")
			}
			return data, resCode, err
		}()
		if aiErr != nil {
			err = aiErr
			res = _emptyWxArc
			return
		}

		for _, v := range data {
			if v.Goto == _av {
				card := &model.WxArchiveCard{}
				card.ID = v.ID
				if v.RcmdReason != nil {
					card.CornerMark = v.RcmdReason.CornerMark
					card.Desc = v.RcmdReason.Content
				}
				pageCards = append(pageCards, card)
				pageAids = append(pageAids, v.ID)
			}
		}
		count = len(pageAids)
	} else {
		arcCard := s.wxHotAids
		pageCards, count = wxHotPage(arcCard, pn, ps)
		if len(pageCards) == 0 {
			res = _emptyWxArc
			return
		}
		for _, v := range pageCards {
			pageAids = append(pageAids, v.ID)
		}
	}

	if arcs, err = s.archiveWithTag(c, pageAids); err != nil {
		res = _emptyWxArc
		return
	}
	for _, card := range pageCards {
		if arc, ok := arcs[card.ID]; ok && arc != nil {
			arc.HotDesc = card.Desc
			arc.CornerMark = card.CornerMark
			res = append(res, arc)
		}
	}

	if platform == "wechat" {
		s.wxHotFilterBindOid(c, &res)
	}
	return
}

func (s *Service) wxHotFilterBindOid(c context.Context, containOidSlice *[]*model.WxArchive) {
	if len(*containOidSlice) == 0 {
		return
	}
	var oidList []int64
	for _, v := range *containOidSlice {
		if v != nil {
			oidList = append(oidList, v.Aid)
		}
	}

	bindOidList, err := s.dao.TagBind(c, oidList)
	k := 0
	for _, v := range *containOidSlice {
		if err != nil || bindOidList == nil || v == nil || !inIntSlice(bindOidList, v.Aid) {
			(*containOidSlice)[k] = v
			k++
		}
	}
	*containOidSlice = (*containOidSlice)[:k]
}

func wxHotPage(cards []*model.WxArchiveCard, pn, ps int) (res []*model.WxArchiveCard, count int) {
	count = len(cards)
	start := (pn - 1) * ps
	end := start + ps - 1
	if count == 0 || count < start {
		return
	}
	if count > end {
		res = cards[start : end+1]
	} else {
		res = cards[start:]
	}
	return
}

func (s *Service) archiveWithTag(c context.Context, aids []int64) (list map[int64]*model.WxArchive, err error) {
	var arcsReply *arcmdl.ArcsReply
	if arcsReply, err = s.arcGRPC.Arcs(c, &arcmdl.ArcsRequest{Aids: aids}); err != nil {
		log.Error("WxHot s.arcGRPC.Arcs(%d) error %v", aids, err)
		return
	}
	list = make(map[int64]*model.WxArchive, len(aids))
	for _, aid := range aids {
		if arc, ok := arcsReply.Arcs[aid]; ok && arc.IsNormal() {
			wxArc := new(model.WxArchive)
			wxArc.FromArchive(arc, s.avToBv(arc.Aid))
			wxArc.Tags = _emptyWxArcTag
			list[aid] = wxArc
		}
	}
	return
}

func (s *Service) loadWxHot() {
	if s.wxHotRunning {
		return
	}
	s.wxHotRunning = true
	defer func() {
		s.wxHotRunning = false
	}()
	tmp, err := s.dao.WxHot(context.Background())
	if err != nil {
		log.Warn("loadWxHot s.dao.WxHot error(%v)", err)
		return
	}
	if len(tmp) == 0 {
		log.Warn("loadWxHot s.dao.WxHot len(tmp) == 0")
		return
	}
	s.wxHotAids = tmp
}
