package college

import (
	"git.bilibili.co/bapis/bapis-go/account/service"

	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"

	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/college"
	"go-gateway/app/web-svr/activity/interface/dao/invite"
	"go-gateway/app/web-svr/activity/interface/dao/rank"
	mdlcollege "go-gateway/app/web-svr/activity/interface/model/college"
	mdlrank "go-gateway/app/web-svr/activity/interface/model/rank"
	"go-gateway/app/web-svr/activity/interface/service/account"
	"go-gateway/app/web-svr/activity/interface/service/archive"
	"go-gateway/app/web-svr/activity/interface/service/like"

	relationAPI "git.bilibili.co/bapis/bapis-go/account/service/relation"
	passportinfoapi "git.bilibili.co/bapis/bapis-go/passport/service/user"
	"github.com/robfig/cron"
)

// Service ...
type Service struct {
	c                     *conf.Config
	college               college.Dao
	invite                invite.Dao
	rank                  rank.Dao
	cron                  *cron.Cron
	accClient             api.AccountClient
	like                  *like.Service
	passportClient        passportinfoapi.PassportUserClient
	provinceRank          *mdlcollege.ProvinceCollegeRank
	account               *account.Service
	archive               *archive.Service
	tagRPC                tagrpc.TagRPCClient
	version               *mdlcollege.Version
	nationLastVersion     int
	provinceLastVersion   int
	provinceCollegeRank   *mdlcollege.ProvinceCollegeRank
	nationWideCollegeRank []*mdlrank.Redis
	allCollege            []*mdlcollege.Detail
	allProvince           []*mdlcollege.Province
	relationClient        relationAPI.RelationClient
}

// New ...
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:       c,
		college: college.New(c),
		rank:    rank.New(c),
		account: account.New(c),
		archive: archive.New(c),
		invite:  invite.New(c),
		cron:    cron.New(),
	}
	// 启动时上一次版本号为0
	s.nationLastVersion = -1
	s.provinceLastVersion = -1
	s.provinceCollegeRank = &mdlcollege.ProvinceCollegeRank{}
	s.provinceCollegeRank.Init()
	s.allProvince = make([]*mdlcollege.Province, 0)
	s.allCollege = make([]*mdlcollege.Detail, 0)
	s.nationWideCollegeRank = make([]*mdlrank.Redis, 0)
	var err error
	if s.accClient, err = api.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	if s.passportClient, err = passportinfoapi.NewClient(c.PassportClient); err != nil {
		panic(err)
	}
	if s.relationClient, err = relationAPI.NewClient(c.RelationClient); err != nil {
		panic(err)
	}
	if s.tagRPC, err = tagrpc.NewClient(c.TagClient); err != nil {
		panic(err)
	}
	// s.loadAllProvince()
	// s.loadVersion()
	// s.loadProvinceRank()
	// s.loadAllCollege()
	// s.loadNationwideRank()
	// s.createCron()
	return s
}

func (s *Service) createCron() {
	var err error
	if s.c.College.VersionChangeCron != "" {
		if err = s.cron.AddFunc(s.c.College.VersionChangeCron, s.loadVersion); err != nil {
			panic(err)
		}
		if err = s.cron.AddFunc(s.c.College.VersionChangeCron, s.loadProvinceRank); err != nil {
			panic(err)
		}
		if err = s.cron.AddFunc(s.c.College.VersionChangeCron, s.loadNationwideRank); err != nil {
			panic(err)
		}

	}
	if s.c.College.CollegeChangeCron != "" {
		if err = s.cron.AddFunc(s.c.College.CollegeChangeCron, s.loadAllProvince); err != nil {
			panic(err)
		}
		if err = s.cron.AddFunc(s.c.College.CollegeChangeCron, s.loadAllCollege); err != nil {
			panic(err)
		}
	}
	s.cron.Start()
}

// Close ...
func (s *Service) Close() {
	s.college.Close()
	s.rank.Close()
}
