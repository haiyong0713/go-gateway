package native

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"go-common/library/log"
	"go-main/app/archive/aegis/admin/server/databus"

	pb "go-gateway/app/web-svr/native-page/interface/api"
	dynmdl "go-gateway/app/web-svr/native-page/interface/model/dynamic"
)

const (
	_aegisBusinessID int64 = 28
)

// AegisAdd .
func (d *Dao) AegisAdd(c context.Context, ts *dynmdl.TsSendReq) error {
	metadata, err := buildAegisMetadata(ts)
	if err != nil {
		return err
	}
	info := &databus.AddInfo{
		BusinessID: _aegisBusinessID,
		NetID:      _aegisBusinessID,
		OID:        strconv.FormatInt(ts.TsID, 10), //待审核id
		Content:    ts.Title,                       //tag name
		MID:        ts.Uid,
		Extra1:     ts.State,                            //话题状态 -1:待审核 0:话题待上线 1:话题已上线
		Extra2:     ts.Pid,                              //话题id
		Extra1s:    strconv.FormatInt(ts.AuditTime, 10), //送审时间戳
		Extra2s:    ts.Template,
		MetaData:   metadata,
		OCtime:     time.Now(),
	}
	log.Info("aegisAdd(%+v)", info)
	// request
	if err = databus.Add(info); err != nil {
		log.Error("dao.aegis.aegisAdd.err(%v)", err)
	}
	return err
}

func (d *Dao) AegisUpdate(c context.Context, ts *dynmdl.TsSendReq) error {
	metadata, err := buildAegisMetadata(ts)
	if err != nil {
		return err
	}
	info := &databus.UpdateInfo{
		BusinessID: _aegisBusinessID,
		NetID:      _aegisBusinessID,
		OID:        strconv.FormatInt(ts.TsID, 10),
		Update: map[string]interface{}{
			"metadata": metadata,
			"extra1s":  strconv.FormatInt(ts.AuditTime, 10),
		},
	}
	log.Info("aegisUpdate(%+v)", info)
	if err = databus.Update(info); err != nil {
		log.Error("Fail to update aegis, info=%+v error=%+v", info, err)
		return err
	}
	return nil
}

func buildAegisMetadata(ts *dynmdl.TsSendReq) (string, error) {
	if ts == nil {
		return "", nil
	}
	data := &dynmdl.SendModule{
		BgColor:      ts.BgColor,
		Url:          ts.Url,
		ShareImage:   ts.ShareImage,
		Partitions:   ts.Partitions,
		AuditContent: int64(ts.AuditContent),
		Dynamic:      ts.Dynamic,
	}
	for _, v := range ts.Modules {
		if v == nil {
			continue
		}
		categoryTmp := &pb.NativeModule{Category: v.Category}
		switch {
		case categoryTmp.IsCarouselImg():
			data.Meta = func() string {
				if len(v.Resources) == 0 || v.Resources[0].Ext == "" {
					return ""
				}
				rawExt := &dynmdl.ResourceExt{}
				if err := json.Unmarshal([]byte(v.Resources[0].Ext), rawExt); err != nil {
					log.Error("Fail to unmarshal ResourceExt, ext=%s error=%+v", v.Resources[0].Ext, err)
					return ""
				}
				return rawExt.ImgUrl
			}()
		case categoryTmp.IsStatement(): //文本组件
			data.Remark = v.Remark
		case categoryTmp.IsClick(): //自定义点击组件
			data.Meta = v.Meta
		default:
			continue
		}
	}
	metadata, err := json.Marshal(data)
	if err != nil {
		log.Error("Fail to marshal aegis metadata, data=%+v error=%+v", data, err)
		return "", err
	}
	return string(metadata), nil
}
