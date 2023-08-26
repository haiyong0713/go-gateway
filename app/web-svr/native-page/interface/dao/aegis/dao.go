package aegis

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	aegisgrpc "git.bilibili.co/bapis/bapis-go/aegis/service"
	"go-common/library/log"

	"go-gateway/app/web-svr/native-page/interface/api"
	"go-gateway/app/web-svr/native-page/interface/conf"
	dynmdl "go-gateway/app/web-svr/native-page/interface/model/dynamic"
)

const (
	_businessID = 28
)

type Dao struct {
	client aegisgrpc.AegisServiceClient
}

func New(cfg *conf.Config) *Dao {
	client, err := aegisgrpc.NewClient(cfg.AegisClient)
	if err != nil {
		panic(err)
	}
	return &Dao{client: client}
}

func (d *Dao) AegisResourceAdd(c context.Context, req *aegisgrpc.AegisResourceAddReq) (*aegisgrpc.AegisResourceAddResp, error) {
	rly, err := d.client.AegisResourceAdd(c, req)
	if err != nil {
		log.Errorc(c, "Fail to request aegisgrpc.AegisResourceAdd, req=%+v error=%+v", req, err)
		return nil, err
	}
	return rly, nil
}

func (d *Dao) AegisResourceUpdate(c context.Context, req *aegisgrpc.AegisResourceUpdateReq) error {
	if _, err := d.client.AegisResourceUpdate(c, req); err != nil {
		log.Errorc(c, "Fail to request aegisgrpc.AegisResourceUpdate, req=%+v error=%+v", req, err)
		return err
	}
	return nil
}

func (d *Dao) AegisAddByTsReq(c context.Context, ts *dynmdl.TsSendReq) (err error) {
	defer func() {
		if err != nil {
			log.Errorc(c, "日志告警 UP主发起活动送审失败, tsReq={%+v} error=%+v", ts, err)
		}
	}()
	metadata, err := buildMetadata(ts)
	if err != nil {
		return err
	}
	var isFirstAudit int64
	if ts.IsFirstAudit {
		isFirstAudit = 1
	}
	req := &aegisgrpc.AegisResourceAddReq{
		Resource: &aegisgrpc.Resource{
			BusinessId: _businessID,
			Oid:        strconv.FormatInt(ts.TsID, 10), //待审核id
			Mid:        ts.Uid,
			Content:    ts.Title,                            //tag name
			Extra1:     ts.State,                            //话题状态 -1:待审核 0:话题待上线 1:话题已上线
			Extra2:     ts.Pid,                              //话题id
			Extra3:     isFirstAudit,                        //是否是首次审核
			Extra1S:    strconv.FormatInt(ts.AuditTime, 10), //送审时间戳
			Extra2S:    ts.Template,
			Metadata:   metadata,
			Octime:     time.Now().Format("2006-01-02 15:04:05"),
		},
	}
	for i := 0; i < 3; i++ {
		if _, err = d.AegisResourceAdd(c, req); err == nil {
			log.Warnc(c, "UP主发起活动送审成功, resource={%+v}", req.Resource)
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return err
}

func (d *Dao) AegisUpdateByTsReq(c context.Context, ts *dynmdl.TsSendReq) (err error) {
	defer func() {
		if err != nil {
			log.Errorc(c, "日志告警 UP主发起活动更新送审信息失败, tsReq={%+v} error=%+v", ts, err)
		}
	}()
	metadata, err := buildMetadata(ts)
	if err != nil {
		return err
	}
	rawUpdate := struct {
		Metadata string `json:"metadata"`
		Extra1S  string `json:"extra1s"`
	}{
		Metadata: metadata,
		Extra1S:  strconv.FormatInt(ts.AuditTime, 10),
	}
	update, err := json.Marshal(rawUpdate)
	if err != nil {
		log.Errorc(c, "Fail to marshal AegisResourceUpdateReq.Update, update=%+v error=%+v", rawUpdate, err)
		return err
	}
	req := &aegisgrpc.AegisResourceUpdateReq{
		BusinessId: _businessID,
		Oid:        strconv.FormatInt(ts.TsID, 10),
		Update:     string(update),
	}
	for i := 0; i < 3; i++ {
		if err = d.AegisResourceUpdate(c, req); err == nil {
			log.Warnc(c, "UP主发起活动更新送审信息成功, req={%+v}", req)
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return err
}

func buildMetadata(ts *dynmdl.TsSendReq) (string, error) {
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
		categoryTmp := &api.NativeModule{Category: v.Category}
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
