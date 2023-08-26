package fawkes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	xtime "go-common/library/time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"go-gateway/app/app-svr/fawkes/service/conf"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	bmdl "go-gateway/app/app-svr/fawkes/service/model/broadcast"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	_type "git.bilibili.co/bapis/bapis-go/push/service/broadcast/type"
	v2 "git.bilibili.co/bapis/bapis-go/push/service/broadcast/v2"
)

type Broadcast struct {
	broadcastClient    v2.BroadcastAPIClient
	broadcastProxyHTTP BroadcastProxyHTTP
}

// getBroadcastClient 获取APP设置的Broadcast长链示例
func (d *Dao) getBroadcastClient(serverZone int64) (client v2.BroadcastAPIClient) {
	// 星辰国际版链路
	if serverZone == appmdl.AppServerZone_Abroad {
		return &d.broadcastClient.broadcastProxyHTTP
	}
	// 默认国内Broadcast链路
	return d.broadcastClient.broadcastClient
}

// NewBroadcast 初始化
func NewBroadcast() (broadcast *Broadcast) {
	cfg := &warden.ClientConfig{
		Dial:    xtime.Duration(time.Second * 3),
		Timeout: xtime.Duration(time.Second * 3),
	}
	conn, err := warden.NewClient(cfg).Dial(context.Background(), "discovery://default/push.service.broadcast")
	if err != nil {
		panic(err)
	}
	return &Broadcast{
		broadcastClient:    v2.NewBroadcastAPIClient(conn),
		broadcastProxyHTTP: BroadcastProxyHTTP{httpClient: bm.NewClient(conf.Conf.HTTPClient)},
	}
}

// BroadcastPushOne 单设备推送
func (d *Dao) BroadcastPushOne(c context.Context, appKey, buvid string, mid int64, msg *_type.Message, expired int64) (msgId int64, err error) {
	var (
		labelFilters []*_type.LabelFilter
		vidResp      *v2.PushBuvidsReply
		midResp      *v2.PushMidsReply
	)
	// 当存在buvid. 优先使用buvid管道（ 若同时存在mid，则mid为内置参数 ）
	if buvid != "" {
		if mid != 0 {
			labelFilters = append(labelFilters, &_type.LabelFilter{
				Key:     "mid",
				Pattern: fmt.Sprintf("%d", mid),
			})
		}
		vidResp, err = d.BroadcastPushBuvids(c, appKey, []string{buvid}, msg, labelFilters, expired)
		if vidResp != nil {
			msgId = vidResp.MsgId
		}
	} else if mid != 0 {
		midResp, err = d.BroadcastPushMids(c, appKey, []int64{mid}, msg, labelFilters, expired)
		if midResp != nil {
			msgId = midResp.MsgId
		}
	}
	return
}

func (d *Dao) getBroadcastToken(targetPath string) (token string) {
	switch targetPath {
	case d.c.BroadcastGrpc.Laser.TargetPath:
		token = d.c.BroadcastGrpc.Laser.Token
	case d.c.BroadcastGrpc.LaserCommand.TargetPath:
		token = d.c.BroadcastGrpc.LaserCommand.Token
	default:
		token = ""
	}
	return
}

