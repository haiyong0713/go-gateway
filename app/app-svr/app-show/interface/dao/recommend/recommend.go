package recommend

import (
	"context"
	"encoding/json"
	"fmt"

	"net/url"
	"strconv"
	"time"

	"go-common/component/metadata/restriction"
	"go-common/library/ecode"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	xtime "go-common/library/time"
	"go-common/library/xstr"

	api "go-gateway/app/app-svr/app-show/interface/api/popular"
	"go-gateway/app/app-svr/app-show/interface/conf"
	"go-gateway/app/app-svr/app-show/interface/model/recommend"

	"git.bilibili.co/go-tool/libbdevice/pkg/pd"

	"github.com/pkg/errors"
)

const (
	// _hotUrl           = "/y3kflg2k/ranking-m.json"
	_hotUrl    = "/data/rank/reco-tmzb.json"
	_regionUrl = "/8669rank/mobile_random/%s/1.json" // %s must be replaced to concrete tid
	// _regionHotUrl     = "/y3kflg2k/catalogy/%d-recommend-m.json"
	_regionListUrl = "/list"
	// _regionChildHotUrl = "/y3kflg2k/catalogy/catalogy-%d-3-m.json"
	_regionChildHotUrl = "/data/rank/recent_region-%d-3.json"
	_regionArcListUrl  = "/x/v2/archive/rank"
	_rankRegionUrl     = "/y3kflg2k/rank/%s-03-%d.json"
	_rankOriginalUrl   = "/y3kflg2k/rank/%s-03.json"
	_rankBangumiUrl    = "/y3kflg2k/rank/all-3-33.json"
	_feedDynamicUrl    = "/feed/tag/top"
	_rankAllAppUrl     = "/data/rank/recent_all-app.json"
	_rankOriginAppUrl  = "/data/rank/recent_origin-app.json"
	_rankRegionAppUrl  = "/data/rank/recent_region-%d-app.json"
	_rankBangumiAppUrl = "/data/rank/all_region-33-app.json"
	_hottabURL         = "/data/rank/reco-app-remen.json"
	_hotrcmd           = "/recommand"
)

// Dao is recommend dao.
type Dao struct {
	c                 *conf.Config
	client            *httpx.Client
	clientAsyn        *httpx.Client
	clientParam       *httpx.Client
	clientHotData     *httpx.Client
	hotUrl            string
	regionUrl         string
	regionChildHotUrl string
	regionListUrl     string
	regionArcListUrl  string
	rankRegionUrl     string
	rankOriginalUrl   string
	rankBangumilUrl   string
	feedDynamicUrl    string
	rankAllAppUrl     string
	rankOriginAppUrl  string
	rankRegionAppUrl  string
	rankBangumiAppUrl string
	hottabURL         string
	hotrcmd           string
}

// New recommend dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:             c,
		client:        httpx.NewClient(c.HTTPClient),
		clientAsyn:    httpx.NewClient(c.HTTPClientAsyn),
		clientParam:   httpx.NewClient(c.HTTPClient),
		clientHotData: httpx.NewClient(c.HTTPHotData),
		// hotUrl:       c.Host.Hetongzi + _hotUrl,
		hotUrl:    c.Host.HetongziRank + _hotUrl,
		regionUrl: c.Host.Hetongzi + _regionUrl,
		// regionHotUrl: c.Host.Hetongzi + _regionHotUrl,
		// regionChildHotUrl: c.Host.Hetongzi + _regionChildHotUrl,
		regionChildHotUrl: c.Host.HetongziRank + _regionChildHotUrl,
		regionListUrl:     c.Host.ApiCo + _regionListUrl,
		regionArcListUrl:  c.Host.ApiCoX + _regionArcListUrl,
		rankRegionUrl:     c.Host.Hetongzi + _rankRegionUrl,
		rankOriginalUrl:   c.Host.Hetongzi + _rankOriginalUrl,
		rankBangumilUrl:   c.Host.Hetongzi + _rankBangumiUrl,
		feedDynamicUrl:    c.Host.Data + _feedDynamicUrl,
		rankAllAppUrl:     c.Host.HetongziRank + _rankAllAppUrl,
		rankOriginAppUrl:  c.Host.HetongziRank + _rankOriginAppUrl,
		rankRegionAppUrl:  c.Host.HetongziRank + _rankRegionAppUrl,
		rankBangumiAppUrl: c.Host.HetongziRank + _rankBangumiAppUrl,
		hottabURL:         c.Host.Data + _hottabURL,
		hotrcmd:           c.Host.Data + _hotrcmd,
	}
	return
}

