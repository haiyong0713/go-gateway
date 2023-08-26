package dynamic

import (
	"context"
	"go-common/library/log"

	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"

	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
)

// UpListViewMore get UpListViewMore.
func (d *Dao) UpListViewMore(c context.Context, mid int64) (*dyngrpc.UpListViewMoreRsp, error) {
	UpListViewMoreReply, err := d.dynaGRPC.UpListViewMore(c, &dyngrpc.UpListViewMoreReq{Uid: mid, SortType: mdlv2.UplistMoreSortTypeRcmd})
	if err != nil || UpListViewMoreReply == nil {
		log.Errorc(c, "Dao.UpListViewMore(mid: %+v) failed. error(%+v)", mid, err)
		return nil, err
	}
	return UpListViewMoreReply, nil
}

// UpListSearch get UpListSearch.
func (d *Dao) UpListSearch(c context.Context, mid int64, name string, realIp string) (*dyngrpc.UpListSearchRsp, error) {
	UpListSearchReply, err := d.dynaGRPC.UpListSearch(c, &dyngrpc.UpListSearchReq{Mid: mid, Name: name, RealIp: realIp})
	if err != nil || UpListSearchReply == nil {
		log.Errorc(c, "Dao.UpListViewMore(mid: %+v) failed. error(%+v)", mid, err)
		return nil, err
	}
	return UpListSearchReply, nil
}
