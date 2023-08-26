package dao

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-listener/interface/internal/model"

	accSvc "git.bilibili.co/bapis/bapis-go/account/service"
)

func (d *dao) UpInfoByMids(ctx context.Context, mids interface{}, ip string) (ret map[int64]model.MemberInfo, err error) {
	var midSlc []int64
	switch typd := mids.(type) {
	case []int64:
		midSlc = typd
	case map[int64]struct{}:
		for k := range typd {
			midSlc = append(midSlc, k)
		}
	default:
		panic(fmt.Sprintf("programmer error: unknown type %T", mids))
	}
	resp, err := d.accGRPC.Infos3(ctx, &accSvc.MidsReq{Mids: midSlc, RealIp: ip})
	if err != nil {
		return nil, wrapDaoError(err, "accGRPC.Infos3", midSlc)
	}
	ret = make(map[int64]model.MemberInfo)
	for k, v := range resp.GetInfos() {
		ret[k] = model.MemberInfo{Info: v}
	}
	return
}

func (d *dao) UpInfoStatByMid(ctx context.Context, mid int64, ip string) (ret *model.MemberInfo, err error) {
	resp, err := d.accGRPC.ProfileWithStat3(ctx, &accSvc.MidReq{Mid: mid, RealIp: ip})
	if err != nil {
		return nil, wrapDaoError(err, "accGRPC.ProfileWithStat3", mid)
	}

	return &model.MemberInfo{ProfileStatReply: resp}, err
}
