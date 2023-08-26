package search

import (
	"context"
	"strconv"
	"time"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	accdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/account"
	arcdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/archive"
	bangumidao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/bangumi"
	gallerydao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/gallery"
	livedao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/live"
	srchdao "go-gateway/app/app-svr/app-interface/interface-legacy/dao/search"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/search"
	"go-gateway/app/app-svr/archive/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	memberAPI "git.bilibili.co/bapis/bapis-go/account/service/member"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	gallerygrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"
	pgcsearch "git.bilibili.co/bapis/bapis-go/pgc/service/card/search/v1"
	upgrpc "git.bilibili.co/bapis/bapis-go/up-archive/service"

	"github.com/pkg/errors"
)

type Service struct {
	c          *conf.Config
	srchDao    *srchdao.Dao
	accDao     *accdao.Dao
	galleryDao *gallerydao.Dao
	arcDao     *arcdao.Dao
	liveDao    *livedao.Dao
	bangumiDao *bangumidao.Dao
}

// New is search service initial func
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:          c,
		srchDao:    srchdao.New(c),
		accDao:     accdao.New(c),
		galleryDao: gallerydao.New(c),
		liveDao:    livedao.New(c),
		arcDao:     arcdao.New(c),
		bangumiDao: bangumidao.New(c),
	}
	return
}