// Hots get recommends.
func (d *Dao) Hots(c context.Context) (arcids []int64, err error) {
	var res struct {
		Code int `json:"code"`
		List []struct {
			Aid int64 `json:"aid"`
		} `json:"list"`
	}
	if err = d.clientAsyn.Get(c, d.hotUrl, "", nil, &res); err != nil {
		log.Error("recommend hots url(%s) error(%v)", d.hotUrl, err)
		return
	}
	b, _ := json.Marshal(&res)
	log.Info("recommend hots url(%s) json(%s)", d.hotUrl, b)
	if res.Code != 0 {
		log.Error("recommend hots url(%s) error(%v)", d.hotUrl, res.Code)
		err = fmt.Errorf("recommend api response code(%v)", res)
		return
	}
	for _, arcs := range res.List {
		arcids = append(arcids, arcs.Aid)
	}
	return
}

// Region get region recommend.
func (d *Dao) Region(c context.Context, tid string) (arcids []int64, err error) {
	var res struct {
		Code int `json:"code"`
		Data []struct {
			Aid string `json:"aid"`
		} `json:"list"`
	}
	api := fmt.Sprintf(d.regionUrl, tid)
	if err = d.clientAsyn.Get(c, api, "", nil, &res); err != nil {
		log.Error("recommend region url(%s) error(%v)", api, err)
		return
	}
	if res.Code != 0 {
		log.Error("recommend region url(%s) error(%v)", api, res.Code)
		err = fmt.Errorf("recommend region api response code(%v)", res)
		return
	}
	for _, arcs := range res.Data {
		arcids = append(arcids, aidToInt(arcs.Aid))
	}
	return
}

// RegionHots get hots recommend
func (d *Dao) RegionHots(c context.Context, tid int) (arcids []int64, err error) {
	var res struct {
		Code int `json:"code"`
		List []struct {
			Aid int64 `json:"aid"`
		} `json:"list"`
	}
	api := fmt.Sprintf(d.rankRegionAppUrl, tid)
	if err = d.clientAsyn.Get(c, api, "", nil, &res); err != nil {
		log.Error("recommend region hots url(%s) error(%v)", api, err)
		return
	}
	if res.Code != 0 {
		log.Error("recommend region hots url(%s) error(%v)", api, res.Code)
		err = fmt.Errorf("recommend region hots api response code(%v)", res)
		return
	}
	for _, arcs := range res.List {
		arcids = append(arcids, arcs.Aid)
	}
	return
}

// RegionList
func (d *Dao) RegionList(c context.Context, rid, tid, audit, pn, ps int, order string) (arcids []int64, err error) {
	params := url.Values{}
	params.Set("order", order)
	params.Set("filtered", strconv.Itoa(audit))
	params.Set("page", strconv.Itoa(pn))
	params.Set("pagesize", strconv.Itoa(ps))
	params.Set("tid", strconv.Itoa(rid))
	if tid > 0 {
		params.Set("tag_id", strconv.Itoa(tid))
	}
	params.Set("apiver", "2")
	params.Set("ver", "2")
	var res struct {
		Code int `json:"code"`
		List []struct {
			Aid interface{} `json:"aid"`
		} `json:"list"`
	}
	if err = d.client.Get(c, d.regionListUrl, "", params, &res); err != nil {
		log.Error("recommend region news url(%s) error(%v)", d.regionListUrl+"?"+params.Encode(), err)
		return
	}
	if res.Code != 0 && res.Code != -1 {
		log.Error("recommend region news url(%s) error(%v)", d.regionListUrl+"?"+params.Encode(), res.Code)
		err = fmt.Errorf("recommend region news api response code(%v)", res)
		return
	}
	for _, arcs := range res.List {
		var aidInt int64
		switch aid := arcs.Aid.(type) {
		case string:
			aidInt = aidToInt(aid)
		case float64:
			aidInt = int64(aid)
		}
		arcids = append(arcids, aidInt)
	}
	return
}

