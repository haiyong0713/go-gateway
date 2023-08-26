package fawkes

import (
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-job/job/conf"
)

// Dao macross dao.
type Dao struct {
	// http client
	client *bm.Client
	// url
	laserAll           string
	laserReport        string
	broadcastPushAll   string
	laserReportSilence string
	laserAllSilence    string
}

// New dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:             bm.NewClient(c.HTTPClient),
		laserAll:           c.Host.Fawkes + _laserAll,
		laserReport:        c.Host.Fawkes + _laserReport,
		broadcastPushAll:   c.Host.APICo + _broadcastPushAll,
		laserReportSilence: c.Host.Fawkes + _laserReportSilence,
		laserAllSilence:    c.Host.Fawkes + _laserAllSilence,
	}
	return
}
