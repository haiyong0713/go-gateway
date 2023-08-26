package fawkes

import (
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-resource/interface/conf"
)

// Dao macross dao.
type Dao struct {
	// conf
	c *conf.Config
	// http client
	client *httpx.Client
	// url
	version            string
	upgrade            string
	pack               string
	filter             string
	patch              string
	channel            string
	flow               string
	hfUpgrade          string
	laser              string
	laserReport        string
	laserReport2       string
	laserReportSilence string
	laserCmdReport     string
	apkList            string
	tribeList          string
	tribeRelation      string
	testFlight         string
}

// New dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:                  c,
		client:             httpx.NewClient(c.HTTPClient),
		version:            c.Host.Fawkes + _version,
		upgrade:            c.Host.Fawkes + _upgrade,
		pack:               c.Host.Fawkes + _pack,
		filter:             c.Host.Fawkes + _filter,
		patch:              c.Host.Fawkes + _patch,
		channel:            c.Host.Fawkes + _channel,
		flow:               c.Host.Fawkes + _flow,
		hfUpgrade:          c.Host.Fawkes + _hfUpgrade,
		laser:              c.Host.Fawkes + _laser,
		laserReport:        c.Host.Fawkes + _laserReport,
		laserReport2:       c.Host.Fawkes + _laserReport2,
		laserReportSilence: c.Host.Fawkes + _laserReportSilence,
		laserCmdReport:     c.Host.Fawkes + _laserCmdReport,
		apkList:            c.Host.Fawkes + _apkList,
		tribeList:          c.Host.Fawkes + _tribeList,
		tribeRelation:      c.Host.Fawkes + _tribeRelation,
		testFlight:         c.Host.Fawkes + _testFlight,
	}
	return
}