// TwoRegionHots
func (d *Dao) RegionChildHots(c context.Context, rid int) (arcids []int64, err error) {
	var res struct {
		Code int `json:"code"`
		List []struct {
			Aid int64 `json:"aid"`
		} `json:"list"`
	}
	api := fmt.Sprintf(d.regionChildHotUrl, rid)
	if err = d.clientAsyn.Get(c, api, "", nil, &res); err != nil {
		log.Error("recommend region child hots url(%s) error(%v)", api, err)
		return
	}
	if res.Code != 0 {
		log.Error("recommend region child hots url(%s) error(%v)", api, res.Code)
		err = fmt.Errorf("recommend region child hots api response code(%v)", res)
		return
	}
	for _, arcs := range res.List {
		arcids = append(arcids, arcs.Aid)
	}
	return
}

func (d *Dao) RegionArcList(c context.Context, rid, pn, ps int, now time.Time) (arcids []int64, err error) {
	params := url.Values{}
	params.Set("rid", strconv.Itoa(rid))
	params.Set("pn", strconv.Itoa(pn))
	params.Set("ps", strconv.Itoa(ps))
	var res struct {
		Code int `json:"code"`
		Data struct {
			List []struct {
				Aid int64 `json:"aid"`
			} `json:"archives"`
		} `json:"data"`
	}
	if err = d.client.Get(c, d.regionArcListUrl, "", params, &res); err != nil {
		log.Error("recommend regionArc news url(%s) error(%v)", d.regionArcListUrl+"?"+params.Encode(), err)
		return
	}
	if res.Code != 0 && res.Code != -1 {
		log.Error("recommend regionArc news url(%s) error(%v)", d.regionArcListUrl+"?"+params.Encode(), res.Code)
		err = fmt.Errorf("recommend regionArc news api response code(%v)", res)
		return
	}
	for _, arcs := range res.Data.List {
		arcids = append(arcids, arcs.Aid)
	}
	return
}

// RegionRank
func (d *Dao) RankRegion(c context.Context, rid int, order string) (data []*recommend.Arc, err error) {
	var res struct {
		Data struct {
			Code int              `json:"code"`
			List []*recommend.Arc `json:"list"`
		} `json:"rank"`
	}
	api := fmt.Sprintf(d.rankRegionUrl, order, rid)
	if err = d.clientAsyn.Get(c, api, "", nil, &res); err != nil {
		log.Error("recommend region rank hots url(%s) error(%v)", api, err)
		return
	}
	if res.Data.Code != 0 {
		log.Error("recommend region rank hots url(%s) error(%v)", api, res.Data.Code)
		err = fmt.Errorf("recommend region rank hots api response code(%v)", res)
		return
	}
	data = res.Data.List
	return
}

// RankAll
func (d *Dao) RankAll(c context.Context, order string) (data []*recommend.Arc, err error) {
	var res struct {
		Data struct {
			Code int              `json:"code"`
			List []*recommend.Arc `json:"list"`
		} `json:"rank"`
	}
	api := fmt.Sprintf(d.rankOriginalUrl, order)
	if err = d.clientAsyn.Get(c, api, "", nil, &res); err != nil {
		log.Error("recommend region rank hots url(%s) error(%v)", api, err)
		return
	}
	if res.Data.Code != 0 {
		log.Error("recommend region rank hots url(%s) error(%v)", api, res.Data.Code)
		err = fmt.Errorf("recommend region rank hots api response code(%v)", res)
		return
	}
	data = res.Data.List
	return
}

// RankAll
func (d *Dao) RankBangumi(c context.Context) (data []*recommend.Arc, err error) {
	var res struct {
		Data struct {
			Code int              `json:"code"`
			List []*recommend.Arc `json:"list"`
		} `json:"rank"`
	}
	if err = d.clientAsyn.Get(c, d.rankBangumilUrl, "", nil, &res); err != nil {
		log.Error("recommend region rank hots url(%s) error(%v)", d.rankBangumilUrl, err)
		return
	}
	if res.Data.Code != 0 {
		log.Error("recommend region rank hots url(%s) error(%v)", d.rankBangumilUrl, res.Data.Code)
		err = fmt.Errorf("recommend region rank hots api response code(%v)", res)
		return
	}
	data = res.Data.List
	return
}

