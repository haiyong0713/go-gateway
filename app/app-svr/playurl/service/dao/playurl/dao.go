package playurl

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/stat/prom"

	pb "go-gateway/app/app-svr/playurl/service/api"
	v2 "go-gateway/app/app-svr/playurl/service/api/v2"
	"go-gateway/app/app-svr/playurl/service/conf"

	hlsgrpc "git.bilibili.co/bapis/bapis-go/video/vod/playurlhls"
	h5grpc "git.bilibili.co/bapis/bapis-go/video/vod/playurlhtml5"
	hqgrpc "git.bilibili.co/bapis/bapis-go/video/vod/playurltvproj"
	playurlgrpc "git.bilibili.co/bapis/bapis-go/video/vod/playurlugc"
	disastergrpc "git.bilibili.co/bapis/bapis-go/video/vod/playurlugcdisaster"
	volume "git.bilibili.co/bapis/bapis-go/video/vod/playurlvolume"

	"google.golang.org/grpc"
)

const (
	_platformHtml5    = "html5"
	_platformHtml5New = "html5_new"
)

// Dao dao
type Dao struct {
	client              *bm.Client
	playurlGRPC         playurlgrpc.PlayurlServiceClient
	playurlH5GRPC       h5grpc.PlayurlServiceClient
	playurlHqGRPC       hqgrpc.PlayurlServiceClient
	playurlhlsGRPC      hlsgrpc.PlayurlServiceClient
	playurlDisasterGRPC disastergrpc.PlayurlServiceClient
	volumeClient        volume.PlayurlServiceClient
	errProm             *prom.Prom
}

// New init mysql db
func New(c *conf.Config) (dao *Dao) {
	dao = &Dao{
		client:  bm.NewClient(c.HTTPClient),
		errProm: prom.BusinessErrCount,
	}
	var (
		err        error
		opts       []grpc.DialOption
		WindowSize = int32(65535000)
	)
	opts = append(opts, grpc.WithInitialWindowSize(WindowSize))
	opts = append(opts, grpc.WithInitialConnWindowSize(WindowSize))
	if dao.playurlGRPC, err = playurlgrpc.NewClient(c.PlayurlClient, opts...); err != nil {
		panic(err)
	}
	if dao.playurlH5GRPC, err = h5grpc.NewClient(c.H5PlayurlClient, opts...); err != nil {
		panic(err)
	}
	if dao.playurlHqGRPC, err = hqgrpc.NewClient(c.HqPlayurlClient, opts...); err != nil {
		panic(err)
	}
	if dao.playurlhlsGRPC, err = hlsgrpc.NewClient(c.HlsPlayurlClient, opts...); err != nil {
		panic(err)
	}
	if dao.playurlDisasterGRPC, err = disastergrpc.NewClient(c.PlayurlDisasterClient, opts...); err != nil {
		panic(err)
	}
	if dao.volumeClient, err = volume.NewClientPlayurlService(c.VolumeClient, opts...); err != nil {
		panic(err)
	}
	return
}

// Playurl is
func (d *Dao) Playurl(c context.Context, reqParam *pb.PlayURLReq, isSp, isPreview bool, reqURL string) (playurl *pb.PlayURLReply, code int, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("otype", "json")
	params.Set("buvid", reqParam.Buvid)
	params.Set("mid", strconv.FormatInt(reqParam.Mid, 10))
	params.Set("cid", strconv.FormatInt(reqParam.Cid, 10))
	params.Set("session", reqParam.Session)
	params.Set("force_host", strconv.Itoa(int(reqParam.ForceHost)))
	params.Set("fnver", strconv.Itoa(int(reqParam.Fnver)))
	params.Set("fnval", strconv.Itoa(int(reqParam.Fnval)))
	if isSp {
		params.Set("is_sp", "1")
	} else {
		params.Set("is_sp", "0")
	}
	if isPreview {
		params.Set("is_preview", "1")
	} else {
		params.Set("is_preview", "0")
	}
	params.Set("fourk", strconv.Itoa(int(reqParam.Fourk)))
	if reqParam.Type != "" {
		params.Set("type", reqParam.Type)
	}
	if reqParam.MobiApp != "" { // 客户端用mobi_app
		params.Set("platform", reqParam.MobiApp)
	} else if reqParam.Platform != "" { // pc,h5,小程序之类
		params.Set("platform", reqParam.Platform)
	}
	if reqParam.Aid > 0 {
		params.Set("avid", strconv.FormatInt(reqParam.Aid, 10))
	}
	if reqParam.Qn != 0 {
		params.Set("qn", strconv.FormatInt(reqParam.Qn, 10))
	}
	if reqParam.Npcybs != 0 {
		params.Set("npcybs", strconv.Itoa(int(reqParam.Npcybs)))
	}
	if reqParam.Dl != 0 {
		params.Set("dl", strconv.Itoa(int(reqParam.Dl)))
	}
	var res struct {
		Code int `json:"code"`
		*pb.PlayURLReply
	}
	if err = d.client.Get(c, reqURL, ip, params, &res); err != nil {
		return
	}
	playurl = res.PlayURLReply
	code = res.Code
	return
}

