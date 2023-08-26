package databus

import (
	"encoding/json"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-gateway/app/app-svr/archive-push/admin/internal/model"
)

func (d *IDatabus) GetUserAuthorizationSubMessages() <-chan *databus.Message {
	return d.UserAuthorizationSub.Messages()
}

// processUserAuthorizationMsg 处理用户授权状态变更消息
func (d *IDatabus) processUserAuthorizationMsg(msg *databus.Message) (err error) {
	if msg == nil {
		return xecode.NothingFound
	}
	content := &model.UserAuthorizationContent{}
	if d.svc.Cfg.Debug {
		log.Info("Databus: processUserAuthorizationMsg Start")
	}
	if err = json.Unmarshal(msg.Value, content); err != nil {
		log.Error("Databus: processUserAuthorizationMsg Unmarshal error (%v)", msg.Value, err)
		return
	}
	if d.svc.Cfg.Debug {
		log.Info("Databus: processUserAuthorizationMsg content %+v", content)
	}

	// 处理可绑定用户的厂商
	for _, vendor := range model.DefaultVendors {
		_vendor := vendor
		if _vendor.UserBindable {
			// 检查用户是否在白名单中
			if exists, _err := d.svc.CheckIfAuthorInWhiteList(_vendor.ID, content.MID); _err != nil {
				log.Error("Databus: processUserAuthorizationMsg CheckIfUserIntWhiteList(%d, %d) error %v", _vendor.ID, content.MID, _err)
				continue
			} else if !exists {
				continue
			}

			var sid int64
			if sid, err = d.svc.GetUserAuthorizationSIDByUser(_vendor.ID, content.MID); err != nil {
				log.Error("Databus: processUserAuthorizationMsg GetUserAuthorizationSIDByUser(%d) error %v", _vendor.ID, err)
				return
			} else if sid != content.SID {
				if d.svc.Cfg.Debug {
					log.Info("Databus: processUserAuthorizationMsg 活动ID(%d)不对应用户(%d, %d)", content.SID, _vendor.ID, content.MID)
				}
				continue
			}

			req := model.SyncAuthorAuthorizationReq{
				VendorID: _vendor.ID,
				MID:      content.MID,
			}

			if _err := d.svc.SyncAuthorAuthorization(req); _err != nil {
				log.Error("Databus: processUserAuthorizationMsg SyncAuthorAuthorization %+v error %v", req, _err)
			}
		}
	}

	return nil
}
