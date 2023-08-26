package service

import (
	"context"

	relationGRPC "git.bilibili.co/bapis/bapis-go/account/service/relation"
	"go-common/library/log"
)

// Relation .
func (s *Service) Relation(c context.Context, mid int64) (int64, error) {
	stat, err := s.relaGRPC.Stat(c, &relationGRPC.MidReq{Mid: mid})
	if err != nil {
		log.Error("Relation s.relation.Stat(mid:%d) error(%v)", mid, err)
		return 0, err
	}
	return stat.GetFollower(), nil
}

// Fans .
func (s *Service) Fans(c context.Context, mids []int64) (map[int64]*relationGRPC.StatReply, error) {
	res, err := s.relaGRPC.Stats(c, &relationGRPC.MidsReq{Mids: mids})
	if err != nil {
		log.Error("Fans(mid:%v) error(%v)", mids, err)
		return nil, err
	}
	return res.GetStatReplyMap(), nil
}
