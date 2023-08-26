package playurl

import (
	"context"
	"time"

	"go-common/library/log"
	infoc2 "go-common/library/log/infoc.v2"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/common"
	"go-gateway/app/app-svr/app-car/interface/model/playurl"
)

const (
	_dolbyUgcType = "ugc"
	_dolbyScene   = "transfer"
)

func (s *Service) PlayUrl(c context.Context, buvid, referer string, mid int64, param *playurl.Param) (*playurl.Info, string, error) {
	if param.Qn == 0 {
		param.Qn = s.c.Custom.DefaultQn
	}
	//干预检索请求,附带杜比请求
	param.Fnval = param.Fnval | model.FnvalDolby
	switch param.Otype {
	case model.GotoAv, string(common.ItemTypeUGC), string(common.ItemTypeUGCSingle), string(common.ItemTypeUGCMulti),
		string(common.ItemTypeVideoSerial), string(common.ItemTypeVideoChannel), string(common.ItemTypeFmSerial), string(common.ItemTypeFmChannel):
		data, err := s.ugcPlayUrl(c, buvid, mid, param)
		return data, "", err
	default:
		return s.bangumiPlayUrl(c, buvid, referer, param)
	}
}

func (s *Service) ugcPlayUrl(c context.Context, buvid string, mid int64, param *playurl.Param) (*playurl.Info, error) {
	param.Oid = param.Aid
	reply, err := s.player.PlayURL(c, buvid, mid, param)
	if err != nil {
		log.Error("ugcPlayUrl s.player.PlayURL aid(%d) cid(%d) error(%v)", param.Oid, param.Cid, err)
		return nil, err
	}
	res := &playurl.Info{}
	res.FromUGC(reply, false)
	return res, nil
}

func (s *Service) bangumiPlayUrl(c context.Context, buvid, referer string, param *playurl.Param) (*playurl.Info, string, error) {
	param.Oid = param.SeasonID
	param.Cid = param.EpID
	reply, msg, err := s.bgm.PlayurlAPP(c, buvid, referer, param)
	if err != nil {
		log.Error("bangumiPlayUrl s.bgm.Playurl season_id(%d) epid(%d) error(%v)", param.Oid, param.Cid, err)
		return nil, msg, err
	}
	res := &playurl.Info{}
	res.FromPGC(reply, false)
	return res, "", nil
}

// EventReport 上报杜比日志
func (s *Service) EventReport(ctx context.Context, param *common.EventReportReq) {
	payload := infoc2.NewLogStream("007250", param.Buvid, param.Mid,
		time.Now().Format("2006-01-02 15:04:05"), param.MobiApp, param.Platform,
		param.Build, param.Avid, param.Cid, _dolbyUgcType, param.Scene)
	s.makeEventReport(ctx, nil, _dolbyScene)
	infocV2, _ := infoc2.New(nil)
	err := infocV2.Info(context.Background(), payload)
	if err != nil {
		log.Errorc(ctx, "日志告警 Failed to report dolby info error(%+v)", err)
	}
}
func (s *Service) makeEventReport(ctx context.Context, param *playurl.Param, scene string) {
	rep := new(common.EventReportReq)
	rep.Buvid = param.Buvid
	rep.Mid = param.Mid
	rep.Ctime = time.Now().Format("2006-01-02 15:04:05")
	rep.MobiApp = param.MobiApp
	rep.Platform = param.Platform
	rep.Build = param.Build
	rep.Avid = param.Oid
	rep.Cid = param.Cid
	rep.Scene = scene
	s.EventReport(ctx, rep)
}
