package space

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"git.bilibili.co/go-tool/libbdevice/pkg/pd"
	"go-common/library/log"
	errgroupv2 "go-common/library/sync/errgroup.v2"
	"go-common/library/text/translate/chinese.v2"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/space"
	"go-gateway/app/app-svr/archive/service/api"
	ugcSeasonGrpc "go-gateway/app/app-svr/ugc-season/service/api"

	uparcapi "git.bilibili.co/bapis/bapis-go/up-archive/service"

	"github.com/pkg/errors"
)

func (s *Service) UpArcsCursor(c context.Context, param *space.ArchiveCursorParam, mid int64, isHant, isIpad bool) (*space.ArcCursorList, error) {
	var (
		arcs      []*space.ArcPlayerCursor
		arcsExtra *space.ArcPlayerCursorExtra
		uname     string
		position  *space.HistoryPosition
	)
	eg := errgroupv2.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		withoutLivePlayback := s.getWithoutLivePlayback(ctx, param.Vmid)
		arcs, arcsExtra, err = s.getArcPlayerByCursor(ctx, param, withoutLivePlayback, isIpad)
		if err != nil {
			log.Error("s.getArcPlayerByCursor err: %+v", err)
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		uname, err = s.getUserName(ctx, param.Vmid)
		if err != nil {
			log.Error("s.getUserName err: %+v", err)
			return nil
		}
		return nil
	})
	if mid > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			position, err = s.getHistoryPosition(ctx, mid, param.Vmid, _spaceArchiveBusiness)
			if err != nil {
				log.Error("s.getHistoryPosition err: %+v", err)
				return nil
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("UpArcsCursor() eg.Wait err: %+v", err)
		return nil, err
	}
	var seasons map[int64]*ugcSeasonGrpc.Season // 以aid为key的合集信息
	if pd.WithContext(c).Where(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().And().Build(">=", 6780000)
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPhone().And().Build(">=", 67800100)
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPad().And().Build(">=", 68000000)
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPadHD().And().Build(">=", 34500000)
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatAndroidHD().And().Build(">=", 1230000)
	}).MustFinish() {
		if seasonDisplay, err := s.ArcReorderBySeason(c, nil, arcs, param.Order, param.Vmid); err == nil {
			arcs = seasonDisplay.ArcPlayerCursor
			seasons = seasonDisplay.Seasons
		}
	}
	res, err := s.resolveSpaceArcList(arcs, param, arcsExtra, uname, position, seasons)
	if err != nil {
		log.Error("UpArcsCursor() s.resolveSpaceArcList err: %+v", err)
		return nil, err
	}
	if isHant {
		return convertHantItems(c, res), nil
	}
	return res, nil
}

func convertHantItems(ctx context.Context, in *space.ArcCursorList) *space.ArcCursorList {
	for _, v := range in.Item {
		out := chinese.Converts(ctx, v.Title, v.TypeName)
		v.Title = out[v.Title]
		v.TypeName = out[v.TypeName]
	}
	return in
}

func (s *Service) getWithoutLivePlayback(ctx context.Context, vmid int64) bool {
	const (
		_openLivePlaybackSetting = 1
	)
	setting, err := s.spcDao.Setting(ctx, vmid)
	if err != nil {
		log.Error("getWithoutLivePlayback() s.spcDao.Setting vmid: %d, err: %+v", vmid, err)
		return false
	}
	if setting != nil {
		return setting.LivePlayback != _openLivePlaybackSetting
	}
	return false
}

