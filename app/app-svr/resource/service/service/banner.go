package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-feed/interface/model/sets"
	pb "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/app-svr/resource/service/model"

	location "git.bilibili.co/bapis/bapis-go/community/service/location"
	"github.com/pkg/errors"

	farm "go-farm"
)

const (
	_categoryHash = "category"
	_inlineAd     = 4
)

type bannerHash struct {
	Title string `json:"title"`
	Image string `json:"image"`
	URI   string `json:"uri"`
}

// loadBannerCahce load banner cache.
//
//nolint:gocognit
func (s *Service) loadBannerCahce() (err error) {
	var hashbs = make(map[int][]*model.Banner)
	// 强运营帧
	boss, err := s.res.Boss(context.TODO())
	if err != nil {
		log.Error("s.res.Boss error(%v)", err)
		return
	}
	s.bossBannerCache = boss
	for _, bos := range boss {
		for resID, bo := range bos {
			hashbs[resID] = append(hashbs[resID], bo...)
		}
	}
	log.Info("load bossBannerCache success")
	// 固定帧
	nbs, err := s.res.Banner(context.TODO())
	if err != nil {
		log.Error("s.res.Banner error(%v)", err)
		return
	}
	for _, nb := range nbs {
		for resID, n := range nb {
			hashbs[resID] = append(hashbs[resID], n...)
		}
	}
	// 重新聚合固定投放
	var nbsNew = make(map[int8]map[int][]*model.Banner)
	for plat, nb := range nbs {
		for resID, n := range nb {
			var (
				rankWeightBans = make(map[int]map[int][]*model.Banner) // rank->weight->[]*banner
				ranks          []int
			)
			for _, ntmp := range n {
				weightBans, ok := rankWeightBans[ntmp.Rank]
				if !ok {
					weightBans = make(map[int][]*model.Banner)
					rankWeightBans[ntmp.Rank] = weightBans
					ranks = append(ranks, ntmp.Rank)
				}
				weightBans[ntmp.Weight] = append(weightBans[ntmp.Weight], ntmp)
			}
			if len(rankWeightBans) == 0 {
				continue
			}
			sort.Ints(ranks) // rank 正序
			for _, rank := range ranks {
				var weights []int
				for weight := range rankWeightBans[rank] {
					weights = append(weights, weight)
				}
				sort.Sort(sort.Reverse(sort.IntSlice(weights))) // weight倒序
				for _, weight := range weights {
					rand.Seed(time.Now().Unix())
					count := len(rankWeightBans[rank][weight]) // 获取当前权重组包含元素数
					if count == 0 {
						continue
					}
					ban := rankWeightBans[rank][weight][rand.Intn(count)] // 通过随机数获取当前权重组元素
					if ban != nil {
						nbsNewPlat, ok := nbsNew[plat]
						if !ok {
							nbsNewPlat = make(map[int][]*model.Banner)
							nbsNew[plat] = nbsNewPlat
						}
						nbsNewPlat[resID] = append(nbsNewPlat[resID], ban)
						break
					}
				}
			}
		}
	}
	s.bannerCache = nbsNew
	log.Info("load bannerCache success")
	// 合并强运营和固定帧 计算hash
	var bannerHashTpm = make(map[int]string, len(nbs))
	for resID, hashb := range hashbs {
		var bhs []*bannerHash
		for _, hb := range hashb {
			if hb != nil {
				bhs = append(bhs, &bannerHash{Title: hb.Title, Image: hb.Image, URI: hb.Image})
			}
		}
		bannerHashTpm[resID] = hash(bhs)
	}
	s.bannerHashCache = bannerHashTpm
	log.Info("load BannerHashCache success")
	// 推荐池
	cbs, err := s.res.Category(context.TODO())
	if err != nil {
		log.Error("s.res.Category error(%v)", err)
		return
	}
	s.categoryBannerCache = cbs
	log.Info("load categoryBannerCache success")
	// banner limit
	limit, err := s.res.Limit(context.TODO())
	if err != nil {
		log.Error("s.dao.Limit error(%v)", err)
		return
	}
	s.bannerLimitCache = limit
	log.Info("load BannerLimitCache success")
	// all banner
	allBanner, err := s.res.ALLBanner(context.Background())
	if err != nil {
		log.Error("Failed to load add banner: %+v", err)
		return
	}
	s.allBannerCache = allBanner
	log.Info("load ALLBannerCache success")
	return
}