// Suggest3 for search suggest
func (s *Service) Suggest3(c context.Context, mid int64, platform, buvid, keyword, device string, build, highlight int, mobiApp string, now time.Time) (res *search.SuggestionResult3) {
	var (
		suggest   *search.Suggest3
		err       error
		avids     []int64
		avm       map[int64]*api.Arc
		roomIDs   []int64
		entryRoom map[int64]*livexroomgate.EntryRoomInfoResp_EntryList
		ssids     []int32
		seasonm   map[int32]*pgcsearch.SearchCardProto
		mids      []int64
		nftRegion map[int64]*gallerygrpc.NFTRegion
	)
	res = &search.SuggestionResult3{}
	if s.c.Switch.SearchSuggest {
		return
	}
	if suggest, err = s.srchDao.Suggest3(c, mid, platform, buvid, keyword, device, build, highlight, mobiApp, now); err != nil {
		log.Error("%+v", err)
		return
	}
	plat := model.Plat(mobiApp, device)
	var buildLimit bool
	buildLimit = cdm.ShowLive(mobiApp, device, build)
	if s.c.Feature.FeatureBuildLimit.Switch {
		buildLimit = cdm.ShowLiveV2(c, s.c.Feature.FeatureBuildLimit.ShowLive, &feature.OriginResutl{
			BuildLimit: !cdm.ShowLive(mobiApp, device, build),
		})
	}
	for _, v := range suggest.Result {
		if v.TermType == search.SuggestionJump {
			if v.SubType == search.SuggestionAV {
				avids = append(avids, v.Ref)
			}
			if v.SubType == search.SuggestionLive && buildLimit && !model.IsOverseas(plat) {
				roomIDs = append(roomIDs, v.Ref)
			}
		} else if v.TermType == search.SuggestionJumpPGC && !model.IsOverseas(plat) {
			if v.PGC == nil && v.PGC.SeasonID != 0 {
				continue
			}
			ssids = append(ssids, int32(v.PGC.SeasonID))
		} else if v.TermType == search.SuggestionJumpUser {
			mids = append(mids, v.User.Mid)
		}
	}
	g, ctx := errgroup.WithContext(c)
	if len(mids) != 0 {
		g.Go(func() (err error) {
			nftRegion, err = s.getNFTIconInfo(ctx, mids)
			if err != nil {
				log.Error("s.getNFTIconInfo mids=%+v, err=%+v", mids, err)
				return nil
			}
			return
		})
	}
	if len(avids) != 0 {
		g.Go(func() (err error) {
			if avm, err = s.arcDao.Archives(ctx, avids, mobiApp, device, mid); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(roomIDs) != 0 {
		g.Go(func() (err error) {
			req := &livexroomgate.EntryRoomInfoReq{
				EntryFrom: []string{model.DefaultLiveEntry},
				RoomIds:   roomIDs,
				Uid:       mid,
				Uipstr:    metadata.String(ctx, metadata.RemoteIP),
				Platform:  platform,
				Build:     int64(build),
				Network:   "other",
			}
			if entryRoom, err = s.liveDao.EntryRoomInfo(ctx, req); err != nil {
				log.Error("Failed to get entry room info: %+v: %+v", req, err)
				err = nil
				return
			}
			return
		})
	}
	if len(ssids) != 0 {
		g.Go(func() (err error) {
			if seasonm, err = s.bangumiDao.SugOGV(ctx, ssids); err != nil {
				log.Error("%v", err)
				err = nil
			}
			return
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	for _, v := range suggest.Result {
		if v.SubType == search.SuggestionLive && (!buildLimit || model.IsOverseas(plat)) {
			continue
		}
		if v.TermType == search.SuggestionJumpPGC && model.IsOverseas(plat) {
			continue
		}
		si := &search.Item{}
		si.FromSuggest3(v, avm, entryRoom, seasonm, nftRegion)
		res.List = append(res.List, si)
	}
	res.TrackID = suggest.TrackID
	res.ExpStr = suggest.ExpStr
	return
}

// DefaultWords search for default words
func (s *Service) DefaultWords(c context.Context, mid int64, build, from int, buvid, platform, mobiApp, device string, loginEvent int64, extParam *search.DefaultWordsExtParam) (res *search.DefaultWords, err error) {
	if res, err = s.srchDao.DefaultWords(c, mid, build, from, buvid, platform, mobiApp, device, loginEvent, extParam); err != nil {
		log.Error("%+v", err)
	}
	return
}

func (s *Service) getNFTIconInfo(ctx context.Context, mids []int64) (map[int64]*gallerygrpc.NFTRegion, error) {
	req := &memberAPI.NFTBatchInfoReq{
		Mids:   mids,
		Status: "inUsing",
		Source: "face",
	}
	reply, err := s.accDao.NFTBatchInfo(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "s.accDao.NFTBatchInfo req=%+v", req)
	}
	var (
		nftIDs        []string
		nftRegionInfo *gallerygrpc.GetNFTRegionReply
	)
	for _, v := range reply.GetNftInfos() {
		nftIDs = append(nftIDs, v.NftId)
	}
	if len(nftIDs) == 0 {
		return nil, err
	}
	if nftRegionInfo, err = s.galleryDao.GetNFTRegionBatch(ctx, nftIDs); err != nil {
		return nil, errors.Wrapf(err, "s.galleryDao.GetNFTRegion nftIDs=%+v", nftIDs)
	}
	res := make(map[int64]*gallerygrpc.NFTRegion, len(nftIDs))
	for _, info := range reply.GetNftInfos() {
		if v, ok := nftRegionInfo.Region[info.NftId]; ok {
			res[info.Mid] = v
		}
	}
	return res, nil
}

func (s *Service) SpaceSearch(ctx context.Context, vmid int64, keyword string, highlight, isTitle, isIpad bool, pn, ps int) (*search.SpaceResult, error) {
	kwFields := []upgrpc.KwField{upgrpc.KwField_title, upgrpc.KwField_content}
	if isTitle {
		kwFields = []upgrpc.KwField{upgrpc.KwField_title}
	}
	reply, err := s.srchDao.ArcPassedSearch(ctx, vmid, keyword, highlight, kwFields, upgrpc.SearchOrder_pubtime, "desc", int64(pn), int64(ps), isIpad)
	if err != nil {
		return nil, err
	}
	var item []*search.Item
	for _, v := range reply.Archives {
		if v == nil {
			continue
		}
		param := strconv.FormatInt(v.Aid, 10)
		i := &search.Item{
			Title:    v.Title,
			Cover:    v.Pic,
			Param:    param,
			Goto:     model.GotoAv,
			URI:      model.FillURI(model.GotoAv, param, nil),
			Play:     int(v.Stat.View),
			Danmaku:  int(v.Stat.Danmaku),
			Duration: model.DurationString(v.Duration),
			PTime:    v.PubDate,
		}
		// pgc稿件url直接跳转
		if model.AttrVal(v.Attribute, model.AttrBitIsPGC) == model.AttrYes && v.RedirectURL != "" {
			i.URI = v.RedirectURL
		}
		item = append(item, i)
	}
	return &search.SpaceResult{
		Total: int(reply.Total),
		Page:  pn,
		Item:  item,
	}, nil
}
