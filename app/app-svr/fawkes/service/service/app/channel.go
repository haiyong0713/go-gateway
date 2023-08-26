package app

import (
	"context"

	"go-common/library/database/sql"
	"go-common/library/ecode"

	"go-gateway/app/app-svr/fawkes/service/model"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// ChannelList get static channel list
func (s *Service) ChannelList(c context.Context, size, page int, filterKey string) (result model.ChannelResult, err error) {
	var (
		total    int
		chLists  []*appmdl.Channel
		pageInfo *model.PageInfo
	)
	pageInfo = &model.PageInfo{}
	result = model.ChannelResult{}
	pageInfo.Pn = page
	pageInfo.Ps = size
	if total, err = s.fkDao.GetChannelCount(c, filterKey); err != nil {
		log.Error("%v", err)
		return
	}
	pageInfo.Total = total
	if chLists, err = s.fkDao.ChannelList(c, page, size, filterKey); err != nil {
		log.Error("%v", err)
		return
	}
	result.Page = pageInfo
	result.Channel = chLists
	return
}

// ChannelAdd add static channel
func (s *Service) ChannelAdd(c context.Context, code, name, plate, operator string, status int8, isSync bool) (err error) {
	var (
		tx                 *sql.Tx
		channelID, groupID int64
	)
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
	if channelID, err = s.fkDao.GetChannelIDByCode(c, code, name, plate); err != nil {
		log.Error("s.fsDao.CheckChannelExist() error(%v)", err)
		return
	}
	if channelID != 0 {
		if _, err = s.fkDao.TxChannelToStatic(tx, channelID, operator); err != nil {
			log.Error("%v", err)
			return
		}
	} else if channelID, err = s.fkDao.TxChannelAdd(tx, code, name, plate, operator, status, appmdl.ChannelNormal); err != nil {
		log.Error("%v", err)
		return
	}
	if isSync {
		var (
			apps []*appmdl.APP
		)
		if apps, err = s.fkDao.AppsPass(c, []string{}, operator, 0); err != nil {
			log.Error("%v", err)
			return
		}
		for _, app := range apps {
			if app.Platform == "ios" {
				continue
			}
			var count int
			if count, err = s.fkDao.CheckAppChannel(c, app.AppKey, channelID); err != nil {
				log.Error("%v", err)
				return
			}
			if count != 0 {
				continue
			}
			if _, err = s.fkDao.AppChannelAdd(tx, channelID, groupID, app.AppKey, operator); err != nil {
				log.Error("%v", err)
				return
			}
		}
	}
	return
}

// ChannelDelete delete static channel
func (s *Service) ChannelDelete(c context.Context, channelID int64, operator string) (err error) {
	var mc *appmdl.Channel
	if mc, err = s.fkDao.ChannelByCode(c, "master"); err != nil {
		log.Error("%v", err)
		return
	}
	if mc != nil {
		if mc.ID == channelID {
			return ecode.Error(ecode.NothingFound, "master渠道禁止删除")
		}
	}
	count, err := s.fkDao.CheckChannelByID(c, channelID)
	if err != nil {
		log.Error("s.fkDao.CheckChannelFromID %v", err)
		return
	}
	if count == 0 {
		return ecode.Error(ecode.NothingFound, "静态渠道不存在，请核对后操作")
	}
	var tx *sql.Tx
	if tx, err = s.fkDao.BeginTran(c); err != nil {
		log.Error("s.fdDao.BeginTran failed. %v", err)
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
				log.Error("rx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if _, err = s.fkDao.TxChannelDelete(tx, channelID, operator); err != nil {
		log.Error("%v", err)
	}
	return
}

// AppChannelListV2 get app channel list pagination
func (s *Service) AppChannelListV2(c context.Context, appkey, filterKey string, groupID int64, pn, ps int) (result model.ChannelResult, err error) {
	var (
		total    int
		chLists  []*appmdl.Channel
		pageInfo *model.PageInfo
	)
	pageInfo = &model.PageInfo{}
	result = model.ChannelResult{}
	total, err = s.fkDao.AppChannelListCount(c, appkey, filterKey, groupID)
	if err != nil {
		log.Error("s.fdDao.AppChannelListCount failed. %v", err)
		return
	}
	pageInfo.Pn = pn
	pageInfo.Ps = ps
	pageInfo.Total = total
	if chLists, err = s.fkDao.AppChannelList(c, appkey, filterKey, "", "", pn, ps, groupID); err != nil {
		log.Error("s.fdDao.AppChannelList failed. %v", err)
	}
	result.Channel = chLists
	result.Page = pageInfo
	return
}

// AppChannelList get app channel list
func (s *Service) AppChannelList(c context.Context, appkey, filterKey string, groupID int64) (result []*appmdl.Channel, err error) {
	if result, err = s.fkDao.AppChannelList(c, appkey, filterKey, "", "", -1, -1, groupID); err != nil {
		log.Error("s.fdDao.AppChannelList failed. %v", err)
	}
	return
}

// AppChannelAdd add app channel
func (s *Service) AppChannelAdd(c context.Context, chType int, channelID, groupID int64, code, name, plate, operator, appKey string) (err error) {
	var tx *sql.Tx
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
	// nolint:gomnd
	if chType == 1 {
		var chel *appmdl.Channel
		if chel, err = s.fkDao.GetChannelByID(c, channelID); err != nil {
			log.Error("s.fkDao.GetChannelByID() failed. %v", err)
			return
		}
		if chel == nil {
			return ecode.Error(ecode.NothingFound, "静态渠道不存在，请优先配置静态渠道")
		}
		var count int
		if count, err = s.fkDao.CheckAppChannel(c, appKey, channelID); err != nil {
			log.Error("s.fkDao.CheckAppChannel() failed. %v", err)
			return
		}
		if count != 0 {
			return ecode.Error(ecode.Conflict, "App已与该渠道关联，请勿重复添加")
		}
	} else if chType == 2 {
		if channelID, err = s.fkDao.GetChannelIDByCode(c, code, name, plate); err != nil {
			log.Error("s.fsDao.GetChannelIDByCode() failed. %v", err)
			return
		}
		if channelID == 0 {
			if channelID, err = s.fkDao.TxChannelAdd(tx, code, name, plate, operator, appmdl.ChannelCustom, appmdl.ChannelNormal); err != nil {
				log.Error("s.fkDao.TxChannelAdd() failed. %v", err)
				return
			}
		} else {
			var count int
			if count, err = s.fkDao.CheckAppChannel(c, appKey, channelID); err != nil {
				log.Error("s.fkDao.CheckAppChannel() failed. %v", err)
				return
			}
			if count != 0 {
				return ecode.Error(ecode.Conflict, "App已与该渠道关联，请勿重复添加")
			}
		}
	}
	_, err = s.fkDao.AppChannelAdd(tx, channelID, groupID, appKey, operator)
	if err != nil {
		log.Error("s.fkDao.AppChannelAdd() failed. %v", err)
	}
	return
}

// AppChannelDelete delete app channel
func (s *Service) AppChannelDelete(c context.Context, appKey string, channelID int64) (err error) {
	var mc *appmdl.Channel
	if mc, err = s.fkDao.ChannelByCode(c, "master"); err != nil {
		log.Error("%v", err)
		return
	}
	if mc != nil {
		if mc.ID == channelID {
			return ecode.Error(ecode.NothingFound, "master渠道禁止删除")
		}
	}
	count, err := s.fkDao.CheckAppChannel(c, appKey, channelID)
	if err != nil {
		log.Error("s.fkDao.CheckAppChannel failed. %v", err)
		return
	}
	if count == 0 {
		err = ecode.Error(ecode.NothingFound, "APP未关联该渠道，请核对后操作")
		return
	}
	var tx *sql.Tx
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
				log.Error("rx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	_, err = s.fkDao.AppChannelDelete(tx, appKey, channelID)
	if err != nil {
		log.Error("s.fkDao.AppChannelDelete failed. %v", err)
		return
	}
	var chel *appmdl.Channel
	if chel, err = s.fkDao.GetChannelByID(c, channelID); err != nil {
		log.Error("s.fkDao.GetChannelByID failed. %v", err)
		return
	}
	if chel.Status == appmdl.ChannelCustom {
		if count, err = s.fkDao.GetAppCountByID(c, channelID); err != nil {
			log.Error("s.fkDao.GetAppCountByID() failed. %v", err)
			return
		}
		if count <= 1 {
			if _, err = s.fkDao.TxCustomChannelDeleteByID(tx, channelID); err != nil {
				log.Error("s.fkDao.TxCustomChannelDeleteByID() failed. %v", err)
			}
		}
	}
	return
}

// AppChannelGroupRelate relate app channel to group
func (s *Service) AppChannelGroupRelate(c context.Context, appChannelIDs []int64, groudID int64, userName string) (err error) {
	var tx *sql.Tx
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
	if err = s.fkDao.TxAppChannelGroupRelate(tx, appChannelIDs, groudID, userName); err != nil {
		log.Error("TxAppChannelGroupRelate  error")
	}
	return
}

// AppChannelGroupList get app channel group list
func (s *Service) AppChannelGroupList(c context.Context, appKey, filterKey string) (res []*appmdl.ChannelGroup, err error) {
	if res, err = s.fkDao.AppChannelGroupList(c, appKey, filterKey); err != nil {
		log.Error("AppChannelGroupList failed")
	}
	return
}

// AppChannelGroupAdd add app channelgroup
func (s *Service) AppChannelGroupAdd(c context.Context, appKey, name, description, username string, autoPushCdn, isAutoGen int64, qaOwner, marketOwner string, priority int64) (err error) {
	var tx *sql.Tx
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
	if err := s.fkDao.TxAppChannelGroupAdd(tx, appKey, name, description, username, qaOwner, marketOwner, autoPushCdn, isAutoGen, priority); err != nil {
		log.Error("AppChannelGroupAdd Tx  error")
	}
	return
}

// AppChannelGroupUpdate update APP channel group
func (s *Service) AppChannelGroupUpdate(c context.Context, id int64, name, description, userName string, autoPushCdn, autoGen int64, qaOwner, marketOwner string, priority int64) (err error) {
	var tx *sql.Tx
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
	if err = s.fkDao.TxAppChannelGroupUpdate(tx, id, name, description, userName, autoPushCdn, autoGen, qaOwner, marketOwner, priority); err != nil {
		log.Error("AppChannelGroupUpdate tx error %s", err)
		return
	}
	return
}

// AppChannelGroupDel del app channel group
func (s *Service) AppChannelGroupDel(c context.Context, id int64, userName string) (err error) {
	var tx *sql.Tx
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
	if err = s.fkDao.TxResetAppChannelGroup(tx, id, userName); err != nil {
		log.Error("TxResetAppChannelGroup error %s", err)
		return
	}
	if err = s.fkDao.TxAppChannelGroupDel(tx, id, userName); err != nil {
		log.Error("TxAppChannelGroupDel tx error %s", err)
		return
	}
	return
}
