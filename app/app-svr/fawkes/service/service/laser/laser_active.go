package laser

import (
	"context"

	"go-common/library/database/sql"

	"go-gateway/app/app-svr/fawkes/service/model"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	"go-gateway/app/app-svr/fawkes/service/model/business"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

func (s *Service) AppLaserActiveList(c context.Context, appKey, buvid string, mid, laserId int64, pn, ps int) (res *model.LaserResult, err error) {
	var (
		total int
		ls    []*appmdl.Laser
	)
	// 如果 mid 和 buvid 都为空. 则不去查询count(*)
	// 存在查询效率异常
	if mid != 0 || buvid != "" || laserId != 0 {
		if total, err = s.fkDao.LaserActiveCount(c, appKey, buvid, mid, laserId, 0, 0); err != nil {
			log.Error("%v", err)
			return
		}
	}
	if ls, err = s.fkDao.LaserActiveList(c, appKey, buvid, mid, laserId, 0, 0, pn, ps); err != nil {
		log.Error("%v", err)
		return
	}
	if total == 0 && len(ls) > 0 {
		total = 1000
	}
	res = &model.LaserResult{
		PageInfo: &model.PageInfo{
			Total: total,
			Pn:    pn,
			Ps:    ps,
		},
		Items: ls,
	}
	return
}

func (s *Service) AppLaserAdd2(c context.Context, appKey, buvid string, url, recallMobiApp, build, errorMessage, md5, rawUposUri string, status int, mid, taskID int64) (res *business.ActiveLaser2Result, err error) {
	var (
		tx      *sql.Tx
		rTaskID int64
	)
	// 初始化res
	res = &business.ActiveLaser2Result{TaskID: taskID}
	// 设置默认值
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	// 旧版本不传TaskID. 则进行直接写入
	// 新版本，传了TaskID. 则判断是否为已有数据. 若存在则修改； 若不存在则写入
	var activeLasers []*appmdl.Laser
	if taskID != 0 {
		if activeLasers, err = s.fkDao.LaserActiveByID(c, appKey, taskID); err != nil {
			log.Error("tx.Commit() error(%v)", err)
			return
		}
	}
	// 因为用主键id查询. 数据count必为 0/1
	if len(activeLasers) > 1 {
		log.Error("AppLaserAdd2 查询异常 %v %v %v", len(activeLasers), appKey, taskID)
		return
	}
	// 若存在：则进行数据修改
	if len(activeLasers) == 1 {
		laser := activeLasers[0]
		// 已经为成功/失败的数据源. 不支持覆盖操作
		if laser.Status == appmdl.StatusUpSuccess || laser.Status == appmdl.StatusUpFaild {
			log.Error("AppLaserAdd2 修改异常. 状态已为成功失败 %v %v %v", appKey, taskID, status)
			return
		}
		if _, err = s.fkDao.TxUpdateLaser2(tx, appKey, url, errorMessage, md5, rawUposUri, status, taskID); err != nil {
			log.Error("%v", err)
			return
		}
	} else {
		// 若不存在：直接进行数据插入
		if rTaskID, err = s.fkDao.TxAddLaser2(tx, appKey, buvid, url, recallMobiApp, build, errorMessage, md5, rawUposUri, status, mid); err != nil {
			log.Error("%v", err)
			return
		}
		res.TaskID = rTaskID
	}
	return
}
