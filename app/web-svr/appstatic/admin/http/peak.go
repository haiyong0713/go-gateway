package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/time"
	peakModel "go-gateway/app/web-svr/appstatic/admin/model/peak"
)

const (
	_timeLimit    time.Time = 10000000000
	_timeTransfer time.Time = 1000
)

func addPeak(c *bm.Context) {
	var (
		err    error
		appVer []peakModel.AppVer
	)
	res := map[string]interface{}{}
	param := new(peakModel.AddPeakParam)
	if err = c.Bind(param); err != nil {
		return
	}
	uid, name := managerInfo(c)
	peak := &peakModel.Peak{
		FileName:     param.FileName,
		Priority:     param.Priority,
		Type:         param.Type,
		Url:          param.Url,
		Md5:          param.Md5,
		Size:         param.Size,
		ExpectDw:     param.ExpectDw,
		EffectTime:   timeStampTransfer(param.EffectTime),
		ExpireTime:   timeStampTransfer(param.ExpireTime),
		OnlineStatus: peakModel.NotOnline,
		IsDeleted:    peakModel.NotDeleted,
		Person:       name,
		Uid:          uid,
	}
	if err = json.Unmarshal([]byte(param.AppVer), &appVer); err != nil {
		res["message"] = "app_ver参数有误:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = peakSvc.AddPeak(peak, appVer); err != nil {
		res["message"] = "添加失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func indexPeak(c *bm.Context) {
	var (
		err      error
		peakList *peakModel.IndexPager
	)
	res := map[string]interface{}{}
	param := &peakModel.IndexParam{}
	if err = c.Bind(param); err != nil {
		return
	}
	param.EffectTime = timeStampTransfer(param.EffectTime)
	param.ExpireTime = timeStampTransfer(param.ExpireTime)
	if peakList, err = peakSvc.IndexPeak(param); err != nil {
		res["message"] = "查询失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(peakList, nil)
}

func updatePeak(c *bm.Context) {
	var (
		err    error
		appVer []peakModel.AppVer
	)
	res := map[string]interface{}{}
	param := new(peakModel.UpdatePeakParam)
	if err = c.Bind(param); err != nil {
		return
	}
	uid, name := managerInfo(c)
	peak := &peakModel.Peak{
		ID:           param.ID,
		Priority:     param.Priority,
		FileName:     param.FileName,
		Type:         param.Type,
		Url:          param.Url,
		Md5:          param.Md5,
		Size:         param.Size,
		ExpectDw:     param.ExpectDw,
		EffectTime:   timeStampTransfer(param.EffectTime),
		ExpireTime:   timeStampTransfer(param.ExpireTime),
		OnlineStatus: peakModel.NotOnline,
		IsDeleted:    peakModel.NotDeleted,
		Person:       name,
		Uid:          uid,
	}
	if err = json.Unmarshal([]byte(param.AppVer), &appVer); err != nil {
		res["message"] = "参数有误:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if err = peakSvc.UpdatePeak(peak, appVer); err != nil {
		res["message"] = "更新失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func publishPeak(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	param := &struct {
		ID           uint  `form:"id" validate:"required"`
		OnlineStatus uint8 `form:"online_status"`
	}{}
	if err = c.Bind(param); err != nil {
		return
	}
	if err = peakSvc.PublishPeak(param.ID, param.OnlineStatus); err != nil {
		res["message"] = "发布失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func deletePeak(c *bm.Context) {
	var (
		err error
	)
	res := map[string]interface{}{}
	param := &struct {
		ID           uint  `form:"id" validate:"required"`
		OnlineStatus uint8 `form:"online_status"`
	}{}
	if err = c.Bind(param); err != nil {
		return
	}
	if err = peakSvc.PublishPeak(param.ID, param.OnlineStatus); err != nil {
		res["message"] = "发布失败:" + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

func uploadPeak(c *bm.Context) {
	var (
		req = c.Request
		md5 string
		url string
	)

	//req.ParseMultipartForm(int64(10485760))
	file, _, err := req.FormFile("file")
	if err != nil {
		log.Error("c.Request().FormFile(\"file\") error(%v) | ", err)
		c.JSON(nil, err)
		return
	}
	bs, err := ioutil.ReadAll(file)
	file.Close()
	if err != nil {
		log.Error("ioutil.ReadAll(c.Request().Body) error(%v)", err)
		c.JSON("ioutil.ReadAll(file)："+err.Error(), ecode.RequestErr)
		return
	}
	if md5, err = peakSvc.FileMd5(bs); err != nil {
		log.Error("peakSvc.FileMd5 error(%v)", err)
		c.JSON("peakSvc.FileMd5："+err.Error(), ecode.RequestErr)
		return
	}
	ftype := http.DetectContentType(bs)
	if ftype == "application/octet-stream" {
		ftype = req.Form.Get("type")
		if ftype == "" {
			log.Error("clientUpload ftype is null")
			c.JSON("clientUpload ftype is nil："+err.Error(), ecode.RequestErr)
		}
	}
	if url, err = peakSvc.UploadPeak(c, ftype, bs); err != nil {
		log.Error("bfsSvc.ClientUpCover error(%v)", err)
		c.JSON("文件上传有问题："+err.Error(), ecode.RequestErr)
		return
	}
	data := map[string]interface{}{
		"url":  url,
		"md5":  md5,
		"size": len(bs),
	}
	c.JSON(data, nil)
}

func timeStampTransfer(current time.Time) time.Time {
	if current > _timeLimit {
		return current / _timeTransfer
	}
	return current
}
