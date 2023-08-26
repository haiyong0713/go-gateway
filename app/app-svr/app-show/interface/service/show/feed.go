package show

import (
	"context"
	"hash/crc32"
	"strconv"
	"time"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-show/interface/model/rank"
	"go-gateway/app/app-svr/archive/service/api"

	"go-gateway/app/app-svr/app-show/interface/model"
	"go-gateway/app/app-svr/app-show/interface/model/card"
	"go-gateway/app/app-svr/app-show/interface/model/feed"
)

var (
	_emptyList = []*feed.Item{}
)

// FeedIndex feed index
// nolint:gomnd
func (s *Service) FeedIndex(c context.Context, mid, idx int64, plat int8, build, loginEvent int, lastParam, mobiApp, device, buvid string, now time.Time) (res []*feed.Item) {
	var (
		ps     = 10
		isIpad = plat == model.PlatIPad
		cards  []*card.PopularCard
	)
	if isIpad {
		ps = 20
	}
	var key int
	if mid > 0 {
		key = int((mid / 1000) % 10)
	} else {
		key = int((crc32.ChecksumIEEE([]byte(buvid)) / 1000) % 10)
	}
	cards = s.PopularCardTenList(c, key, int(idx), ps)
	if len(cards) == 0 {
		res = _emptyList
		return
	}
	res = s.dealItem(c, plat, build, ps, cards, idx, lastParam, now, mid, mobiApp, device)
	if len(res) == 0 {
		res = _emptyList
		return
	}
	for i := 0; i < len(res); i++ { // 老接口全部用全部热门
		res[i].ChannelID = _allHotID
		res[i].ChannelOrder = _allHotOrder
		res[i].ChannelName = _allHotEntrance
	}
	//infoc
	infoc := &feedInfoc{
		mobiApp:    mobiApp,
		device:     device,
		build:      strconv.Itoa(build),
		now:        now.Format("2006-01-02 15:04:05"),
		loginEvent: strconv.Itoa(loginEvent),
		mid:        strconv.FormatInt(mid, 10),
		buvid:      buvid,
		page:       strconv.Itoa((int(idx) / ps) + 1),
		feed:       res,
		url:        "/x/v2/show/popular",
	}
	s.infocfeed(infoc)
	return
}

// dealItem feed item
func (s *Service) dealItem(c context.Context, plat int8, build, ps int, cards []*card.PopularCard, idx int64, _ string, _ time.Time, mid int64, mobiApp, device string) (is []*feed.Item) {
	const _rankCount = 3
	var (
		uri map[int64]string
		// key
		max             = int64(100)
		_fTypeOperation = "operation"
		aids            []int64
		am              map[int64]*api.Arc
		feedcards       []*card.PopularCard
		innerAttr       = make(map[int64]*rank.InnerAttr)
		err             error
	)
LOOP:
	for pos, ca := range cards {
		var cardIdx = idx + int64(pos+1)
		if cardIdx > max && ca.FromType != _fTypeOperation {
			continue
		}
		if config, ok := ca.PopularCardPlat[plat]; ok {
			for _, l := range config {
				if model.InvalidBuild(build, l.Build, l.Condition) {
					continue LOOP
				}
			}
		} else if ca.FromType == _fTypeOperation {
			continue LOOP
		}
		tmp := &card.PopularCard{}
		*tmp = *ca
		tmp.Idx = cardIdx
		feedcards = append(feedcards, tmp)
		switch ca.Type {
		case model.GotoAv:
			aids = append(aids, ca.Value)
		}
		if len(feedcards) == ps {
			break
		}
	}
	if len(aids) != 0 {
		eg := errgroup.WithContext(c)
		eg.Go(func(c context.Context) error {
			if am, err = s.arc.ArchivesPB(c, aids, mid, mobiApp, device); err != nil {
				s.pMiss.Incr("popularcard_Archives")
			} else {
				s.pHit.Incr("popularcard_Archives")
			}
			return nil
		})
		eg.Go(func(c context.Context) error {
			innerAttr = s.controld.CircleReqInternalAttr(c, aids)
			return nil
		})
		_ = eg.Wait()

	}
	for _, ca := range feedcards {
		i := &feed.Item{}
		i.FromType = ca.FromType
		i.Idx = ca.Idx
		i.Pos = ca.Pos
		switch ca.Type {
		case model.GotoAv:
			a := am[ca.Value]
			isOsea := model.IsOverseas(plat)
			var overSeaFlag bool
			if inva, k := innerAttr[ca.Value]; k && inva != nil {
				overSeaFlag = inva.OverSeaBlock
			}
			if a != nil && a.IsNormal() && (!isOsea || (isOsea && !overSeaFlag)) {
				i.FromPlayerAv(a, uri[a.Aid])
				i.FromRcmdReason(ca)
				i.Goto = ca.Type
				is = append(is, i)
			}
		case model.GotoRank:
			if rankAids := s.rankAidsCache; len(rankAids) >= _rankCount {
				i.FromRank(rankAids, s.rankScoreCache, s.rankArchivesCache)
				if i.Goto != "" {
					is = append(is, i)
				}
			}
		}
	}
	if rl := len(is); rl == 0 {
		is = _emptyList
		return
	}
	return
}
