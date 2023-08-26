package article

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/app-svr/hkt-note/common"
)

func (d *Dao) UnbindNoteReplyTaishan(ctx context.Context, cvid, rpid int64) (err error) {
	// 解关联关系 设置status为2
	cvidKey := fmt.Sprintf(common.Cvid_Mapping_Rpid_Taishan_Key, cvid)
	rpidKey := fmt.Sprintf(common.Rpid_Mapping_Cvid_Taishan_Key, rpid)
	cvidMappingRpidInfo := &common.TaishanCvidMappingRpidInfo{
		Rpid:   rpid,
		Status: common.Cvid_Rpid_Non_Attached,
	}
	rpidMappingCvidInfo := &common.TaishanRpidMappingCvidInfo{
		Cvid:   cvid,
		Status: common.Cvid_Rpid_Non_Attached,
	}
	cvidMappingRpidValue, err := json.Marshal(cvidMappingRpidInfo)
	if err != nil {
		log.Errorc(ctx, "cvidMappingRpidInfo marshal err %v and cvidMappingRpidInfo %v", err, cvidMappingRpidInfo)
		return
	}
	rpidMappingCvidValue, err := json.Marshal(rpidMappingCvidInfo)
	if err != nil {
		log.Errorc(ctx, "rpidMappingCvidInfo marshal err %v and cvidMappingRpidInfo %v", err, rpidMappingCvidInfo)
		return
	}
	if err := d.PutTaishan(ctx, cvidKey, cvidMappingRpidValue, common.TaishanConfig.NoteReply); err != nil {
		return err
	}
	if err := d.PutTaishan(ctx, rpidKey, rpidMappingCvidValue, common.TaishanConfig.NoteReply); err != nil {
		return err
	}
	return nil
}

func (d *Dao) UnbindArticleCommentTaishan(ctx context.Context, cvid, opid int64) (err error) {
	cvidKey := fmt.Sprintf(common.Cvid_Mapping_Opid_Taishan_Key, cvid)
	cvidMappingOpidInfo := &common.TaishanCvidMappingOpidInfo{
		Opid:   opid,
		Status: common.Cvid_Opid_Non_Attached,
	}
	cvidMappingOpidValue, err := json.Marshal(cvidMappingOpidInfo)
	if err != nil {
		log.Errorc(ctx, "cvidMappingRpidInfo marshal err %v and cvidMappingOpidInfo %v", err, cvidMappingOpidInfo)
		return
	}
	if err := d.PutTaishan(ctx, cvidKey, cvidMappingOpidValue, common.TaishanConfig.NoteReply); err != nil {
		return err
	}
	return nil
}
