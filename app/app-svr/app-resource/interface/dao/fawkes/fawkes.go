package fawkes

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-resource/interface/model/fawkes"
	fkmdl "go-gateway/app/app-svr/fawkes/service/model"
	fkappmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	bizapkmdl "go-gateway/app/app-svr/fawkes/service/model/bizapk"
	fkcdmdl "go-gateway/app/app-svr/fawkes/service/model/cd"
	tribemdl "go-gateway/app/app-svr/fawkes/service/model/tribe"

	"github.com/pkg/errors"
)

const (

	// 	_opt = "1007"

	_version            = "/x/admin/fawkes/business/version/all"
	_upgrade            = "/x/admin/fawkes/business/upgrade/all"
	_pack               = "/x/admin/fawkes/business/pack/all"
	_filter             = "/x/admin/fawkes/business/filter/all"
	_patch              = "/x/admin/fawkes/business/patch/all2"
	_channel            = "/x/admin/fawkes/business/channel/all"
	_flow               = "/x/admin/fawkes/business/flow/all"
	_hfUpgrade          = "/x/admin/fawkes/business/hotfix/all"
	_laser              = "/x/admin/fawkes/business/laser"
	_laserReport        = "/x/admin/fawkes/business/laser/report"
	_laserReport2       = "/x/admin/fawkes/business/laser/report2"
	_laserReportSilence = "/x/admin/fawkes/business/laser/report/silence"
	_laserCmdReport     = "/x/admin/fawkes/business/laser/cmd/report"
	_apkList            = "/x/admin/fawkes/business/bizapk/list/all"
	_tribeList          = "/x/admin/fawkes/business/tribe/list/all"
	_tribeRelation      = "/x/admin/fawkes/business/tribe/relation/all"

	_testFlight = "/x/admin/fawkes/business/testflight"
)

// Versions get all version.
func (d *Dao) Versions(c context.Context) (res map[string]map[int64]*fkmdl.Version, err error) {
	var re struct {
		Code int                                 `json:"code"`
		Data map[string]map[int64]*fkmdl.Version `json:"data"`
	}
	if err = d.client.Get(c, d.version, "", nil, &re); err != nil {
		return
	}
	if re.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(re.Code), d.version)
		return
	}
	res = re.Data
	return
}

// UpgradConfig get all upgrad config.
func (d *Dao) UpgradConfig(c context.Context) (res map[string]map[int64]*fkcdmdl.UpgradConfig, err error) {
	var re struct {
		Code int                                        `json:"code"`
		Data map[string]map[int64]*fkcdmdl.UpgradConfig `json:"data"`
	}
	if err = d.client.Get(c, d.upgrade, "", nil, &re); err != nil {
		return
	}
	if re.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(re.Code), d.version)
		return
	}
	res = re.Data
	return
}

// Packs get all pack.
func (d *Dao) Packs(c context.Context) (res map[string]map[int64][]*fkcdmdl.Pack, err error) {
	var re struct {
		Code int                                  `json:"code"`
		Data map[string]map[int64][]*fkcdmdl.Pack `json:"data"`
	}
	if err = d.client.Get(c, d.pack, "", nil, &re); err != nil {
		return
	}
	if re.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(re.Code), d.version)
		return
	}
	res = re.Data
	return
}

// Patch get patch.
func (d *Dao) Patch(c context.Context) (res map[string]map[string]*fkcdmdl.Patch, err error) {
	var re struct {
		Code int                                  `json:"code"`
		Data map[string]map[string]*fkcdmdl.Patch `json:"data"`
	}
	if err = d.client.Get(c, d.patch, "", nil, &re); err != nil {
		return
	}
	if re.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(re.Code), d.version)
		return
	}
	res = re.Data
	return
}

// FilterConfig get all filter config.
func (d *Dao) FilterConfig(c context.Context) (res map[string]map[int64]*fkcdmdl.FilterConfig, err error) {
	var re struct {
		Code int                                        `json:"code"`
		Data map[string]map[int64]*fkcdmdl.FilterConfig `json:"data"`
	}
	if err = d.client.Get(c, d.filter, "", nil, &re); err != nil {
		return
	}
	if re.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(re.Code), d.version)
		return
	}
	res = re.Data
	return
}

