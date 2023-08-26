package service

import (
	"context"
	"fmt"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/space/admin/model"
	"go-gateway/app/web-svr/space/admin/util"
	"go-gateway/pkg/idsafe/bvid"

	"github.com/robfig/cron"
)

func (s *Service) UpdateWhitelistState() (err error) {
	c := cron.New()
	// 每10秒更新下白名单状态
	err = c.AddFunc("*/10 * * * *", func() {
		var lists []*model.WhitelistAdd
		if lists, err = s.dao.FindWhiteByStatus(model.StatusValid); err != nil {
			log.Error("service.UpdateWhitelistState arg(%v) error(%v)", model.StatusValid, err)
			return
		}
		for _, item := range lists {
			if err = s.dao.ChangeStatus(item.ID, model.StatusFailed); err != nil {
				log.Error("service.UpadateWhitelistState.ChangeStatus arg(%v) error(%v)", item, err)
				return
			}
			if err = s.dao.ClearWhitelistInfo(context.Background(), item.Mid); err != nil {
				log.Error("service.UpdateWhitelistState.ClearWhitelistInfo mid(%v) error(%v)", item.Mid, err)
			}
		}
		if lists, err = s.dao.FindWhiteByStatus(model.StatusReady); err != nil {
			log.Error("service.UpdateWhitelistState arg(%v) error(%v)", model.StatusValid, err)
			return
		}
		for _, item := range lists {
			if err = s.dao.ChangeStatus(item.ID, model.StatusValid); err != nil {
				log.Error("service.UpadateWhitelistState.ChangeStatus arg(%v) error(%v)", item, err)
				return
			}
		}

	})
	c.Start()
	return
}

// WhitelistAdd add whitelist
func (s *Service) WhitelistAdd(c context.Context, arg *model.WhitelistReq) (failedList []int64, err error) {
	var (
		midInfo *model.MidInfoReply
		state   int
		ok      bool
	)
	now := time.Now().Unix()
	if arg.Stime.Time().Unix() <= now && arg.Etime.Time().Unix() > now {
		state = model.StatusValid
	} else if arg.Etime.Time().Unix() < now {
		state = model.StatusFailed
	} else if arg.Stime.Time().Unix() > now {
		state = model.StatusReady
	}
	arg.Mids = util.RemoveRep(arg.Mids)
	whitelist := []*model.WhitelistAdd{}
	for _, mid := range arg.Mids {
		if ok, err = s.dao.ValidWhitelistMid(mid); err != nil {
			log.Error("service.WhitelistAdd.ValidWhitelistMid arg(%v) error(%v)", arg, err)
			return
		}
		if !ok {
			failedList = append(failedList, mid)
			continue
		}
		if midInfo, err = s.MidInfo(c, mid); err != nil {
			err = fmt.Errorf("无效的Mid(%v)", mid)
			log.Error("service.WhitelistAdd arg(%v) error(%v)", arg, err)
			return
		}
		whitelist = append(whitelist, &model.WhitelistAdd{
			Mid:      mid,
			MidName:  midInfo.MidName,
			Stime:    arg.Stime,
			Etime:    arg.Etime,
			State:    state,
			Username: arg.Username,
		})
	}
	if err = s.dao.WhitelistAdd(whitelist); err != nil {
		log.Error("service.WhitelistAdd error(%v)", err)
		return
	}
	return
}

// WhitelistUp update whitelist
func (s *Service) WhitelistUp(arg *model.WhitelistAdd) (err error) {
	var (
		tmp   *model.WhitelistAdd
		state int
	)
	if tmp, err = s.dao.WhitelistFindById(arg.ID); err != nil {
		log.Error("s.WhitelistUp.WhitelistFindById arg(%v) error(%v)", arg, err)
		return
	}
	state = tmp.State
	//nolint:gomnd
	if tmp.State == 1 {
		if arg.Stime != tmp.Stime {
			err = fmt.Errorf("开始时间不可更改")
			return err
		}
	} else if tmp.State == 3 {
		err = fmt.Errorf("已失效的配置")
		return err
	} else {
		now := time.Now().Unix()
		if arg.Stime.Time().Unix() <= now && arg.Etime.Time().Unix() > now {
			state = model.StatusValid
		} else if arg.Etime.Time().Unix() < now {
			state = model.StatusFailed
			if err = s.dao.ClearWhitelistInfo(context.Background(), tmp.Mid); err != nil {
				log.Error("s.WhitelistUp.ClearWhitelistInfo mid(%v) error(%v)", tmp.Mid, err)
				return
			}
		} else if arg.Stime.Time().Unix() > now {
			state = model.StatusReady
		}
	}
	arg.State = state
	if err = s.dao.WhitelistUp(arg); err != nil {
		log.Error("s.WhitelistUp.WhitelistUp arg(%v) error(%v)", arg, err)
		return
	}
	return
}

func (s *Service) WhitelistDel(id int64, t int) (err error) {
	if err = s.dao.WhitelistDelete(id, t); err != nil {
		log.Error("serviece.WhitelistDelete id(%v) error(%v)", id, err)
		return
	}
	return
}

// WhitelistIndex .
func (s *Service) WhitelistIndex(mid int64, pn, ps, status int) (pager *model.WhitelistPager, err error) {
	var (
		mids         []int64
		topPhotoInfo map[int64]*model.TopPhotoArc
	)
	if pager, err = s.dao.WhitelistIndex(mid, pn, ps, status); err != nil {
		return
	}
	if len(pager.Item) == 0 {
		return
	}
	for _, item := range pager.Item {
		mids = append(mids, item.Mid)
	}
	if topPhotoInfo, err = s.TopPhotoArcs(context.Background(), mids); err != nil {
		log.Error("serviece.WhitelistIndex.TopPhotoArcs mids(%v) error(%v)", mids, err)
		return
	}
	for _, item := range pager.Item {
		if _, ok := topPhotoInfo[item.Mid]; ok {
			var bvstring string
			if bvstring, err = bvid.AvToBv(topPhotoInfo[item.Mid].Aid); err != nil {
				log.Error("service.WhitelistIndex.AvtoBv(%v) error(%v)", topPhotoInfo[item.Mid].Aid, err)
				err = nil
			}
			item.MidConf = model.TopPhotoConf{
				Mid:      topPhotoInfo[item.Mid].Mid,
				Bvid:     bvstring,
				ImageUrl: topPhotoInfo[item.Mid].ImageURL,
			}
		}
	}
	return
}