// hash get banner hash.
func hash(v []*bannerHash) (value string) {
	bs, err := json.Marshal(v)
	if err != nil {
		log.Error("json.Marshal error(%v)", err)
		return
	}
	value = strconv.FormatUint(farm.Hash64(bs), 10)
	return
}

// Banners get banners by plat, build channel, ip for app-feed.
//
//nolint:gocognit
func (s *Service) Banners(c context.Context, plat int8, build int, aid, mid, splashID int64, resIdsStr, channel, ip, buvid, network, mobiApp, device, openEvent, adExtra, version string, isAd bool) (res *model.Banners) {
	res = &model.Banners{}
	var (
		cpmResBus map[int]map[int]*model.Banner
		resIds    []string
		banner    map[int][]*model.Banner
		newHash   string
	)
	if resIds = strings.Split(resIdsStr, ","); len(resIds) > 1 {
		version = ""
		isAd = false
	}
	banner = map[int][]*model.Banner{}
	for _, resIDStr := range resIds {
		if resIDStr == "" {
			continue
		}
		resID, e := strconv.Atoi(resIDStr)
		if e != nil {
			//log.Warn("strconv.Atoi(%s) error(%v)", resIDStr, e)
			continue
		}
		if (version != "" && (version == s.bannerHashCache[resID])) || (version == _categoryHash && (s.bannerHashCache[resID] == "")) {
			//log.Warn("Banners() plat(%v) res(%v) version(%v) same as hash cache, return nil", plat, resID, version)
			continue
		}
		if isAd {
			cpmResBus = s.cpmBanners(c, aid, mid, splashID, build, resIDStr, mobiApp, device, buvid, network, ip, openEvent, adExtra)
		}
		var (
			resBs, cbcs, resAll []*model.Banner
			cbc                 = s.categoryBannerCache[plat] // operater category banner
			bArea               []string
			ok                  bool
			maxBannerIndex      int
		)
		// 推荐池
		if len(cbc) > 0 {
			if cbcs, ok = cbc[resID]; ok {
				btime := strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
				for _, b := range cbcs {
					if s.filterBs(c, plat, build, channel, b) {
						continue
					}
					if b.Area != "" {
						bArea = append(bArea, b.Area)
					}
					tmp := &model.Banner{}
					*tmp = *b
					tmp.ServerType = 0
					tmp.RequestId = btime
					if tmp.Rank > maxBannerIndex {
						maxBannerIndex = tmp.Rank
					}
					resBs = append(resBs, tmp)
				}
			}
		}
		if len(resBs) > maxBannerIndex {
			maxBannerIndex = len(resBs)
		}
		var (
			bbcs    []*model.Banner
			bbc     = s.bossBannerCache[plat] // operater boss banner
			tmpBbcs []*model.Banner
			isBoss  bool
		)
		// 插入帧逻辑
		if len(bbc) > 0 {
			if bbcs, ok = bbc[resID]; ok {
				btime := strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
				for _, b := range bbcs {
					if s.filterBs(c, plat, build, channel, b) {
						continue
					}
					if b.Area != "" {
						bArea = append(bArea, b.Area)
					}
					tmp := &model.Banner{}
					*tmp = *b
					tmp.ServerType = 0
					tmp.RequestId = btime
					if tmp.Rank > maxBannerIndex {
						maxBannerIndex = tmp.Rank
					}
					tmpBbcs = append(tmpBbcs, tmp)
					isBoss = true
				}
			}
		}
		if (len(resBs) + len(tmpBbcs)) > maxBannerIndex {
			maxBannerIndex = len(resBs) + len(tmpBbcs)
		}
		var (
			cpmBus              map[int]*model.Banner // cpm ad
			allRank             []int
			tmpCmps, tmpTopView []*model.Banner
		)
		// 广告逻辑
		if cpmBus, ok = cpmResBus[resID]; ok && len(cpmBus) > 0 {
			var cpmMs = map[int]*model.Banner{}
			for _, cpm := range cpmBus {
				if cpm.IsAdReplace {
					if cpm.SplashID == 0 {
						cpmMs[cpm.Rank] = cpm
						allRank = append(allRank, cpm.Rank)
					} else if splashID == cpm.SplashID {
						tmpTopView = append(tmpTopView, cpm)
					}
					delete(cpmBus, cpm.Rank)
				}
			}
			if len(allRank) > 0 {
				sort.Ints(allRank)
				for _, key := range allRank {
					if cpmMs[key].Rank > maxBannerIndex {
						maxBannerIndex = cpmMs[key].Rank
					}
					tmpCmps = append(tmpCmps, cpmMs[key])
				}
			}
		}
		if (len(resBs) + len(tmpCmps) + len(tmpBbcs)) > maxBannerIndex {
			maxBannerIndex = len(resBs) + len(tmpCmps) + len(tmpBbcs)
		}
		var (
			plm         = s.bannerCache[plat] // operater normal banner
			tmpBs, plbs []*model.Banner
		)
		// 固定投放
		if len(plm) > 0 {
			if plbs, ok = plm[resID]; ok {
				btime := strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
				for _, b := range plbs {
					if s.filterBs(c, plat, build, channel, b) {
						continue
					}
					if b.Area != "" {
						bArea = append(bArea, b.Area)
					}
					tmp := &model.Banner{}
					*tmp = *b
					tmp.ServerType = 0
					tmp.RequestId = btime
					if tmp.Rank > maxBannerIndex {
						maxBannerIndex = tmp.Rank
					}
					tmpBs = append(tmpBs, tmp)
				}
			}
		}
		if (len(resBs) + len(tmpCmps) + len(tmpBs) + len(tmpBbcs)) > maxBannerIndex {
			maxBannerIndex = len(resBs) + len(tmpCmps) + len(tmpBs) + len(tmpBbcs)
		}
		var bossIndex, tcIndex, tbIndex, cIndex int
		for i := 1; i <= maxBannerIndex; i++ {
			if bossIndex < len(tmpBbcs) {
				boss := tmpBbcs[bossIndex]
				resAll = append(resAll, boss)
				bossIndex++
				continue
			}
			if tcIndex < len(tmpCmps) {
				tc := tmpCmps[tcIndex]
				if tc.Rank <= i {
					resAll = append(resAll, tc)
					tcIndex++
					continue
				}
			}
			if tbIndex < len(tmpBs) {
				tb := tmpBs[tbIndex]
				if tb.Rank <= i {
					resAll = append(resAll, tb)
					tbIndex++
					continue
				}
			}
			if cIndex < len(resBs) {
				cb := resBs[cIndex]
				resAll = append(resAll, cb)
				cIndex++
			}
		}
		for i, b := range resAll {
			if cpm, ok := cpmBus[i+1]; ok && !b.IsAdReplace { // NOTE: surplus cpm is ad loc
				b.IsAdLoc = true
				b.IsAd = cpm.IsAd
				b.CmMark = cpm.CmMark
				b.SrcId = cpm.SrcId
				b.RequestId = cpm.RequestId
				b.ClientIp = cpm.ClientIp
			}
		}
		// 插入topview
		if len(tmpTopView) > 0 {
			resAll = append([]*model.Banner{}, append(tmpTopView, resAll...)...)
		}
		if resID == 467 || resID == 631 || resID == 771 {
			if isBoss && len(resAll) > 6 {
				resAll = resAll[:6]
			} else if !isBoss && len(resAll) > 5 {
				resAll = resAll[:5]
			}
		} else if resID == 3143 || resID == 3150 || resID == 3179 {
			if isBoss && len(resAll) > 4 {
				resAll = resAll[:4]
			} else if !isBoss && len(resAll) > 3 {
				resAll = resAll[:3]
			}
		} else {
			if max, ok := s.bannerLimitCache[resID]; ok && len(resAll) > max {
				resAll = resAll[:max]
			}
		}
		for i := 0; i < len(resAll); i++ {
			resAll[i].Index = i + 1
			resAll[i].ResourceID = resID
		}
		if len(resAll) > 0 {
			var (
				auths  map[int64]*location.Auth
				resBs2 []*model.Banner
			)
			if len(bArea) > 0 {
				reply, e := s.locGRPC.AuthPIDs(c, &location.AuthPIDsReq{Pids: strings.Join(bArea, ","), IpAddr: ip, InvertedMode: true})
				if e != nil {
					log.Error("%v", e)
				}
				if reply != nil {
					auths = reply.Auths
				}
			}
			for _, resB := range resAll {
				if resB.Area != "" {
					var (
						pid int64
						e   error
					)
					if pid, e = strconv.ParseInt(resB.Area, 10, 64); e != nil {
						log.Warn("banner strconv.ParseInt(%v) error(%v)", resB.Area, e)
					} else {
						if auth, ok := auths[pid]; ok && auth.Play == int64(location.Status_Forbidden) {
							//log.Warn("resID(%v) pid(%v) ip(%v) in zone limit", resID, resB.Area, ip)
							continue
						}
					}
				}
				resBs2 = append(resBs2, resB)
			}
			banner[resID] = resBs2
		}
		if newHash = s.bannerHashCache[resID]; newHash == "" {
			newHash = _categoryHash
		}
	}
	res.Banner = banner
	res.Version = newHash
	return
}

