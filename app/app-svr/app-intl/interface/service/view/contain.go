package view

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup"
	egv2 "go-common/library/sync/errgroup.v2"
	"go-common/library/text/translate/chinese.v2"

	"go-gateway/app/app-svr/app-intl/interface/model"
	"go-gateway/app/app-svr/app-intl/interface/model/bangumi"
	"go-gateway/app/app-svr/app-intl/interface/model/tag"
	"go-gateway/app/app-svr/app-intl/interface/model/view"
	"go-gateway/app/app-svr/archive/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accApi "git.bilibili.co/bapis/bapis-go/account/service"
	location "git.bilibili.co/bapis/bapis-go/community/service/location"
	thumbup "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	v1 "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

var (
	_rate = map[int]int64{15: 464, 16: 464, 32: 1028, 48: 1328, 64: 2192, 74: 3192, 80: 3192, 112: 6192, 116: 6192, 66: 1820}
)

const (
	_dmformat     = "http://comment.bilibili.com/%d.xml"
	_videoChannel = 3
)

// initReqUser init Req User
// nolint:gocognit
func (s *Service) initReqUser(c context.Context, v *view.View, mid int64) {
	// owner ext
	var (
		owners []int64
		cards  map[int64]*accApi.Card
		fls    map[int64]int8
	)
	g, ctx := errgroup.WithContext(c)
	if v.Author.Mid > 0 {
		owners = append(owners, v.Author.Mid)
		for _, staffInfo := range v.StaffInfo {
			owners = append(owners, staffInfo.Mid)
		}
		g.Go(func() (err error) {
			v.OwnerExt.OfficialVerify.Type = -1
			cards, err = s.accDao.Cards3(ctx, owners)
			if err != nil {
				log.Error("%+v", err)
				err = nil
				return
			}
			if card, ok := cards[v.Author.Mid]; ok && card != nil {
				otp := -1
				odesc := ""
				if card.Official.Role != 0 {
					if card.Official.Role <= 2 || card.Official.Role == 7 {
						otp = 0
					} else {
						otp = 1
					}
					odesc = card.Official.Title
				}
				v.OwnerExt.OfficialVerify.Type = otp
				v.OwnerExt.OfficialVerify.Desc = odesc
				v.OwnerExt.Vip.Type = int(card.Vip.Type)
				v.OwnerExt.Vip.VipStatus = int(card.Vip.Status)
				v.OwnerExt.Vip.DueDate = card.Vip.DueDate
				v.Author.Name = card.Name
				v.Author.Face = card.Face
			}
			return
		})
		g.Go(func() (err error) {
			stat, err := s.relDao.Stat(c, v.Author.Mid)
			if err != nil {
				log.Error("%+v", err)
				err = nil
				return
			}
			if stat != nil {
				v.OwnerExt.Fans = int(stat.Follower)
			}
			return
		})
		g.Go(func() error {
			if ass, err := s.assDao.Assist(ctx, v.Author.Mid); err != nil {
				log.Error("%+v", err)
			} else {
				v.OwnerExt.Assists = ass
			}
			return nil
		})
	}
	// req user
	v.ReqUser = &view.ReqUser{Favorite: 0, Attention: -999, Like: 0, Dislike: 0}
	// check req user
	if mid > 0 {
		g.Go(func() error {
			if is, _ := s.favDao.IsFav(ctx, mid, v.Aid); is {
				v.ReqUser.Favorite = 1
			}
			return nil
		})
		g.Go(func() error {
			res, err := s.thumbupDao.HasLike(ctx, mid, _businessLike, []int64{v.Aid})
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			if typ, ok := res[v.Aid]; ok {
				if typ == thumbup.State_STATE_LIKE {
					v.ReqUser.Like = 1
				} else if typ == thumbup.State_STATE_DISLIKE {
					v.ReqUser.Dislike = 1
				}
			}
			return nil
		})
		g.Go(func() (err error) {
			res, err := s.coinDao.ArchiveUserCoins(ctx, v.Aid, mid, _avTypeAv)
			if err != nil {
				log.Error("%+v", err)
				err = nil
			}
			if res > 0 {
				v.ReqUser.Coin = 1
			}
			return
		})
		if v.Author.Mid > 0 {
			g.Go(func() error {
				fls = s.accDao.IsAttention(ctx, owners, mid)
				if _, ok := fls[v.Author.Mid]; ok {
					v.ReqUser.Attention = 1
				}
				return nil
			})
		}
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
	}
	// fill staff
	if v.AttrVal(arcgrpc.AttrBitIsCooperation) == arcgrpc.AttrYes {
		for _, owner := range owners {
			if card, ok := cards[owner]; ok && card != nil {
				staff := &view.Staff{Mid: owner}
				if owner == v.Author.Mid {
					staff.Title = "UP主"
				} else {
					for _, s := range v.StaffInfo {
						if s.Mid == owner {
							staff.Title = s.Title
						}
						if s.StaffAttrVal(api.StaffAttrBitAdOrder) == api.AttrYes {
							staff.LabelStyle = model.StaffLabelAd
						}
					}
				}
				staff.Name = card.Name
				staff.Face = card.Face
				staff.OfficialVerify.Type = -1
				otp := -1
				odesc := ""
				if card.Official.Role != 0 {
					if card.Official.Role <= 2 || card.Official.Role == 7 {
						otp = 0
					} else {
						otp = 1
					}
					odesc = card.Official.Title
				}
				staff.OfficialVerify.Type = otp
				staff.OfficialVerify.Desc = odesc
				staff.Vip.Type = int(card.Vip.Type)
				staff.Vip.VipStatus = int(card.Vip.Status)
				staff.Vip.DueDate = card.Vip.DueDate
				staff.Vip.ThemeType = int(card.Vip.ThemeType)
				if _, ok := fls[owner]; ok {
					staff.Attention = 1
				}
				v.Staff = append(v.Staff, staff)
			}
		}
	}
}