// AppChannel get all app channel.
func (d *Dao) AppChannel(c context.Context) (res map[string]map[int64]*fkappmdl.Channel, err error) {
	var re struct {
		Code int                                    `json:"code"`
		Data map[string]map[int64]*fkappmdl.Channel `json:"data"`
	}
	if err = d.client.Get(c, d.channel, "", nil, &re); err != nil {
		return
	}
	if re.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(re.Code), d.version)
		return
	}
	res = re.Data
	return
}

// FlowConfig get all flow config.
func (d *Dao) FlowConfig(c context.Context) (res map[string]map[int64]*fkcdmdl.FlowConfig, err error) {
	var re struct {
		Code int                                      `json:"code"`
		Data map[string]map[int64]*fkcdmdl.FlowConfig `json:"data"`
	}
	if err = d.client.Get(c, d.flow, "", nil, &re); err != nil {
		return
	}
	if re.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(re.Code), d.version)
		return
	}
	res = re.Data
	return
}

// HfUpgrade get all hotfix upgrade information
func (d *Dao) HfUpgrade(c context.Context) (res map[string]map[int64][]*fkappmdl.HfUpgrade, err error) {
	var re struct {
		Code int                                        `json:"code"`
		Data map[string]map[int64][]*fkappmdl.HfUpgrade `json:"data"`
	}
	if err = d.client.Get(c, d.hfUpgrade, "", nil, &re); err != nil {
		return
	}
	if re.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(re.Code), d.version)
		return
	}
	res = re.Data
	return
}

// LaserReport report laser status and url.
func (d *Dao) LaserReport(c context.Context, taskID int64, status int, logURL, errMsg, mobiApp, build, md5, rawUposUri string) (err error) {
	var ip = metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("task_id", strconv.FormatInt(taskID, 10))
	params.Set("status", strconv.Itoa(status))
	params.Set("url", logURL)
	params.Set("error_msg", errMsg)
	params.Set("recall_mobi_app", mobiApp)
	params.Set("build", build)
	params.Set("md5", md5)
	params.Set("raw_upos_uri", rawUposUri)
	var re struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
	}
	if err = d.client.Post(c, d.laserReport, ip, params, &re); err != nil {
		return
	}
	if re.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(re.Code), d.laserReport)
		log.Error("%v", err)
	}
	err = nil
	return
}

// Laser get laser.
func (d *Dao) Laser(c context.Context, taskID int64) (res *fkappmdl.Laser, err error) {
	params := url.Values{}
	params.Set("task_id", strconv.FormatInt(taskID, 10))
	var re struct {
		Code int             `json:"code"`
		Data *fkappmdl.Laser `json:"data"`
	}
	if err = d.client.Get(c, d.laser, "", params, &re); err != nil {
		return
	}
	if re.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(re.Code), d.version)
		return
	}
	res = re.Data
	return
}

// LaserReport2 添加主动触发的laser任务.
func (d *Dao) LaserReport2(c context.Context, appkey, buvid, uri, errMsg, mobiApp, build, md5, rawUposUri string, mid, taskID int64, status int) (fkTaskID int64, err error) {
	var ip = metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("app_key", appkey)
	params.Set("buvid", buvid)
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("url", uri)
	params.Set("error_msg", errMsg)
	params.Set("recall_mobi_app", mobiApp)
	params.Set("build", build)
	params.Set("status", strconv.Itoa(status))
	params.Set("md5", md5)
	params.Set("raw_upos_uri", rawUposUri)
	if taskID > 0 {
		params.Set("task_id", strconv.FormatInt(taskID, 10))
	}
	var re struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
		Data struct {
			TaskID int64 `json:"task_id"`
		}
	}
	if err = d.client.Post(c, d.laserReport2, ip, params, &re); err != nil {
		return
	}
	if re.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(re.Code), d.laserReport2)
		log.Error("%v", err)
		return
	}
	return re.Data.TaskID, nil
}