// SteinsPlayurl is
func (d *Dao) SteinsPlayurl(c context.Context, reqParam *pb.SteinsPreviewReq, isSp, isPreview int, reqURL string) (playurl *pb.PlayURLInfo, code int, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("otype", "json")
	params.Set("buvid", reqParam.Buvid)
	params.Set("mid", strconv.FormatInt(reqParam.Mid, 10))
	params.Set("cid", strconv.FormatInt(reqParam.Cid, 10))
	params.Set("fnver", strconv.Itoa(int(reqParam.Fnver)))
	params.Set("fnval", strconv.Itoa(int(reqParam.Fnval)))
	params.Set("is_sp", strconv.Itoa(isSp))
	params.Set("is_preview", strconv.Itoa(isPreview))
	if reqParam.Platform != "" { // pc,h5,小程序之类
		params.Set("platform", reqParam.Platform)
	}
	if reqParam.Aid > 0 {
		params.Set("avid", strconv.FormatInt(reqParam.Aid, 10))
	}
	if reqParam.Qn != 0 {
		params.Set("qn", strconv.FormatInt(reqParam.Qn, 10))
	}
	var res struct {
		Code int `json:"code"`
		*pb.PlayURLInfo
	}
	if err = d.client.Get(c, reqURL, ip, params, &res); err != nil {
		return
	}
	playurl = res.PlayURLInfo
	code = res.Code
	return
}

// PlayurlV2 .
func (d *Dao) PlayurlV2(c context.Context, param *playurlgrpc.RequestMsg, h5Hq, isView, isSp, isFreeSp bool) (res *v2.ResponseMsg, code int, err error) {
	res = new(v2.ResponseMsg)
	if param.Platform == _platformHtml5 {
		if h5Hq {
			var hqdata *hqgrpc.ResponseMsg
			hqReq := &hqgrpc.RequestMsg{
				Cid:       param.Cid,
				Qn:        param.Qn,
				Uip:       param.Uip,
				Platform:  param.Platform,
				Fnver:     param.Fnver,
				Fnval:     param.Fnval,
				Mid:       param.Mid,
				BackupNum: param.BackupNum,
				Preview:   param.Preview,
				Download:  param.Download,
				ForceHost: param.ForceHost,
				IsSp:      param.IsSp,
				Fourk:     param.Fourk,
			}
			if hqdata, err = d.playurlHqGRPC.ProtobufPlayurl(c, hqReq); err != nil {
				return nil, 0, err
			}
			res.FromPlayurlHQ(hqdata)
			code = int(res.Code)
			return
		}
		var h5data *h5grpc.ResponseMsg
		h5Req := &h5grpc.RequestMsg{
			Cid:       param.Cid,
			Qn:        param.Qn,
			Uip:       param.Uip,
			Platform:  param.Platform,
			Fnver:     param.Fnver,
			Fnval:     param.Fnval,
			Mid:       param.Mid,
			BackupNum: param.BackupNum,
			Preview:   param.Preview,
			Download:  param.Download,
			ForceHost: param.ForceHost,
			IsSp:      param.IsSp,
			Fourk:     param.Fourk,
		}
		if h5data, err = d.playurlH5GRPC.ProtobufPlayurl(c, h5Req); err != nil {
			return nil, 0, err
		}
		res.FromPlayurlH5(h5data)
		code = int(res.Code)
		return
	}
	if param.Platform == _platformHtml5New {
		param.Platform = _platformHtml5
	}
	var data *playurlgrpc.ResponseMsg
	//是新的playview页面走新接口
	if isView {
		//增加灾备逻辑
		if data, err = d.Playurl2(c, param); err != nil {
			return nil, 0, err
		}
	} else {
		//增加灾备逻辑
		if data, err = d.ProtobufPlayurl(c, param); err != nil {
			return nil, 0, err
		}
	}
	res.FromPlayurlV2(data, isView, isSp, isFreeSp)
	code = int(res.Code)
	return
}

