package popups

import (
	"context"
	"fmt"
	crowd "git.bilibili.co/bapis/bapis-go/platform/service/bgroup"
	"github.com/robfig/cron"
	"go-common/library/ecode"
	"go-common/library/log"
	pb "go-gateway/app/app-svr/resource/service/api/v1"
	"go-gateway/app/app-svr/resource/service/model"
	"time"
)

const (
	dataFormat    = "2006-01-02 15:04:00"
	_getPopUpsSQL = `select id, img_url, img_url_ipad, description, redirect_type, redirect_target,
		builds, crowd_type,crowd_base, crowd_value, teenage_push,  auto_hide_status,auto_hide_countdown from bilibili_manager.popup_config 
		where deleted_flag=0 and stime <= ? and etime >= ? and audit_state=1 order by stime asc`
)

func crowdTimeValid(bgGroups map[int64]*crowd.BGroup, crow_val int64) (bool, error) {
	now := time.Now().Unix()
	if len(bgGroups) == 0 || bgGroups[crow_val].EndTime < now {
		return false, nil
	} else {
		return true, nil
	}
}

func (d *Dao) CheckCrowd(req *pb.PopUpsReq, crow_base int, crow_value int64) (valid bool, err error) {
	var (
	//bgids = []int64{39}
	)
	//nolint:gomnd
	if crow_base == 1 {
		if req.Mid < 1 {
			err = fmt.Errorf("请求MID格式不合法")
			log.Error("resource.Popups.CheckCrow request mid and crowd_base not match")
			return false, err
		}
		arg := &crowd.MidBGroupsReq{
			Mid:          req.Mid,
			BusinessName: "TIANMA_POPUP",
			//BgroupIds: bgids,
		}
		ret, err := d.CrowdGRPC.MidBGroups(context.Background(), arg)
		if err != nil {
			log.Error("resource.Popups.checkCrow mid(%d) error(%d)", req.Mid, err)
			return false, err
		}
		return crowdTimeValid(ret.Bgroups, crow_value)
	} else if crow_base == 2 {
		if req.Buvid == "" {
			err = fmt.Errorf("请求Buvid格式不合法")
			log.Error("resource.Popups.CheckCrow request buvid and crowd_base not match")
			return false, err
		}
		arg := &crowd.BuvidBGroupsReq{
			Buvid:        req.Buvid,
			BusinessName: "TIANMA_POPUP",
			//BgroupIds: bgids,
		}
		ret, err := d.CrowdGRPC.BuvidBGroups(context.Background(), arg)
		if err != nil {
			log.Error("resource.Popups.checkCrow mid(%d) error(%d)", req.Mid, err)
			return false, err
		}
		return crowdTimeValid(ret.Bgroups, crow_value)
	}
	return false, ecode.RequestErr
}

func (d *Dao) GetMysqlPopUps(c context.Context) (ret []*model.PopUps, err error) {
	now := time.Now().Format(dataFormat)
	rows, err := d.db.Query(c, _getPopUpsSQL, now, now)
	if err != nil {
		log.Error("GetMysqlPopups Query error: %s", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		item := &model.PopUps{}
		if err = rows.Scan(&item.ID, &item.Pic, &item.PicIpad, &item.Description, &item.LinkType, &item.Link,
			&item.Builds, &item.CrowdType, &item.CrowdBase, &item.CrowdValue, &item.TeenagePush, &item.AutoHideStatus, &item.CloseTime); err != nil {
			log.Error("GetMysqlPopups rows Scan error: %s", err)
			return
		}
		ret = append(ret, item)
	}
	err = rows.Err()
	if err != nil {
		log.Error("GetMysqlPopups rows error: %s", err)
	}
	return ret, err
}

func (d *Dao) GetEffectivePopUps(c context.Context) (ret []*model.PopUps, err error) {
	// 查找缓存
	if len(d.popCache) == 0 {
		err = ecode.NothingFound
		log.Error("Get dao.GetEffectivePopUps error: %s", err)
		return
	}
	return d.popCache, err
}

func (d *Dao) UpdatePopUpsCache(c context.Context) (err error) {
	// 从mysql中取出数据并存储
	var popups []*model.PopUps
	if popups, err = d.GetMysqlPopUps(c); err != nil {
		log.Error("dao.UpdatePopUpsCache error: %s", err.Error())
		return
	}
	d.popCache = popups
	return
}

func (d *Dao) FlushPopUpsCache() {
	var err error
	c := cron.New()
	// 每10秒刷新一下缓存
	err = c.AddFunc("*/10 * * * *", func() {
		if err = d.UpdatePopUpsCache(context.Background()); err != nil {
			log.Error("popupsDao.FlushPopUpsCache error: (%v)", err)
		}
	})
	if err != nil {
		log.Error("popupsDao.FlushPopUpsCache error: (%v)", err)
	}
	c.Start()
}
