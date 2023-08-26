package vogue

import (
	"context"
	"encoding/json"
	"time"

	accountAPI "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	archiveApi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/lottery"
	dao "go-gateway/app/web-svr/activity/interface/dao/vogue"
	model "go-gateway/app/web-svr/activity/interface/model/vogue"
	"go-gateway/pkg/idsafe/bvid"

	silver "git.bilibili.co/bapis/bapis-go/silverbullet/service/silverbullet-proxy"
)

// Service struct
type Service struct {
	c                  *conf.Config
	dao                *dao.Dao
	lottDao            *lottery.Dao
	accClient          accountAPI.AccountClient
	silverBulletClient silver.SilverbulletProxyClient
}

// New Service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:       c,
		dao:     dao.New(c),
		lottDao: lottery.New(c),
	}
	var err error
	if s.accClient, err = accountAPI.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	if s.silverBulletClient, err = silver.NewClient(c.SilverBullet); err != nil {
		panic(err)
	}

	return
}

func (s *Service) playList(c context.Context) (res map[int64]struct{}, err error) {
	var (
		data string
		play []*model.PlayListItem
	)
	res = map[int64]struct{}{}
	if data, err = s.dao.Config(c, "play_list"); err != nil {
		log.Error("s.dao.Config(%v)", err)
		return nil, err
	}
	if data == "" {
		return nil, err
	}
	if err = json.Unmarshal([]byte(data), &play); err != nil {
		log.Error("json.Unmarshal(%v)", err)
		return nil, err
	}
	for _, n := range play {
		list, err := s.dao.FavList(c, n.Id)
		if err != nil {
			return nil, err
		}
		for _, m := range list {
			var aid int64
			if aid, err = bvid.BvToAv(m.Bvid); err != nil {
				return nil, err
			}
			pageReply, err := client.ArchiveClient.Page(c, &archiveApi.PageRequest{Aid: aid})
			if err != nil {
				log.Error("s.archiveRPC.Page(aid:%v),error(%v)", aid, err)
				return nil, err
			}
			for _, l := range pageReply.Pages {
				res[l.Cid] = struct{}{}
			}
		}
		time.Sleep(time.Second)
	}
	return
}

func (s *Service) taskList(c context.Context) (res []int64, err error) {
	res = make([]int64, 0, 0)
	var id int64
	for {
		var data []int64
		data, id, err = s.dao.TaskMid(c, id)
		if err != nil {
			return nil, err
		}
		res = append(res, data...)
		if len(data) < 100 {
			return
		}
		time.Sleep(time.Second)
	}
}

func (s *Service) inActive(c context.Context) (err error) {
	start, end, err := s.actTime(c)
	if err != nil {
		return nil
	}
	now := time.Now().Unix()
	if now < start {
		return ecode.ActivityNotStart
	}
	if end < now {
		return ecode.ActivityOverEnd
	}
	return
}
