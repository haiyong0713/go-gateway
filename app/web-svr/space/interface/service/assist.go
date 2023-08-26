package service

import (
	"context"

	assist "git.bilibili.co/bapis/bapis-go/assist/service"
	"go-common/library/log"
)

var _emptyAssists = make([]*assist.AssistAssistUp, 0)

// RiderList get rider list by mid
func (s *Service) RiderList(c context.Context, mid int64, pn, ps int) (res *assist.AssistAssistUpsPager, err error) {
	var (
		reply *assist.AssistUpsReply
	)
	res = &assist.AssistAssistUpsPager{}
	arg := &assist.AssistUpsReq{AssistMid: mid, Pn: int64(pn), Ps: int64(ps)}
	if reply, err = s.ass.AssistUps(c, arg); err != nil {
		log.Error("s.ass.AssistUps(%d,%d,%d) error(%v)", mid, pn, ps, err)
		return
	}
	if reply == nil || reply.AssistUpsPager == nil || len(reply.AssistUpsPager.Data) == 0 {
		res.Data = _emptyAssists
		res.Pager = &assist.AssistPager{}
	} else {
		res.Data = reply.AssistUpsPager.Data
		res.Pager = reply.AssistUpsPager.Pager
	}
	return
}

// ExitRider del rider with mid and upMid
func (s *Service) ExitRider(c context.Context, mid, upMid int64) (err error) {
	if _, err = s.ass.Exit(c, &assist.ExitReq{Mid: upMid, AssistMid: mid}); err != nil {
		log.Error("s.ass.Exit(%d,%d) error(%v)", mid, upMid, err)
	}
	return
}
