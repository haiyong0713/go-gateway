package service

import (
	"context"

	acccli "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	lmdl "go-gateway/app/web-svr/activity/admin/model"
	"go-gateway/pkg/idsafe/bvid"
)

// Archives get achives info .
func (s *Service) Archives(c context.Context, aids []int64) (res map[int64]*lmdl.BvArc, err error) {
	var (
		arcs *arcmdl.ArcsReply
	)
	if arcs, err = s.arcClient.Arcs(c, &arcmdl.ArcsRequest{Aids: aids}); err != nil {
		log.Error("s.arcClient.Archives3(%v) error(%v)", aids, err)
		return
	}
	res = make(map[int64]*lmdl.BvArc, len(aids))
	for _, aid := range aids {
		if arc, ok := arcs.Arcs[aid]; ok && arc.IsNormal() {
			res[aid] = &lmdl.BvArc{Arc: arc, Bvid: s.avToBv(arc.Aid)}
		}
	}
	return
}

func (s *Service) avToBv(aid int64) (bvID string) {
	var err error
	if bvID, err = bvid.AvToBv(aid); err != nil {
		log.Warn("avToBv(%d) error(%v)", aid, err)
	}
	return
}

// Accounts .
func (s *Service) Accounts(c context.Context, mids []int64) (res map[int64]*acccli.Info, err error) {
	var (
		rly *acccli.InfosReply
	)
	if rly, err = s.accClient.Infos3(c, &acccli.MidsReq{Mids: mids}); err != nil {
		return
	}
	res = rly.Infos
	for _, v := range res {
		//不展示生日信息
		v.Birthday = 0
	}
	return
}
