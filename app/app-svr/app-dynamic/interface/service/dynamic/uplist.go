package dynamic

import (
	"context"
	"strconv"
	"sync"

	"go-common/component/metadata/auth"
	"go-common/library/log"
	"go-common/library/net/metadata"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-dynamic/interface/api"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	dynmdl "go-gateway/app/app-svr/app-dynamic/interface/model/dynamic"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
)

const (
	infoBatchNum = 200
	retryNum     = 1
)

// 动态综合-查看更多-列表
func (s *Service) DynMixUpListViewMore(c context.Context, header *dynmdl.Header, _ *api.NoReq) (*api.DynMixUpListViewMoreReply, error) {
	au, _ := auth.FromContext(c)
	upListViewMoreReply, err := s.upListViewMore(c, header, au.Mid)
	if err != nil {
		log.Error("日志告警 DynMixUpListViewMore upListViewMore error:(%+v)", err)
		return upListViewMoreReply, err
	}
	return upListViewMoreReply, nil
}

// 动态综合-查看更多-搜索
func (s *Service) DynMixUpListSearch(c context.Context, header *dynmdl.Header, arg *api.DynMixUpListSearchReq) (*api.DynMixUpListSearchReply, error) {
	res := &api.DynMixUpListSearchReply{}
	ip := metadata.String(c, metadata.RemoteIP)
	au, _ := auth.FromContext(c)
	upListSearchItems, err := s.UpListSearch(c, header, au.Mid, arg.Name, ip)
	if err != nil {
		log.Error("日志告警 DynMixUpListSearch UpListSearch error:(%+v)", err)
		return res, err
	}
	res.Items = upListSearchItems
	return res, nil
}

// upListViewMore 查看更多
func (s *Service) upListViewMore(c context.Context, header *dynmdl.Header, mid int64) (*api.DynMixUpListViewMoreReply, error) {
	var (
		uidsMap             map[int64]struct{}
		cardsRes            []*accountgrpc.CardsReply
		upListViewMoreReply = &api.DynMixUpListViewMoreReply{SearchDefaultText: "搜索我的关注"}
		mixUplistRes        []*api.MixUpListItem
		relationUltima      map[int64]*relationgrpc.InterrelationReply
	)
	// 动态获取关注up主信息
	upListViewMoreRsp, err := s.dynDao.UpListViewMore(c, mid)
	if err != nil {
		log.Errorc(c, "upListFollowings UpListViewMore(mid: %+v) failed. error(%+v)", mid, err)
		return upListViewMoreReply, err
	}
	if upListViewMoreRsp == nil || len(upListViewMoreRsp.Items) == 0 {
		return upListViewMoreReply, nil
	}
	uidsMap, mixUplistRes = s.procViewMoreFollowingParams(upListViewMoreRsp)
	uidsNum := len(uidsMap)
	if uidsNum == 0 && mixUplistRes == nil {
		log.Info("upListViewMore mixUplistResMap is empty.")
		return upListViewMoreReply, nil
	}
	// 账号信息
	if uidsNum != 0 {
		var (
			uidSubsMap map[int][]int64
			mids       []int64
		)
		uidSubsMap, mids = s.procUidSubsMap(uidsMap)
		eg := errgroup.WithCancel(c)
		mutex := sync.Mutex{}
		// 账号分批处理，拉取账号信息
		for _, uidSubs := range uidSubsMap {
			tmpUidSubs := uidSubs
			eg.Go(func(ctx context.Context) error {
				if err = s.withRetry(retryNum, func() error {
					cardRes, err := s.accDao.Cards3(ctx, tmpUidSubs)
					if err != nil {
						log.Errorc(ctx, "upListViewMore Cards3(tmpUidSubs: %+v) failed. error(%+v)", tmpUidSubs, err)
						return err
					}
					mutex.Lock()
					defer mutex.Unlock()
					cardsRes = append(cardsRes, cardRes)
					return nil
				}); err != nil {
					log.Errorc(ctx, "upListViewMore withRetry(tmpUidSubs: %+v) failed. error(%+v)", tmpUidSubs, err)
					return err
				}
				return nil
			})
		}
		if len(mids) > 0 {
			eg.Go(func(ctx context.Context) error {
				relationUltima, err = s.accDao.Interrelations(ctx, mid, mids)
				if err != nil {
					log.Error("Interrelations mid(%v) mids(%v) error %v", mid, mids, err)
				}
				return nil
			})
		}
		if err := eg.Wait(); err != nil {
			return upListViewMoreReply, err
		}
	}
	mixUplistRes = s.procMixUpListItem(c, header, mid, mixUplistRes, cardsRes, relationUltima)
	upListViewMoreReply.Items = mixUplistRes
	return upListViewMoreReply, nil
}