// initRelateCMTag is.
// nolint:staticcheck
func (s *Service) initRelateCMTag(c context.Context, v *view.View, plat int8, build, autoplay int, mid int64, buvid, from, trackid, filtered string, isTW bool) {
	var (
		rls               []*view.Relate
		err               error
		autoplayCountdown int
		returnPage        int
		autoplayToast     string
	)
	s.initTag(c, v, mid, plat)
	// 审核版本，和有屏蔽推荐池属性的稿件下 不出相关推荐任何信息
	if filtered == "1" || v.ForbidRec == 1 {
		log.Warn("no relates aid(%d) filtered(%s) ForbidRec(%d)", v.Aid, filtered, v.ForbidRec)
		return
	}
	if mid > 0 || buvid != "" {
		if rls, v.UserFeature, v.ReturnCode, v.PlayParam, autoplayCountdown, returnPage, _, autoplayToast, v.PvFeature, err = s.newRcmdRelate(c, plat, v.Aid, mid, buvid, from, trackid, build, autoplay); err != nil {
			log.Error("s.newRcmdRelate(%d) error(%+v)", v.Aid, err)
		}
		if v.Config != nil {
			if autoplayCountdown > 0 {
				v.Config.AutoplayCountdown = autoplayCountdown
			} else {
				v.Config.AutoplayCountdown = s.c.ViewConfig.AutoplayCountdown
			}
			v.Config.PageRefresh = returnPage
			v.Config.AutoplayDesc = autoplayToast
		}
	}
	//ai：code=-3表示无有效结果稿;code=5表示屏蔽用户黑名单;code=-2表示内部拉用户信息缺失
	if len(rls) == 0 && v.ReturnCode != "-3" && v.ReturnCode != "-5" { //-3和-5不要取灾备数据
		// nolint:ineffassign
		rls, err = s.dealRcmdRelate(c, plat, v.Aid)
		log.Warn("s.dealRcmdRelate aid(%d) mid(%d) buvid(%s)", v.Aid, mid, buvid)
		return
	}
	v.IsRec = 1
	log.Warn("s.newRcmdRelate returncode(%s) aid(%d) mid(%d) buvid(%s)", v.ReturnCode, v.Aid, mid, buvid)
	if len(rls) == 0 {
		s.prom.Incr("zero_relates")
		return
	}
	for _, rl := range rls {
		if rl.Aid == v.Aid {
			continue
		}
		v.Relates = append(v.Relates, rl)
	}
	if isTW {
		for _, rl := range v.Relates {
			rl.Title = chinese.Convert(c, rl.Title)
		}
	}
}

