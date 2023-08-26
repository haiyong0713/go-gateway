package static

import (
	"context"
	"encoding/json"
	"go-common/library/stat/prom"
	"go-gateway/app/app-svr/app-resource/interface/dao/dwtime"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-resource/interface/conf"
	eggdao "go-gateway/app/app-svr/app-resource/interface/dao/egg"
	"go-gateway/app/app-svr/app-resource/interface/model"
	resMdl "go-gateway/app/app-svr/app-resource/interface/model/resource"
	"go-gateway/app/app-svr/app-resource/interface/model/static"

	farm "go-farm"

	"github.com/robfig/cron"
)

const (
	_initVersion = "static_version"
)

var (
	_emptyStatics = []*static.Static{}
)

// Service static service.
type Service struct {
	c          *conf.Config
	dao        *eggdao.Dao
	tick       time.Duration
	cache      map[int8][]*static.Static
	cachePic   map[int8][]*static.Static
	staticPath string
	// cron
	cron     *cron.Cron
	dwDao    *dwtime.Dao
	dwCache  map[string]*resMdl.CdnDwTime
	infoProm *prom.Prom
}

// New new a static service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:          c,
		dao:        eggdao.New(c),
		tick:       time.Duration(c.Tick),
		cache:      map[int8][]*static.Static{},
		cachePic:   map[int8][]*static.Static{},
		staticPath: c.StaticJSONFile,
		// cron
		cron:     cron.New(),
		dwDao:    dwtime.New(c),
		infoProm: prom.BusinessInfoCount,
	}
	s.initCron()
	s.cron.Start()
	return
}

func (s *Service) initCron() {
	s.loadCache()
	if err := s.cron.AddFunc(s.c.Cron.LoadStaticCache, func() { s.loadCache() }); err != nil {
		panic(err)
	}
	s.loadDwTime()
	if err := s.cron.AddFunc(s.c.Cron.LoadStaticCache, func() { s.loadDwTime() }); err != nil {
		panic(err)
	}
}

func (s *Service) Download(c context.Context, req *resMdl.ResourceDownloadRequest) (res *resMdl.ResourceDownloadResponse, err error) {
	res = &resMdl.ResourceDownloadResponse{}
	resources := make([]*resMdl.Resource, 0)

	switch req.Type {
	case "egg":
		//获取彩蛋信息
		resources = append(resources, s.getEggs(req.Mid, req.Build, req.Platform)...)
	default:
		//获取彩蛋信息
		resources = append(resources, s.getEggs(req.Mid, req.Build, req.Platform)...)
		//获取错峰时间
		res.Dwtime = s.getDwtime()
	}

	res.Ver, err = getVerHash(resources)
	if err != nil {
		return nil, err
	}

	//版本号相同，不再下发
	if res.Ver == req.Ver {
		s.infoProm.Incr("Download-Ver相同")
		res.Resource = make([]*resMdl.Resource, 0)
		return res, nil
	}

	//设置资源以及打点
	res.Resource = resources
	for _, resource := range res.Resource {
		for _, item := range resource.List {
			s.infoProm.Incr("taskId-" + item.TaskId)
		}
	}

	return res, nil
}

func (s *Service) getEggs(mid int64, build int, platform int8) []*resMdl.Resource {
	//获取彩蛋信息
	var (
		tmps    = s.cache[platform]
		tmpPics = s.cachePic[platform]
	)
	eggRes := make([]*static.Static, 0)
	for _, tmp := range tmps {
		if model.InvalidBuild(build, tmp.Build, tmp.Condition) {
			continue
		}
		//针对指定用户下发
		if len(tmp.Mids) > 0 {
			if tmp.Whitelist[mid] == 1 {
				eggRes = append(eggRes, tmp)
			}
		} else {
			eggRes = append(eggRes, tmp)
		}
	}
	for _, tmpPic := range tmpPics {
		if model.InvalidBuild(build, tmpPic.Build, tmpPic.Condition) {
			continue
		}
		//针对指定用户下发
		if len(tmpPic.Mids) > 0 {
			if tmpPic.Whitelist[mid] == 1 {
				eggRes = append(eggRes, tmpPic)
			}
		} else {
			eggRes = append(eggRes, tmpPic)
		}
	}

	resources := make([]*resMdl.Resource, 0)
	//构造彩蛋
	resources = append(resources, buildEggResource(eggRes))
	return resources
}

func (s *Service) getDwtime() map[string]*resMdl.DwTime {
	res := make(map[string]*resMdl.DwTime)

	for k, v := range s.dwCache {
		res[k] = buildDwTime(v)
	}

	return res
}

func buildDwTime(cdn *resMdl.CdnDwTime) *resMdl.DwTime {
	res := &resMdl.DwTime{
		Type: 1,
		Peak: cdn.Peak,
		Low:  cdn.Low,
	}
	return res
}

