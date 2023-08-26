package splash

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/model"
	"go-gateway/app/app-svr/app-resource/interface/model/manager"
	"go-gateway/app/app-svr/app-resource/interface/model/splash"

	"go-common/library/log/infoc.v2"

	"git.bilibili.co/go-tool/libbdevice/pkg/pd"
)

func (s *Service) BrandList(c context.Context, param *splash.SplashParam) (*splash.Brand, error) {
	var (
		res = &splash.Brand{PullInterval: s.c.BrandSplash.PullInterval}
		now = time.Now()
	)
	if s.brandCache == nil {
		return res, nil
	}
	bcImgMap := s.chooseImgMap(param.MobiApp, param.Build)
	config := s.chooseConfig(c, s.c.Feature, param.MobiApp, param.Build)
	imgMap := BuildImgMap(bcImgMap)
	res.List = splash.BrandListChange(c, s.c.Feature, bcImgMap, s.c.BrandSplash.LogoURL, param)
	res.Show, res.Forcibly, res.Rule, res.ForceShowTimes = splash.BrandListShowChange(c, s.c.Feature, config, imgMap, s.c.BrandSplash, param, now)
	res.ShowHash = calcShowHash(res.Show)
	res.Preload = make([]*splash.BrandPreload, 0, len(s.brandCache.PrepareDefaultConfigs))
	for _, v := range s.brandCache.PrepareDefaultConfigs {
		cfg := &splash.BrandPreload{
			PullInterval: s.c.BrandSplash.PullInterval,
			Start:        v.Stime,
			End:          v.Etime,
		}
		cfg.Show, cfg.Forcibly, cfg.Rule, cfg.ForceShowTimes = splash.BrandListShowChange(c, s.c.Feature, v, imgMap, s.c.BrandSplash, param, now)
		cfg.ShowHash = calcShowHash(cfg.Show)
		res.Preload = append(res.Preload, cfg)
	}
	res.HasNewSplashSet, res.NewSplashHash = s.hasNewSplashSet(c, param)
	if res.HasNewSplashSet {
		res.BadgeFrom = "new_splash"
	}
	if !res.HasNewSplashSet {
		bso := newBrandSetOption(now)
		res.HasNewSplashSet, res.NewSplashHash = bso.HasBadge, bso.BadgeHash
		if res.HasNewSplashSet {
			res.BadgeFrom = "new_brand_set"
		}
	}
	res.CollectionList = s.BrandCollectionList(c, param)
	return res, nil
}

func brandHttpsBuild(mobiApp string, build int) bool {
	return (mobiApp == "android" && build >= 6820000) || (mobiApp == "iphone" && build >= 68200000)
}

func (s *Service) BrandCollectionList(ctx context.Context, param *splash.SplashParam) []*splash.BrandList {
	list, ok := s.collectionSplashList(ctx, param)
	if !ok {
		return nil
	}
	out := make([]*splash.BrandList, 0, len(list))
	for _, id := range list {
		img, ok := s.collectionBrandCache.ImgMap[id]
		if !ok {
			log.Warn("Failed to match collection splash in list: %d", id)
			continue
		}
		out = append(out, &splash.BrandList{
			ID:      img.ID,
			Image:   splash.BuildImgURL(img, param),
			LogoURL: img.LogoConfig.ImgURL,
			Mode:    splash.SwitchMode(img.Mode),
		})
	}
	return out
}

func newBrandSetOption(now time.Time) *splash.BrandSetOption {
	const (
		maxSelected = 5
		// badgeStartAt = 1627228800 // 产品说从这个时间点开始 2021-07-26 00:00:00
		badgeEndAt = 1630339200 // 产品说截止到这个时间点 2021-08-31 00:00:00
	)

	out := &splash.BrandSetOption{
		Prompt:         "自选模式下可以选择多张开屏进行展示",
		MaxSelected:    maxSelected,
		MaxPrompt:      fmt.Sprintf("最多同时选择%d张", maxSelected),
		SelectedPrompt: fmt.Sprintf("已选择{{selected}}/%d张", maxSelected),
		SelectedText:   "选择",
		OverflowToast:  fmt.Sprintf("最多只能选%d张哦", maxSelected),
		EmptyToast:     "至少选择1张",
	}
	out.ExitDialog.Unsaved.Text = "你还未保存选择的开屏画面，确定要退出吗？"
	out.ExitDialog.Unsaved.YES = "退出"
	out.ExitDialog.Unsaved.NO = "取消"
	out.ExitDialog.Empty.Text = "你还未选择开屏画面，确定要退出吗？"
	out.ExitDialog.Empty.YES = "退出"
	out.ExitDialog.Empty.NO = "取消"
	out.BottomSaveButton.Text = "保存"
	out.BottomSaveButton.SuccessToast = "开屏设置成功"
	out.HasBadge = now.Unix() < badgeEndAt
	if out.HasBadge {
		hashV := uint64(crc32.ChecksumIEEE([]byte(strconv.FormatInt(badgeEndAt, 10))))
		out.BadgeHash = strconv.FormatUint(hashV, 10)
	}
	return out
}