// UpListSearch 获取搜索信息
func (s *Service) UpListSearch(ctx context.Context, header *dynmdl.Header, mid int64, name string, ip string) ([]*api.MixUpListItem, error) {
	var (
		upListSearchRsp    *dyngrpc.UpListSearchRsp
		mixUpListSearchRes []*api.MixUpListItem
		cardsRes           []*accountgrpc.CardsReply
		uidsMap            map[int64]struct{}
		err                error
		relationUltima     map[int64]*relationgrpc.InterrelationReply
	)
	upListSearchRsp, err = s.dynDao.UpListSearch(ctx, mid, name, ip)
	if err != nil {
		log.Errorc(ctx, "UpListSearch(mid: %+v) failed. error(%+v)", mid, err)
		return nil, err
	}
	if upListSearchRsp == nil || len(upListSearchRsp.Items) == 0 {
		return mixUpListSearchRes, nil
	}
	uidsMap, mixUpListSearchRes = s.procSearchFollowingParams(upListSearchRsp)
	uidsNum := len(uidsMap)
	if uidsNum == 0 && mixUpListSearchRes == nil {
		return mixUpListSearchRes, nil
	}
	// 账号信息
	if uidsNum != 0 {
		var (
			uidSubsMap map[int][]int64
			mids       []int64
		)
		uidSubsMap, mids = s.procUidSubsMap(uidsMap)
		eg := errgroup.WithCancel(ctx)
		mutex := sync.Mutex{}
		// 账号分批处理，拉取账号信息
		for _, uidSubs := range uidSubsMap {
			tmpUidSubs := uidSubs
			eg.Go(func(ctx context.Context) error {
				if err = s.withRetry(retryNum, func() error {
					cardRes, err := s.accDao.Cards3(ctx, tmpUidSubs)
					if err != nil {
						log.Errorc(ctx, "UpListSearch Cards3(tmpUidSubs: %+v) failed. error(%+v)", tmpUidSubs, err)
						return err
					}
					mutex.Lock()
					defer mutex.Unlock()
					cardsRes = append(cardsRes, cardRes)
					return nil
				}); err != nil {
					log.Errorc(ctx, "UpListSearch withRetry(tmpUidSubs: %+v) failed. error(%+v)", tmpUidSubs, err)
					return err
				}
				return nil
			})
		}
		if len(mids) > 0 {
			eg.Go(func(ctx context.Context) error {
				relationUltima, err = s.accDao.Interrelations(ctx, mid, mids)
				if err != nil {
					log.Error("Interrelations mid(%v) mids(%v) error %v", mid, mids, err)
				}
				return nil
			})
		}
		if err = eg.Wait(); err != nil {
			return nil, err
		}
	}
	mixUpListSearchRes = s.procMixUpListItem(ctx, header, mid, mixUpListSearchRes, cardsRes, relationUltima)
	return mixUpListSearchRes, nil
}

