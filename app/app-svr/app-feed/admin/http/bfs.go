package http

import (
	"encoding/base64"
	"io/ioutil"
	"net/http"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
)

func clientUpload(c *bm.Context) {
	var (
		req = c.Request
		md5 string
		url string
	)
	//nolint:errcheck
	req.ParseMultipartForm(int64(bfsSvc.BfsMaxSize))
	file, _, err := req.FormFile("file")
	if err != nil {
		log.Error("c.Request().FormFile(\"file\") error(%v) | ", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	bs, err := ioutil.ReadAll(file)
	file.Close()
	if err != nil {
		log.Error("ioutil.ReadAll(c.Request().Body) error(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if md5, err = bfsSvc.FileMd5(bs); err != nil {
		log.Error("bfsSvc.FileMd5 error(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	ftype := http.DetectContentType(bs)
	if ftype == "application/octet-stream" {
		ftype = req.Form.Get("type")
		if ftype == "" {
			log.Error("clientUpload ftype is null")
			c.JSON(nil, ecode.RequestErr)
		}
	}
	if url, err = bfsSvc.ClientUpCover(c, ftype, bs); err != nil {
		log.Error("bfsSvc.ClientUpCover error(%v)", err)
		c.JSON("文件上传错误："+err.Error(), ecode.RequestErr)
		return
	}
	data := map[string]interface{}{
		"url":  url,
		"md5":  md5,
		"size": len(bs),
	}
	c.JSON(data, nil)
}

func specialUpload(c *bm.Context) {
	var (
		req      = c.Request
		md5      string
		url      string
		fileType string
	)
	fileType = req.Form.Get("content_type")
	// 文件大小不超过300k
	//nolint:errcheck
	req.ParseMultipartForm(int64(300 * 1024))
	file, _, err := req.FormFile("file")
	if err != nil {
		log.Error("c.Request().FormFile(\"file\") error(%v) | ", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	bs, err := ioutil.ReadAll(file)
	file.Close()
	if err != nil {
		log.Error("ioutil.ReadAll(c.Request().Body) error(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if md5, err = bfsSvc.FileMd5(bs); err != nil {
		log.Error("bfsSvc.FileMd5 error(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	// json 格式无法从函数中获取 需要特殊处理
	if fileType == "" {
		fileType = http.DetectContentType(bs)
	}
	if fileType != "application/json" && fileType != "image/png" && fileType != "application/zlib" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if url, err = bfsSvc.ClientUpCover(c, fileType, bs); err != nil {
		log.Error("bfsSvc.ClientUpCover error(%v)", err)
		c.JSON("文件上传错误："+err.Error(), ecode.RequestErr)
		return
	}
	data := struct {
		URL  string `json:"url"`
		Md5  string `json:"md5"`
		Size int    `json:"size"`
	}{
		URL:  url,
		Md5:  md5,
		Size: len(bs),
	}
	c.JSON(data, nil)
}

// uploadBase64 .
func uploadBase64(c *bm.Context) {
	var (
		md5 string
		url string
		err error
		bs  []byte
	)
	res := map[string]interface{}{}
	type File struct {
		File string `form:"file" validate:"required"`
	}
	file := &File{}
	if err = c.Bind(file); err != nil {
		return
	}
	if bs, err = base64.StdEncoding.DecodeString(file.File); err != nil {
		res["message"] = "base64解析失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if md5, err = bfsSvc.FileMd5(bs); err != nil {
		res["message"] = "FileMd5解析失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	ftype := http.DetectContentType(bs)
	if url, err = bfsSvc.ClientUpCover(c, ftype, bs); err != nil {
		res["message"] = "文件上传错误 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	data := map[string]interface{}{
		"url":  url,
		"md5":  md5,
		"size": len(bs),
	}
	c.JSON(data, nil)
}

// validatGif .
func validatGif(c *bm.Context) {
	var (
		req = c.Request
	)
	frame := req.Form.Get("frame")
	if frame == "" {
		res := map[string]interface{}{}
		res["message"] = "frame参数不能为空"
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	file, _, err := req.FormFile("file")
	if err != nil {
		res := map[string]interface{}{}
		res["message"] = "文件上传错误 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	defer file.Close()
	bs, err := ioutil.ReadAll(file)
	if err != nil {
		res := map[string]interface{}{}
		res["message"] = "文件上传错误 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	err = bfsSvc.ValidGif(c, frame, bs)
	if err != nil {
		res := map[string]interface{}{}
		res["message"] = "文件验证失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}