func (s *Service) chooseImgMap(mobiApp string, build int) map[string]*manager.ImgInfo {
	if brandHttpsBuild(mobiApp, build) {
		return s.brandCache.ImgMapV2
	}
	return s.brandCache.ImgMap
}

func (s *Service) chooseConfig(c context.Context, config *conf.Feature, mobiApp string, build int) *manager.SplashConfig {
	if model.SplashUseBaseDefaultConfig(c, config, mobiApp, build) {
		return s.brandCache.BaseDefaultConfig
	}
	return s.brandCache.DefaultConfig
}

func BuildImgMap(in map[string]*manager.ImgInfo) map[int64]*manager.ImgInfo {
	out := make(map[int64]*manager.ImgInfo, len(in))
	for _, v := range in {
		out[v.ID] = v
	}
	return out
}

func (s *Service) hasNewSplashSet(ctx context.Context, param *splash.SplashParam) (bool, string) {
	bSet, err := s.BrandSet(ctx, param)
	if err != nil {
		log.Error("Failed to get brand set meta: %+v", err)
		return false, ""
	}
	newSplashID := []string{}
	for _, v := range bSet.Show {
		if v.NewSplash {
			newSplashID = append(newSplashID, strconv.FormatInt(v.ID, 10))
		}
	}
	if len(newSplashID) <= 0 {
		return false, ""
	}
	hashV := uint64(crc32.ChecksumIEEE([]byte(strings.Join(newSplashID, ""))))
	hash := strconv.FormatUint(hashV, 10)
	return true, hash
}

func calcShowHash(showSet []*splash.BrandShow) string {
	splashID := []string{}
	for _, v := range showSet {
		splashID = append(splashID, strconv.FormatInt(v.ID, 10))
	}
	hashV := uint64(crc32.ChecksumIEEE([]byte(strings.Join(splashID, ""))))
	hash := strconv.FormatUint(hashV, 10)
	return hash
}

func sortCategroies(in []*manager.SplashCategory) {
	sort.Slice(in, func(i, j int) bool {
		return in[i].Sort > in[j].Sort
	})
}

func fillAllCategories(in *splash.BrandSet) {
	allCategories := map[int64]*manager.SplashCategory{}
	for _, s := range in.Show {
		for _, c := range s.Categories {
			if _, ok := allCategories[c.ID]; !ok {
				dup := *c
				dup.Count = 0
				allCategories[dup.ID] = &dup
			}
			allCategories[c.ID].Count += 1
		}
	}
	for _, c := range allCategories {
		in.AllCategories = append(in.AllCategories, c)
	}
	// 如果算下来全部分类里有东西的话，就加上一个假的全部分类
	if len(in.AllCategories) > 0 {
		for _, s := range in.Show {
			s.Categories = append(s.Categories, &manager.SplashCategory{
				ID:   0,
				Name: "全部",
				Sort: math.MaxInt32, // 这里不能用 i64 的最大值
			})
		}
		in.AllCategories = append(in.AllCategories, &manager.SplashCategory{
			ID:    0,
			Name:  "全部",
			Sort:  math.MaxInt32,
			Count: int64(len(in.Show)),
		})
	}
	sortCategroies(in.AllCategories)
}

func (s *Service) BrandSet(c context.Context, param *splash.SplashParam) (*splash.BrandSet, error) {
	var (
		res = &splash.BrandSet{
			Config: []*splash.ConfigItem{
				{
					Title:     s.c.BrandSplash.DefaultTitle,
					Type:      s.c.BrandSplash.DefaultType,
					MainTitle: "默认模式",
					Subtitle:  "（随机展示）",
				},
				{
					Title:     s.c.BrandSplash.ProbabilityTitle,
					Type:      s.c.BrandSplash.ProbabilityType,
					MainTitle: "自选模式",
					Subtitle:  probabilitySubtitle(c),
				},
			},
			Desc:      s.c.BrandSplash.Desc,
			ShowTitle: s.c.BrandSplash.ShowTitle,
		}
		now = time.Now()
	)
	if s.brandCache == nil {
		return res, nil
	}
	imgMap := BuildImgMap(s.brandCache.ImgMap)
	res.Show = splash.BrandSetShowChange(c, s.c.Feature, s.brandCache.SelectConfig, imgMap, s.c.BrandSplash, param, now, s.brandCache.Categories)
	fillAllCategories(res)
	res.BrandSetOption = newBrandSetOption(now)
	res.CollectionShow = s.BrandCollectionSet(c, param, s.c.BrandSplash)
	if len(res.CollectionShow) > 0 {
		res.CollectionShowTitle = s.c.BrandSplash.CollectionShowTitle
	}
	return res, nil
}