// FeedDynamic
func (d *Dao) FeedDynamic(c context.Context, pull bool, rid, tid int, ctime, mid int64, now time.Time) (hotAids, newAids []int64, ctop, cbottom xtime.Time, err error) {
	var pn string
	if pull {
		pn = "1"
	} else {
		pn = "2"
	}
	params := url.Values{}
	params.Set("src", "2")
	params.Set("pn", pn)
	params.Set("mid", strconv.FormatInt(mid, 10))
	if ctime != 0 {
		params.Set("ctime", strconv.FormatInt(ctime, 10))
	}
	if rid != 0 {
		params.Set("rid", strconv.Itoa(rid))
	}
	if tid != 0 {
		params.Set("tag", strconv.Itoa(tid))
	}
	var res struct {
		Code    int        `json:"code"`
		Data    []int64    `json:"data"`
		Hot     []int64    `json:"hot"`
		CTop    xtime.Time `json:"ctop"`
		CBottom xtime.Time `json:"cbottom"`
	}
	if err = d.client.Get(c, d.feedDynamicUrl, "", params, &res); err != nil {
		log.Error("region feed dynamic d.client.Get(%s) error(%v)", d.feedDynamicUrl+"?"+params.Encode(), err)
		return
	}
	b, _ := json.Marshal(&res)
	log.Info("region feed dynamic url(%s) response(%s)", d.feedDynamicUrl+"?"+params.Encode(), b)
	if res.Code != 0 {
		log.Error("region feed dynamic d.client.Get(%s) error(%v)", d.regionArcListUrl+"?"+params.Encode(), res.Code)
		err = fmt.Errorf("region feed dynamicapi response code(%v)", res)
		return
	}
	hotAids = res.Hot
	newAids = res.Data
	ctop = res.CTop
	cbottom = res.CBottom
	return
}

func (d *Dao) RankAppRegion(c context.Context, rid int) (aids []int64, others, scores map[int64]int64, err error) {
	var res struct {
		Code int `json:"code"`
		List []struct {
			Aid    int64 `json:"aid"`
			Score  int64 `json:"score"`
			Others []struct {
				Aid   int64 `json:"aid"`
				Score int64 `json:"score"`
			} `json:"others"`
		} `json:"list"`
	}
	api := fmt.Sprintf(d.rankRegionAppUrl, rid)
	if err = d.client.Get(c, api, "", nil, &res); err != nil {
		log.Error("recommend region rank hots url(%s) error(%v)", api, err)
		return
	}
	if res.Code != 0 && res.Code != -1 {
		log.Error("recommend region rank hots url(%s) error(%v)", api, res.Code)
		err = fmt.Errorf("recommend region rank hots api response code(%v)", res)
		return
	}
	scores = map[int64]int64{}
	others = map[int64]int64{}
	for _, arcs := range res.List {
		aids = append(aids, arcs.Aid)
		scores[arcs.Aid] = arcs.Score
		for _, o := range arcs.Others {
			aids = append(aids, o.Aid)
			scores[o.Aid] = o.Score
			others[o.Aid] = arcs.Aid
		}
	}
	return
}

func (d *Dao) RankAppOrigin(c context.Context) (aids []int64, others, scores map[int64]int64, err error) {
	var res struct {
		Code int `json:"code"`
		List []struct {
			Aid    int64 `json:"aid"`
			Score  int64 `json:"score"`
			Others []struct {
				Aid   int64 `json:"aid"`
				Score int64 `json:"score"`
			} `json:"others"`
		} `json:"list"`
	}
	if err = d.client.Get(c, d.rankOriginAppUrl, "", nil, &res); err != nil {
		log.Error("recommend Origin rank hots url(%s) error(%v)", d.rankOriginAppUrl, err)
		return
	}
	if res.Code != 0 && res.Code != -1 {
		log.Error("recommend Origin rank hots url(%s) error(%v)", d.rankOriginAppUrl, res.Code)
		err = fmt.Errorf("recommend Origin rank hots api response code(%v)", res)
		return
	}
	scores = map[int64]int64{}
	others = map[int64]int64{}
	for _, arcs := range res.List {
		aids = append(aids, arcs.Aid)
		scores[arcs.Aid] = arcs.Score
		for _, o := range arcs.Others {
			aids = append(aids, o.Aid)
			scores[o.Aid] = o.Score
			others[o.Aid] = arcs.Aid
		}
	}
	return
}

