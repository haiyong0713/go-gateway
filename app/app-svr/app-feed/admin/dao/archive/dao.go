package archive

import (
	"git.bilibili.co/bapis/bapis-go/archive/service"
	"go-gateway/app/app-svr/app-feed/admin/conf"

	flowCtrlGrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"

	bm "go-common/library/net/http/blademaster"
)

// Dao is archive dao.
type Dao struct {
	// rpc
	arcClient         api.ArchiveClient
	flowControlClient flowCtrlGrpc.FlowControlClient
	client            *bm.Client
	archiveBanURL     string
	archiveAuditURL   string
	userFeed          *conf.UserFeed
	feedFlowCtrlConf  *conf.FlowCtrl
}

// New account dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:           bm.NewClient(c.HTTPClient.Read),
		archiveBanURL:    c.Host.Archive + _arcBanURL,
		archiveAuditURL:  c.Host.Archive + _arcAuditURL,
		userFeed:         c.UserFeed,
		feedFlowCtrlConf: c.FeedConfig.FlowCtrl,
	}
	var err error
	if d.flowControlClient, err = flowCtrlGrpc.NewClient(c.FlowCtrlGRPCClient); err != nil {
		panic(err)
	}
	if d.arcClient, err = api.NewClient(c.ArcClient); err != nil {
		panic(err)
	}
	return
}
