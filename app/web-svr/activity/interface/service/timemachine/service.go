package timemachine

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/log"
	arcapi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/like"
	"go-gateway/app/web-svr/activity/interface/dao/timemachine"
	timemdl "go-gateway/app/web-svr/activity/interface/model/timemachine"

	"github.com/robfig/cron"
)

type Service struct {
	c             *conf.Config
	dao           *timemachine.Dao
	likeDao       *like.Dao
	typeNames     map[int32]*arcapi.Tp
	tagDescs      map[int64]*timemdl.Tag
	regionDescs   map[int32]*timemdl.RegionDesc
	dftRegionDesc *timemdl.RegionDesc
	events        map[int64]*timemdl.Event
	tmMidMap      map[int64]struct{}
	cron          *cron.Cron
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:       c,
		dao:     timemachine.New(c),
		likeDao: like.New(c),
		cron:    cron.New(),
	}
	s.loadTypes()
	s.loadAdminMids()
	s.loadWebDatas()
	s.createCron()
	s.LoadUserReportBaseData()
	go func() {
		for range time.Tick(time.Minute) {
			s.LoadUserReportBaseData()
		}
	}()
	return s
}

func (s *Service) createCron() {
	var err error
	if err = s.cron.AddFunc("@every 1h", s.loadTypes); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc("@every 1h", s.loadAdminMids); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc("@every 59m", s.loadWebDatas); err != nil {
		panic(err)
	}
	s.cron.Start()
}

// Close .
func (s *Service) Close() {
	s.cron.Stop()
	s.dao.Close()
}

func (s *Service) loadAdminMids() {
	tmp := make(map[int64]struct{}, len(s.c.Rule.TmMids))
	for _, mid := range s.c.Rule.TmMids {
		tmp[mid] = struct{}{}
	}
	s.tmMidMap = tmp
	log.Info("loadAdminMids() success")
}

func (s *Service) loadTypes() {
	types, err := client.ArchiveClient.Types(context.Background(), &arcapi.NoArgRequest{})
	if err != nil || types == nil {
		log.Error("loadTypes s.arcClient.Types error(%v)", err)
		return
	}
	s.typeNames = types.Types
	log.Info("loadTypes() success")
}

func (s *Service) loadWebDatas() {
	var c = context.Background()
	s.dftRegionDesc = new(timemdl.RegionDesc)
	func() {
		if s.c.Timemachine.EventSid == 0 {
			return
		}
		data, err := s.likeDao.SourceItem(c, s.c.Timemachine.EventSid)
		if err != nil {
			log.Error("loadWebDatas s.likeDao.SourceItem(%d) error(%v)", s.c.Timemachine.EventSid, err)
			return
		}
		var event = new(struct {
			List []*struct {
				Data *struct {
					EventID  int64  `json:"event_id"`
					Desc     string `json:"desc"`
					LongDesc string `json:"long_desc"`
					PreTime  string `json:"pre_time"`
				}
			} `json:"list"`
		})
		if err = json.Unmarshal(data, &event); err != nil {
			log.Error("loadWebDatas event json.Unmarshal error(%v)", err)
			return
		}
		tmp := make(map[int64]*timemdl.Event, len(event.List))
		for _, v := range event.List {
			if v == nil || v.Data == nil {
				continue
			}
			tmp[v.Data.EventID] = &timemdl.Event{
				Title:   v.Data.Desc,
				Desc:    v.Data.LongDesc,
				PreTime: v.Data.PreTime,
			}
		}
		if len(tmp) > 0 {
			s.events = tmp
		}
	}()
	func() {
		if s.c.Timemachine.TagSid == 0 {
			return
		}
		data, err := s.likeDao.SourceItem(c, s.c.Timemachine.TagSid)
		if err != nil {
			log.Error("loadWebDatas s.likeDao.SourceItem(%d) error(%v)", s.c.Timemachine.TagSid, err)
			return
		}
		var tag = new(struct {
			List []*struct {
				Data *struct {
					TagID   int64  `json:"tag_id"`
					TagName string `json:"tag_name"`
					DescOne string `json:"desc_one"`
					DescTwo string `json:"desc_two"`
					Pic     string `json:"pic"`
				}
			} `json:"list"`
		})
		if err = json.Unmarshal(data, &tag); err != nil {
			log.Error("loadWebDatas tag json.Unmarshal error(%v)", err)
			return
		}
		tmp := make(map[int64]*timemdl.Tag, len(tag.List))
		for _, v := range tag.List {
			if v.Data == nil {
				continue
			}
			tmp[v.Data.TagID] = &timemdl.Tag{
				Name:    v.Data.TagName,
				DescOne: v.Data.DescOne,
				DescTwo: v.Data.DescTwo,
				Pic:     v.Data.Pic,
			}
		}
		if len(tmp) > 0 {
			s.tagDescs = tmp
		}
	}()
	func() {
		if s.c.Timemachine.RegionSid == 0 {
			return
		}
		data, err := s.likeDao.SourceItem(c, s.c.Timemachine.RegionSid)
		if err != nil {
			log.Error("loadWebDatas s.likeDao.SourceItem(%d) error(%v)", s.c.Timemachine.RegionSid, err)
			return
		}
		var region = new(struct {
			List []*struct {
				Data *struct {
					SubTid   int32  `json:"sub_tid"`
					Name     string `json:"name"`
					DescOne  string `json:"desc_one"`
					DescTwo  string `json:"desc_two"`
					Pic      string `json:"pic"`
					FlagDesc string `json:"flag_desc"`
				}
			} `json:"list"`
		})
		if err = json.Unmarshal(data, &region); err != nil {
			log.Error("loadWebDatas json.Unmarshal error(%v)", err)
			return
		}
		tmp := make(map[int32]*timemdl.RegionDesc, len(region.List))
		for _, v := range region.List {
			if v == nil || v.Data == nil {
				continue
			}
			if v.Data.SubTid == 0 {
				s.dftRegionDesc = &timemdl.RegionDesc{
					Tid:      v.Data.SubTid,
					Name:     v.Data.Name,
					DescOne:  v.Data.DescOne,
					DescTwo:  v.Data.DescTwo,
					Pic:      v.Data.Pic,
					FlagDesc: v.Data.FlagDesc,
				}
			}
			tmp[v.Data.SubTid] = &timemdl.RegionDesc{
				Tid:      v.Data.SubTid,
				Name:     v.Data.Name,
				DescOne:  v.Data.DescOne,
				DescTwo:  v.Data.DescTwo,
				Pic:      v.Data.Pic,
				FlagDesc: v.Data.FlagDesc,
			}
		}
		if len(tmp) > 0 {
			s.regionDescs = tmp
		}
	}()
	log.Info("loadWebDatas() success")
}

// StartTmproc start tm proc.
func (s *Service) StartTmproc(c context.Context) (err error) {
	s.dao.StartTmProc()
	return
}

// StopTmproc start tm proc.
func (s *Service) StopTmproc(c context.Context) (err error) {
	s.dao.StopTmproc()
	return
}
