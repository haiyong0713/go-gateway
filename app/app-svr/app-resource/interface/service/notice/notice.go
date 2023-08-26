package notice

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	locdao "go-gateway/app/app-svr/app-resource/interface/dao/location"
	ntcdao "go-gateway/app/app-svr/app-resource/interface/dao/notice"
	"go-gateway/app/app-svr/app-resource/interface/model"
	"go-gateway/app/app-svr/app-resource/interface/model/location"
	"go-gateway/app/app-svr/app-resource/interface/model/notice"

	"go-farm"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	bGroup "git.bilibili.co/bapis/bapis-go/platform/service/bgroup/v2"

	"github.com/robfig/cron"
)

const (
	_initNoticeKey = "notice_key_%d_%d"
	_initNoticeVer = "notice_version"
)

var (
	_emptyNotice = &notice.Notice{}
)

// Service notice service.
type Service struct {
	c   *conf.Config
	dao *ntcdao.Dao
	loc *locdao.Dao
	// tick
	tick time.Duration
	// cache
	cache map[string][]*notice.Notice
	// cron
	cron *cron.Cron
	//版本包推送
	packagePushMap map[string]*notice.PushDetail
	bGroupClient   bGroup.BGroupServiceClient
}

// New new a notice service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: ntcdao.New(c),
		loc: locdao.New(c),
		// tick
		tick: time.Duration(c.Tick),
		// cache
		cache: map[string][]*notice.Notice{},
		// cron
		cron: cron.New(),
	}
	s.initCron()
	s.cron.Start()
	var err error
	if s.bGroupClient, err = bGroup.NewClient(s.c.BGroupClient); err != nil {
		panic(err)
	}
	return
}

func (s *Service) initCron() {
	s.load()
	if err := s.cron.AddFunc(s.c.Cron.LoadNotice, s.load); err != nil {
		panic(err)
	}
	if err := s.cron.AddFunc("@every 10s", s.loadPackagePush); err != nil {
		panic(err)
	}
}

func (s *Service) loadPackagePush() {
	temp, err := s.dao.PackagePushList(context.Background())
	if err != nil {
		log.Error("s.dao.PackagePushList error(%+v)", err)
		return
	}
	s.packagePushMap = temp
	log.Info("load package push success")
}

// Notice return Notice to json
func (s *Service) Notice(c context.Context, plat int8, build, typeInt int, ver string) (res *notice.Notice, version string, err error) {
	var (
		ip    = metadata.String(c, metadata.RemoteIP)
		pids  []string
		auths map[int64]*locgrpc.Auth
	)
	for _, ntc := range s.cache[fmt.Sprintf(_initNoticeKey, plat, typeInt)] {
		if model.InvalidBuild(build, ntc.Build, ntc.Condition) {
			continue
		}
		if ntc.Area != "" {
			pids = append(pids, ntc.Area)
		}
	}
	if len(pids) > 0 {
		auths, _ = s.loc.AuthPIDs(c, strings.Join(pids, ","), ip)
	}
	for _, ntc := range s.cache[fmt.Sprintf(_initNoticeKey, plat, typeInt)] {
		if model.InvalidBuild(build, ntc.Build, ntc.Condition) {
			continue
		}
		area, _ := strconv.ParseInt(ntc.Area, 10, 64)
		if auth, ok := auths[area]; ok && auth.Play == location.Forbidden {
			log.Warn("s.invalid area(%v) ip(%v) error(%v)", ntc.Area, ip, err)
			continue
		}
		res = ntc
		break
	}
	if res == nil {
		res = _emptyNotice
	}
	if version = s.hash(res); ver == version {
		err = ecode.NotModified
		res = nil
	}
	return
}

func (s *Service) GetPackagePushMsg(ctx context.Context, buvid, model string) (*notice.PushDetail, error) {
	if _, ok := s.c.PackagePushModel[model]; !ok {
		return nil, nil
	}
	groups := s.fetchCrowedGroup()
	if len(groups) == 0 {
		return nil, nil
	}
	reply, err := s.bGroupClient.MemberIn(ctx, &bGroup.MemberInReq{
		Member:    buvid,
		Groups:    groups,
		Dimension: bGroup.Buvid,
	})
	if err != nil {
		log.Error("s.bGroupClient.MemberIn error(%+v), buvid(%s)", err, buvid)
		return nil, err
	}
	for _, v := range reply.Results {
		if v == nil || !v.In {
			continue
		}
		if detail, ok := s.packagePushMap[fmt.Sprintf("%s:%s", v.Business, v.Name)]; ok {
			return detail, nil
		}
		log.Warn("GetPackagePushMsg business(%s) or name(%s) not match", v.Business, v.Name)
	}
	return nil, nil
}

func (s *Service) fetchCrowedGroup() []*bGroup.MemberInReq_MemberInReqSingle {
	var crowedGroups []*bGroup.MemberInReq_MemberInReqSingle
	for _, v := range s.packagePushMap {
		if v == nil {
			continue
		}
		crowedGroups = append(crowedGroups, &bGroup.MemberInReq_MemberInReqSingle{
			Business: v.CrowedBusiness,
			Name:     v.CrowedName,
		})
	}
	return crowedGroups
}

// load
func (s *Service) load() {
	log.Info("cronLog start notice load")
	now := time.Now()
	// get notice
	ntcs, err := s.dao.All(context.TODO(), now)
	if err != nil {
		log.Error("s.dao.GetAll() error(%v)", err)
		return
	}
	// copy cache
	tmp := map[string][]*notice.Notice{}
	for _, v := range ntcs {
		key := fmt.Sprintf(_initNoticeKey, v.Plat, v.Type)
		tmp[key] = append(tmp[key], v)
	}
	s.cache = tmp
	log.Info("notice cacheproc success")
}

func (s *Service) hash(v *notice.Notice) string {
	bs, err := json.Marshal(v)
	if err != nil {
		log.Error("json.Marshal error(%v)", err)
		return _initNoticeVer
	}
	return strconv.FormatUint(farm.Hash64(bs), 10)
}

// Close dao
func (s *Service) Close() {
	s.dao.Close()
}