// Banners get banners by plat, build channel, ip for app-feed.
func (s *Service) Banners2(c context.Context, arg *pb.BannersRequest) (resp *pb.BannersReply, err error) {
	plat := int8(arg.Plat)
	build := int(arg.Build)
	aid := arg.Aid
	mid := arg.Mid
	resIdsStr := arg.ResIDs
	channel := arg.Channel
	ip := arg.Ip
	buvid := arg.Buvid
	network := arg.Network
	mobiApp := arg.MobiApp
	device := arg.Device
	openEvent := arg.OpenEvent
	adExtra := arg.AdExtra
	version := arg.Version
	isAd := arg.IsAd
	splashID := arg.SplashId
	res := s.Banners(c, int8(plat), build, aid, mid, splashID, resIdsStr, channel, ip, buvid, network, mobiApp, device, openEvent, adExtra, version, isAd)
	resp = new(pb.BannersReply)
	resp.Banners = make(map[int32]*pb.Banners)
	for k, ban := range res.Banner {
		for _, result := range ban {
			tres := &pb.Banner{
				Id:           int64(result.ID),
				ParentId:     int64(result.ParentID),
				Plat:         int32(result.Plat),
				Module:       result.Module,
				Position:     result.Position,
				Title:        result.Title,
				Image:        result.Image,
				Hash:         result.Hash,
				URI:          result.URI,
				Goto:         result.Goto,
				Value:        result.Value,
				Param:        result.Param,
				Channel:      result.Channel,
				Build:        int32(result.Build),
				Condition:    result.Condition,
				Area:         result.Area,
				Rank:         int64(result.Rank),
				Rule:         result.Rule,
				Type:         int32(result.Type),
				Start:        int64(result.Start),
				End:          int64(result.End),
				MTime:        int64(result.MTime),
				ResourceId:   int64(result.ResourceID),
				RequestId:    result.RequestId,
				CreativeId:   int64(result.CreativeId),
				SrcId:        int64(result.SrcId),
				IsAd:         result.IsAd,
				IsAdReplace:  result.IsAdReplace,
				IsAdLoc:      result.IsAdLoc,
				CmMark:       int64(result.CmMark),
				AdCb:         result.AdCb,
				ShowUrl:      result.ShowUrl,
				ClickUrl:     result.ClickUrl,
				ClientIp:     result.ClientIp,
				Index:        int64(result.Index),
				ServerType:   int64(result.ServerType),
				Extra:        result.Extra,
				CreativeType: int64(result.CreativeType),
				SubTitle:     result.SubTitle,
				SplashId:     result.SplashID,
			}
			var (
				bs *pb.Banners
				ok bool
			)
			if bs, ok = resp.Banners[int32(k)]; !ok {
				bs = &pb.Banners{}
				resp.Banners[int32(k)] = bs
			}
			bs.Banners = append(bs.Banners, tres)
		}
	}
	resp.Version = res.Version
	return
}

