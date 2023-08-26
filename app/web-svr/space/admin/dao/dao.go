package dao

import (
	"context"

	"go-common/library/database/orm"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/space/admin/conf"

	midrpc "git.bilibili.co/bapis/bapis-go/account/service"
	controlGRPC "git.bilibili.co/bapis/bapis-go/account/service/account_control_plane"
	moralrpc "git.bilibili.co/bapis/bapis-go/account/service/member"
	relationGRPC "git.bilibili.co/bapis/bapis-go/account/service/relation"
	sysMsgGRPC "git.bilibili.co/bapis/bapis-go/system-msg/interface"
	"github.com/jinzhu/gorm"
)

// Dao .
type Dao struct {
	c                  *conf.Config
	DB                 *gorm.DB
	http               *httpx.Client
	messageURL         string
	clearMsgURL        string
	clearTopPhotoURL   string
	midClient          midrpc.AccountClient
	moralClient        moralrpc.MemberClient
	relationClient     relationGRPC.RelationClient
	controlClient      controlGRPC.AccountControlPlaneClient
	systemMsgClient    sysMsgGRPC.SystemMsgClient
	usertabURL         string
	fansURL            string
	actionLogURL       string
	vipInfoURL         string
	bfsMoveURL         string
	notifySendURL      string
	accountBlockURL    string
	creditBlockInfoURL string
	delMoralURL        string
	purgeCacheURL      string
	//
	clearCacheTopPhotoURL string
}

// New .
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// conf
		c: c,
		// db
		DB: orm.NewMySQL(c.ORM),
		// http
		http: httpx.NewClient(c.HTTPClient),
	}
	var err error
	if d.midClient, err = midrpc.NewClient(nil); err != nil {
		panic(err)
	}
	if d.relationClient, err = relationGRPC.NewClient(nil); err != nil {
		panic(err)
	}
	if d.moralClient, err = moralrpc.NewClient(nil); err != nil {
		panic(err)
	}
	if d.controlClient, err = controlGRPC.NewClient(nil); err != nil {
		panic(err)
	}
	if d.systemMsgClient, err = sysMsgGRPC.NewClient(nil); err != nil {
		panic(err)
	}
	d.DB.LogMode(true)
	d.messageURL = d.c.Host.Message + _sysMessageURI
	d.clearMsgURL = d.c.Host.Api + _clearCacheURI
	d.clearTopPhotoURL = d.c.Host.Api + _topPhotoUrl
	d.usertabURL = d.c.Host.Manager + _offlineUsertab

	d.fansURL = d.c.Host.Manager + _fansURL
	d.actionLogURL = d.c.Host.Manager + _actionLog
	d.vipInfoURL = d.c.Host.Vip + _vipInfo
	d.bfsMoveURL = d.c.Host.Api + _bfsMove
	d.notifySendURL = d.c.Host.Message + _notifySend
	d.accountBlockURL = d.c.Host.Api + _accountBlock
	d.creditBlockInfoURL = d.c.Host.Api + _creditBlock
	d.delMoralURL = d.c.Host.Api + _delMoral
	d.purgeCacheURL = d.c.Host.Space + _purgeCache

	d.clearCacheTopPhotoURL = d.c.Host.Api + _clearCacheTopPhoto
	return
}

// Ping .
func (d *Dao) Ping(c context.Context) error {
	return d.DB.DB().PingContext(c)
}
