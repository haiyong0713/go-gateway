package service

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"

	v1 "go-gateway/app/web-svr/space/interface/api/v1"
	"go-gateway/app/web-svr/space/interface/model"

	membergrpc "git.bilibili.co/bapis/bapis-go/account/service/member"
	gallerygrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"
	spacegrpc "git.bilibili.co/bapis/bapis-go/space/service/v1"

	"github.com/pkg/errors"
)

const (
	_allowShowingFollow    = 0
	_disallowShowingFollow = 1
	_defaultPrivacy        = 1
)

func (s *Service) SpaceSetting(ctx context.Context, req *v1.SpaceSettingReq) (*v1.SpaceSettingReply, error) {
	res, err := s.SettingInfo(ctx, req.Mid)
	if err != nil {
		log.Error("s.SettingInfo req=%+v, err=%+v", req, err)
		return nil, err
	}
	return &v1.SpaceSettingReply{
		Channel:           int64(res.Privacy[model.PcyChannel]),
		FavVideo:          int64(res.Privacy[model.PcyFavVideo]),
		CoinsVideo:        int64(res.Privacy[model.PcyCoinVideo]),
		LikesVideo:        int64(res.Privacy[model.PcyLikeVideo]),
		Bangumi:           int64(res.Privacy[model.PcyBangumi]),
		PlayedGame:        int64(res.Privacy[model.PcyGame]),
		Groups:            int64(res.Privacy[model.PcyGroup]),
		Comic:             int64(res.Privacy[model.PcyComic]),
		BBQ:               int64(res.Privacy[model.PcyBbq]),
		DressUp:           int64(res.Privacy[model.PcyDressUp]),
		DisableFollowing:  int64(res.Privacy[model.PcyDisableFollowing]),
		LivePlayback:      int64(res.Privacy[model.LivePlayback]),
		CloseSpaceMedal:   int64(res.Privacy[model.PcyCloseSpaceMedal]),
		OnlyShowWearing:   int64(res.Privacy[model.PcyOnlyShowWearing]),
		DisableShowSchool: int64(res.Privacy[model.PcyDisableShowSchool]),
		DisableShowNft:    int64(res.Privacy[model.PcyDisableShowNft]),
	}, nil
}

func (s *Service) SettingInfo(c context.Context, mid int64) (data *model.Setting, err error) {
	data = &model.Setting{
		Privacy: make(map[string]int),
	}
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		privacy := s.privacy(ctx, mid)
		if privacy != nil {
			data.Privacy = privacy
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		data.IndexOrder = s.indexOrder(ctx, mid)
		return nil
	})
	var (
		followingPcy, closeSpaceMedalPcy, onlyShowWearingPcy, disableShowSchoolPcy, disableShowNftPcy int
		showNftSwitch                                                                                 bool
	)
	group.Go(func(ctx context.Context) error {
		followingPcy = s.followingPrivacy(ctx, mid)
		return nil
	})
	group.Go(func(ctx context.Context) error {
		closeSpaceMedalPcy, onlyShowWearingPcy = s.spaceLiveMedalPrivacy(ctx, mid)
		return nil
	})
	group.Go(func(ctx context.Context) error {
		disableShowSchoolPcy = s.disableShowSchoolPrivacy(ctx, mid)
		return nil
	})
	group.Go(func(ctx context.Context) error {
		showNftSwitch, disableShowNftPcy = s.getShowNftPrivacy(ctx, mid)
		return nil
	})
	if err = group.Wait(); err != nil {
		log.Error("SettingInfo group err(%+v)", err)
	}
	data.Privacy[model.PcyDisableFollowing] = followingPcy
	data.Privacy[model.PcyCloseSpaceMedal] = closeSpaceMedalPcy
	data.Privacy[model.PcyOnlyShowWearing] = onlyShowWearingPcy
	data.Privacy[model.PcyDisableShowSchool] = disableShowSchoolPcy
	if showNftSwitch {
		data.ShowNftSwitch = true
		data.Privacy[model.PcyDisableShowNft] = disableShowNftPcy
	}
	return
}

// PrivacySetting .
func (s *Service) PrivacySetting(c context.Context, mid int64) (res map[string]int) {
	return s.privacy(c, mid)
}

