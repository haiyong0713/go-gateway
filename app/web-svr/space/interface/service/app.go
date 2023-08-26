package service

import (
	"context"
	"time"

	"go-common/library/conf/env"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/space/interface/model"
	mainEcode "go-gateway/ecode"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	memapi "git.bilibili.co/bapis/bapis-go/account/service/member"
	relaapi "git.bilibili.co/bapis/bapis-go/account/service/relation"
	payrank "git.bilibili.co/bapis/bapis-go/account/service/ugcpay-rank"
	artmdl "git.bilibili.co/bapis/bapis-go/article/service"
	favapi "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	livemdl "git.bilibili.co/bapis/bapis-go/live/xfansmedal"
	uparcmdl "git.bilibili.co/bapis/bapis-go/up-archive/service"
)

const (
	_devicePad        = "pad"
	_devicePhone      = "phone"
	_platShopH5       = 1
	_shopGoodsOn      = 1
	_blockStatusLimit = 2
)

// AppIndex app index info.
func (s *Service) AppIndex(c context.Context, arg *model.AppIndexArg) (data *model.AppIndex, err error) {
	if env.DeployEnv == env.DeployEnvProd {
		if _, ok := s.BlacklistValue[arg.Vmid]; ok {
			err = xecode.NothingFound
			return
		}
	}
	var appInfo *model.AppAccInfo
	if appInfo, err = s.appAccInfo(c, arg.Mid, arg.Vmid, arg.Platform, arg.Device); err != nil {
		return
	}
	data = new(model.AppIndex)
	data.Info = appInfo
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		data.Tab = s.appTabInfo(ctx, arg.Mid, arg.Vmid, arg.Device, arg.Platform)
		return nil
	})
	if arg.Device == _devicePad {
		group.Go(func(ctx context.Context) error {
			data.Archive, _ = s.UpArcs(ctx, arg.Vmid, _samplePn, arg.Ps)
			return nil
		})
	}
	group.Go(func(ctx context.Context) error {
		dyListArg := &model.DyListArg{Mid: arg.Mid, Vmid: arg.Vmid, Qn: arg.Qn, Pn: _samplePn}
		data.Dynamic, _ = s.DynamicList(ctx, dyListArg)
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	if arg.Device == _devicePad {
		if data.Archive != nil && len(data.Archive.List) > 0 {
			data.Tab.Archive = true
		}
	}
	if data.Dynamic != nil && len(data.Dynamic.List) > 0 {
		data.Tab.Dynamic = true
	}
	return
}

// nolint:gocognit
func (s *Service) appAccInfo(ctx context.Context, mid, vmid int64, platform, device string) (*model.AppAccInfo, error) {
	data := &model.AppAccInfo{
		Relation:   struct{}{},
		BeRelation: struct{}{},
		Live:       struct{}{},
		Elec:       struct{}{},
		Shop:       struct{}{},
	}
	var (
		accBlock *model.AccBlock
		cert     *model.AudioUpperCert
	)
	ip := metadata.String(ctx, metadata.RemoteIP)
	group := errgroup.WithContext(ctx)
	group.Go(func(ctx context.Context) error {
		reply, err := s.accClient.ProfileWithStat3(ctx, &accapi.MidReq{Mid: vmid})
		if err != nil {
			if xecode.EqualError(xecode.UserNotExist, err) || xecode.EqualError(mainEcode.MemberNotExist, err) {
				return xecode.NothingFound
			}
			log.Error("s.accClient.ProfileWithStat3(%d) error(%v)", vmid, err)
		}
		if reply.GetProfile() == nil {
			reply = model.DefaultProfileStat
		}
		data.FromProfile(reply)
		if data.Mid == 0 {
			data.Mid = vmid
		}
		if mid == vmid {
			data.LevelInfo = reply.GetLevelInfo()
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if mid <= 0 || mid == vmid {
			return nil
		}
		reply, err := s.relationClient.Relation(ctx, &relaapi.RelationReq{Mid: mid, Fid: vmid, RealIp: ip})
		if err != nil {
			log.Error("s.relation.Relation(%d,%d,%s) error %v", mid, vmid, ip, err)
			return nil
		}
		if reply != nil {
			data.Relation = reply
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if mid <= 0 || mid == vmid {
			return nil
		}
		reply, err := s.relationClient.Relation(ctx, &relaapi.RelationReq{Mid: vmid, Fid: mid, RealIp: ip})
		if err != nil {
			log.Error("s.relation.Relation(%d,%d,%s) error %v", vmid, mid, ip, err)
			return nil
		}
		if reply != nil {
			data.BeRelation = reply
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if mid != vmid {
			return nil
		}
		reply, err := s.memberClient.BlockInfo(ctx, &memapi.MemberMidReq{Mid: mid})
		if err != nil {
			log.Error("s.memberClient.BlockInfo mid(%d) error(%v)", mid, err)
			accBlock = &model.AccBlock{Status: _accBlockDefault}
			return nil
		}
		accBlock = &model.AccBlock{
			Status: int(reply.GetBlockStatus()),
		}
		if reply.GetBlockStatus() == _blockStatusLimit {
			if time.Now().Unix() >= reply.GetEndTime() {
				accBlock.IsDue = _accBlockDue
			}
			status, err := s.dao.IsAnswered(ctx, mid, reply.GetStartTime())
			if err != nil {
				log.Error("s.dao.IsAnswered mid(%d) startTime(%v) error(%v)", mid, reply.GetStartTime(), err)
				return nil
			}
			accBlock.IsAnswered = status
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		data.TopPhoto, _ = s.dao.TopPhoto(ctx, mid, vmid, platform, device)
		return nil
	})
	group.Go(func(ctx context.Context) error {
		reply, err := s.dao.Live(ctx, vmid, "")
		if err != nil {
			log.Error("s.dao.Live vmid:%d,error:%+v", vmid, err)
			return nil
		}
		if reply != nil {
			data.Live = reply
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		reply, err := s.liveClient.QueryMedal(ctx, &livemdl.QueryMedalReq{UpUid: vmid})
		if err != nil {
			log.Error("s.liveClient.QueryMedal error(%+v)", err)
			return nil
		}
		if reply.UpMedal.Id > 0 {
			data.FansBadge = true
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		reply, err := s.dao.AudioCard(ctx, vmid)
		if err != nil {
			log.Error("s.dao.AudioCard error(%+v)", err)
			return nil
		}
		if v, ok := reply[vmid]; ok && v.Type == _audioCardOn && v.Status == 1 {
			data.Audio = 1
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		reply, err := s.payRankGRPC.UPRankWithPanelByUPMid(ctx, &payrank.RankElecUPReq{UPMID: vmid})
		if err != nil {
			log.Error("appAccInfo s.payRankGRPC.UPRankWithPanelByUPMid vmid:%d,error:%+v", vmid, err)
			return nil
		}
		if reply == nil || reply.RankElecUPProto == nil {
			return nil
		}
		var list []*model.ElecUserList
		for _, v := range reply.RankElecUPProto.List {
			if v == nil {
				continue
			}
			item := &model.ElecUserList{
				Mid:       v.UpMID,
				PayMid:    v.MID,
				Rank:      v.Rank,
				Uname:     v.Nickname,
				Avatar:    v.Avatar,
				Message:   v.Message,
				TrendType: v.TrendType,
			}
			if v.VIP != nil {
				item.VipInfo = model.ElecVipInfo{
					VipType:    v.VIP.Type,
					VipDueMsec: v.VIP.DueDate,
					VipStatus:  v.VIP.Status,
				}
			}
			if v.Hidden {
				item.MsgDeleted = 1
			}
			list = append(list, item)
		}
		data.Elec = &model.ElecInfo{
			Show:    true,
			Total:   reply.RankElecUPProto.CountUPTotalElec,
			Count:   reply.RankElecUPProto.Count,
			List:    list,
			ElecSet: reply.ElecSet,
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		shop, err := s.dao.ShopLink(ctx, vmid, _platShopH5)
		if err != nil {
			log.Error("s.dao.ShopLink vmid(%d) error(%+v)", vmid, err)
			return nil
		}
		if shop != nil {
			data.Shop = &model.ShopInfo{ID: shop.ShopID, Name: shop.Name, URL: shop.JumpURL}
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		var err error
		if cert, err = s.dao.AudioUpperCert(ctx, vmid); err != nil {
			log.Error("s.dao.AudioUpperCert vmid(%d) error(%+v)", vmid, err)
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		var err error
		if data.FansGroup, err = s.dao.GroupsCount(ctx, mid, vmid); err != nil {
			log.Error("s.dao.GroupsCount mid(%d) vmid(%d) error(%v)", mid, vmid, err)
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	if data.Silence == _silenceForbid {
		data.Block = accBlock
	}
	if cert != nil && cert.Cert != nil && cert.Cert.Type != -1 && cert.Cert.Desc != "" {
		if data.OfficialInfo.Type == _officialNoType {
			data.OfficialInfo.Type = cert.Cert.Type
		}
		if data.OfficialInfo.Desc != "" {
			data.OfficialInfo.Desc = data.OfficialInfo.Desc + "、" + cert.Cert.Desc
		} else {
			data.OfficialInfo.Desc = cert.Cert.Desc
		}
	}
	return data, nil
}

// AppTabInfo get app tab info.
// nolint:gocognit
func (s *Service) appTabInfo(c context.Context, mid, vmid int64, device, platform string) (tab *model.AppTab) {
	ip := metadata.String(c, metadata.RemoteIP)
	tab = new(model.AppTab)
	privacy := s.privacy(c, vmid)
	group := errgroup.WithContext(c)
	// pad tab dy,arc value out this func
	if device != _devicePad {
		group.Go(func(ctx context.Context) error {
			if dyCnt, err := s.dao.DynamicCnt(ctx, vmid); err != nil {
				log.Error("s.dao.DynamicCnt error(%+v)", err)
			} else if dyCnt > 0 {
				tab.Dynamic = true
			}
			return nil
		})
		group.Go(func(ctx context.Context) error {
			if shop, err := s.dao.ShopLink(ctx, vmid, _platShopH5); err != nil {
				log.Error("s.dao.ShopInfo error(%+v)", err)
			} else if shop != nil && shop.ShowItemsTab == _shopGoodsOn {
				tab.Shop = true
			}
			return nil
		})
		group.Go(func(ctx context.Context) error {
			if reply, err := s.upArcClient.ArcPassedTotal(ctx, &uparcmdl.ArcPassedTotalReq{Mid: vmid}); err != nil {
				log.Error("s.upArcClient.ArcPassedTotal mid(%d) error(%v)", vmid, err)
			} else if reply.Total > 0 {
				tab.Archive = true
			}
			return nil
		})
		group.Go(func(ctx context.Context) error {
			if article, err := s.artClient.UpArtMetas(ctx, &artmdl.UpArtMetasReq{Mid: vmid, Pn: 1, Ps: 10, Ip: ip}); err != nil {
				log.Error("s.artClient.UpArtMetas(%d) error(%v)", vmid, err)
			} else if article != nil && len(article.Articles) > 0 {
				tab.Article = true
			}
			return nil
		})
		group.Go(func(ctx context.Context) error {
			if audioCnt, err := s.dao.AudioCnt(ctx, vmid); err != nil {
				log.Error("s.dao.AudioCnt error(%+v)", err)
			} else if audioCnt > 0 {
				tab.Audio = true
			}
			return nil
		})
		group.Go(func(ctx context.Context) error {
			if albumCnt, err := s.dao.AlbumCount(ctx, vmid); err == nil && albumCnt > 0 {
				tab.Album = true
			}
			return nil
		})
		if value, ok := privacy[model.PcyGame]; (ok && value == _defaultPrivacy) || mid == vmid {
			group.Go(func(ctx context.Context) error {
				if _, gameCnt, err := s.dao.AppPlayedGame(ctx, vmid, platform, _samplePn, _samplePs); err == nil && gameCnt > 0 {
					tab.Game = true
				}
				return nil
			})
		}
	}
	if value, ok := privacy[model.PcyFavVideo]; (ok && value == _defaultPrivacy) || mid == vmid {
		group.Go(func(ctx context.Context) error {
			if favReply, err := s.favClient.UserFolders(ctx, &favapi.UserFoldersReq{Typ: _typeFavArchive, Mid: mid, Vmid: vmid}); err != nil {
				log.Error("s.favClient.UserFolders error(%+v)", err)
			} else if favReply != nil && len(favReply.Res) > 0 {
				for _, v := range favReply.Res {
					if v != nil && v.Count > 0 {
						tab.Favorite = true
						break
					}
				}
			}
			return nil
		})
	}
	if value, ok := privacy[model.PcyBangumi]; (ok && value == _defaultPrivacy) || mid == vmid {
		group.Go(func(ctx context.Context) error {
			if reply, err := s.dao.BangumiList(ctx, mid, vmid, _samplePn, _samplePs); err != nil {
				log.Error("s.dao.BangumiList mid(%d) vmid(%d) error(%v)", mid, vmid, err)
			} else if reply.Total > 0 {
				tab.Bangumi = true
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return
}

// AppTopPhoto get app top photo.
func (s *Service) AppTopPhoto(c context.Context, mid, vmid int64, platform, device string) (imgURL string) {
	imgURL, _ = s.dao.TopPhoto(c, mid, vmid, platform, device)
	return
}
