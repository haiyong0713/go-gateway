package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/app-thirdsdk/interface/internal/model"
	"go-gateway/pkg/idsafe/bvid"

	arcpush "git.bilibili.co/bapis/bapis-go/manager/service/archive-push"

	camp "git.bilibili.co/bapis/bapis-go/video/vod/playurlcamp"

	dm "git.bilibili.co/bapis/bapis-go/bilibili/community/service/dm/v1"
)

func (s *Service) PlayURL(ctx context.Context, params *model.PlayURLParam) (*model.PlayURLMsg, error) {
	var plat arcpush.Plat_Enum
	switch params.Platform {
	case "android":
		plat = arcpush.Plat_ANDROID
	case "ios":
		plat = arcpush.Plat_IOS
	default:
	}
	validReply, err := s.arcPushClient.ValidateArchiveToPlay(ctx, &arcpush.ValidateArchiveToPlayReq{Bvid: params.Bvid, Plat: plat, Bundle: params.SDKIdentifier})
	if err != nil {
		log.Error("%+v", err)
	}
	if validReply != nil && !validReply.GetValid() {
		return nil, ecode.NothingFound
	}
	aid, err := bvid.BvToAv(params.Bvid)
	if err != nil {
		return nil, err
	}
	arc, err := s.dao.Archive(ctx, aid)
	if err != nil {
		return nil, err
	}
	if !arc.IsNormal() {
		return nil, ecode.NothingFound
	}
	if arc.GetFirstCid() == 0 {
		return nil, ecode.NothingFound
	}
	req := &camp.RequestMsg{
		Cid:       uint64(arc.GetFirstCid()),
		Qn:        params.Qn,
		Uip:       metadata.String(ctx, metadata.RemoteIP),
		Platform:  params.Platform,
		Fnver:     params.Fnver,
		Fnval:     params.Fnval,
		ForceHost: params.ForceHost,
		BackupNum: 2, //客户端请求默认2个
	}
	reply, err := s.dao.ProtobufPlayurl(ctx, req)
	if err != nil {
		return nil, err
	}
	res := &model.PlayURLMsg{
		Aid:          aid,
		Cid:          arc.GetFirstCid(),
		Quality:      reply.Quality,
		Format:       reply.Format,
		Timelength:   reply.Timelength,
		VideoCodecid: reply.VideoCodecid,
	}
	if len(reply.GetSupportFormats()) == 0 {
		return res, nil
	}
	if len(reply.GetDurl()) != 0 {
		res.StreamList = formatPlayDurl(reply)
		return res, nil
	}
	res.StreamList, res.DashAudio = formatPlayDash(reply, params)
	return res, nil
}

// nolint:gomnd
func formatPlayDurl(plVod *camp.ResponseMsg) []*model.Stream {
	var res []*model.Stream
	tmpDurl := &model.SegmentVideo{}
	for _, v := range plVod.Durl {
		backupURL := v.BackupUrl
		if len(v.BackupUrl) > 2 {
			backupURL = v.BackupUrl[:2]
		}
		tmpDurl.Segment = append(tmpDurl.Segment, &model.Segment{
			Order:     v.Order,
			Length:    v.Length,
			Size:      v.Size_,
			URL:       v.Url,
			BackupURL: backupURL,
			Md5:       v.Md5,
		})
	}
	// 清晰度列表
	for _, tVal := range plVod.SupportFormats {
		if tVal == nil {
			continue
		}
		var attr int64
		tmpStream := &model.Stream{
			StreamInfo: &model.StreamInfo{
				Quality:        tVal.Quality,
				Format:         tVal.Format,
				Attribute:      model.SetQnAttr(attr, tVal.Quality),
				NewDescription: tVal.NewDescription,
				DisplayDesc:    tVal.DisplayDesc,
			},
		}
		// 当前可播放清晰度
		if tVal.Quality == plVod.Quality {
			tmpStream.StreamInfo.Intact = true
			tmpStream.SegmentVideo = tmpDurl
		}
		res = append(res, tmpStream)
	}
	return res
}