// cpmBanners
func (s *Service) cpmBanners(c context.Context, aid, mid, splashID int64, build int, resource, mobiApp, device, buvid, network, ipaddr, openEvent, adExtra string) (banners map[int]map[int]*model.Banner) {
	ipInfo, err := s.locGRPC.Info(c, &location.InfoReq{Addr: ipaddr})
	if err != nil || ipInfo == nil {
		log.Error("CpmsBanners s.locationRPC.Zone(%s) error(%v) or ipinfo is nil", ipaddr, err)
		ipInfo = &location.InfoReply{Addr: ipaddr}
	}
	adr, err := s.cpm.CpmsAPP(c, aid, mid, splashID, build, resource, mobiApp, device, buvid, network, openEvent, adExtra, ipInfo)
	if err != nil || adr == nil {
		log.Error("s.ad.ADRequest error(%v)", err)
		return
	}
	banners = adr.ConvertBanner(ipInfo.Addr, mobiApp, build)
	return
}

// filterBs filter banner.
func (s *Service) filterBs(_ context.Context, plat int8, build int, channel string, b *model.Banner) bool {
	if model.InvalidBuild(build, b.Build, b.Condition) {
		return true
	}
	if model.InvalidChannel(plat, channel, b.Channel) && b.Channel != "" {
		return true
	}
	return false
}

