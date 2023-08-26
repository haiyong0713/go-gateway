// Package common
// FM和视频的频道业务
package common

import (
	"context"
	"encoding/json"
	"sync"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/model"
	comm "go-gateway/app/app-svr/app-car/interface/model/common"
	"go-gateway/app/app-svr/app-car/interface/model/fm_v2"
)

func (s *Service) channelInfoIntegrate(ctx context.Context, req *comm.ChannelInfosReq, mid int64, buvid string,
	dev model.DeviceInfo) (*comm.ChannelInfosResp, error) { // nolint:unparam
	var (
		res = new(comm.ChannelInfosResp)
	)
	// FM频道信息
	if len(req.Fm) > 0 {
		fmInfos := make(map[int64]*comm.ChannelInfo)
		for _, chanId := range req.Fm {
			r := &fm_v2.HandleTabItemsReq{
				DeviceInfo: dev,
				Mid:        mid,
				Buvid:      buvid,
				FmType:     fm_v2.AudioVertical,
				FmId:       chanId,
			}
			tabResp, localErr := TabItemsStrategy(ctx, r)
			//log.Warnc(ctx, "channelInfoIntegrate debug TabItemsStrategy, req:%s, resp:%s, errTmp:%+v", toJson(r), toJson(tabResp), localErr)
			if localErr != nil {
				log.Errorc(ctx, "【P2】channelInfoIntegrate TabItemsStrategy localErr:%+v, req:%+v", localErr, r)
				continue
			}
			if tabResp == nil || len(tabResp.TabItems) == 0 {
				log.Errorc(ctx, "【P2】channelInfoIntegrate TabItemsStrategy empty tabItems, req:%+v", r)
				continue
			}
			fmInfos[chanId] = &comm.ChannelInfo{
				Title:    tabResp.TabItems[0].Title,
				Cover:    tabResp.TabItems[0].Cover,
				SubTitle: tabResp.TabItems[0].SubTitle,
			}
		}
		res.Fm = fmInfos
	}
	// 视频频道信息 todo del mock
	if s.v23debug(mid, dev.Build) {
		if len(req.Video) > 0 {
			infos := make(map[int64]*comm.ChannelInfo)
			for _, chanId := range req.Video {
				r := &fm_v2.HandleTabItemsReq{
					DeviceInfo: dev,
					Mid:        mid,
					Buvid:      buvid,
					FmType:     fm_v2.AudioVertical,
					FmId:       chanId,
				}
				tabResp, localErr := TabItemsStrategy(ctx, r)
				if localErr != nil {
					log.Errorc(ctx, "【P2】channelInfoIntegrate TabItemsStrategy localErr:%+v, req:%+v", localErr, r)
					continue
				}
				if tabResp == nil || len(tabResp.TabItems) == 0 {
					log.Errorc(ctx, "【P2】channelInfoIntegrate TabItemsStrategy empty tabItems, req:%+v", r)
					continue
				}
				infos[chanId] = &comm.ChannelInfo{
					Title:    tabResp.TabItems[0].Title,
					Cover:    tabResp.TabItems[0].Cover,
					SubTitle: tabResp.TabItems[0].SubTitle,
					HotRate:  2032,
					Count:    250,
				}
			}
			res.Video = infos
		}
	}
	s.fillChannelInfoAI(ctx, res)
	return res, nil
}

// fillChannelInfoAI 填充算法侧的频道信息
func (s *Service) fillChannelInfoAI(ctx context.Context, resp *comm.ChannelInfosResp) {
	if resp == nil {
		return
	}
	// FM频道的信息
	var fmChanIds = make([]int64, 0)
	for k := range resp.Fm {
		fmChanIds = append(fmChanIds, k)
	}
	infoAI, err := s.fmDao.FmChannelInfoAI(ctx, fmChanIds)
	if err != nil {
		log.Errorc(ctx, "【P2】fillChannelInfoAI s.fmDao.FmChannelInfoAI err:%+v, fmChanIds:%+v", err, fmChanIds)
		return
	}

	// 若频道信息中包含外露稿件，则需拉取稿件标题
	var (
		arcIdMap = make(map[int64]int64)
		arcIds   = make([]int64, 0)
	)
	for k, v := range resp.Fm {
		if info, ok := infoAI[k]; ok {
			v.HotRate = info.HeatScore
			v.Count = info.ArchiveCount
			if info.SubTitle != "" {
				v.SubTitle = info.SubTitle
			}
			if info.Cover != "" {
				v.Cover = info.Cover
			}
			if len(info.Avids) > 0 {
				arcIdMap[k] = info.Avids[0] // 2.3版本只露出一个稿件
				arcIds = append(arcIds, info.Avids[0])
			}
		}
	}
	if len(arcIds) == 0 {
		return
	}
	archives, err := s.archiveDao.Archives(ctx, arcIds)
	if err != nil {
		log.Errorc(ctx, "【P2】fillChannelInfoAI s.archiveDao.Archives err:%+v, arcIds:%+v", err, arcIds)
		return
	}
	for k, v := range resp.Fm {
		if _, ok := arcIdMap[k]; !ok {
			continue
		}
		aid := arcIdMap[k]
		if _, ok := archives[aid]; !ok {
			continue
		}
		v.SubTitle = archives[aid].Title // 内容透出卡，副标题为透出稿件的标题
	}

	// 视频频道的信息 TODO DELAY
}

