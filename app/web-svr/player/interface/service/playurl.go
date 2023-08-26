package service

import (
	"context"
	hisapi "git.bilibili.co/bapis/bapis-go/community/interface/history"
	"go-common/library/ecode"
	"go-common/library/log"
	v2 "go-gateway/app/app-svr/playurl/service/api/v2"
	"go-gateway/app/web-svr/player/interface/model"
	xecode "go-gateway/ecode"

	"github.com/pkg/errors"
)

const (
	_ugcPayOtypeArc     = "archive"
	_relationPaid       = "paid"
	_defaultForceHost   = 2
	_defaultBackupNum   = 2
	_platformPC         = "pc"
	_typeMp4            = "mp4"
	_playUrlNotPayEcode = 87005
)

// Playurl get pc playurl data.
func (s *Service) Playurl(c context.Context, mid int64, arg *model.PlayurlArg) (*model.PlayurlRes, error) {
	if arg.Platform != _platformPC {
		arg.Platform = _platformPC
	}
	playurlArg := &v2.PlayURLReq{
		Aid:          arg.Aid,
		Cid:          arg.Cid,
		Qn:           arg.Qn,
		Platform:     arg.Platform,
		Fnver:        arg.Fnver,
		Fnval:        arg.Fnval,
		Mid:          mid,
		Fourk:        arg.Fourk == 1,
		VoiceBalance: arg.VoiceBalance,
		ForceHost:    _defaultForceHost,
		BackupNum:    _defaultBackupNum,
	}
	if arg.Type == _typeMp4 && playurlArg.Fnver == 0 && playurlArg.Fnval == 0 {
		// no type,use fnver and fnval
		playurlArg.Fnver = 0
		playurlArg.Fnval = 1
	}
	// min auto qn
	if playurlArg.Mid > 0 && playurlArg.Qn == 0 {
		playurlArg.Qn = s.c.Rule.AutoQn
	}
	reply, err := s.playurlV2GRPC.PlayURL(c, playurlArg)
	if err != nil {
		return nil, s.slbRetryCode(err)
	}
	if reply.Playurl.Code == _playUrlNotPayEcode {
		// 未付款
		return nil, ecode.Int(_playUrlNotPayEcode)
	}
	if reply.GetPlayurl() == nil {
		return nil, ecode.NothingFound
	}
	data := &model.PlayurlRes{}
	var lpt int64
	var lpc int64
	if mid > 0 {
		lpt, lpc = s.getLastPlayTime(c, mid, arg)
	}
	data.LastPlayTime = lpt
	data.LastPlayCid = lpc
	data.FromPlayurlV2(reply, playurlArg.Mid > 0)

	return data, nil
}

func (s *Service) PlayurlH5(c context.Context, mid int64, arg *model.PlayurlArg) (*model.PlayurlRes, error) {
	if arg.Type == _typeMp4 && arg.Fnver == 0 && arg.Fnval == 0 {
		// no type,use fnver and fnval
		arg.Fnver = 0
		arg.Fnval = 1
	}
	req := &v2.PlayURLReq{
		Aid:          arg.Aid,
		Cid:          arg.Cid,
		Qn:           arg.Qn,
		Platform:     arg.Platform,
		Fnver:        arg.Fnver,
		Fnval:        arg.Fnval,
		Mid:          mid,
		VoiceBalance: arg.VoiceBalance,
		ForceHost:    _defaultForceHost,
	}
	if arg.HighQuality > 0 {
		req.H5Hq = true
	}
	reply, err := s.playurlV2GRPC.PlayURL(c, req)
	if err != nil {
		if arg.HighQuality <= 0 {
			return nil, s.slbRetryCode(err)
		}
		// high hq back up to normal h5
		req.H5Hq = false
		if reply, err = s.playurlV2GRPC.PlayURL(c, req); err != nil {
			return nil, s.slbRetryCode(err)
		}
	}
	if reply.Playurl.Code == _playUrlNotPayEcode {
		// 未付款
		return nil, ecode.Int(_playUrlNotPayEcode)
	}
	data := &model.PlayurlRes{}
	var lpt int64
	var lpc int64
	if mid > 0 {
		lpt, lpc = s.getLastPlayTime(c, mid, arg)
	}
	data.LastPlayTime = lpt
	data.LastPlayCid = lpc
	data.FromPlayurlV2(reply, req.Mid > 0)
	return data, nil
}