// initPGC is.
func (s *Service) initPGC(c context.Context, v *view.View, mid int64, build int, mobiApp, device string) (err error) {
	s.pHit.Incr("is_PGC")
	var season *bangumi.Season
	if season, err = s.banDao.PGC(c, v.Aid, mid, build, mobiApp, device); err != nil {
		log.Error("%+v", err)
		err = ecode.NothingFound
		s.pMiss.Incr("err_is_PGC")
		return
	}
	if season != nil {
		if season.Player != nil {
			if len(v.Pages) != 0 {
				if season.Player.Cid != 0 {
					v.Pages[0].Cid = season.Player.Cid
				}
				if season.Player.From != "" {
					v.Pages[0].From = season.Player.From
				}
				if season.Player.Vid != "" {
					v.Pages[0].Vid = season.Player.Vid
				}
			}
			season.Player = nil
		}
		if season.AllowDownload == "1" {
			v.Rights.Download = 1
		} else {
			v.Rights.Download = 0
		}
		if season.SeasonID != "" {
			season.AllowDownload = ""
			v.Season = season
		}
	}
	if v.Rights.HD5 == 1 && !s.checkVIP(c, mid) {
		v.Rights.HD5 = 0
	}
	v.Rights.Bp = 0
	return
}

// initPages is.
func (s *Service) initPages(_ context.Context, vs *view.ViewStatic, ap []*api.Page) {
	pages := make([]*view.Page, 0, len(ap))
	for _, v := range ap {
		page := &view.Page{}
		metas := make([]*view.Meta, 0, 4)
		for q, r := range _rate {
			meta := &view.Meta{
				Quality: q,
				Size:    int64(float64(r*v.Duration) * 1.1 / 8.0),
			}
			metas = append(metas, meta)
		}
		if vs.AttrVal(arcgrpc.AttrBitIsBangumi) == arcgrpc.AttrYes {
			v.From = "bangumi"
		}
		page.Page = v
		page.Metas = metas
		page.DMLink = fmt.Sprintf(_dmformat, v.Cid)
		pages = append(pages, page)
	}
	vs.Pages = pages
}

// initDownload is.
func (s *Service) initDownload(c context.Context, v *view.View, mid int64, cdnIP string) (err error) {
	var download int64
	if v.AttrVal(arcgrpc.AttrBitLimitArea) == arcgrpc.AttrYes {
		if download, err = s.ipLimit(c, mid, v.Aid, cdnIP); err != nil {
			return
		}
	} else {
		download = int64(location.StatusDown_AllowDown)
	}
	if download == int64(location.StatusDown_ForbiddenDown) {
		v.Rights.Download = int32(download)
		return
	}
	for _, p := range v.Pages {
		if p.From == "qq" {
			download = int64(location.StatusDown_ForbiddenDown)
			break
		}
	}
	v.Rights.Download = int32(download)
	return
}

// initAudios is.
// nolint:gomnd
func (s *Service) initAudios(c context.Context, v *view.View) {
	pLen := len(v.Pages)
	if pLen == 0 || pLen > 100 {
		return
	}
	if pLen > 50 {
		pLen = 50
	}
	cids := make([]int64, 0, len(v.Pages[:pLen]))
	for _, p := range v.Pages[:pLen] {
		cids = append(cids, p.Cid)
	}
	vam, err := s.audioDao.AudioByCids(c, cids)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	if len(vam) != 0 {
		for _, p := range v.Pages[:pLen] {
			if va, ok := vam[p.Cid]; ok {
				p.Audio = va
			}
		}
		if len(v.Pages) == 1 {
			if va, ok := vam[v.Pages[0].Cid]; ok {
				v.Audio = va
			}
		}
	}
}