func probabilitySubtitle(ctx context.Context) string {
	if pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().Or().IsPlatAndroidB().Or().IsPlatAndroidI().And().Build("<", int64(6530000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPhone().Or().IsPlatIPhoneB().Or().IsPlatIPhoneI().Or().IsPlatIPad().And().Build("<", int64(65300000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPadHD().And().Build("<", int64(33200000))
	}).MustFinish() {
		return ""
	}
	return "（可选择多张开屏画面进行展示）"
}

func (s *Service) BrandCollectionSet(ctx context.Context, param *splash.SplashParam, brandConf *conf.BrandSplash) []*splash.BrandShow {
	list, ok := s.collectionSplashList(ctx, param)
	if !ok {
		return nil
	}
	out := make([]*splash.BrandShow, 0, len(list))
	for _, id := range list {
		img, ok := s.collectionBrandCache.ImgMap[id]
		if !ok {
			log.Warn("Failed to match collection splash in set: %d", id)
			continue
		}
		out = append(out, &splash.BrandShow{
			ID:       img.ID,
			Duration: brandConf.Duration,
			Mode:     splash.SwitchMode(img.Mode),
			ShowLogo: img.LogoConfig.Show,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func (s *Service) collectionSplashList(ctx context.Context, param *splash.SplashParam) ([]int64, bool) {
	if param.Mid == 0 {
		return nil, false
	}
	if s.collectionBrandCache == nil {
		return nil, false
	}
	list, err := s.dao.UserCollectionSplashList(ctx, param.Mid)
	if err != nil {
		log.Error("Failed to UserCollectionSplashList: %d, %+v", param.Mid, err)
		return nil, false
	}
	return list, true
}

func (s *Service) loadCollectionSplash() {
	list, err := s.dao.CollectionSplash(context.Background())
	if err != nil {
		log.Error("Failed to CollectionSplash: %+v", err)
		return
	}
	imgMap := make(map[int64]*manager.ImgInfo, len(list))
	for _, v := range list {
		img := &manager.ImgInfo{
			ID:      v.Id,
			ImgName: v.ImgName,
			ImgURL:  v.ImgUrl,
			Mode:    v.Mode,
			LogoConfig: manager.LogoConfig{
				Show:   v.LogoHide == 0,
				Mode:   v.LogoMode,
				ImgURL: v.LogoImgUrl,
			},
			FullScreenImgURL: manager.FullScreenImgURL{
				Normal: v.ImgUrlNormal,
				Full:   v.ImgUrlFull,
				Pad:    v.ImgUrlPad,
			},
		}
		imgMap[img.ID] = img
	}
	s.collectionBrandCache = &manager.CollectionSplashList{ImgMap: imgMap}
	log.Info("loadCollectionSplash success")
}

func (s *Service) BrandSave(c context.Context, param *splash.SplashSaveParam, api, buvid string, mid int64, now time.Time) {
	ip := metadata.String(c, metadata.RemoteIP)
	oneID := ""
	if len(param.ID) == 1 {
		oneID = strconv.FormatInt(param.ID[0], 10)
	}
	idList, err := json.Marshal(param.ID)
	if err != nil {
		idList = []byte("[]")
	}
	csi, err := json.Marshal(param.CollectionSplashID)
	if err != nil {
		csi = []byte("[]")
	}
	event := infoc.NewLogStreamV("005102",
		log.String(ip),
		log.String(now.Format("2006-01-02 15:04:05")),
		log.String(api),
		log.String(buvid),
		log.String(param.MobiApp),
		log.String(param.Device),
		log.String(strconv.Itoa(param.Build)),
		log.String(param.Network),
		log.String(strconv.FormatInt(mid, 10)),
		log.String(oneID),
		log.String(string(idList)),
		log.String(string(csi)),
	)
	if err := s.inf2.Info(c, event); err != nil {
		log.Error("infoc2 splash params(%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s) err(%v)",
			ip, now.Format("2006-01-02 15:04:05"), api, buvid, param.MobiApp, param.Device, strconv.Itoa(param.Build), param.Network, strconv.FormatInt(mid, 10), idList, csi, err)
	}
}

func (s *Service) loadBrandCache() {
	tmp, err := s.mager.SplashList(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	tmp.ImgMapV2 = targetHttpsMap(tmp.ImgMap)
	s.brandCache = tmp
}

func targetHttpsMap(in map[string]*manager.ImgInfo) map[string]*manager.ImgInfo {
	out := make(map[string]*manager.ImgInfo, len(in))
	for k, v := range in {
		img := &manager.ImgInfo{}
		*img = *v
		img.ImgURL = convertHttpToHttps(img.ImgURL)
		img.FullScreenImgURL.Full = convertHttpToHttps(img.FullScreenImgURL.Full)
		img.FullScreenImgURL.Pad = convertHttpToHttps(img.FullScreenImgURL.Pad)
		img.FullScreenImgURL.Normal = convertHttpToHttps(img.FullScreenImgURL.Normal)
		img.LogoConfig.ImgURL = convertHttpToHttps(img.LogoConfig.ImgURL)
		out[k] = img
	}
	return out
}

func convertHttpToHttps(in string) string {
	return strings.Replace(in, "http://", "https://", 1)
}
