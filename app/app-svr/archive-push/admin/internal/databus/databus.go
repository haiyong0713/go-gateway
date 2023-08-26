package databus

import (
	"context"

	"go-common/library/conf/paladin"
	"go-common/library/log"
	xdatabus "go-common/library/queue/databus"

	"go-gateway/app/app-svr/archive-push/admin/internal/model"
	"go-gateway/app/app-svr/archive-push/admin/internal/service"
)

var (
	Cfg = &model.DatabusConfigs{}
)

type IDatabus struct {
	svc                  *service.Service
	ArchiveNotifySub     *xdatabus.Databus
	UserAuthorizationSub *xdatabus.Databus
}

// New 初始化databus配置
func New(_svc *service.Service) (res *IDatabus, cf func(), err error) {
	res = &IDatabus{}
	loadConfig()
	configChangeCh := paladin.WatchEvent(context.Background(), "databus.toml")
	go func() {
		// 监听配置文件变更
		for {
			sig := <-configChangeCh
			if sig.Event >= 0 {
				log.Info("databus config reload")
				loadConfig()
			}
		}
	}()
	res.ArchiveNotifySub = xdatabus.New(Cfg.ArchiveNotifySub)
	res.UserAuthorizationSub = xdatabus.New(Cfg.UserAuthorizationSub)

	res.svc = _svc
	cf = res.Close

	return
}

func loadConfig() {
	if err := paladin.Get("databus.toml").UnmarshalTOML(Cfg); err != nil {
		panic(err)
	}
}

func (d *IDatabus) Start() {
	var (
		msg *xdatabus.Message
		ok  bool
	)
	log.Info("Databus: Start consumer")
	// 稿件审核状态变更
	go func() {
		for {
			if msg, ok = <-d.GetArchiveNotifySubMessages(); !ok {
				log.Error("Databus: Archive Notify GetSubMessages Error (%v)", msg)
				break
			} else {
				if err := msg.Commit(); err != nil {
					log.Error("Databus: GetArchiveNotifySubMessages msg.Commit Error(%+v)", err)
				}
				if err := d.processArchiveNotifyMsg(msg); err != nil {
					log.Error("Databus: GetArchiveNotifySubMessages ProcessArchiveNotifyMsg(%v) Error(%+v)", msg, err)
				}
			}
		}
	}()
	// 用户授权状态变更
	go func() {
		for {
			if msg, ok = <-d.GetUserAuthorizationSubMessages(); !ok {
				log.Error("Author Notify GetSubMessages Error (%v)", msg)
				break
			} else {
				if err := msg.Commit(); err != nil {
					log.Error("Databus: GetUserAuthorizationSubMessages msg.Commit Error(%+v)", err)
				}
				if err := d.processUserAuthorizationMsg(msg); err != nil {
					log.Error("Databus: GetUserAuthorizationSubMessages processUserAuthorizationMsg(%v) Error(%+v)", msg, err)
				}
			}
		}
	}()
}

func (d *IDatabus) Close() {
	if d.ArchiveNotifySub != nil {
		if err := d.ArchiveNotifySub.Close(); err != nil {
			log.Error("archive-push-admin.databus.ArchiveNotifySub.Close Error %v", err)
		}
	}
}