// initTag is.
func (s *Service) initTag(c context.Context, v *view.View, mid int64, plat int8) (tids []int64) {
	var (
		actTag     []*tag.Tag
		arcTag     []*tag.Tag
		actTagName string
	)
	if v.MissionID > 0 {
		protocol, err := s.actDao.ActProtocol(c, v.MissionID)
		if err != nil {
			log.Error("s.actDao.ActProtocol err(%+v)", err)
			// nolint:ineffassign
			err = nil
		} else {
			if protocol.Subject != nil {
				v.ActivityURL = protocol.Subject.AndroidURL
				if model.IsIOS(plat) {
					v.ActivityURL = protocol.Subject.IosURL
				}
			}
			if protocol.Protocol != nil {
				actTagName = protocol.Protocol.Tags
			}
		}
	}
	chans, err := s.channelDao.ResourceChannels(c, v.Aid, mid, _videoChannel)
	if err != nil {
		log.Error("s.channelDao.ResourceChannels err(%+v)", err)
		return
	}
	if len(chans) == 0 {
		return
	}
	tids = make([]int64, 0, len(chans))
	// 优先级  话题活动的链接模式 > 话题活动 > 新频道 > 旧频道
	for _, t := range chans {
		tempTag := &tag.Tag{TagID: t.ID, Name: t.Name, Cover: t.Cover, Likes: t.Likes, Hates: t.Hates, Liked: t.Liked, Hated: t.Hated, Attribute: t.Attribute}
		if t.CType == model.ChannelCtypeNew {
			tempTag.URI = "bilibili://pegasus/channel/v2/" + strconv.FormatInt(tempTag.TagID, 10) + "?tab=select"
			tempTag.TagType = "new"
		} else {
			tempTag.URI = "bilibili://pegasus/channel/" + strconv.FormatInt(tempTag.TagID, 10)
			tempTag.TagType = "common"
		}
		if actTagName == tempTag.Name {
			tempTag.IsActivity = 1
			tempTag.TagType = "act"
			actTag = append(actTag, tempTag)
		} else {
			arcTag = append(arcTag, tempTag)
		}
		tids = append(tids, tempTag.TagID)
	}
	//活动稿件tag放在第一位
	v.Tag = append(actTag, arcTag...)
	v.TIcon = make(map[string]*tag.TIcon)
	v.TIcon["act"] = &tag.TIcon{Icon: s.c.TagConfig.ActIcon}
	if s.c.TagConfig.OpenIcon {
		v.TIcon["new"] = &tag.TIcon{Icon: s.c.TagConfig.NewIcon}
	}
	return
}

// initDM is.
// nolint:gomnd
func (s *Service) initDM(c context.Context, v *view.View) {
	const (
		_dmTypeAv    = 1
		_dmPlatMobie = 1
	)
	pLen := len(v.Pages)
	if pLen == 0 || pLen > 100 {
		return
	}
	if pLen > 50 {
		pLen = 50
	}
	cids := make([]int64, 0, len(v.Pages[:pLen]))
	for _, p := range v.Pages[:pLen] {
		cids = append(cids, p.Cid)
	}
	res, err := s.dmDao.SubjectInfos(c, _dmTypeAv, _dmPlatMobie, cids...)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	if len(res) == 0 {
		return
	}
	for _, p := range v.Pages[:pLen] {
		if r, ok := res[p.Cid]; ok {
			p.DM = r
		}
	}
}

// dealRcmdRelate is.
func (s *Service) dealRcmdRelate(c context.Context, plat int8, aid int64) (rls []*view.Relate, err error) {
	var aids []int64
	if aids, err = s.arcDao.RelateAids(c, aid); err != nil {
		return
	}
	if len(aids) == 0 {
		return
	}
	var as map[int64]*api.Arc
	if as, err = s.arcDao.Archives(c, aids); err != nil {
		return
	}

	reply, err := s.ContentFlowControlInfosV2Slices(c, aids)
	if err != nil {
		return
	}

	for _, aid := range aids {
		if a, ok := as[aid]; ok {
			if s.overseaCheck(reply, aid, plat) || !a.IsNormal() {
				continue
			}
			r := &view.Relate{}
			r.FromAv(a, "", nil, "")
			rls = append(rls, r)
		}
	}
	return
}

