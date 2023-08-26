package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-player/interface/model"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

// PlayurlHls 获取播放列表 .
func (s *Service) PlayurlHls(c context.Context, mid int64, params *model.ParamHls) (*model.PlayurlHlsReply, error) {
	reply, err := s.playURLDao.HlsScheduler(c, params, mid)
	if err != nil {
		log.Error("s.playURLDao.HlsScheduler aid(%d,%d) error(%+v) ", params.AID, params.CID, err)
		return nil, err
	}
	if reply == nil || reply.Playurl == nil {
		return nil, ecode.NothingFound
	}
	rly := &model.PlayurlHlsReply{}
	rly.FormatPlayHls(reply.Playurl, params, s.c.HlsSign.Key, s.c.HlsSign.Secret)
	return rly, nil
}

// HlsMaster 获取hls 音频和视频qn信息.
func (s *Service) HlsMaster(c context.Context, mid int64, params *model.ParamHls) (*model.HlsMasterReply, error) {
	reply, err := s.playURLDao.MasterScheduler(c, params, mid)
	if err != nil {
		log.Error("s.playURLDao.MasterScheduler aid(%d,%d) error(%+v) ", params.AID, params.CID, err)
		return nil, err
	}
	if reply == nil || reply.Info == nil {
		err = ecode.NothingFound
		log.Error("s.playURLDao.MasterScheduler aid(%d,%d) return error(%+v) ", params.AID, params.CID, err)
		return nil, err
	}
	rly := &model.HlsMasterReply{}
	//版本判断
	isMulitVer := feature.GetBuildLimit(c, "service.playurlMulitHls", &feature.OriginResutl{
		MobiApp:    params.MobiApp,
		Build:      int64(params.Build),
		Device:     params.Device,
		BuildLimit: (params.MobiApp == "iphone" && params.Build >= 62800300) || (params.MobiApp == "ipad" && params.Build >= 31900100),
	})
	if isMulitVer && len(reply.Info.Videos) > 0 {
		rly.FormatMultPlayMaster(reply.Info, params, s.c.HlsSign.Key, s.c.HlsSign.Secret)
	} else {
		//audio为空，使用新模板
		if reply.Info.Audio == nil {
			rly.FormatPlayNoAudioMaster(reply.Info, params, s.c.HlsSign.Key, s.c.HlsSign.Secret)
		} else {
			rly.FormatPlayMaster(reply.Info, params, s.c.HlsSign.Key, s.c.HlsSign.Secret)
		}
	}
	return rly, nil
}

// M3U8Scheduler 根据qn获取音频或者视频的文件地址.
func (s *Service) M3U8Scheduler(c context.Context, mid int64, params *model.ParamHls) (*model.HlsMasterReply, error) {
	reply, err := s.playURLDao.M3U8Scheduler(c, params, mid)
	if err != nil {
		log.Error("s.playURLDao.MasterScheduler aid(%d,%d) error(%+v) ", params.AID, params.CID, err)
		return nil, err
	}
	if reply == nil || reply.Info == nil {
		err = ecode.NothingFound
		log.Error("s.playURLDao.MasterScheduler aid(%d,%d) return error(%+v) ", params.AID, params.CID, err)
		return nil, err
	}
	rly := &model.HlsMasterReply{M3u8Data: []byte(reply.Info.M3U8Data)}
	return rly, nil
}