// LaserReportSilence report laser status and url.
func (d *Dao) LaserReportSilence(c context.Context, taskID int64, status int, logURL, errMsg, mobiApp, build string) (err error) {
	var ip = metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("task_id", strconv.FormatInt(taskID, 10))
	params.Set("status", strconv.Itoa(status))
	params.Set("url", logURL)
	params.Set("error_msg", errMsg)
	params.Set("recall_mobi_app", mobiApp)
	params.Set("build", build)
	var re struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
	}
	if err = d.client.Post(c, d.laserReportSilence, ip, params, &re); err != nil {
		return
	}
	if re.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(re.Code), d.laserReportSilence)
		log.Error("%v", err)
	}
	err = nil
	return
}

func (d *Dao) ApkList(c context.Context) (res map[int64]map[string]map[string][]*bizapkmdl.Apk, err error) {
	var re struct {
		Code int                                              `json:"code"`
		Data map[int64]map[string]map[string][]*bizapkmdl.Apk `json:"data"`
	}
	if err = d.client.Get(c, d.apkList, "", nil, &re); err != nil {
		return
	}
	if re.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(re.Code), d.version)
		return
	}
	res = re.Data
	return
}

func (d *Dao) TribeList(c context.Context) (res map[string]map[int64]map[string]map[string][]*tribemdl.TribeApk, err error) {
	var re struct {
		Code int                                                             `json:"code"`
		Data map[string]map[int64]map[string]map[string][]*tribemdl.TribeApk `json:"data"`
	}
	if err = d.client.Get(c, d.tribeList, "", nil, &re); err != nil {
		return
	}
	if re.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(re.Code), d.version)
		return
	}
	res = re.Data
	return
}

func (d *Dao) TribeRelation(c context.Context) (res map[int64]int64, err error) {
	var re struct {
		Code int             `json:"code"`
		Data map[int64]int64 `json:"data"`
	}
	if err = d.client.Get(c, d.tribeRelation, "", nil, &re); err != nil {
		return
	}
	if re.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(re.Code), d.version)
		return
	}
	res = re.Data
	return
}

func (d *Dao) TestFlight(c context.Context, env string) (map[string]*fawkes.TestFlight, error) {
	var resTmp struct {
		Code int                  `json:"code"`
		Data []*fawkes.TestFlight `json:"data"`
	}
	params := url.Values{}
	params.Set("env", env)
	if err := d.client.Get(c, d.testFlight, "", params, &resTmp); err != nil {
		err = errors.Wrap(ecode.Int(resTmp.Code), d.testFlight)
		return nil, err
	}
	var res = make(map[string]*fawkes.TestFlight)
	for _, tf := range resTmp.Data {
		if tf == nil {
			continue
		}
		if tf.MobiApp == "" {
			continue
		}
		res[tf.MobiApp] = tf
	}
	return res, nil
}

// LaserCmdReport report laser cmd status and url.
func (d *Dao) LaserCmdReport(c context.Context, taskID int64, status int, mobiApp, build, fileUrl, errorMsg, result, md5, rawUposUri string) error {
	var ip = metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("task_id", strconv.FormatInt(taskID, 10))
	params.Set("status", strconv.Itoa(status))
	params.Set("url", fileUrl)
	params.Set("error_msg", errorMsg)
	params.Set("mobi_app", mobiApp)
	params.Set("build", build)
	params.Set("result", result)
	params.Set("md5", md5)
	params.Set("raw_upos_uri", rawUposUri)
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
	}
	if err := d.client.Post(c, d.laserCmdReport, ip, params, &res); err != nil {
		return err
	}
	if res.Code != ecode.OK.Code() {
		log.Error("%+v", errors.Wrap(ecode.Int(res.Code), d.laserCmdReport+"?"+params.Encode()))
	}
	return nil
}