// banner type
const (
	BannerTypeFirstFrame = "FIRST_FRAME"
	BannerTypeFixedFrame = "FIXED_FRAME"
	BannerTypeRcmdPool   = "RCMD_POOL"
	BannerTypeCPM        = "AD_CPM_FRAME"
	BannerTypeTopView    = "TOP_VIEW"
	BannerTypeCPMInline  = "AD_CPM_INLINE"
)

var (
	BannerCPMTypeSet = sets.NewString(BannerTypeCPM, BannerTypeTopView, BannerTypeCPMInline)
)

type bannerSlot struct {
	Type       string
	Banner     *model.Banner
	InlineType string
	InlineID   string
}

type cpmStatus struct {
	TopView *model.Banner
	Banners []*model.Banner
}

type bannerMeta struct {
	pb.BannerMeta
	banner *model.Banner
}

type feedBannerCtx struct {
	staticMeta []*bannerMeta
	cpmStatus  *cpmStatus
}

func (ctx feedBannerCtx) hasTopView() bool {
	return ctx.cpmStatus.TopView != nil
}
func (ctx feedBannerCtx) hasFirstFrame() bool {
	return ctx.staticMeta[0].Type == BannerTypeFirstFrame
}

func InsertBannerSlot(s []*bannerSlot, k int, vs ...*bannerSlot) []*bannerSlot {
	if n := len(s) + len(vs); n <= cap(s) {
		s2 := s[:n]
		copy(s2[k+len(vs):], s[k:])
		copy(s2[k:], vs)
		return s2
	}
	s2 := make([]*bannerSlot, len(s)+len(vs))
	copy(s2, s[:k])
	copy(s2[k:], vs)
	copy(s2[k+len(vs):], s[k:])
	return s2
}

func ReplaceBannerSlot(s []*bannerSlot, k int, target *bannerSlot) []*bannerSlot {
	if len(s) <= k || k < 0 {
		return s
	}
	s[k] = target
	return s
}

func resolverKey(hasTopView, hasFirstFrame bool) string {
	return fmt.Sprintf("%t-%t", hasTopView, hasFirstFrame)
}

type resolver func(*feedBannerCtx) ([]*bannerSlot, int, error)

