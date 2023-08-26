package dynamic

import (
	"context"

	"go-common/library/log"
	xecode "go-gateway/app/app-svr/app-dynamic/ecode"

	pplApi "go-gateway/app/app-svr/app-show/interface/api"
)

func (d *Dao) PopularIndexSv(c context.Context, idx int64, entranceID int64) (res *pplApi.IndexSVideoReply, err error) {
	req := &pplApi.IndexSVideoReq{
		EntranceId: entranceID,
		Index:      idx,
	}
	res, err = d.popularGRPC.IndexSVideo(c, req)
	if err != nil {
		log.Errorc(c, "PopularIndexSv idx(%d) entranceID(%d) error(%+v) ", idx, entranceID, err)
		return nil, err
	}
	if res == nil {
		log.Errorc(c, "PopularIndexSv idx(%d) entranceID(%d) res nil", idx, entranceID)
		return nil, xecode.PopularGRPCRecordNotFound
	}
	return res, nil
}

func (d *Dao) PopularAggrSv(c context.Context, idx int64, hotwordID int64) (res *pplApi.AggrSVideoReply, err error) {
	req := &pplApi.AggrSVideoReq{
		HotwordId: hotwordID,
		Index:     idx,
	}
	res, err = d.popularGRPC.AggrSVideo(c, req)
	if err != nil {
		log.Errorc(c, "PopularAggrSv idx(%d) hotwordID(%d) error(%+v) ", idx, hotwordID, err)
		return nil, err
	}
	if res == nil {
		log.Errorc(c, "PopularIndexSv idx(%d) hotwordID(%d) res nil", idx, hotwordID)
		return nil, xecode.PopularGRPCRecordNotFound
	}
	return res, nil
}