// PrivacyModify privacy modify.
func (s *Service) PrivacyModify(c context.Context, mid int64, field string, value int) error {
	switch field {
	case model.PcyDisableFollowing:
		if err := s.updateFollowingPrivacy(c, mid, value); err != nil {
			return err
		}
		return nil
	case model.PcyCloseSpaceMedal, model.PcyOnlyShowWearing:
		if err := s.dao.UpdateLiveUserMedalStatusPrivacy(c, mid, field, int64(value)); err != nil {
			log.Error("s.dao.UpdateFollowingPrivacy error(%+v)", err)
			return err
		}
		return nil
	case model.PcyDisableShowSchool:
		args := constructUpdatePrivacySettingArgs(mid, value)
		if _, err := s.dao.UpdateAccMemberPrivacySetting(c, args); err != nil {
			log.Error("PrivacyModify() s.dao.UpdateAccMemberPrivacySetting args:%+v, err:%+v", args, err)
			return err
		}
		return nil
	case model.PcyDisableShowNft:
		if _, err := s.dao.UpdateNftGalleryPrivacySetting(c, &gallerygrpc.UpdatePrivacySettingReq{Mid: mid, Switch: value != 1}); err != nil {
			log.Error("PrivacyModify() s.dao.UpdateNftGalleryPrivacySetting mid:%d, err:%+v", mid, err)
			return err
		}
		return nil
	default:
	}
	// update inner privacy
	privacy := s.privacy(c, mid)
	for k, v := range privacy {
		if field == k && value == v {
			return ecode.NotModified
		}
	}
	var settings []*spacegrpc.PrivacySetting
	option, ok := spacegrpc.PrivacyOption_value[field]
	if !ok {
		return errors.WithMessage(ecode.RequestErr, "field 不合法")
	}
	state := spacegrpc.PrivacyState_closed
	if value == 1 {
		state = spacegrpc.PrivacyState_opened
	}
	item := &spacegrpc.PrivacySetting{
		Option: spacegrpc.PrivacyOption(option),
		State:  state,
	}
	settings = append(settings, item)
	return s.dao.UpdatePrivacySetting(c, mid, settings)
}

func (s *Service) PrivacyBatchModify(c context.Context, mid int64, batch map[string]int, outerBatch map[string]int) error {
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		if len(batch) == 0 {
			return nil
		}
		var settings []*spacegrpc.PrivacySetting
		for key, val := range batch {
			option, ok := spacegrpc.PrivacyOption_value[key]
			if !ok {
				continue
			}
			state := spacegrpc.PrivacyState_closed
			if val == 1 {
				state = spacegrpc.PrivacyState_opened
			}
			item := &spacegrpc.PrivacySetting{
				Option: spacegrpc.PrivacyOption(option),
				State:  state,
			}
			settings = append(settings, item)
		}
		if len(settings) == 0 {
			return nil
		}
		return s.dao.UpdatePrivacySetting(ctx, mid, settings)
	})
	group.Go(func(ctx context.Context) error {
		if len(outerBatch) == 0 {
			return nil
		}
		if _, ok := outerBatch[model.PcyDisableFollowing]; !ok {
			return nil
		}
		if err := s.updateFollowingPrivacy(ctx, mid, outerBatch[model.PcyDisableFollowing]); err != nil {
			return err
		}
		return nil
	})
	if len(outerBatch) != 0 {
		group.Go(func(ctx context.Context) error {
			if _, ok := outerBatch[model.PcyCloseSpaceMedal]; !ok {
				return nil
			}
			if err := s.dao.UpdateLiveUserMedalStatusPrivacy(ctx, mid, model.PcyCloseSpaceMedal, int64(outerBatch[model.PcyCloseSpaceMedal])); err != nil {
				log.Error("s.dao.UpdateFollowingPrivacy error(%+v)", err)
				return err
			}
			return nil
		})
		group.Go(func(ctx context.Context) error {
			if _, ok := outerBatch[model.PcyOnlyShowWearing]; !ok {
				return nil
			}
			if err := s.dao.UpdateLiveUserMedalStatusPrivacy(ctx, mid, model.PcyOnlyShowWearing, int64(outerBatch[model.PcyOnlyShowWearing])); err != nil {
				log.Error("s.dao.UpdateFollowingPrivacy error(%+v)", err)
				return err
			}
			return nil
		})
		group.Go(func(ctx context.Context) error {
			val, ok := outerBatch[model.PcyDisableShowSchool]
			if !ok {
				return nil
			}
			args := constructUpdatePrivacySettingArgs(mid, val)
			if _, err := s.dao.UpdateAccMemberPrivacySetting(ctx, args); err != nil {
				log.Error("PrivacyBatchModify() s.dao.UpdateAccMemberPrivacySetting args:%+v, err:%+v", args, err)
				return err
			}
			return nil
		})
		group.Go(func(ctx context.Context) error {
			val, ok := outerBatch[model.PcyDisableShowNft]
			if !ok {
				return nil
			}
			if _, err := s.dao.UpdateNftGalleryPrivacySetting(ctx, &gallerygrpc.UpdatePrivacySettingReq{Mid: mid, Switch: val != 1}); err != nil {
				log.Error("PrivacyBatchModify() s.dao.UpdateNftGalleryPrivacySetting mid:%d, err:%+v", mid, err)
				return err
			}
			return nil
		})
	}
	return group.Wait()
}

