package service

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"

	"go-common/library/conf/env"
	"go-common/library/log"
	"go-common/library/log/infoc.v2"

	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	fkmdl "go-gateway/app/app-svr/fawkes/service/model/business"
	playsvcgrpc "go-gateway/app/app-svr/playurl/service/api/v2"
	playurlgrpc "go-gateway/app/app-svr/playurl/service/api/v2"
	resgrpc "go-gateway/app/app-svr/resource/service/api/v1"
	resmdl "go-gateway/app/app-svr/resource/service/model"
	resrpc "go-gateway/app/app-svr/resource/service/rpc/client"
	sggrpc "go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/web-svr/player/interface/conf"
	"go-gateway/app/web-svr/player/interface/dao"
	"go-gateway/app/web-svr/player/interface/model"
	"go-gateway/pkg/idsafe/bvid"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	memgrpc "git.bilibili.co/bapis/bapis-go/account/service/member"
	ugcpaygrpc "git.bilibili.co/bapis/bapis-go/account/service/ugcpay"
	assgrpc "git.bilibili.co/bapis/bapis-go/assist/service"
	pugvgrpc "git.bilibili.co/bapis/bapis-go/cheese/service/auth"
	ansgrpc "git.bilibili.co/bapis/bapis-go/community/interface/answer"
	dmgrpc "git.bilibili.co/bapis/bapis-go/community/interface/dm"
	hisgrpc "git.bilibili.co/bapis/bapis-go/community/interface/history"
	appconfiggrpc "git.bilibili.co/bapis/bapis-go/community/service/appconfig"
	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	videogrpc "git.bilibili.co/bapis/bapis-go/videoup/open/service"

	"github.com/robfig/cron"
)

const (
	_resourceID         = 2319
	_paramPlatWebPlayer = 34
	_bgColor            = "#000000"
)

// Service is a service.
type Service struct {
	// config
	c *conf.Config
	// dao
	dao *dao.Dao
	// rpc
	res *resrpc.Service
	tag *dao.TagRPCService
	// memory cache
	caItems   []*model.Item
	params    string
	paramsMap map[string]string
	// template
	tWithU *template.Template
	tNoU   *template.Template
	// broadcast
	BrBegin time.Time
	BrEnd   time.Time
	// 拜年祭相关
	matOn    bool
	matTime  time.Time
	pastView *model.View
	matView  *model.View
	// grpc client
	ansGRPC        ansgrpc.AnswerClient
	accGRPC        accgrpc.AccountClient
	arcGRPC        arcgrpc.ArchiveClient
	ugcPayGRPC     ugcpaygrpc.UGCPayClient
	playurlV2GRPC  playurlgrpc.PlayURLClient
	steinsGateGRPC sggrpc.SteinsGateClient
	dmGRPC         dmgrpc.DMClient
	assistGRPC     assgrpc.AssistClient
	hisGRPC        hisgrpc.HistoryClient
	locGRPC        locgrpc.LocationClient
	pugvGRPC       pugvgrpc.AuthClient
	memberGRPC     memgrpc.MemberClient
	resGRPC        resgrpc.ResourceClient
	appConfigGRPC  appconfiggrpc.AppConfigClient
	videoUpGRPC    videogrpc.VideoUpOpenClient
	playsvcGRPC    playsvcgrpc.PlayURLClient
	cfcGRPC        cfcgrpc.FlowControlClient
	// stein gate guide cids
	steinGuideCids map[int64]struct{}
	// bnj view map
	bnjViewMap map[int64]*arcgrpc.ViewReply
	// limit free map
	limitFreeMap map[int64]*model.LimitFreeInfo
	//Cron
	cron *cron.Cron
	// load check
	loadResourceRunning  bool
	loadParamRunning     bool
	loadMatRunning       bool
	loadGuideCidRunning  bool
	loadBnjViewRunning   bool
	loadLimitFreeRunning bool
	// local cache
	fawkesVersionCache map[string]map[string]*fkmdl.Version
	playInfoc          infoc.Infoc
}