func (d *Dao) RankAppAll(c context.Context) (aids []int64, others, scores map[int64]int64, err error) {
	var res struct {
		Code int `json:"code"`
		List []struct {
			Aid    int64 `json:"aid"`
			Score  int64 `json:"score"`
			Others []struct {
				Aid   int64 `json:"aid"`
				Score int64 `json:"score"`
			} `json:"others"`
		} `json:"list"`
	}
	if err = d.client.Get(c, d.rankAllAppUrl, "", nil, &res); err != nil {
		log.Error("recommend All rank hots url(%s) error(%v)", d.rankAllAppUrl, err)
		return
	}
	if res.Code != 0 && res.Code != -1 {
		log.Error("recommend All rank hots url(%s) error(%v)", d.rankAllAppUrl, res.Code)
		err = fmt.Errorf("recommend All rank hots api response code(%v)", res)
		return
	}
	scores = map[int64]int64{}
	others = map[int64]int64{}
	for _, arcs := range res.List {
		aids = append(aids, arcs.Aid)
		scores[arcs.Aid] = arcs.Score
		for _, o := range arcs.Others {
			aids = append(aids, o.Aid)
			scores[o.Aid] = o.Score
			others[o.Aid] = arcs.Aid
		}
	}
	return
}

func (d *Dao) RankAppBangumi(c context.Context) (aids []int64, others, scores map[int64]int64, err error) {
	var res struct {
		Code int `json:"code"`
		List []struct {
			Aid    int64 `json:"aid"`
			Score  int64 `json:"score"`
			Others []struct {
				Aid   int64 `json:"aid"`
				Score int64 `json:"score"`
			} `json:"others"`
		} `json:"list"`
	}
	if err = d.client.Get(c, d.rankBangumiAppUrl, "", nil, &res); err != nil {
		log.Error("recommend bangumi rank hots url(%s) error(%v)", d.rankBangumiAppUrl, err)
		return
	}
	if res.Code != 0 && res.Code != -1 {
		log.Error("recommend bangumi rank hots url(%s) error(%v)", d.rankBangumiAppUrl, res.Code)
		err = fmt.Errorf("recommend bangumi rank hots api response code(%v)", res)
		return
	}
	scores = map[int64]int64{}
	others = map[int64]int64{}
	for _, arcs := range res.List {
		aids = append(aids, arcs.Aid)
		scores[arcs.Aid] = arcs.Score
		for _, o := range arcs.Others {
			aids = append(aids, o.Aid)
			scores[o.Aid] = o.Score
			others[o.Aid] = arcs.Aid
		}
	}
	return
}

func aidToInt(aidstr string) (aid int64) {
	aid, _ = strconv.ParseInt(aidstr, 10, 64)
	return
}

// HotTab hot tab
func (d *Dao) HotTab(c context.Context) (list []*recommend.List, err error) {
	var res struct {
		Code int               `json:"code"`
		List []*recommend.List `json:"list"`
	}
	if err = d.client.Get(c, d.hottabURL, "", nil, &res); err != nil {
		log.Error("hottab hots url(%s) error(%v)", d.hottabURL, err)
		return
	}
	b, _ := json.Marshal(&res)
	log.Info("hottab list url(%s) response(%s)", d.hottabURL, b)
	if res.Code != 0 {
		err = ecode.Int(res.Code)
		log.Error("hottab hots url(%s) code(%d)", d.hottabURL, res.Code)
		return
	}
	list = res.List
	return
}

// HotRcmd get hot rcmd from new ai api.
func (d *Dao) HotRcmd(c context.Context, mid int64, plat int8, build, offset int, mobiApp, device, buvid string) (list []*recommend.CardList, err error) {
	var (
		ip     = metadata.String(c, metadata.RemoteIP)
		params = url.Values{}
	)
	params.Set("cmd", "hot_tab")
	params.Set("from", "10")
	timeout := time.Duration(d.c.HTTPData.Timeout) / time.Millisecond
	params.Set("timeout", strconv.FormatInt(int64(timeout), 10))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("buvid", buvid)
	params.Set("build", strconv.Itoa(build))
	params.Set("plat", strconv.Itoa(int(plat)))
	params.Set("request_cnt", "10")
	params.Set("page_no", strconv.Itoa(offset))
	var res struct {
		Code int                   `json:"code"`
		Data []*recommend.CardList `json:"data"`
	}
	if err = d.client.Get(c, d.hotrcmd, ip, params, &res); err != nil {
		log.Error("%v", err)
		return
	}
	if code := ecode.Int(res.Code); !code.Equal(ecode.OK) {
		err = errors.Wrap(ecode.Int(res.Code), d.hotrcmd+"?"+params.Encode())
		log.Error("%v", err)
		return
	}
	list = res.Data
	return
}

