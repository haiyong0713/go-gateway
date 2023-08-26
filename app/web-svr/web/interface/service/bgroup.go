package service

import (
	"context"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/web/interface/model"
	"strconv"

	bgroupgrpc "git.bilibili.co/bapis/bapis-go/platform/service/bgroup/v2"
)

func (s *Service) MemberIn(c context.Context, req *model.MemberInReq) (*bgroupgrpc.MemberInReply_MemberInReplySingle, error) {
	var (
		err        error
		member     string
		res        *bgroupgrpc.MemberInReply
		groupInfos []*bgroupgrpc.MemberInReq_MemberInReqSingle
	)
	groupInfo := &bgroupgrpc.MemberInReq_MemberInReqSingle{
		Name:     req.Name,
		Business: req.Business,
		Version:  req.Version,
	}
	groupInfos = append(groupInfos, groupInfo)
	if req.Dimension == int(bgroupgrpc.Mid) {
		member = strconv.FormatInt(req.Mid, 10)
	} else {
		member = req.Buvid
	}
	memReq := &bgroupgrpc.MemberInReq{
		Groups:    groupInfos,
		Member:    member,
		Dimension: bgroupgrpc.Dimension(req.Dimension),
	}
	if res, err = s.bgroupGRPC.MemberIn(c, memReq); err != nil {
		log.Error("【@MemberIn】bgroupGRPC.MemberIn error: (%v)", err)
		return nil, err
	}
	if len(res.Results) > 0 {
		return res.Results[0], nil
	}
	return nil, ecode.Error(ecode.NothingFound, "没有找到要查找的人群包")
}