// New  new and return service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:    c,
		dao:  dao.New(c),
		res:  resrpc.New(c.ResourceRPC),
		tag:  dao.NewTagRPC(c.TagRPC),
		cron: cron.New(),
		// local cache
		fawkesVersionCache: make(map[string]map[string]*fkmdl.Version),
	}
	var err error
	if s.accGRPC, err = accgrpc.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	if s.arcGRPC, err = arcgrpc.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	if s.ugcPayGRPC, err = ugcpaygrpc.NewClient(c.UGCPayClient); err != nil {
		panic(err)
	}
	if s.playurlV2GRPC, err = playurlgrpc.NewClient(c.PlayURLClient); err != nil {
		panic(err)
	}
	if s.steinsGateGRPC, err = sggrpc.NewClient(c.SteinsClient); err != nil {
		panic(err)
	}
	if s.dmGRPC, err = dmgrpc.NewClient(c.DMClient); err != nil {
		panic(err)
	}
	if s.ansGRPC, err = ansgrpc.NewClient(c.AnswerClient); err != nil {
		panic(err)
	}
	if s.assistGRPC, err = assgrpc.NewClient(c.AssistClient); err != nil {
		panic(err)
	}
	if s.hisGRPC, err = hisgrpc.NewClient(c.HistoryClient); err != nil {
		panic(err)
	}
	if s.locGRPC, err = locgrpc.NewClient(c.LocationClient); err != nil {
		panic(err)
	}
	if s.pugvGRPC, err = pugvgrpc.NewClient(c.PugvClient); err != nil {
		panic(err)
	}
	if s.memberGRPC, err = memgrpc.NewClient(c.MemberClient); err != nil {
		panic(err)
	}
	if s.resGRPC, err = resgrpc.NewClient(c.ResourceClient); err != nil {
		panic(err)
	}
	if s.appConfigGRPC, err = appconfiggrpc.NewClient(c.AppConfigClient); err != nil {
		panic(err)
	}
	if s.videoUpGRPC, err = videogrpc.NewClient(c.VideoUpGRPC); err != nil {
		panic(err)
	}
	if s.playInfoc, err = infoc.New(c.Infoc2); err != nil {
		panic(err)
	}
	if s.playsvcGRPC, err = playsvcgrpc.NewClient(c.PlayURLGRPC); err != nil {
		panic(err)
	}
	if s.cfcGRPC, err = cfcgrpc.NewClient(c.CfcGRPC); err != nil {
		panic(err)
	}
	s.caItems = []*model.Item{}
	// 模板
	s.tWithU, _ = template.New("with_user").Parse(model.TpWithUinfo)
	s.tNoU, _ = template.New("no_user").Parse(model.TpWithNoUinfo)
	// broadcast
	s.BrBegin, _ = time.Parse("2006-01-02 15:04:05", c.Broadcast.Begin)
	s.BrEnd, _ = time.Parse("2006-01-02 15:04:05", c.Broadcast.End)
	s.policyproc()
	// 拜年祭
	s.matTime, _ = time.Parse(time.RFC3339, c.Matsuri.MatTime)
	s.matOn = true
	if err = s.initCron(); err != nil {
		panic(err)
	}
	return
}

// Ping check service health
func (s *Service) Ping(_ context.Context) (err error) {
	return
}

func (s *Service) loadMat() {
	if s.loadMatRunning {
		return
	}
	defer func() {
		s.loadMatRunning = false
	}()
	s.loadMatRunning = true
	ctx := context.Background()
	tmp, err := s.arcGRPC.View(ctx, &arcgrpc.ViewRequest{Aid: s.c.Matsuri.PastID})
	if err != nil || tmp == nil {
		log.Error("loadMat s.arcGRPC.View(%d) tmp(%v) error(%v) ", s.c.Matsuri.PastID, tmp, err)
		return
	}
	s.pastView = &model.View{Arc: tmp.Arc, Pages: tmp.Pages}
	tmp, err = s.arcGRPC.View(ctx, &arcgrpc.ViewRequest{Aid: s.c.Matsuri.MatID})
	if err != nil || tmp == nil {
		log.Error("loadMat s.arcGRPC.View(%d) tmp(%v) error(%v)", s.c.Matsuri.MatID, tmp, err)
		return
	}
	s.matView = &model.View{Arc: tmp.Arc, Pages: tmp.Pages}
}

func (s *Service) loadResource() {
	if s.loadResourceRunning {
		return
	}
	defer func() {
		s.loadResourceRunning = false
	}()
	s.loadResourceRunning = true
	res, err := s.res.Resource(context.Background(), &resmdl.ArgRes{ResID: _resourceID})
	if err != nil {
		log.Error("loadResource s.res.Resource(%d) error(%v)", _resourceID, err)
		return
	}
	if res == nil || len(res.Assignments) == 0 {
		log.Warn("loadResource s.res.Resource(%d) res(%v) is nil || res.Assignments is nil", _resourceID, res)
		return
	}
	var items []*model.Item
	for _, v := range res.Assignments {
		if v == nil {
			continue
		}
		item := &model.Item{
			Bgcolor:    _bgColor,
			ResourceID: strconv.Itoa(_resourceID),
			SrcID:      strconv.Itoa(v.ResID),
			ID:         strconv.Itoa(v.ID),
		}
		if catalog, ok := model.Catalog[v.PlayerCategory]; ok {
			item.Catalog = catalog
		}
		item.Content = string(rune(10)) + string(rune(13)) + fmt.Sprintf(_content, v.URL, v.Name) + string(rune(10)) + string(rune(13))
		items = append(items, item)
	}
	log.Info("loadResource success len(items):%d", len(items))
	s.caItems = items
}

