package note

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/app-svr/hkt-note/common"
)

func (s *Service) recordCvidRpidMapping(ctx context.Context, cvid int64, rpid int64) (err error) {
	// 记录cvid到rpid的映射关系
	_ = s.recordCvidToRpid(ctx, cvid, rpid)
	// 记录rpid到cvid的映射关系
	err = s.recordRpidToCvid(ctx, cvid, rpid)
	return err
}

func (s *Service) recordCvidToRpid(ctx context.Context, cvid int64, rpid int64) (err error) {
	taiShanKey := fmt.Sprintf(common.Cvid_Mapping_Rpid_Taishan_Key, cvid)
	var value []byte
	curInfo := &common.TaishanCvidMappingRpidInfo{
		Rpid:   rpid,
		Status: common.Cvid_Rpid_Attached,
	}
	value, err = json.Marshal(curInfo)
	if err != nil {
		log.Errorc(ctx, "recordCvidMappingRpid marshal err %v and cvid %v", err, cvid)
		return err
	}
	err = s.artDao.PutTaishan(ctx, taiShanKey, value, common.TaishanConfig.NoteReply)
	return err
}

func (s *Service) recordRpidToCvid(ctx context.Context, cvid int64, rpid int64) (err error) {
	taiShanKey := fmt.Sprintf(common.Rpid_Mapping_Cvid_Taishan_Key, rpid)
	var value []byte
	curInfo := &common.TaishanRpidMappingCvidInfo{
		Cvid:   cvid,
		Status: common.Cvid_Rpid_Attached,
	}
	value, err = json.Marshal(curInfo)
	if err != nil {
		log.Errorc(ctx, "recordRpidMappingCvid marshal err %v and cvid %v", err, cvid)
		return err
	}
	err = s.artDao.PutTaishan(ctx, taiShanKey, value, common.TaishanConfig.NoteReply)
	return err
}
