package service

import (
	"context"
	"fmt"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	pb "go-gateway/app/web-svr/space/interface/api/v1"
	"go-gateway/app/web-svr/space/interface/model"

	accclient "git.bilibili.co/bapis/bapis-go/account/service"
)

// 白名单
func (s Service) Whitelist(c context.Context, req *pb.WhitelistReq) (res *pb.WhitelistReply, err error) {
	if res, err = s.dao.Whitelist(c, req); err != nil {
		log.Error("Whitelist req(%v) error(%v)", req, err)
		return
	}
	if res == nil {
		err = ecode.NothingFound
	}
	return
}

func (s *Service) MidInfoReply(c context.Context, mid int64) (res *accclient.CardReply, err error) {
	var midinfo *accclient.CardReply
	arg := &accclient.MidReq{
		Mid: mid,
	}
	midinfo, err = s.accClient.Card3(c, arg)
	if err != nil {
		err = fmt.Errorf("Get MidInfo error")
		log.Error("MidInfoReply req(%v) err(%v)", mid, err)
		return nil, err
	}
	if midinfo == nil {
		err = fmt.Errorf("无效Mid(%v)", mid)
		return nil, err
	}
	res = midinfo
	return
}

// 增加空间白名单
func (s *Service) AddWhitelist(c context.Context, req *pb.WhitelistAddReq) (res *pb.WhitelistAddReply, err error) {
	ok, err := s.dao.ValidWhitelistMid(c, req.Mid)
	if err != nil {
		log.Error("space.AddWhitelist mid(%d) error(%+v)", req.Mid, err)
		return
	}
	if !ok {
		err = fmt.Errorf("当前MID已存在")
		log.Error("space.AddWhitelist mid(%d) err(%+v)", req.Mid, err)
		return
	}
	midInfo, err := s.MidInfoReply(c, req.Mid)
	if err != nil {
		err = fmt.Errorf("无效的MID")
		log.Error("space.AddWhitelist get Mid(%+v) error(%+v)", req.Mid, err)
		return
	}
	now := time.Now().Unix()
	var state int
	if req.Stime.Time().Unix() <= now && req.Etime.Time().Unix() >= now {
		state = model.StatusValid
	} else if req.Etime.Time().Unix() < now {
		state = model.StatusFailed
	} else if req.Stime.Time().Unix() > now {
		state = model.StatusReady
	}
	arg := &model.WhitelistAdd{
		Mid:      req.Mid,
		MidName:  midInfo.Card.Name,
		Stime:    req.Stime,
		Etime:    req.Etime,
		State:    state,
		Username: "bussiness",
	}
	if _, err = s.dao.AddWhiteList(c, arg); err != nil {
		log.Error("space.AddWhitelist.dao.AddWhitelist add(%+v) error(%+v)", arg, err)
		return
	}
	res = &pb.WhitelistAddReply{
		AddOk: true,
	}
	return
}

// 修改白名单生效时间
func (s *Service) UpWhitelist(c context.Context, req *pb.WhitelistAddReq) (res *pb.WhitelistUpReply, err error) {
	if _, err = s.dao.UpWhitelist(c, req); err != nil {
		log.Error("space.UpWhitelist Up(%+v) error(%+v)", req, err)
		return
	}
	if err := s.dao.DelCacheWhitelist(c, req.Mid); err != nil {
		return nil, err
	}
	res = &pb.WhitelistUpReply{
		UpOk: true,
	}
	return
}

// 白名单查询时间
func (s *Service) QueryWhitelistValid(c context.Context, req *pb.WhitelistReq) (res *pb.WhitelistValidTimeReply, err error) {
	if res, err = s.dao.QueryWhitelistValid(c, req); err != nil {
		log.Error("Whitelist req(%v) error(%v)", req, err)
		return
	}
	if res == nil {
		err = ecode.NothingFound
	}
	return
}