// newRcmdRelate is.
// nolint:gocognit
func (s *Service) newRcmdRelate(c context.Context, plat int8, aid, mid int64, buvid, from, trackid string, build, autoplay int) (rls []*view.Relate, userFeature, returnCode string, playParam, autoplayCountdown, returnPage, gamecardStyleExp int, autoplayToast string, pvFeature json.RawMessage, err error) {
	zoneID := int64(0)
	loc, _ := s.locDao.Info2(c)
	// 相关推荐AI使用zoneID取zoneID[3]
	if loc != nil && len(loc.ZoneId) >= 4 {
		zoneID = loc.ZoneId[3]
	}
	res, returnCode, err := s.arcDao.NewRelateAids(c, aid, mid, build, autoplay, buvid, from, trackid, plat, zoneID)
	if err != nil || res == nil || len(res.Data) == 0 {
		return
	}
	userFeature = res.UserFeature
	playParam = res.PlayParam
	pvFeature = res.PvFeature
	autoplayCountdown = res.AutoplayCountdown
	returnPage = res.ReturnPage
	gamecardStyleExp = res.GamecardStyleExp // 是否展示游戏新卡
	autoplayToast = res.AutoplayToast
	var (
		aids             []int64
		ssIDs            []int32
		arcm             map[int64]*api.ArcPlayer
		banm             map[int32]*v1.CardInfoProto
		flowInfosV2Reply *cfcgrpc.FlowCtlInfosV2Reply
	)
	for _, rec := range res.Data {
		switch rec.Goto {
		case model.GotoAv:
			aids = append(aids, rec.Oid)
		case model.GotoBangumi:
			ssIDs = append(ssIDs, int32(rec.Oid))
		}
	}
	eg := egv2.WithContext(c)
	if len(aids) > 0 {
		var aidPlays []*api.PlayAv
		for _, aval := range aids {
			aidPlays = append(aidPlays, &api.PlayAv{Aid: aval})
		}
		eg.Go(func(ctx context.Context) (err error) {
			if (plat == model.PlatAndroidI && build > s.c.ViewBuildLimit.ArcWithPlayerAndroid) || (plat == model.PlatIPhoneI && build > s.c.ViewBuildLimit.ArcWithPlayerIOS) {
				if arcm, err = s.arcDao.ArcsPlayer(ctx, aidPlays); err != nil {
					log.Error("%+v", err)
				}
			} else {
				if arcm, err = s.arcDao.Arcs(ctx, aids); err != nil {
					log.Error("%+v", err)
				}
			}
			return nil
		})
		eg.Go(func(ctx context.Context) error {
			flowInfosV2Reply, err = s.ContentFlowControlInfosV2Slices(ctx, aids)
			if err != nil {
				log.Error("s.ContentFlowControlInfosV2Slices err=%+v", err)
				return nil
			}
			return nil
		})
	}
	if len(ssIDs) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if banm, err = s.banDao.CardsInfoReply(ctx, ssIDs); err != nil {
				log.Error("s.banDao.CardsInfoReply err(%+v)", err)
			}
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("newRcmdRelate errGroup err(%+v)", err)
	}
	for _, rec := range res.Data {
		r := &view.Relate{AvFeature: rec.AvFeature, Source: rec.Source}
		switch rec.Goto {
		case model.GotoAv:
			arc, ok := arcm[rec.Oid]
			if !ok || arc == nil || arc.Arc == nil {
				continue
			}
			if s.overseaCheck(flowInfosV2Reply, arc.Arc.Aid, plat) || !arc.Arc.IsNormal() {
				continue
			}
			if rec.IsDalao == 1 {
				r.FromOperate(rec, arc.Arc, model.FromOperation, rec.TrackID)
			} else {
				firstPlay := arc.PlayerInfo[arc.DefaultPlayerCid]
				r.FromAv(arc.Arc, "", firstPlay, rec.TrackID)
			}
		case model.GotoBangumi:
			ban, ok := banm[int32(rec.Oid)]
			if !ok {
				continue
			}
			r.FromBangumi(ban, aid)
		}
		rls = append(rls, r)
	}
	return
}

func (s *Service) ContentFlowControlInfosV2Slices(ctx context.Context, aids []int64) (*cfcgrpc.FlowCtlInfosV2Reply, error) {
	const (
		_maxAids = 30
	)
	var (
		aidsLen = len(aids)
		mutex   = sync.Mutex{}
	)
	aidMap := make(map[int64]*cfcgrpc.FlowCtlInfoV2Reply, aidsLen)
	eg := egv2.WithContext(ctx)
	for i := 0; i < aidsLen; i += _maxAids {
		var partAids []int64
		if i+_maxAids > aidsLen {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+_maxAids]
		}
		eg.Go(func(ctx context.Context) error {
			tmpRes, err := s.cfcDao.ContentFlowControlInfosV2(ctx, partAids)
			if err != nil {
				log.Error("s.cfc.ContentFlowControlInfosV2 partAids=%+v err=%+v", partAids, err)
				return nil
			}
			if tmpRes == nil || tmpRes.ItemsMap == nil {
				log.Error("CircleReqInternalAttr is nil(%v)", partAids)
				return nil
			}
			if len(tmpRes.ItemsMap) > 0 {
				mutex.Lock()
				for aid, arc := range tmpRes.ItemsMap {
					if arc == nil {
						continue
					}
					aidMap[aid] = arc
				}
				mutex.Unlock()
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return &cfcgrpc.FlowCtlInfosV2Reply{ItemsMap: aidMap}, nil
}

func (s *Service) initLabel(v *view.View) {
	var (
		_hot      = int8(1)
		_activity = int8(2)
	)
	if _, ok := s.hotAids[v.Aid]; ok {
		v.Label = &view.Label{
			Type: _hot,
			URI:  model.FillURI(model.GotoHotPage, "", nil),
		}
		return
	}
	if v.ActivityURL != "" {
		v.Label = &view.Label{
			Type: _activity,
			URI:  v.ActivityURL,
		}
	}
}