func withTopViewWithFirstFrame(ctx *feedBannerCtx) ([]*bannerSlot, int, error) {
	//nolint:gomnd
	if len(ctx.staticMeta) < 2 {
		return nil, 0, errors.Errorf("insufficient staticMeta: %+v", ctx.staticMeta)
	}

	cpmStatus := ctx.cpmStatus
	meta := ctx.staticMeta

	template := []*bannerSlot{}
	template = append(template, &bannerSlot{
		Type:   BannerTypeTopView,
		Banner: cpmStatus.TopView,
	})
	template = append(template, asBannerSlot(meta[0]))
	template = append(template, asBannerSlot(meta[1]))
	if len(cpmStatus.Banners) > 0 {
		template = append(template, &bannerSlot{
			Type:   constructCpmType(cpmStatus.Banners[0]),
			Banner: cpmStatus.Banners[0],
		})
	}
	for _, b := range meta[2:] {
		template = append(template, asBannerSlot(b))
	}
	return template, calcVisiableLen(4, template), nil
}

func asBannerSlot(in *bannerMeta) *bannerSlot {
	return &bannerSlot{
		Type:       in.Type,
		Banner:     in.banner,
		InlineID:   in.InlineId,
		InlineType: in.InlineType,
	}
}

func withTopViewNoFirstFrame(ctx *feedBannerCtx) ([]*bannerSlot, int, error) {
	cpmStatus := ctx.cpmStatus
	meta := ctx.staticMeta

	template := []*bannerSlot{}
	template = append(template, &bannerSlot{
		Type:   BannerTypeTopView,
		Banner: cpmStatus.TopView,
	})
	template = append(template, asBannerSlot(meta[0]))
	if len(cpmStatus.Banners) > 0 {
		template = append(template, &bannerSlot{
			Type:   constructCpmType(cpmStatus.Banners[0]),
			Banner: cpmStatus.Banners[0],
		})
	}
	for _, b := range meta[1:] {
		template = append(template, asBannerSlot(b))
	}
	return template, calcVisiableLen(3, template), nil
}

func noTopViewWithFirstFrame(ctx *feedBannerCtx) ([]*bannerSlot, int, error) {
	cpmStatus := ctx.cpmStatus
	meta := ctx.staticMeta

	template := []*bannerSlot{}
	template = append(template, asBannerSlot(meta[0]))
	for _, b := range meta[1:] {
		template = append(template, asBannerSlot(b))
	}
	for _, b := range cpmStatus.Banners {
		if len(template) < b.Rank {
			//log.Warn("Insufficient template to insert ad banner: %+v: %+v", jsonify(template), b)
			continue
		}
		template = InsertBannerSlot(template, b.Rank-1, &bannerSlot{
			Type:   constructCpmType(b),
			Banner: b,
		})
	}
	return template, calcVisiableLen(4, template), nil
}

func noTopViewNoFirstFrame(ctx *feedBannerCtx) ([]*bannerSlot, int, error) {
	cpmStatus := ctx.cpmStatus
	meta := ctx.staticMeta

	template := []*bannerSlot{}
	template = append(template, asBannerSlot(meta[0]))
	for _, b := range meta[1:] {
		template = append(template, asBannerSlot(b))
	}
	for _, b := range cpmStatus.Banners {
		if len(template) < b.Rank {
			//log.Warn("Insufficient template to insert ad banner: %+v: %+v", jsonify(template), b)
			continue
		}
		type_ := constructCpmType(b)
		index := b.Rank - 1
		if type_ == BannerTypeCPMInline && index == 0 { // inline广告替换内容帧
			template = ReplaceBannerSlot(template, index, &bannerSlot{
				Type:   type_,
				Banner: b,
			})
			continue
		}
		template = InsertBannerSlot(template, index, &bannerSlot{
			Type:   type_,
			Banner: b,
		})
	}
	return template, calcVisiableLen(3, template), nil
}

func constructCpmType(b *model.Banner) string {
	if b.CreativeStyle == _inlineAd && b.SplashID == 0 {
		return BannerTypeCPMInline
	}
	return BannerTypeCPM
}

func calcVisiableLen(byDefault int, actually []*bannerSlot) int {
	if len(actually) < byDefault {
		return len(actually)
	}
	return byDefault
}

