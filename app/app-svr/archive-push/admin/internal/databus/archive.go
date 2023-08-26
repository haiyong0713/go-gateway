package databus

import (
	"encoding/json"
	archiveGRPC "git.bilibili.co/bapis/bapis-go/archive/service"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-gateway/app/app-svr/archive-push/admin/api"
	"go-gateway/app/app-svr/archive-push/admin/internal/model"
	"go-gateway/app/app-svr/archive-push/admin/internal/util"
	archiveEcode "go-gateway/app/app-svr/archive/ecode"
)

func (d *IDatabus) GetArchiveNotifySubMessages() <-chan *databus.Message {
	return d.ArchiveNotifySub.Messages()
}

// processArchiveNotifyMsg 处理稿件审核状态变更消息
func (d *IDatabus) processArchiveNotifyMsg(msg *databus.Message) (err error) {
	if msg == nil {
		return xecode.NothingFound
	}
	content := &model.ArchiveNotify{}
	if d.svc.Cfg.Debug {
		log.Info("Databus: processArchiveNotifyMsg")
	}
	if err = json.Unmarshal(msg.Value, content); err != nil {
		log.Error("Databus: processArchiveNotifyMsg.Unmarshal error (%v)", msg.Value, err)
		return
	}
	if d.svc.Cfg.Debug {
		log.Info("Databus: processArchiveNotifyMsg Old %+v", content.Old)
		log.Info("Databus: processArchiveNotifyMsg New %+v", content.New)
	}

	if content.New != nil && content.New.State >= 0 {
		return d.checkAndPushArchives(content.New)
	} else if content.New != nil && content.New.State < 0 {
		return d.checkAndWithdrawArchives(content.New)
	}
	return
}

// checkAndPushArchives 只有稿件在白名单且没有推送成功且符合可推送了才做上架操作
func (d *IDatabus) checkAndPushArchives(archive *model.ArchiveNotifyArchive) (err error) {
	if _err := d.checkAndPushArchivesQQCMC(archive); _err != nil {
		log.Error("Databus: checkAndPushArchives checkAndPushArchivesQQCMC %+v error %v", archive, _err)
	}
	if _err := d.checkAndPushArchivesQQTGL(archive); _err != nil {
		log.Error("Databus: checkAndPushArchives checkAndPushArchivesQQTGL %+v error %v", archive, _err)
	}
	if _err := d.checkAndPushArchivesBlizzard(archive); _err != nil {
		log.Error("Databus: checkAndPushArchives checkAndPushArchivesBlizzard %+v error %v", archive, _err)
	}

	return nil
}

// checkAndWithdrawArchives 只有稿件推送成功了才做下架操作
func (d *IDatabus) checkAndWithdrawArchives(archive *model.ArchiveNotifyArchive) (err error) {
	var (
		bvid string
	)
	if bvid, err = util.AvToBv(archive.AID); err != nil {
		log.Error("Databus: checkAndWithdrawArchives Got wrong AID(%d) (%v)", archive.AID, err)
		err = nil
		return
	}
	for _, vendor := range model.DefaultVendors {
		vendorID := vendor.ID
		var (
			details     model.ArchivePushDetailByBVIDSlice
			inWhiteList = false
		)

		// 首先检查bvid是否在白名单中
		if inWhiteList, err = d.svc.CheckIfBVIDInWhiteList(vendorID, bvid, archive.MID); err != nil {
			log.Error("Databus: checkAndWithdrawArchives (%d) %+v error %v", vendorID, archive, err)
			return
		} else if !inWhiteList {
			if d.svc.Cfg.Debug {
				log.Warn("Databus: checkAndWithdrawArchives (%d) %+v 稿件不在白名单中", vendorID, archive)
			}
			continue
		}

		// 获取稿件当前状态
		if details, err = d.svc.GetArchivesByPushStatus([]string{bvid}, vendorID, []api.ArchivePushDetailPushStatus_Enum{api.ArchivePushDetailPushStatus_SUCCESS}, 0); err != nil {
			log.Error("Databus: checkAndWithdrawArchives GetArchivesByPushStatus Error (%v)", err)
			return
		} else if len(details) == 0 {
			if d.svc.Cfg.Debug {
				log.Info("Databus: checkAndWithdrawArchives GetArchivesByPushStatus vendor: %d cannot find BVID (%s)", vendorID, bvid)
			}
			continue
		} else {
			// 只有稿件推送成功了才做下架操作
			if details[0].PushStatus == api.ArchivePushDetailPushStatus_Enum_name[int32(api.ArchivePushDetailPushStatus_SUCCESS)] && details[0].ArchiveStatus == api.ArchiveStatus_Enum_name[int32(api.ArchiveStatus_OPEN)] {
				var rawArchive *archiveGRPC.Arc
				if rawArchive, err = d.svc.GetArcByAID(archive.AID); err != nil {
					log.Error("Databus: checkAndWithdrawArchives (%d) GetArcByAID(%d) Error (%v)", vendorID, archive.AID, err)
					return
				} else if rawArchive == nil {
					log.Error("Databus: checkAndWithdrawArchives (%d) GetArcByAID(%d) not found", vendorID, archive.AID)
					err = archiveEcode.ArchiveNotExist
					continue
				}

				log.Info("Databus: checkAndWithdrawArchives vendor: %d bvid: %s 稿件状态变为不可用，下线", vendorID, bvid)
				err = d.svc.WithdrawArchive(bvid, "稿件状态变更不可用", vendorID, false, "system", 0)
			}
		}
	}

	return
}