// Static return statics
func (s *Service) Static(plat int8, build int, ver string, now time.Time) (res []*static.Static, version string, err error) {
	var (
		tmps    = s.cache[plat]
		tmpPics = s.cachePic[plat]
	)
	for _, tmp := range tmps {
		if model.InvalidBuild(build, tmp.Build, tmp.Condition) {
			continue
		}
		res = append(res, tmp)
	}
	for _, tmpPic := range tmpPics {
		if model.InvalidBuild(build, tmpPic.Build, tmpPic.Condition) {
			continue
		}
		res = append(res, tmpPic)
	}
	if len(res) == 0 {
		res = _emptyStatics
	}
	if version = s.hash(res); version == ver {
		err = ecode.NotModified
		res = nil
	}

	//设置资源以及打点
	for _, re := range res {
		s.infoProm.Incr("static-" + re.URL)
	}

	return
}

func (s *Service) hash(v []*static.Static) string {
	log.Info("cronLog start static hash")
	bs, err := json.Marshal(v)
	if err != nil {
		log.Error("json.Marshal error(%v)", err)
		return _initVersion
	}
	return strconv.FormatUint(farm.Hash64(bs), 10)
}

// loadCache update egg
func (s *Service) loadCache() {
	tmp, err := s.dao.Egg(context.TODO(), time.Now())
	if err != nil {
		log.Error("s.dao.Egg error(%v)", err)
		return
	}
	s.cache = tmp
	tmpPic, err := s.dao.EggPic(context.TODO(), time.Now())
	if err != nil {
		log.Error("s.dao.EggPic error(%v)", err)
		return
	}
	s.cachePic = tmpPic
	aa, _ := json.Marshal(s.cache)
	bb, _ := json.Marshal(s.cachePic)
	log.Info("loadCache s.cache(%s) s.cachePic(%s)", string(aa), string(bb))
}

func (s *Service) loadDwTime() {
	res := make(map[string]*resMdl.CdnDwTime)

	for k := range s.c.DWConfig.DomainList {
		//获取当天的错峰时间段
		today := time.Now()
		cph, err := s.dwDao.CdnPeakHours(context.TODO(), k, today.Format("20060102"))
		if err != nil {
			log.Info("loadDwTime s.dwDao.CdnPeakHours error(%+v)", err)
			continue
		}
		//获取第二天的错峰时间段
		tomorrow := today.AddDate(0, 0, 1)
		tcph, err := s.dwDao.CdnPeakHours(context.TODO(), k, tomorrow.Format("20060102"))
		if err != nil {
			log.Info("loadDwTime s.dwDao.CdnPeakHours error(%+v)", err)
		}

		dt := cph[k]
		if tcph != nil {
			dt.Peak = append(dt.Peak, tcph[k].Peak...)
			dt.Low = append(dt.Low, tcph[k].Low...)
		}
		res[k] = dt
	}

	s.dwCache = res
	msg, _ := json.Marshal(s.dwCache)
	log.Info("loadDwTime s.dwCache(%s)", string(msg))
}

func buildEggResource(eggRes []*static.Static) *resMdl.Resource {
	res := &resMdl.Resource{
		Type: "egg",
	}
	list := make([]*resMdl.ResourceItem, 0, len(eggRes))
	taskIdMap := make(map[string]*resMdl.ResourceItem, len(eggRes))
	for _, egg := range eggRes {
		item := &resMdl.ResourceItem{
			TaskId:         egg.Hash,
			FileName:       egg.Name,
			Type:           egg.Type,
			URL:            replaceHttp(egg.URL),
			Hash:           egg.Hash,
			Size:           egg.Size,
			FileEffectTime: egg.Start.Time().Unix(),
			FileExpireTime: egg.End.Time().Unix(),
			FileUploadTime: egg.Mtime.Time().Unix(),
		}
		list = append(list, item)
		if _, ok := taskIdMap[item.TaskId]; ok {
			log.Error("日志告警 taskId相同 already has item(%+v), but item2(%+v)", taskIdMap[item.TaskId], item)
		}
		taskIdMap[item.TaskId] = item
	}
	res.List = list
	return res
}

func replaceHttp(oldURL string) string {
	const (
		_httpPrefix  = "http"
		_httpsPrefix = "https"
		_domins      = "hdslb.com"
	)
	newURL, err := url.Parse(oldURL)
	if err != nil {
		log.Error("replaceHttp url.Parse() err(%+v)", err)
		return oldURL
	}
	if newURL.Scheme == _httpPrefix && strings.Contains(newURL.Host, _domins) {
		newURL.Scheme = _httpsPrefix
		return newURL.String()
	}
	return oldURL
}

func getVerHash(resources []*resMdl.Resource) (string, error) {
	rs, err := json.Marshal(resources)
	if err != nil {
		log.Error("getVerHash resources(%+v) json.Marshal error(%+v)", resources, err)
		return "", err
	}
	return strconv.FormatUint(farm.Hash64(rs), 10), nil
}