func constructUpdatePrivacySettingArgs(mid int64, val int) *membergrpc.UpdatePrivacySettingReq {
	const (
		_schoolSwitchType = 1
	)
	return &membergrpc.UpdatePrivacySettingReq{
		Mid:        mid,
		SwitchType: _schoolSwitchType,
		Switch:     val != 1,
	}
}

// IndexOrderModify index order modify
func (s *Service) IndexOrderModify(c context.Context, mid int64, orderNum []string) (err error) {
	var orderStr []byte
	if orderStr, err = json.Marshal(orderNum); err != nil {
		log.Error("index order modify json.Marshal(%v) error(%v)", orderNum, err)
		err = ecode.RequestErr
		return
	}
	if err = s.dao.IndexOrderModify(c, mid, string(orderStr)); err == nil {
		s.cache.Do(c, func(c context.Context) {
			var cacheData []*model.IndexOrder
			for _, v := range orderNum {
				i, _ := strconv.Atoi(v)
				cacheData = append(cacheData, &model.IndexOrder{ID: i, Name: model.IndexOrderMap[i]})
			}
			_ = s.dao.SetIndexOrderCache(c, mid, cacheData)
		})
	}
	return
}

func (s *Service) privacy(ctx context.Context, mid int64) map[string]int {
	reply, err := s.dao.PrivacySetting(ctx, mid, nil)
	if err != nil {
		log.Error("%+v", err)
		return nil
	}
	res := make(map[string]int, len(reply))
	for key, setting := range reply {
		var state int
		if setting.State == spacegrpc.PrivacyState_opened {
			state = 1
		}
		res[key] = state
	}
	return res
}

func (s *Service) getShowNftPrivacy(ctx context.Context, mid int64) (bool, int) {
	midArgs := &gallerygrpc.MidReq{Mid: mid, RealIp: metadata.String(ctx, metadata.RemoteIP)}
	showNftSwitch, err := s.dao.HasNFT(ctx, midArgs)
	if err != nil {
		log.Error("getShowNftPrivacy() s.dao.HasNFT midArgs: %+v, err: %+v", midArgs, err)
		return false, 0
	}
	if showNftSwitch.Status == gallerygrpc.OwnerStatus_NOTOWNER {
		return false, 0
	}
	// disable_show_nft 默认态 0:公开拥有的数字艺术品,对应后端返回为true 1:隐藏拥有的数字艺术品，对应后端返回为false
	disableShowNftPcy, err := s.dao.NftGalleryPrivacySetting(ctx, midArgs)
	if err != nil {
		log.Error("getShowNftPrivacy() s.dao.NftGalleryPrivacySetting midArgs: %+v, err: %+v", midArgs, err)
		return true, 0
	}
	if disableShowNftPcy.Switch {
		return true, 0
	}
	return true, 1
}