func (d *Dao) Playurl2(c context.Context, param *playurlgrpc.RequestMsg) (*playurlgrpc.ResponseMsg, error) {
	d.errProm.Incr("disaster_playurl2")
	data, err := d.playurlGRPC.Playurl2(c, param)
	if err == nil && data != nil {
		//除了ResponseMsg.code等于10005&0之外
		if int(data.Code) == ecode.OK.Code() || data.Code == 10005 {
			return data, nil
		}
		//主服务有返回失败信息时，需要切灾备的error code 50004 50000
		if data.Code != 50004 && data.Code != 50000 {
			d.errProm.Incr("disaster_playurl2_other_err")
			return data, nil
		}
	}
	//走灾备
	d.errProm.Incr("disaster_playurl2_err")
	log.Error("日志告警:disaster d.playurlGRPC.Playurl2 灾备 req(%v) error(%v) res(%v)", param, err, data)
	disData, disErr := d.playurlDisasterGRPC.Playurl2(c, param)
	if disErr != nil || (disData != nil && (int(disData.Code) != ecode.OK.Code() && disData.Code != 10005)) {
		//灾备失败
		log.Error("日志告警:app端灾备接口出错了! disData(%v), disErr(%v)", disData, disErr)
		d.errProm.Incr("disaster_playurl2_fail_err")
	}
	return disData, disErr
}

func (d *Dao) ProtobufPlayurl(c context.Context, param *playurlgrpc.RequestMsg) (*playurlgrpc.ResponseMsg, error) {
	d.errProm.Incr("disaster_pbPlayurl")
	data, err := d.playurlGRPC.ProtobufPlayurl(c, param)
	//除了ResponseMsg.code等于10005&0之外，其他都可以走灾备
	if err == nil && data != nil {
		if int(data.Code) == ecode.OK.Code() || data.Code == 10005 {
			return data, nil
		}
		//主服务有返回失败信息时，需要切灾备的error code 50004 50000
		if data.Code != 50004 && data.Code != 50000 {
			d.errProm.Incr("disaster_pbPlayurl_other_err")
			return data, nil
		}
	}
	//走灾备
	d.errProm.Incr("disaster_pbPlayurl_err")
	log.Error("日志告警:disaster d.playurlGRPC.ProtobufPlayurl 灾备 req(%v) error(%v) res(%v)", param, err, data)
	disData, disErr := d.playurlDisasterGRPC.ProtobufPlayurl(c, param)
	if disErr != nil || (disData != nil && (int(disData.Code) != ecode.OK.Code() && disData.Code != 10005)) {
		//灾备失败
		log.Error("日志告警:web端灾备接口出错了! disData(%v), disErr(%v)", disData, disErr)
		d.errProm.Incr("disaster_pbPlayurl_fail_err")
	}
	return disData, disErr
}

// Project .
func (d *Dao) Project(c context.Context, param *hqgrpc.RequestMsg) (res *v2.ResponseMsg, code int, err error) {
	res = new(v2.ResponseMsg)
	var hqdata *hqgrpc.ResponseMsg
	if hqdata, err = d.playurlHqGRPC.ProtobufPlayurl(c, param); err != nil {
		return nil, 0, err
	}
	res.FromPlayurlHQ(hqdata)
	code = int(res.Code)
	return
}

// PlayurlVolume .
func (d *Dao) PlayurlVolume(c context.Context, cid uint64, mid uint64) (*volume.VolumeItem, error) {
	req := &volume.RequestMsg{
		Cids: []uint64{cid},
		Mid:  mid,
	}
	res, err := d.volumeClient.Volume(c, req)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, err
	}
	if int(res.Code) != ecode.OK.Code() {
		return nil, err
	}
	return res.Data[cid], nil
}