func constructBannerSlot(meta []*bannerMeta, cpmStatus *cpmStatus) ([]*bannerSlot, int, error) {
	if len(meta) <= 0 {
		return nil, 0, errors.Errorf("insufficient banner meta length: %+v", meta)
	}

	ctx := &feedBannerCtx{
		staticMeta: meta,
		cpmStatus:  cpmStatus,
	}
	templateSolver := map[string]resolver{}
	templateSolver[resolverKey(true, true)] = withTopViewWithFirstFrame
	templateSolver[resolverKey(true, false)] = withTopViewNoFirstFrame
	templateSolver[resolverKey(false, true)] = noTopViewWithFirstFrame
	templateSolver[resolverKey(false, false)] = noTopViewNoFirstFrame

	resolver, ok := templateSolver[resolverKey(ctx.hasTopView(), ctx.hasFirstFrame())]
	if !ok {
		return nil, 0, errors.Errorf("unable to get banner template: %+v: hasTopView: %+v: hasFirstFrame: %+v", ctx, ctx.hasTopView(), ctx.hasFirstFrame())
	}
	solt, visableCount, err := resolver(ctx)
	if err != nil {
		log.Error("Failed to resolve banner ordering: %+v", err)
		return nil, 0, err
	}
	if visableCount > len(solt) {
		return nil, 0, errors.Errorf("insufficient banner slot with visable count: %+v: %d", solt, visableCount)
	}
	return solt, visableCount, nil
}

func jsonify(in interface{}) string {
	bs, _ := json.Marshal(in)
	return string(bs)
}

// FeedBanners is
func (s *Service) FeedBanners(ctx context.Context, arg *pb.FeedBannersRequest) (*pb.FeedBannersReply, error) {
	if len(arg.Meta) <= 0 {
		return nil, ecode.Errorf(ecode.RequestErr, "empty banner meta: %+v", arg)
	}
	staticMeta := make([]*bannerMeta, 0, len(arg.Meta))
	for _, meta := range arg.Meta {
		b, ok := s.allBannerCache[meta.Id]
		if !ok {
			//log.Warn("Failed to get available banner with id: %d: %+v", meta.Id, meta)
			continue
		}
		staticMeta = append(staticMeta, &bannerMeta{
			BannerMeta: *meta,
			banner:     b,
		})
	}
	hashVersion := s.bannerHashCache[int(arg.ResId)]
	if (arg.Version != "" && (arg.Version == hashVersion)) ||
		(arg.Version == _categoryHash && (hashVersion == "")) {
		//log.Warn("Version matched banner request, will return empty banner response.")
		return &pb.FeedBannersReply{}, nil
	}

	cpmStatus := &cpmStatus{}
	cpmBanners := func() map[int]*model.Banner {
		groupedCpmBanners := s.cpmBanners(
			ctx, 0, arg.Mid, arg.SplashId, int(arg.Build),
			strconv.FormatInt(arg.ResId, 10), arg.MobiApp, arg.Device, arg.Buvid,
			arg.Network, arg.Ip, arg.OpenEvent, arg.AdExtra,
		)
		return groupedCpmBanners[int(arg.ResId)]
	}()
	cpmStatus.Banners = make([]*model.Banner, 0, len(cpmBanners))
	for _, cpmb := range cpmBanners {
		if !cpmb.IsAdReplace {
			continue
		}
		if cpmb.SplashID == 0 {
			cpmStatus.Banners = append(cpmStatus.Banners, cpmb)
			continue
		}
		if cpmb.SplashID == arg.SplashId {
			cpmStatus.TopView = cpmb
			continue
		}
	}
	sort.Slice(cpmStatus.Banners, func(i, j int) bool {
		return cpmStatus.Banners[i].Rank < cpmStatus.Banners[j].Rank
	})
	//nolint:gomnd
	if len(cpmStatus.Banners) > 2 {
		cpmStatus.Banners = cpmStatus.Banners[:2]
	}

	slot, visableCount, err := constructBannerSlot(staticMeta, cpmStatus)
	if err != nil {
		log.Error("Failed to construct banner slot: %+v", err)
		return nil, err
	}
	log.Info("Resolved banner slot on mid: %d: %+v: with visable count: %d", arg.Mid, jsonify(slot), visableCount)

	reply := &pb.FeedBannersReply{}
	topViewCount := 0
	if slot[0].Type == BannerTypeTopView {
		topViewCount = 1
	}
	for idx, b := range slot[:visableCount] {
		dup := *b.Banner
		dup.Index = idx + 1
		dup.ResourceID = int(arg.ResId)
		// 标记为一个广告库存
		func() {
			// 广告卡不用动
			if BannerCPMTypeSet.Has(b.Type) {
				return
			}

			// 非广告卡填充
			btime := strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
			dup.ServerType = 0
			dup.RequestId = btime

			cpmSlotAt := idx + 1
			if cpmSlotAt <= (topViewCount + 1) {
				cpmSlotAt -= topViewCount
			}
			cpmB, ok := cpmBanners[cpmSlotAt]
			if !ok {
				return
			}
			// 已经是一张广告卡了
			if cpmB.IsAdReplace {
				return
			}
			dup.IsAdLoc = true
			dup.IsAd = cpmB.IsAd
			dup.CmMark = cpmB.CmMark
			dup.SrcId = cpmB.SrcId
			dup.RequestId = cpmB.RequestId
			dup.ClientIp = cpmB.ClientIp
		}()
		meta := &pb.BannerMeta{
			Id:         int64(b.Banner.ID),
			Type:       b.Type,
			InlineType: b.InlineType,
			InlineId:   b.InlineID,
		}
		if b.Banner.InlineURL != b.InlineID {
			log.Error("Banner has different inline id, banner: %d, inline id: %s, %s, args: %+v", b.Banner.ID, b.Banner.InlineURL, b.InlineID, arg)
		}
		reply.Banner = append(reply.Banner, copyAsProtoBanner(&dup, meta))
	}
	reply.Version = hashVersion
	if reply.Version == "" {
		reply.Version = _categoryHash
	}
	log.Info("Finally fulfilled banners to mid: %d: %s", arg.Mid, jsonify(reply))
	return reply, nil
}