func (s *Service) loadParam() {
	if s.loadParamRunning {
		return
	}
	defer func() {
		s.loadParamRunning = false
	}()
	s.loadParamRunning = true
	c := context.Background()
	reply, err := s.resGRPC.ParamList(c, &resgrpc.ParamReq{Plats: []int64{_paramPlatWebPlayer}})
	if err != nil {
		log.Error("loadParam s.resGRPC.ParamList() error(%v)", err)
		return
	}
	if len(reply.GetList()) == 0 {
		return
	}
	var items []string
	tmp := make(map[string]string)
	for _, pa := range reply.GetList() {
		nameBy := bytes.NewBuffer(nil)
		valueBy := bytes.NewBuffer(nil)
		if err = xml.EscapeText(nameBy, []byte(pa.Name)); err != nil {
			log.Error("loadParam xml.EscapeText(%s) error(%v)", pa.Name, err)
			continue
		} else {
			pa.Name = nameBy.String()
		}
		if err = xml.EscapeText(valueBy, []byte(pa.Value)); err != nil {
			log.Error("loadParam xml.EscapeText(%s) error(%v)", pa.Value, err)
			continue
		} else {
			pa.Value = valueBy.String()
		}
		item := "<" + pa.Name + ">" + pa.Value + "</" + pa.Name + ">"
		items = append(items, item)
		tmp[pa.Name] = pa.Value
	}
	if len(items) == 0 {
		return
	}
	log.Info("loadParam success len(items):%d", len(items))
	s.params = strings.Join(items, "\n")
	s.paramsMap = tmp
}

func (s *Service) policyproc() {
	s.c.Policy.StartTime, _ = time.Parse("2006-01-02 15:04:05", s.c.Policy.Start)
	s.c.Policy.EndTime, _ = time.Parse("2006-01-02 15:04:05", s.c.Policy.End)
	s.c.Policy.MtimeTime, _ = time.Parse("2006-01-02 15:04:05", s.c.Policy.Mtime)
}

func (s *Service) avToBv(aid int64) (bvID string) {
	var err error
	if bvID, err = bvid.AvToBv(aid); err != nil {
		log.Warn("avToBv(%d) error(%v)", aid, err)
	}
	return
}

func (s *Service) loadBnjView() {
	if s.loadBnjViewRunning {
		return
	}
	defer func() {
		s.loadBnjViewRunning = false
	}()
	s.loadBnjViewRunning = true
	aids := s.c.Bnj.ListAids
	aids = append(aids, s.c.Bnj.MainAid)
	aids = append(aids, s.c.Bnj.SpAid)
	views, err := s.arcGRPC.Views(context.Background(), &arcgrpc.ViewsRequest{Aids: aids})
	if err != nil || views == nil {
		log.Error("loadBnjView s.arcGRPC.Views(%v) error(%v)", aids, err)
		return
	}
	tmp := make(map[int64]*arcgrpc.ViewReply, len(aids))
	for _, aid := range aids {
		if view, ok := views.Views[aid]; ok && view != nil && view.Arc != nil && view.Arc.IsNormal() {
			tmp[aid] = view
		}
	}
	s.bnjViewMap = tmp
}

func (s *Service) initCron() error {
	s.loadResource()
	err := s.cron.AddFunc(s.c.Cron.Resource, s.loadResource)
	if err != nil {
		return err
	}
	s.loadParam()
	err = s.cron.AddFunc(s.c.Cron.Param, s.loadParam)
	if err != nil {
		return err
	}
	s.loadMat()
	err = s.cron.AddFunc(s.c.Cron.Mat, s.loadMat)
	if err != nil {
		return err
	}
	s.loadGuideCid()
	err = s.cron.AddFunc(s.c.Cron.GuideCid, s.loadGuideCid)
	if err != nil {
		return err
	}
	s.loadBnjView()
	err = s.cron.AddFunc(s.c.Cron.BnjView, s.loadBnjView)
	if err != nil {
		return err
	}
	s.loadLimitFreeList()
	err = s.cron.AddFunc(s.c.Cron.LimitFree, s.loadLimitFreeList)
	if err != nil {
		return err
	}
	if env.DeployEnv == "pre" || env.DeployEnv == "prod" {
		s.loadFawkesVersion()
		if err = s.cron.AddFunc(s.c.Cron.Fawkes, s.loadFawkesVersion); err != nil {
			return err
		}
	}
	s.cron.Start()
	return nil
}

func (s *Service) Close() (err error) {
	s.cron.Stop()
	return nil
}