// nolint:gomnd
func formatPlayDash(plVod *camp.ResponseMsg, arg *model.PlayURLParam) ([]*model.Stream, []*model.DashItem) {
	var (
		res       []*model.Stream
		dashAudio []*model.DashItem
	)
	// 音频信息默认给第一个
	var defaultAudioId uint32
	if len(plVod.Dash.Audio) > 0 {
		for _, aVal := range plVod.Dash.Audio {
			backupURL := aVal.BackupUrl
			if len(aVal.BackupUrl) > 2 {
				backupURL = aVal.BackupUrl[:2]
			}
			tmpAudio := &model.DashItem{
				ID:        aVal.Id,
				BaseURL:   aVal.BaseUrl,
				BackupURL: backupURL,
				Bandwidth: aVal.Bandwidth,
				Codecid:   aVal.Codecid,
				Md5:       aVal.Md5,
				Size:      aVal.Size_,
			}
			dashAudio = append(dashAudio, tmpAudio)
		}
		defaultAudioId = plVod.Dash.Audio[0].Id
	}
	// 每路清晰度对应的播放地址和音频信息
	fnVideo := make(map[uint32]*model.DashVideo)
	if len(plVod.Dash.Video) > 0 {
		var tmpVideo = make(map[uint32][]*model.DashVideo)
		for _, v := range plVod.Dash.Video {
			backupURL := v.BackupUrl
			if len(v.BackupUrl) > 2 {
				backupURL = v.BackupUrl[:2]
			}
			tmpVideo[v.Id] = append(tmpVideo[v.Id], &model.DashVideo{
				BaseURL:   v.BaseUrl,
				BackupURL: backupURL,
				Bandwidth: v.Bandwidth,
				Codecid:   v.Codecid,
				Md5:       v.Md5,
				Size:      v.Size_,
				AudioID:   defaultAudioId, // 音频信息
				NoRexcode: v.NoRexcode == 1,
			})
		}
		// 找到期望的编码格式,则过滤多余的codecid
		needPv := make(map[uint32]*model.DashVideo)
		// 兜底清晰度
		defaultPv := make(map[uint32]*model.DashVideo)
		for pk, pv := range tmpVideo {
			for _, tv := range pv {
				if tv == nil {
					continue
				}
				if tv.Codecid == arg.PreferCodecID {
					needPv[pk] = tv
				}
				if tv.Codecid == model.CodeH264 {
					defaultPv[pk] = tv
				}
			}
		}
		for pk := range tmpVideo {
			if nVal, ok := needPv[pk]; ok {
				fnVideo[pk] = nVal
				continue
			}
			if dval, kk := defaultPv[pk]; kk {
				fnVideo[pk] = dval
			}
		}
	}
	// 清晰度列表 拼接信息
	for _, tVal := range plVod.SupportFormats {
		if tVal == nil {
			continue
		}
		var attr int64
		tmpStream := &model.Stream{
			StreamInfo: &model.StreamInfo{
				Quality:        tVal.Quality,
				Format:         tVal.Format,
				Attribute:      model.SetQnAttr(attr, tVal.Quality),
				NewDescription: tVal.NewDescription,
				DisplayDesc:    tVal.DisplayDesc,
			},
		}
		// 视频云返回了对应的播放地址
		if _, ok := fnVideo[tVal.Quality]; ok {
			tmpStream.StreamInfo.Intact = true
			tmpStream.StreamInfo.NoRexcode = fnVideo[tVal.Quality].NoRexcode
			// 同一清晰度只返回一路
			tmpStream.DashVideo = fnVideo[tVal.Quality]
		}
		res = append(res, tmpStream)
	}
	return res, dashAudio
}

func (s *Service) DmSeg(ctx context.Context, params *model.DmSegParam) (*dm.DmSegSDKReply, error) {
	reply, err := s.arcPushClient.GetVendorConfigs(ctx, &arcpush.GetVendorConfigsReq{Bundles: []string{params.SDKIdentifier}})
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	if len(reply.GetItems()) == 0 {
		return nil, ecode.NothingFound
	}
	var ok bool
	for _, v := range reply.GetItems()[0].GetAppConfigs() {
		if v.GetBundle() == params.SDKIdentifier {
			ok = true
			break
		}
	}
	var showDanmaku bool
	if ok {
		showDanmaku = reply.GetItems()[0].GetDanmaku()
	}
	if !showDanmaku {
		return nil, ecode.NothingFound
	}
	return s.dmClient.DmSegSDK(ctx, &dm.DmSegSDKReq{Pid: params.Pid, Oid: params.Oid, Type: params.Type, SegmentIndex: params.SegmentIndex})
}