func (s *Service) disableShowSchoolPrivacy(ctx context.Context, mid int64) int {
	args := &membergrpc.MidReq{
		Mid:    mid,
		RealIP: metadata.String(ctx, metadata.RemoteIP),
	}
	disableShowSchool, err := s.dao.AccMemberPrivacySetting(ctx, args)
	if err != nil {
		log.Error("disableShowSchoolPrivacy() s.dao.AccMemberPrivacySetting args: %+v, err: %+v", args, err)
		return 0
	}
	// disable_show_school 默认态 0: 展示学校信息,对应后端返回为true 1:不展示学校信息，对应后端返回为false
	if !disableShowSchool.SchoolSwitch {
		return 1
	}
	return 0
}

func (s *Service) spaceLiveMedalPrivacy(c context.Context, mid int64) (int, int) {
	liveMedalStatus, err := s.dao.LiveUserMedalStatusPrivacy(c, mid)
	if err != nil {
		log.Error("s.dao.LiveUserMedalStatusPrivacy error(%+v)", err)
		return 0, 0
	}
	// close_space_medal 默认态 0: 展示佩戴的粉丝勋章, 1:关闭佩戴的粉丝勋章
	// only_show_wearing 默认态 0: 粉丝勋章列表全部显示 1:粉丝勋章列表仅显示佩戴
	closeSpaceMedalStatus, onlyShowWearingStatus := 0, 0
	if val, ok := liveMedalStatus.Configs[model.PcyCloseSpaceMedal]; ok && val.Status == 1 {
		closeSpaceMedalStatus = 1
	}
	if val, ok := liveMedalStatus.Configs[model.PcyOnlyShowWearing]; ok && val.Status == 1 {
		onlyShowWearingStatus = 1
	}
	return closeSpaceMedalStatus, onlyShowWearingStatus
}

func (s *Service) updateFollowingPrivacy(c context.Context, mid int64, value int) error {
	state := false
	if value == _disallowShowingFollow {
		state = true
	}
	if err := s.dao.UpdateFollowingPrivacy(c, mid, state); err != nil {
		log.Error("s.dao.UpdateFollowingPrivacy mid(%d), state(%t), error(%+v)", mid, state, err)
		return err
	}
	return nil
}

func (s *Service) followingPrivacy(c context.Context, mid int64) int {
	followingSetting, err := s.dao.FollowingPrivacy(c, mid)
	if err != nil {
		log.Error("s.dao.followingPrivacy mid(%d) error(%+v)", mid, err)
		return _allowShowingFollow
	}
	if followingSetting.FollowingSwitch {
		return _disallowShowingFollow
	}
	return _allowShowingFollow
}

func (s *Service) indexOrder(c context.Context, mid int64) (data []*model.IndexOrder) {
	var (
		indexOrderStr string
		err           error
		addCache      = true
	)
	if data, err = s.dao.IndexOrderCache(c, mid); err != nil {
		addCache = false
	} else if len(data) != 0 {
		return s.mergeIndexOrder(data)
	}
	if indexOrderStr, err = s.dao.IndexOrder(c, mid); err != nil || indexOrderStr == "" {
		data = model.DefaultIndexOrder
	} else {
		orderNum := make([]string, 0)
		if err = json.Unmarshal([]byte(indexOrderStr), &orderNum); err != nil {
			log.Error("indexOrder mid(%d) json.Unmarshal(%s) error(%v)", mid, indexOrderStr, err)
			addCache = false
			s.cache.Do(c, func(c context.Context) {
				s.fixIndexOrder(c, mid, indexOrderStr)
			})
			data = model.DefaultIndexOrder
		} else {
			extraOrder := make(map[int]string)
			for _, v := range orderNum {
				index, err := strconv.Atoi(v)
				if err != nil {
					continue
				}
				name, ok := model.IndexOrderMap[index]
				if !ok {
					continue
				}
				data = append(data, &model.IndexOrder{ID: index, Name: name})
				extraOrder[index] = name
			}
			data = fixIndexOrder(data, extraOrder)
		}
	}
	if addCache {
		s.cache.Do(c, func(c context.Context) {
			_ = s.dao.SetIndexOrderCache(c, mid, data)
		})
	}
	return
}

func (s *Service) mergeIndexOrder(old []*model.IndexOrder) []*model.IndexOrder {
	data := old
	extraOrder := map[int]string{}
	for _, val := range old {
		name, ok := model.IndexOrderMap[val.ID]
		if !ok {
			continue
		}
		extraOrder[val.ID] = name
	}
	return fixIndexOrder(data, extraOrder)
}

