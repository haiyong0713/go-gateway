package playurl

import (
	"context"

	"go-common/library/log"
	xecode "go-gateway/app/app-svr/app-car/ecode"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/bangumi"
	"go-gateway/app/app-svr/app-car/interface/model/common"
	"go-gateway/app/app-svr/app-car/interface/model/playurl"
)

const (
	_platformHtml5    = "html5"
	_platformHtml5New = "html5_new"
	_mp4              = "mp4"
	_flv              = "FLV"
)

func (s *Service) PlayUrlWeb(c context.Context, buvid, cookie, referer string, mid int64, param *playurl.Param) (*playurl.Info, string, error) {
	if param.Qn == 0 {
		param.Qn = s.c.Custom.DefaultQn
	}
	switch param.Otype {
	case model.GotoAv, string(common.ItemTypeUGC), string(common.ItemTypeUGCSingle), string(common.ItemTypeUGCMulti),
		string(common.ItemTypeVideoSerial), string(common.ItemTypeVideoChannel), string(common.ItemTypeFmSerial), string(common.ItemTypeFmChannel):
		data, err := s.ugcPlayUrlWeb(c, buvid, mid, param)
		return data, "", err
	default:
		return s.bangumiPlayUrlWeb(c, buvid, cookie, referer, param)
	}
}

func (s *Service) bangumiPlayUrlWeb(c context.Context, buvid, cookie, referer string, param *playurl.Param) (*playurl.Info, string, error) {
	var (
		reply *bangumi.PlayInfo
		msg   string
		err   error
	)
	if param.VideoType == _mp4 {
		reply, msg, err = s.bgm.PlayurlProj(c, buvid, cookie, referer, param)
	} else {
		reply, msg, err = s.bgm.PlayurlH5(c, buvid, cookie, referer, param)
	}
	if err != nil {
		log.Error("bangumiPlayUrlWeb s.bgm.Playurl season_id(%d) epid(%d) error(%v)", param.Oid, param.Cid, err)
		return nil, msg, err
	}
	res := &playurl.Info{}
	res.FromPGC(reply, true)
	// 如果是flv直接返回不能播放
	if res.VideoType == _flv {
		return nil, "", xecode.AppCannotPlay
	}
	return res, "", nil
}

func (s *Service) ugcPlayUrlWeb(c context.Context, buvid string, mid int64, param *playurl.Param) (*playurl.Info, error) {
	if param.Platform == _platformHtml5 && param.VideoType != _mp4 {
		param.Platform = _platformHtml5New
	}
	reply, err := s.player.PlayURL(c, buvid, mid, param)
	if err != nil {
		log.Error("ugcPlayUrlWeb s.player.PlayURL aid(%d) cid(%d) error(%v)", param.Oid, param.Cid, err)
		return nil, err
	}
	res := &playurl.Info{}
	res.FromUGC(reply, true)
	// 如果是flv直接返回不能播放
	if res.VideoType == _flv {
		return nil, xecode.AppCannotPlay
	}
	return res, nil
}