func (s *Service) channelArcIntegrate(ctx context.Context, req *comm.ChannelArcsReq, mid int64, buvid string, dev model.DeviceInfo) (*comm.ChannelArcsResp, error) {
	var (
		resp = new(comm.ChannelArcsResp)
		err  error
	)
	eg := errgroup.WithContext(ctx)
	// FM频道稿件拉取
	if len(req.Fm) > 0 {
		var lock sync.Mutex
		resp.Fm = make(map[int64]*comm.ChannelArcs)
		for _, v := range req.Fm {
			chReq := v
			eg.Go(func(ctx context.Context) error {
				var (
					nextBytes = make([]byte, 0)
					errTmp    error
					hdrResp   *fm_v2.HandlerResp
					pageNext  *comm.ChannelPageInfo
				)
				if chReq.PageNext != nil {
					nextBytes, _ = json.Marshal(chReq.PageNext)
				}
				r := &fm_v2.FmListParam{
					DeviceInfo: dev,
					Mid:        mid,
					Buvid:      buvid,
					FmType:     fm_v2.AudioVertical,
					FmId:       chReq.ChanId,
					PageNext:   string(nextBytes),
					Ps:         chReq.Ps,
				}

				hdrResp, errTmp = FmListParamStrategy(ctx, r)
				//log.Warnc(ctx, "channelArcIntegrate debug FmListParamStrategy, req:%s, resp:%s, errTmp:%+v", toJson(r), toJson(hdrResp), errTmp)
				if errTmp != nil {
					log.Warnc(ctx, "【P2】channelArcIntegrate FmListParamStrategy err:%+v, req:%+v", errTmp, r)
					return nil
				}
				if hdrResp.HasNext && hdrResp.PageNext != nil {
					pageNext = &comm.ChannelPageInfo{
						Ps: hdrResp.PageNext.Ps,
						Pn: int(hdrResp.PageNext.Pn),
					}
				}
				// 社区风险内容过滤
				hdrResp.OidList = s.SixLimitFilter(ctx, hdrResp.OidList)

				lock.Lock()
				resp.Fm[chReq.ChanId] = &comm.ChannelArcs{
					Aids:     hdrResp.OidList,
					PageNext: pageNext,
					HasNext:  hdrResp.HasNext,
				}
				lock.Unlock()
				return nil
			})
		}
	}
	// 视频频道稿件拉取 todo del mock
	if s.v23debug(mid, dev.Build) {
		if len(req.Video) > 0 {
			var lock sync.Mutex
			resp.Video = make(map[int64]*comm.ChannelArcs)
			for _, v := range req.Video {
				chReq := v
				eg.Go(func(ctx context.Context) error {
					var (
						nextBytes = make([]byte, 0)
						errTmp    error
						hdrResp   *fm_v2.HandlerResp
						pageNext  *comm.ChannelPageInfo
					)
					if chReq.PageNext != nil {
						nextBytes, _ = json.Marshal(chReq.PageNext)
					}
					r := &fm_v2.FmListParam{
						DeviceInfo: dev,
						Mid:        mid,
						Buvid:      buvid,
						FmType:     fm_v2.AudioVertical,
						FmId:       chReq.ChanId,
						PageNext:   string(nextBytes),
						Ps:         chReq.Ps,
					}

					hdrResp, errTmp = FmListParamStrategy(ctx, r)
					if errTmp != nil {
						log.Warnc(ctx, "【P2】channelArcIntegrate FmListParamStrategy err:%+v, req:%+v", errTmp, r)
						return nil
					}
					if hdrResp.HasNext && hdrResp.PageNext != nil {
						pageNext = &comm.ChannelPageInfo{
							Ps: hdrResp.PageNext.Ps,
							Pn: int(hdrResp.PageNext.Pn),
						}
					}

					lock.Lock()
					resp.Video[chReq.ChanId] = &comm.ChannelArcs{
						Aids:     hdrResp.OidList,
						PageNext: pageNext,
						HasNext:  hdrResp.HasNext,
					}
					lock.Unlock()
					return nil
				})
			}
		}
	}

	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "channelArcIntegrate err:%+v, req:%+v", err, req)
		return nil, err
	}
	return resp, nil
}
