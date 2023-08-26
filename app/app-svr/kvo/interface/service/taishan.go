package service

import (
	"context"
	"encoding/json"
	"fmt"

	xecode "go-gateway/app/app-svr/kvo/ecode"
	pb "go-gateway/app/app-svr/kvo/interface/api"

	"go-common/library/database/taishan"
	"go-common/library/log"
)

func taishanUcKey(mid int64, buvid string, moduleKeyID int) []byte {
	if mid == 0 {
		return []byte(fmt.Sprintf("{%s}_%d_bc", buvid, moduleKeyID))
	}
	return []byte(fmt.Sprintf("{%d}_%d_uc", mid, moduleKeyID))
}

func (s *Service) userDocTaiShan(ctx context.Context, mid int64, buvid string, moduleKeyID int) (rm json.RawMessage, err error) {
	var (
		key = taishanUcKey(mid, buvid, moduleKeyID)
		req = &taishan.GetReq{
			Table: s.cfg.Taishan.Table,
			Auth: &taishan.Auth{
				Token: s.cfg.Taishan.Token,
			},
			Record: &taishan.Record{Key: key},
		}
		resp *taishan.GetResp
	)
	if resp, err = s.taishan.Get(ctx, req); err != nil {
		log.Error("s.userDocTaiShan (key:%s) err(%v)", key, err)
		return
	}
	if resp.GetRecord().GetStatus().GetErrNo() == 404 {
		rm = nil
		err = nil
		return
	}
	if resp.GetRecord().GetStatus().GetErrNo() != 0 {
		log.Error("s.userDocTaiShan (key:%s) err(%v)", key, resp.Record.Status)
		return
	}
	cfg := pb.NewConfig(moduleKeyID, nil)
	if cfg == nil {
		err = xecode.KvoModuleNotExist
		return
	}
	if err = cfg.Unmarshal(resp.Record.Columns[0].Value); err != nil {
		log.Error("s.userDocTaiShan Unmarshal(key:%s) err(%v)", key, err)
		return
	}
	if rm, err = json.Marshal(cfg); err != nil {
		log.Error("s.userDocTaiShan json.Marshal(key:%s) err(%v)", key, err)
		return
	}
	return
}

func (s *Service) addUserDocTaiShan(ctx context.Context, mid int64, buvid string, moduleKeyID int, data interface{}) (err error) {
	var (
		key = taishanUcKey(mid, buvid, moduleKeyID)
		req = &taishan.PutReq{
			Table: s.cfg.Taishan.Table,
			Auth: &taishan.Auth{
				Token: s.cfg.Taishan.Token,
			},
		}
		value []byte
		resp  *taishan.PutResp
	)
	cfg := pb.NewConfig(moduleKeyID, data)
	if cfg == nil {
		err = xecode.KvoModuleNotExist
		return
	}
	if value, err = cfg.Marshal(); err != nil {
		log.Error("pb.DanmuPlayerConfig data(%+v) err(%v)", data, err)
		return
	}
	req.Record = &taishan.Record{Key: key, Columns: []*taishan.Column{{Value: value}}}
	if resp, err = s.taishan.Put(ctx, req); err != nil {
		log.Error("s.addUserDocTaiShan (key:%s, data:%+v) err(%v)", key, data, err)
		return
	}
	if resp.GetStatus().GetErrNo() != 0 {
		log.Error("s.addUserDocTaiShan (key:%s, data:%+v) err(%v)", key, data, resp.Status)
		return
	}
	return
}