// 获取上次观看进度
func (s *Service) getLastPlayTime(c context.Context, mid int64, arg *model.PlayurlArg) (int64, int64) {
	proReply, proErr := s.hisGRPC.Progress(c, &hisapi.ProgressReq{Mid: mid, Aids: []int64{arg.Aid}})
	if proErr != nil || proReply == nil {
		log.Error("PlayUrl s.hisGRPC.Progress mid:%d aid:%d error(%v)", mid, arg.Aid, proErr)
		return 0, 0
	}
	progress, ok := proReply.Res[arg.Aid]
	if !ok || progress == nil || progress.Cid <= 0 {
		return 0, 0
	}
	var (
		LastPlayTime int64
		LastPlayCid  int64
		_sec         int64 = 1000
	)
	if progress.Pro >= 0 && progress.Cid == arg.Cid {
		LastPlayTime = _sec * progress.Pro
		LastPlayCid = progress.Cid
	}
	return LastPlayTime, LastPlayCid
}

// PlayurlHls 获取播放列表 .
func (s *Service) PlayurlHls(c context.Context, mid int64, params *model.ParamHls) (*model.PlayurlHlsReply, error) {
	reply, err := s.dao.HlsScheduler(c, params, mid)
	if err != nil {
		log.Error("s.dao.HlsScheduler aid(%d,%d) error(%+v) ", params.AID, params.CID, err)
		return nil, err
	}
	if reply == nil || reply.Playurl == nil {
		return nil, ecode.NothingFound
	}
	rly := &model.PlayurlHlsReply{}
	rly.FormatPlayHls(reply.Playurl, params)
	return rly, nil
}

// HlsMaster 获取hls 音频和视频qn信息.
func (s *Service) HlsMaster(c context.Context, mid int64, params *model.ParamHls) (*model.HlsMasterReply, error) {
	reply, err := s.dao.MasterScheduler(c, params, mid)
	if err != nil {
		log.Error("s.dao.MasterScheduler aid(%d,%d) error(%+v) ", params.AID, params.CID, err)
		return nil, err
	}
	if reply == nil || reply.Info == nil {
		err = ecode.NothingFound
		log.Error("s.dao.MasterScheduler aid(%d,%d) return error(%+v) ", params.AID, params.CID, err)
		return nil, err
	}
	rly := &model.HlsMasterReply{}
	//audio为空，使用新模板
	if reply.Info.Audio == nil {
		rly.FormatPlayNoAudioMaster(reply.Info, params)
		return rly, nil
	}
	rly.FormatPlayMaster(reply.Info, params)
	return rly, nil
}

// M3U8Scheduler 根据qn获取音频或者视频的文件地址.
func (s *Service) M3U8Scheduler(c context.Context, mid int64, params *model.ParamHls) (*model.HlsMasterReply, error) {
	reply, err := s.dao.M3U8Scheduler(c, params, mid)
	if err != nil {
		log.Error("s.dao.MasterScheduler aid(%d,%d) error(%+v) ", params.AID, params.CID, err)
		return nil, err
	}
	if reply == nil || reply.Info == nil {
		err = ecode.NothingFound
		log.Error("s.dao.MasterScheduler aid(%d,%d) return error(%+v) ", params.AID, params.CID, err)
		return nil, err
	}
	rly := &model.HlsMasterReply{M3u8Data: []byte(reply.Info.M3U8Data)}
	return rly, nil
}

func (s *Service) slbRetryCode(originErr error) error {
	retryCode := []int{-500, -502, -504}
	for _, val := range retryCode {
		if ecode.EqualError(ecode.Int(val), originErr) {
			return errors.Wrapf(xecode.WebSLBRetry, "%v", originErr)
		}
	}
	return originErr
}

func (s *Service) SLBRetry(err error) bool {
	return ecode.EqualError(xecode.WebSLBRetry, err)
}