func (s *Service) fixIndexOrder(c context.Context, mid int64, indexOrderStr string) {
	fixStr := strings.Replace(strings.TrimRight(strings.TrimLeft(indexOrderStr, "["), "]"), "\"", "", -1)
	fixArr := strings.Split(fixStr, ",")
	fixByte, err := json.Marshal(fixArr)
	if err != nil {
		log.Error("fixIndexOrder mid(%d) indexOrder(%s) error(%v)", mid, indexOrderStr, err)
		return
	}
	if err := s.dao.IndexOrderModify(c, mid, string(fixByte)); err == nil {
		_ = s.dao.DelIndexOrderCache(c, mid)
	}
}

// Privacy get privacy info with mid.
func (s *Service) Privacy(c context.Context, req *v1.PrivacyRequest) (reply *v1.PrivacyReply, err error) {
	data := s.privacy(c, req.Mid)
	tmp := make(map[string]int64, len(data))
	for k, v := range data {
		tmp[k] = int64(v)
	}
	reply = &v1.PrivacyReply{Privacy: tmp}
	return
}

func (s *Service) AppSetting(c context.Context, mid int64, mobiApp, device string) (*model.AppSetting, error) {
	setting := &model.AppSetting{
		Privacy: make(map[string]int),
	}
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		privacy := s.privacy(ctx, mid)
		if privacy != nil {
			setting.Privacy = privacy
		}
		return nil
	})
	var (
		userTab                                                                                       *v1.UserTabReply
		isUp, showNftSwitch                                                                           bool
		followingPcy, closeSpaceMedalPcy, onlyShowWearingPcy, disableShowSchoolPcy, disableShowNftPcy int
	)
	eg.Go(func(ctx context.Context) error {
		isUp, _ = s.dao.IsUpActUid(ctx, mid)
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		userTab, _ = s.UserTab(ctx, &v1.UserTabReq{Mid: mid})
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		followingPcy = s.followingPrivacy(ctx, mid)
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		closeSpaceMedalPcy, onlyShowWearingPcy = s.spaceLiveMedalPrivacy(ctx, mid)
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		disableShowSchoolPcy = s.disableShowSchoolPrivacy(ctx, mid)
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		showNftSwitch, disableShowNftPcy = s.getShowNftPrivacy(ctx, mid)
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	// 空间专属页
	if isUp && (userTab == nil || userTab.TabType == model.TabTypeUpAct) && !model.IsIPad(model.Plat(mobiApp, device)) {
		setting.ExclusiveURL = "https://www.bilibili.com/blackboard/up-sponsor.html?act_from=space_setting"
	}
	setting.Privacy[model.PcyDisableFollowing] = followingPcy
	setting.Privacy[model.PcyCloseSpaceMedal] = closeSpaceMedalPcy
	setting.Privacy[model.PcyOnlyShowWearing] = onlyShowWearingPcy
	setting.Privacy[model.PcyDisableShowSchool] = disableShowSchoolPcy
	if showNftSwitch {
		setting.ShowNftSwitch = true
		setting.Privacy[model.PcyDisableShowNft] = disableShowNftPcy
	}
	return setting, nil
}

// nolint:makezero
func fixIndexOrder(src []*model.IndexOrder, extraOrder map[int]string) []*model.IndexOrder {
	var insertLikeVideo bool
	data := make([]*model.IndexOrder, len(src))
	copy(data, src)
	for i, v := range model.IndexOrderMap {
		if _, ok := extraOrder[i]; ok {
			continue
		}
		if i == model.IndexOrderAppointment {
			data = append([]*model.IndexOrder{{ID: i, Name: v}}, data...)
			continue
		}
		if i == model.IndexOrderLikeVideo {
			insertLikeVideo = true
			continue
		}
		data = append(data, &model.IndexOrder{ID: i, Name: v})
	}
	if insertLikeVideo {
		orders := make([]*model.IndexOrder, 0, len(model.IndexOrderMap))
		for _, v := range data {
			orders = append(orders, v)
			if v.ID == model.IndexOrderCoinVideo {
				orders = append(orders, model.OrderItemLikeVideo)
			}
		}
		return orders
	}
	return data
}