func (d *Dao) BroadcastPushBuvids(c context.Context, appKey string, buvids []string, msg *_type.Message, filters []*_type.LabelFilter, expired int64) (resp *v2.PushBuvidsReply, err error) {
	var (
		app          *appmdl.APP
		labelFilters []*_type.LabelFilter
		token        string
	)
	if len(buvids) == 0 {
		err = errors.New("buvids 不能为空")
		return
	}
	if app, err = d.AppPass(c, appKey); err != nil {
		return
	}
	if app == nil || app.MobiApp == "" {
		err = errors.New(fmt.Sprintf("app 信息校验失败。 app_key=%v", appKey))
		return
	}
	if token = d.getBroadcastToken(msg.TargetPath); token == "" {
		err = errors.New(fmt.Sprintf("token尚未配置，请检查Config配置。 targetPath=%v", msg.TargetPath))
		return
	}
	if len(filters) > 0 {
		labelFilters = append(labelFilters, filters...)
	}
	// 默认必传 mobi_app
	labelFilters = append(labelFilters, &_type.LabelFilter{
		Key:     "mobi_app",
		Pattern: app.MobiApp,
	})
	resp, err = d.getBroadcastClient(app.ServerZone).PushBuvids(c, &v2.PushBuvidsReq{
		Opts: &_type.PushOptions{
			AckType:      _type.PushOptions_USRE_ACK,
			LabelFilters: labelFilters,
		},
		Msg:     msg,
		Buvids:  buvids,
		Token:   token,
		Expired: expired,
	})
	log.Infoc(c, "BroadcastPushBuvid resp %v", resp)
	return
}

func (d *Dao) BroadcastPushMids(c context.Context, appKey string, mids []int64, msg *_type.Message, filters []*_type.LabelFilter, expired int64) (resp *v2.PushMidsReply, err error) {
	var (
		app          *appmdl.APP
		labelFilters []*_type.LabelFilter
		token        string
	)
	if len(mids) == 0 {
		err = errors.New("mids 不能为空")
		return
	}
	if app, err = d.AppPass(c, appKey); err != nil {
		return
	}
	if app == nil || app.MobiApp == "" {
		err = errors.New(fmt.Sprintf("app 信息校验失败。 app_key=%v", appKey))
		return
	}
	if token = d.getBroadcastToken(msg.TargetPath); token == "" {
		err = errors.New(fmt.Sprintf("token尚未配置，请检查Config配置。 targetPath=%v", msg.TargetPath))
		return
	}
	if len(filters) > 0 {
		labelFilters = append(labelFilters, filters...)
	}
	// 默认必传 mobi_app
	labelFilters = append(labelFilters, &_type.LabelFilter{
		Key:     "mobi_app",
		Pattern: app.MobiApp,
	})
	resp, err = d.getBroadcastClient(app.ServerZone).PushMids(c, &v2.PushMidsReq{
		Opts: &_type.PushOptions{
			AckType:      _type.PushOptions_USRE_ACK,
			LabelFilters: labelFilters,
		},
		Msg:     msg,
		Mids:    mids,
		Token:   token,
		Expired: expired,
	})
	log.Infoc(c, "BroadcastPushMid resp %v", resp)
	return
}

func (d *Dao) BroadcastPushAll(ctx context.Context, appKey string, msg *_type.Message) (err error) {
	var (
		app          *appmdl.APP
		labelFilters []*_type.LabelFilter
		token        string
		res          *v2.PushAllReply
	)
	if app, err = d.AppPass(ctx, appKey); err != nil {
		return
	}
	if app == nil || app.MobiApp == "" {
		err = errors.New(fmt.Sprintf("app 信息校验失败。 app_key=%v", appKey))
		return
	}
	if token = d.getBroadcastToken(msg.TargetPath); token == "" {
		err = errors.New(fmt.Sprintf("token尚未配置，请检查Config配置。 targetPath=%v", msg.TargetPath))
		return
	}
	// 默认必传 mobi_app
	labelFilters = append(labelFilters, &_type.LabelFilter{
		Key:       "mobi_app",
		Pattern:   app.MobiApp,
		MatchKind: _type.LabelFilter_EQUAL,
	})
	req := &v2.PushAllReq{
		Opts: &_type.PushOptions{
			Ratelimit:    d.c.BroadcastGrpc.Module.Ratelimit,
			LabelFilters: labelFilters,
		},
		Msg:   msg,
		Token: token,
	}
	res, err = d.getBroadcastClient(app.ServerZone).PushAll(ctx, req)
	if err != nil {
		log.Warn("broadcast push fail,req=%+v,error=%+v", req, err)
		return err
	}
	log.Warn("broadcast push success,req=%+v,msgId=%d", req, res.GetMsgId())
	return nil
}