func (s *Service) procMixUpListItem(c context.Context, header *dynmdl.Header, mid int64, mixUpListRes []*api.MixUpListItem, cardsRes []*accountgrpc.CardsReply, relationUltima map[int64]*relationgrpc.InterrelationReply) []*api.MixUpListItem {
	var cards = make(map[int64]*accountgrpc.Card)
	var mixUpListItems []*api.MixUpListItem
	for _, card := range cardsRes {
		for _, item := range card.Cards {
			if item != nil {
				cards[item.Mid] = item
			}
		}
	}
	for _, res := range mixUpListRes {
		if s.checkMidMaxInt32(c, res.Uid, header) {
			continue
		}
		card, ok := cards[res.Uid]
		if !ok {
			continue
		}
		res.Uid = card.Mid
		res.Name = card.Name
		res.Face = card.Face
		official := &api.OfficialVerify{
			Type: int32(card.Official.Type),
			Desc: card.Official.Desc,
		}
		res.Official = official
		vip := &api.VipInfo{
			Type:      card.Vip.Type,
			DueDate:   card.Vip.DueDate,
			ThemeType: card.Vip.ThemeType,
			Status:    card.Vip.Status,
			Label: &api.VipLabel{
				Path: card.Vip.Label.Path,
			},
		}
		res.Vip = vip
		if s.c.Grayscale != nil && s.c.Grayscale.Relation != nil && s.c.Grayscale.Relation.Switch {
			switch s.c.Grayscale.Relation.GrayCheck(mid, "null") {
			case 1:
				res.Relation = dynmdl.RelationChange(card.Mid, relationUltima)
			}
		}
		mixUpListItems = append(mixUpListItems, res)
	}
	return mixUpListItems
}

func (s *Service) procViewMoreFollowingParams(uplistRsp *dyngrpc.UpListViewMoreRsp) (map[int64]struct{}, []*api.MixUpListItem) {
	var (
		mixUplistResMap []*api.MixUpListItem
	)
	uidsMap := make(map[int64]struct{})
	for _, item := range uplistRsp.Items {
		uidsMap[item.Uid] = struct{}{}
		res := &api.MixUpListItem{}
		if item.LiveInfo != nil {
			liveItem := &api.MixUpListLiveItem{
				RoomId: item.LiveInfo.RoomId,
				Uri:    item.LiveInfo.Link,
			}
			if item.LiveInfo.State == dynmdl.UplistMoreLiving {
				liveItem.Status = true
			}
			res.LiveInfo = liveItem
		}
		res.SpecialAttention = item.SpeacialAttention
		res.ReddotState = item.ReddotState
		res.Uid = item.Uid
		mixUplistResMap = append(mixUplistResMap, res)
	}
	return uidsMap, mixUplistResMap
}

func (s *Service) procSearchFollowingParams(uplistRsp *dyngrpc.UpListSearchRsp) (map[int64]struct{}, []*api.MixUpListItem) {
	var (
		mixUplistResMap []*api.MixUpListItem
	)
	uidsMap := make(map[int64]struct{})
	for _, item := range uplistRsp.Items {
		uidsMap[item.Uid] = struct{}{}
		res := &api.MixUpListItem{}
		if item.LiveInfo != nil {
			liveItem := &api.MixUpListLiveItem{
				RoomId: item.LiveInfo.RoomId,
				Uri:    item.LiveInfo.Link,
			}
			if item.LiveInfo.State == dynmdl.UplistMoreLiving {
				liveItem.Status = true
			}
			res.LiveInfo = liveItem
		}
		res.SpecialAttention = item.SpeacialAttention
		res.ReddotState = item.ReddotState
		res.Uid = item.Uid
		res.PremiereState = item.PremiereState
		if item.PremiereState == 1 {
			res.Uri = model.FillURI(model.GotoAv, strconv.FormatInt(item.Avid, 10), model.SuffixHandler("auto_float_layer=7"))
		}
		mixUplistResMap = append(mixUplistResMap, res)
	}
	return uidsMap, mixUplistResMap
}

func (s *Service) procUidSubsMap(uidsMap map[int64]struct{}) (map[int][]int64, []int64) {
	var uids []int64
	for uid := range uidsMap {
		uids = append(uids, uid)
	}
	uidsNum := len(uidsMap)
	var uidSubsMap = make(map[int][]int64, infoBatchNum)
	// 账号信息
	if uidsNum != 0 {
		j := 1
		for i := 0; i < uidsNum; i += infoBatchNum {
			if i+infoBatchNum > uidsNum {
				// 不足一批次
				uidSubsMap[j] = uids[i:]
			} else {
				// 满一批次
				uidSubsMap[j] = uids[i : i+infoBatchNum]
			}
			j++
		}
	}
	return uidSubsMap, uids
}

func (s *Service) withRetry(attempts int, f func() error) error {
	if err := f(); err != nil {
		if attempts--; attempts > 0 {
			return s.withRetry(attempts, f)
		}
		return err
	}
	return nil
}