func (s *Service) getArcPlayerByCursor(c context.Context, param *space.ArchiveCursorParam, withoutLivePlayback, isIpad bool) ([]*space.ArcPlayerCursor, *space.ArcPlayerCursorExtra, error) {
	var (
		reply              *uparcapi.ArcPassedByAidReply
		arcs               []*space.ArcPlayerCursor
		locationCanDisplay bool
	)
	var without []uparcapi.Without
	if !isIpad {
		without = append(without, uparcapi.Without_no_space)

	}
	if withoutLivePlayback {
		without = append(without, uparcapi.Without_live_playback)
	}
	eg := errgroupv2.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		reply, err = s.getArcPassedByAid(ctx, param, without)
		if err != nil {
			log.Error("s.getArcPassedByAid err: %+v", err)
			return err
		}
		arcs, err = s.resolveArcPlayer(ctx, param, reply.Archives)
		if err != nil {
			log.Error("s.resolveArcPlayer err: %+v", err)
			return err
		}
		return nil
	})
	if param.Aid == 0 && param.FromViewAid > 0 {
		// 第一刷执行逻辑
		eg.Go(func(ctx context.Context) (err error) {
			locationCanDisplay, err = s.getLocationCanDisplay(ctx, param, without)
			if err != nil {
				log.Error("s.getLocationCanDisplay err: %+v", err)
				return nil
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, nil, err
	}
	extra := constructArcPlayerCursorExtra(reply, locationCanDisplay)
	return arcs, extra, nil
}

func (s *Service) getLocationCanDisplay(ctx context.Context, param *space.ArchiveCursorParam, without []uparcapi.Without) (bool, error) {
	args := &uparcapi.ArcPassedExistReq{
		Mid:     param.Vmid,
		Aid:     param.FromViewAid,
		Without: without,
	}
	canDisplay, err := s.upArcDao.ArcPassedExist(ctx, args)
	if err != nil {
		return false, errors.Wrapf(err, "getLocationCanDisplay() s.upArcDao.ArcPassedExist args: %+v", args)
	}
	return canDisplay.Exist, nil
}

func (s *Service) getArcPassedByAid(ctx context.Context, param *space.ArchiveCursorParam, without []uparcapi.Without) (*uparcapi.ArcPassedByAidReply, error) {
	args := &uparcapi.ArcPassedByAidReq{
		Mid:     param.Vmid,
		Aid:     param.Aid,
		Ps:      int64(param.Ps),
		Sort:    param.Sort,
		Without: without,
	}
	switch param.Order {
	case space.ArchivePlay:
		args.Order = uparcapi.SearchOrder_click
	default:
		args.Order = uparcapi.SearchOrder_pubtime
	}
	reply, err := s.upArcDao.ArcPassedByAid(ctx, args)
	if err != nil {
		return nil, errors.Wrapf(err, "getArcPassedByAid() s.upArcDao.ArcPassedByAid args: %+v", args)
	}
	return reply, nil
}

func constructArcPlayerCursorExtra(in *uparcapi.ArcPassedByAidReply, locationCanDisplay bool) *space.ArcPlayerCursorExtra {
	extra := &space.ArcPlayerCursorExtra{
		Total:      in.Total,
		CanDisplay: locationCanDisplay,
	}
	if in.Cursor != nil {
		extra.HasMore = in.Cursor.HasMore
	}
	return extra
}

func (s *Service) resolveArcPlayer(ctx context.Context, param *space.ArchiveCursorParam, arcPassed []*uparcapi.ArcPassedWithIndex) ([]*space.ArcPlayerCursor, error) {
	var playAvs []*api.PlayAv
	for _, val := range arcPassed {
		playAvs = append(playAvs, &api.PlayAv{Aid: val.Archive.Aid})
	}
	apm, err := s.arcDao.ArcsPlayer(ctx, playAvs, false)
	if err != nil {
		return nil, errors.Wrapf(err, "resolveArcPlayer() s.arcDao.ArcsPlayer playAvs: %+v", playAvs)
	}
	var arcs []*space.ArcPlayerCursor
	for _, val := range arcPassed {
		if arc, ok := apm[val.Archive.Aid]; ok {
			arcs = append(arcs, &space.ArcPlayerCursor{
				ArcPlayer:  arc,
				CursorAttr: resolveCursorAttr(arc, param, val.Rank),
			})
		}
	}
	return arcs, nil
}

func resolveCursorAttr(arc *api.ArcPlayer, param *space.ArchiveCursorParam, rank int64) *space.CursorAttr {
	return &space.CursorAttr{
		Rank:             rank,
		IsLastWatchedArc: arc.Arc.Aid == param.FromViewAid,
	}
}

func (s *Service) getUserName(ctx context.Context, uid int64) (string, error) {
	account, err := s.accDao.Card(ctx, uid)
	if err != nil {
		return "", errors.Wrapf(err, "getUserName() s.accDao.Card uid: %d", uid)
	}
	return account.Name, nil
}

func (s *Service) getHistoryPosition(ctx context.Context, mid, vmid int64, business string) (*space.HistoryPosition, error) {
	position, err := s.hisDao.Position(ctx, mid, vmid, business)
	if err != nil {
		return nil, errors.Wrapf(err, "getHistoryPosition() s.hisDao.Position mid: %d, vmid: %d", mid, vmid)
	}
	return position, nil
}

//nolint:unparam
func (s *Service) resolveSpaceArcList(arcs []*space.ArcPlayerCursor, param *space.ArchiveCursorParam, extra *space.ArcPlayerCursorExtra, uname string, position *space.HistoryPosition, seasons map[int64]*ugcSeasonGrpc.Season) (*space.ArcCursorList, error) {
	hasPrev, hasNext := constructHasMore(param, extra.HasMore)
	return &space.ArcCursorList{
		EpisodicButton:     s.resolveEpisodicButton(arcs, param, extra.Total, uname, position),
		Order:              constructArcListOrder(),
		Count:              extra.Total,
		Item:               s.resolveArcListItem(arcs, param, seasons),
		LastWatchedLocator: constructLastWatchedLocator(extra.CanDisplay),
		HasPrev:            hasPrev,
		HasNext:            hasNext,
	}, nil
}

func constructHasMore(param *space.ArchiveCursorParam, hasMore bool) (bool, bool) {
	// 第一刷逻辑
	if param.Aid == 0 {
		return false, hasMore
	}
	// 其余刷逻辑
	switch param.Sort {
	case "asc":
		return hasMore, true
	default:
		return true, hasMore
	}
}

func constructArcListOrder() []*space.ArcOrder {
	return []*space.ArcOrder{{Title: "最新发布", Value: space.ArchiveNew}, {Title: "最多播放", Value: space.ArchivePlay}}
}

func constructLastWatchedLocator(canDisplay bool) *space.LastWatchedLocator {
	const (
		_displayThreshold = 10
		_insertRanking    = 6
	)
	return &space.LastWatchedLocator{
		DisplayThreshold: _displayThreshold,
		InsertRanking:    _insertRanking,
		Text:             "定位至上次观看",
		CanDisplay:       canDisplay,
	}
}

func (s *Service) resolveArcListItem(arcs []*space.ArcPlayerCursor, param *space.ArchiveCursorParam, seasons map[int64]*ugcSeasonGrpc.Season) []*space.ArcItem {
	var res []*space.ArcItem
	for _, v := range arcs {
		if v.Arc == nil || !v.Arc.IsNormal() {
			continue
		}
		si := &space.ArcItem{}
		si.FromArc(v.ArcPlayer, s.hotAids, s.c.Custom.UpArcHasShare, true, seasons[v.Arc.Aid])
		si.CursorAttr = v.CursorAttr
		res = append(res, si)
	}
	res = resolveArcItemByParam(res, param)
	return res
}

func resolveArcItemByParam(cursorArcs []*space.ArcItem, param *space.ArchiveCursorParam) []*space.ArcItem {
	if len(cursorArcs) == 0 {
		return nil
	}
	switch param.Sort {
	case "asc":
		// 排序为升序时视频列表仍需要从上至下拼接
		cursorArcs = reverseArcListItem(cursorArcs)
		// 首刷或游标非连续下滑时直接返回数据
		if param.Aid == 0 || param.IncludeCursor {
			return cursorArcs
		}
		// 游标连续下滑时需要截掉已有临界视频
		return removeArcItemCursor(cursorArcs, strconv.FormatInt(param.Aid, 10))
	default:
		if param.Aid == 0 || param.IncludeCursor {
			return cursorArcs
		}
		return removeArcItemCursor(cursorArcs, strconv.FormatInt(param.Aid, 10))
	}
}

// 移除游标已有临界视频
func removeArcItemCursor(cursorArcs []*space.ArcItem, param string) []*space.ArcItem {
	var resArcs []*space.ArcItem
	for i, v := range cursorArcs {
		if v.Param == param {
			if i < len(cursorArcs)-1 {
				resArcs = append(resArcs, cursorArcs[i+1:]...)
			}
			break
		}
		resArcs = append(resArcs, v)
	}
	return resArcs
}

func reverseArcListItem(s []*space.ArcItem) []*space.ArcItem {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func (s *Service) resolveEpisodicButton(arcs []*space.ArcPlayerCursor, param *space.ArchiveCursorParam, total int64, uname string, position *space.HistoryPosition) *space.EpisodicButton {
	if s.c.SpaceArchive.EpisodicOpen && uname != "" && len(arcs) > 1 { // 只有1个稿件的时候不显示一键连播按钮
		for _, item := range s.c.SpaceArchive.EpisodicMid {
			if item == param.Vmid {
				return nil
			}
		}
	}
	res := &space.EpisodicButton{
		Text: s.c.SpaceArchive.EpisodicText,
	}
	var params = url.Values{}
	params.Set("offset", _spaceArchiveOffset)
	params.Set("desc", _spaceArchiveDesc)
	params.Set("oid", _spaceArchiveOid)
	params.Set("ps", _spaceArchivePS)
	params.Set("order", _spaceArchiveOrder)
	params.Set("page_type", _spaceArchivePageType)
	params.Set("user_name", uname)
	params.Set("playlist_intro", s.c.SpaceArchive.EpisodicDesc)
	params.Set("total_count", strconv.FormatInt(total, 10))
	switch param.Order {
	case space.ArchiveNew:
		params.Set("sort_field", "1")
		if position != nil {
			if position.Oid > 0 {
				res.Text = s.c.SpaceArchive.EpisodicText1
			}
			params.Set("offset", strconv.Itoa(position.Offset))
			params.Set("desc", strconv.Itoa(position.Desc))
			params.Set("oid", strconv.Itoa(position.Oid))
		}
	case space.ArchivePlay:
		params.Set("sort_field", "2")
		params.Set("sort_hidden", "1")
	}
	res.Uri = fmt.Sprintf(_spaceArchiveUri, param.Vmid, params.Encode())
	return res
}