//
//func (d *Dao) BroadcastPushBuvids(c context.Context, appKey string, buvids []string, msg *_type.Message, token string) (resp *v2.PushBuvidsReply, err error) {
//	if len(buvids) == 0 {
//		log.Error("BroadcastPushBuvids err: buvid 不能为空")
//		return
//	}
//	resp, err = d.getBroadcastClient(c, appKey).PushBuvids(c, &v2.PushBuvidsReq{
//		Opts: &_type.PushOptions{
//			AckType: _type.PushOptions_USRE_ACK,
//		},
//		Msg:    msg,
//		Buvids: buvids,
//		Token:  token,
//	})
//	if err != nil {
//		log.Error("BroadcastPushBuvids error; %v", err)
//		return
//	}
//	return
//}

//
//func (d *Dao) BroadcastPushMids(c context.Context, appKey string, mids []int64, msg *_type.Message, token string) (resp *v2.PushMidsReply, err error) {
//	log.Warn("BroadcastPushMids start")
//	if len(mids) == 0 {
//		log.Error("BroadcastPushMids err: mid 不能为空")
//		return
//	}
//	resp, err = d.getBroadcastClient(c, appKey).PushMids(c, &v2.PushMidsReq{
//		Opts: &_type.PushOptions{
//			AckType: _type.PushOptions_USRE_ACK,
//		},
//		Msg:   msg,
//		Mids:  mids,
//		Token: token,
//	})
//	if err != nil {
//		log.Error("BroadcastPushMids error; %v", err)
//		return
//	}
//	return
//}

// ========= BroadcastProxyHTTP =========

type BroadcastProxyHTTP struct {
	httpClient *bm.Client
}

func (b *BroadcastProxyHTTP) PushBuvids(ctx context.Context, in *v2.PushBuvidsReq, opts ...grpc.CallOption) (resp *v2.PushBuvidsReply, err error) {
	bs, err := json.Marshal(in)
	if err != nil {
		log.Errorc(ctx, "Decode request body failed. err[%+v]", err)
		return
	}
	var res bmdl.ProxyResp
	if res, err = b.sgpProxy(ctx, bmdl.PushBuvids, bytes.NewReader(bs)); err != nil {
		log.Errorc(ctx, "client request err[%+v]", err)
		return
	}
	if len(res.Err) != 0 {
		err = ecode.Error(ecode.ServerErr, res.Err)
		log.Errorc(ctx, "%v", err)
		return
	}
	resp = new(v2.PushBuvidsReply)
	var msgId int64
	if msgId, err = decodeMsgId(res.Response); err != nil {
		log.Errorc(ctx, "decode magID err[%+v]", err)
		return
	}
	resp.MsgId = msgId
	return
}

func (b *BroadcastProxyHTTP) PushMids(ctx context.Context, in *v2.PushMidsReq, opts ...grpc.CallOption) (resp *v2.PushMidsReply, err error) {
	bs, err := json.Marshal(in)
	if err != nil {
		log.Errorc(ctx, "Decode request body failed. err[%+v]", err)
		return
	}
	var res bmdl.ProxyResp
	if res, err = b.sgpProxy(ctx, bmdl.PushMids, bytes.NewReader(bs)); err != nil {
		log.Errorc(ctx, "client request err[%+v]", err)
		return
	}
	if len(res.Err) != 0 {
		err = ecode.Error(ecode.ServerErr, res.Err)
		log.Errorc(ctx, "%v", err)
		return
	}
	resp = new(v2.PushMidsReply)
	var msgId int64
	if msgId, err = decodeMsgId(res.Response); err != nil {
		log.Errorc(ctx, "decode magID err[%+v]", err)
		return
	}
	resp.MsgId = msgId
	return
}