func copyAsProtoBanner(in *model.Banner, meta *pb.BannerMeta) *pb.Banner {
	return &pb.Banner{
		Id:                  int64(in.ID),
		ParentId:            int64(in.ParentID),
		Plat:                int32(in.Plat),
		Module:              in.Module,
		Position:            in.Position,
		Title:               in.Title,
		Image:               in.Image,
		Hash:                in.Hash,
		URI:                 in.URI,
		Goto:                in.Goto,
		Value:               in.Value,
		Param:               in.Param,
		Channel:             in.Channel,
		Build:               int32(in.Build),
		Condition:           in.Condition,
		Area:                in.Area,
		Rank:                int64(in.Rank),
		Rule:                in.Rule,
		Type:                int32(in.Type),
		Start:               int64(in.Start),
		End:                 int64(in.End),
		MTime:               int64(in.MTime),
		ResourceId:          int64(in.ResourceID),
		RequestId:           in.RequestId,
		CreativeId:          int64(in.CreativeId),
		SrcId:               int64(in.SrcId),
		IsAd:                in.IsAd,
		IsAdReplace:         in.IsAdReplace,
		IsAdLoc:             in.IsAdLoc,
		CmMark:              int64(in.CmMark),
		AdCb:                in.AdCb,
		ShowUrl:             in.ShowUrl,
		ClickUrl:            in.ClickUrl,
		ClientIp:            in.ClientIp,
		Index:               int64(in.Index),
		ServerType:          int64(in.ServerType),
		Extra:               in.Extra,
		CreativeType:        int64(in.CreativeType),
		SubTitle:            in.SubTitle,
		SplashId:            in.SplashID,
		BannerMeta:          *meta,
		InlineUseSame:       in.InlineUseSame,
		InlineBarrageSwitch: in.InlineBarrageSwitch,
	}
}

func (s *Service) VersionMap(c context.Context) map[string]interface{} {
	return map[string]interface{}{
		"version": s.c.Version,
	}
}