// Recommend list
func (d *Dao) Recommend(c context.Context) (rs map[int64]struct{}, err error) {
	params := url.Values{}
	params.Set("cmd", "hot")
	params.Set("from", "10")
	timeout := time.Duration(d.c.Custom.RecommendTimeout) / time.Millisecond
	params.Set("timeout", strconv.FormatInt(int64(timeout), 10))
	params.Set("ignore_custom", "1")
	var res struct {
		Code int `json:"code"`
		Data []struct {
			Id int64 `json:"id"`
		} `json:"data"`
	}
	if err = d.client.Get(c, d.hotrcmd, "", params, &res); err != nil {
		return
	}
	if res.Code != 0 {
		err = errors.Wrapf(ecode.Int(res.Code), "recommend url(%s) code(%d)", d.hotrcmd, res.Code)
		return
	}
	rs = map[int64]struct{}{}
	for _, l := range res.Data {
		if l.Id > 0 {
			rs[l.Id] = struct{}{}
		}
	}
	return
}

func extractDisableRcmd(ctx context.Context) int64 {
	v, ok := restriction.FromContext(ctx)
	if !ok {
		return 0
	}
	if v.DisableRcmd {
		return 1
	}
	return 0
}

func (d *Dao) hotAdResource(ctx context.Context) string {
	var adResource string
	if pd.WithContext(ctx).Where(func(pdContext *pd.PDContext) {
		pdContext.IsPlatAndroid().And().Build(">=", 6660000)
	}).FinishOr(false) {
		adResource = d.c.Custom.PopularAdResourceAndroid
	}
	if pd.WithContext(ctx).Where(func(pdContext *pd.PDContext) {
		pdContext.IsPlatIPhone().And().Build(">=", 66600000)
	}).FinishOr(false) {
		adResource = d.c.Custom.PopularAdResourceIOS
	}
	return adResource
}

// HotAiRcmd .
func (d *Dao) HotAiRcmd(c context.Context, mid int64, sourceID int32, plat int8, build int, buvid, mobiApp string, pageNo, count int, entranceID int64, hotWordID int, locationIDs []int64, ad *api.PopularAd, zoneID int64) (data []*recommend.HotItem, userfeature string, bizData *recommend.BizData, resCode int, err error) {
	adResource := d.hotAdResource(c)
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("cmd", "hot")
	params.Set("from", "10")
	timeout := time.Duration(d.c.HTTPHotData.Timeout) / time.Millisecond
	params.Set("timeout", strconv.FormatInt(int64(timeout), 10))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("buvid", buvid)
	params.Set("build", strconv.Itoa(build))
	params.Set("plat", strconv.Itoa(int(plat)))
	params.Set("request_cnt", strconv.Itoa(count))
	params.Set("page_no", strconv.Itoa(pageNo))
	params.Set("source_id", strconv.Itoa(int(sourceID)))
	params.Set("entrance_id", strconv.FormatInt(entranceID, 10))
	params.Set("hotword_id", strconv.Itoa(hotWordID))
	params.Set("location_ids", xstr.JoinInts(locationIDs))
	params.Set("disable_rcmd", strconv.FormatInt(extractDisableRcmd(c), 10))
	params.Set("zone_id", strconv.FormatInt(zoneID, 10))
	params.Set("ip", metadata.String(c, metadata.RemoteIP))
	if ad != nil {
		params.Set("ad_extra", ad.AdExtra)
	}
	params.Set("ad_resource", adResource)
	params.Set("mobi_app", mobiApp)
	var res struct {
		Code        int                  `json:"code"`
		Data        []*recommend.HotItem `json:"data"`
		UserFeature string               `json:"user_feature"`
		BizData     *recommend.BizData   `json:"biz_data"`
	}
	log.Info("热门 ai host(%s) params(%+v)", d.hotrcmd, params.Encode())
	if err := d.clientHotData.Get(c, d.hotrcmd, ip, params, &res); err != nil {
		log.Error("%v", err)
		return nil, "", nil, ecode.ServerErr.Code(), err
	}
	if code := ecode.Int(res.Code); !code.Equal(ecode.OK) {
		if res.Code == -3 { // code -3 热门数据已经拉到底了
			return []*recommend.HotItem{}, res.UserFeature, nil, res.Code, nil
		}
		err := errors.Wrap(ecode.Int(res.Code), d.hotrcmd+"?"+params.Encode())
		log.Error("%v", err)
		return nil, res.UserFeature, nil, res.Code, err
	}
	return res.Data, res.UserFeature, res.BizData, res.Code, nil
}