func (b *BroadcastProxyHTTP) PushAll(ctx context.Context, in *v2.PushAllReq, opts ...grpc.CallOption) (resp *v2.PushAllReply, err error) {
	bs, err := json.Marshal(in)
	if err != nil {
		log.Errorc(ctx, "Decode request body failed. err[%+v]", err)
		return
	}
	var res bmdl.ProxyResp
	if res, err = b.sgpProxy(ctx, bmdl.PushAll, bytes.NewReader(bs)); err != nil {
		log.Errorc(ctx, "client request err[%+v]", err)
		return
	}
	if len(res.Err) != 0 {
		err = ecode.Error(ecode.ServerErr, res.Err)
		log.Errorc(ctx, "%v", err)
		return
	}
	resp = new(v2.PushAllReply)
	var msgId int64
	if msgId, err = decodeMsgId(res.Response); err != nil {
		log.Errorc(ctx, "decode magID err[%+v]", err)
		return
	}
	resp.MsgId = msgId
	return
}

func (b *BroadcastProxyHTTP) sgpProxy(ctx context.Context, path string, body io.Reader) (res bmdl.ProxyResp, err error) {
	sgpProxyPath := strings.Join([]string{conf.Conf.BroadcastGrpc.SGPProxy.Host, conf.Conf.BroadcastGrpc.SGPProxy.DiscoveryId}, "/") + path
	var req *http.Request
	if req, err = http.NewRequest("POST", sgpProxyPath, body); err != nil {
		log.Errorc(ctx, "client request build err %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if err = b.httpClient.JSON(ctx, req, &res); err != nil {
		log.Errorc(ctx, "client request json err %v", err)
		return
	}
	log.Infoc(ctx, "%v", res)
	return
}

func decodeMsgId(m interface{}) (msgId int64, err error) {
	m2 := m.(map[string]interface{})
	if msgId, err = strconv.ParseInt(m2["msg_id"].(string), 10, 64); err != nil {
		return 0, err
	}
	return
}

func (b *BroadcastProxyHTTP) PushRoom(ctx context.Context, in *v2.PushRoomReq, opts ...grpc.CallOption) (*v2.PushRoomReply, error) {
	return nil, nil
}

func (b *BroadcastProxyHTTP) RoomOnline(ctx context.Context, in *v2.RoomOnlineReq, opts ...grpc.CallOption) (*v2.RoomOnlineReply, error) {
	return nil, nil
}
func (b *BroadcastProxyHTTP) ListSession(ctx context.Context, in *v2.ListSessionReq, opts ...grpc.CallOption) (*v2.ListSessionReply, error) {
	return nil, nil
}

func (b *BroadcastProxyHTTP) ListServer(ctx context.Context, in *v2.ListServerReq, opts ...grpc.CallOption) (*v2.ListServerReply, error) {
	return nil, nil
}

func (b *BroadcastProxyHTTP) ListOfflineByMid(ctx context.Context, in *v2.ListOfflineByMidReq, opts ...grpc.CallOption) (*v2.ListOfflineReply, error) {
	return nil, nil
}

func (b *BroadcastProxyHTTP) DelOfflineByMid(ctx context.Context, in *v2.DelOfflineByMidReq, opts ...grpc.CallOption) (*v2.DelOfflineReply, error) {
	return nil, nil
}

func (b *BroadcastProxyHTTP) ListOfflineByBuvid(ctx context.Context, in *v2.ListOfflineByBuvidReq, opts ...grpc.CallOption) (*v2.ListOfflineReply, error) {
	return nil, nil
}

func (b *BroadcastProxyHTTP) DelOfflineByBuvid(ctx context.Context, in *v2.DelOfflineByBuvidReq, opts ...grpc.CallOption) (*v2.DelOfflineReply, error) {
	return nil, nil
}

func (b *BroadcastProxyHTTP) ListOfflineByRoom(ctx context.Context, in *v2.ListOfflineByRoomReq, opts ...grpc.CallOption) (*v2.ListOfflineReply, error) {
	return nil, nil
}

func (b *BroadcastProxyHTTP) DelOfflineByRoom(ctx context.Context, in *v2.DelOfflineByRoomReq, opts ...grpc.CallOption) (*v2.DelOfflineReply, error) {
	return nil, nil
}
