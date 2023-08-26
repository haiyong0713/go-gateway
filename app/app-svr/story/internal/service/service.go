package service

import (
	"context"
	"fmt"

	"go-common/library/conf/paladin.v2"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	infocV2 "go-common/library/log/infoc.v2"
	"go-common/library/stat/prom"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	pb "go-gateway/app/app-svr/story/api"
	"go-gateway/app/app-svr/story/internal/dao"
	"go-gateway/app/app-svr/story/internal/model"
	gateecode "go-gateway/ecode"

	"github.com/BurntSushi/toml"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
	"github.com/robfig/cron"
)

var Provider = wire.NewSet(New, wire.Bind(new(pb.StoryServer), new(*Service)))

// Service service.
type Service struct {
	ac       *paladin.Map
	dao      dao.Dao
	StorySvc *StoryService
}

// New new a service and return.
func New(d dao.Dao, ic infocV2.Infoc) (s *Service, cf func(), err error) {
	s = &Service{
		ac:       &paladin.TOML{},
		dao:      d,
		StorySvc: newOuterService(d, ic),
	}
	cf = s.Close
	err = paladin.Watch("application.toml", s.ac)
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
	s.StorySvc.cron.Stop()
	s.StorySvc.infocV2Log.Close()
}

type FeatureControl struct {
	DisableAll bool // 总开关
	Feature    map[string][]string
}

type StoryCustom struct {
	IconDrag                   string
	IconDragHash               string
	IconStop                   string
	IconStopHash               string
	IconZoom                   string
	IconZoomHash               string
	DisableStoryLiveReserveMid bool // 直播预约实验组白名单开关
	StoryLiveAttentionMidGroup map[string]int
	StoryLiveAttentionGroup    map[string]int
	DegradeSwitch              bool
	DegradeGroup               []int64
	JumpToViewIcon             string
	ReplyVerticalGroup         []int
	ReplyNoDanmuGroup          []int
	ReplyHighRaisedGroup       []int
	SpeedPlay                  int
	JumpToSeason               int
}

type StoryService struct {
	dao            dao.Dao
	cfg            *Config
	infoProm       *prom.Prom
	storyRcmdCache []*ai.SubItems
	hotAids        map[int64]struct{}
	infocV2Log     infocV2.Infoc
	// infoc
	logCh chan interface{}
	cron  *cron.Cron
}

type Config struct {
	CustomConfig   *StoryCustom
	FeatureControl *FeatureControl
}

func (c *Config) Set(text string) error {
	var tmp Config
	if _, err := toml.Decode(text, &tmp); err != nil {
		return err
	}
	log.Info("progress-service-config changed, old=%+v new=%+v", c, tmp)
	*c = tmp
	return nil
}

func newOuterService(d dao.Dao, ic infocV2.Infoc) *StoryService {
	s := &StoryService{
		dao:      d,
		infoProm: prom.BusinessInfoCount,
		cron:     cron.New(),
		hotAids:  map[int64]struct{}{},
		// infoc
		logCh: make(chan interface{}, 1024),
	}
	s.infocV2Log = ic
	if err := paladin.Get("application.toml").UnmarshalTOML(&s.cfg); err != nil {
		panic(err)
	}
	if err := paladin.Watch("application.toml", s.cfg); err != nil {
		panic(err)
	}
	s.loadCache()
	s.cron.Start()
	go s.infocproc()
	return s
}

func (s *StoryService) loadCache() {
	s.loadStoryRcmdCache()
	s.loadRcmdHotCache()
	checkErr(s.cron.AddFunc("@every 1m", s.loadStoryRcmdCache)) // 间隔1分钟
	checkErr(s.cron.AddFunc("@every 1m", s.loadRcmdHotCache))   // 间隔1分钟
}

func (s *StoryService) loadStoryRcmdCache() {
	aids, err := s.dao.StoryRcmdBackup(context.Background())
	if err != nil {
		log.Error("Failed to get story rcmd backup data: %+v", err)
		return
	}
	//nolint:gomnd
	if len(aids) < 4 {
		log.Warn("Unsufficient story rcmd backup data: %+v", aids)
		return
	}
	items := make([]*ai.SubItems, 0, len(aids))
	for _, aid := range aids {
		items = append(items, &ai.SubItems{
			ID:   aid,
			Goto: model.GotoVerticalAv,
		})
	}
	s.storyRcmdCache = items
}

func (s *StoryService) loadRcmdHotCache() {
	tmp, err := s.dao.RecommendHot(context.Background())
	if err != nil {
		log.Error("Failed to request RecommendHot: %+v", err)
		return
	}
	s.hotAids = tmp
}

func checkErr(err error) {
	if err != nil {
		panic(fmt.Sprintf("cron add func loadCache error(%+v)", err))
	}
}

func (s *StoryService) SLBRetry(err error) bool {
	return xecode.EqualError(gateecode.AppSLBRetry, err)
}

func (s *StoryService) infoc(i interface{}) {
	select {
	case s.logCh <- i:
	default:
		log.Warn("infocproc chan full")
	}
}
