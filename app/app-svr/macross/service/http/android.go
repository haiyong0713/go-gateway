package http

import (
	"bytes"
	"io"
	"path"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/macross/service/model/android"
)

func apklUpload(c *bm.Context) {
	var (
		bundleID, version, apkType string
		isUnzip                    bool
		filePaths                  []string
		err                        error
	)
	// params
	if bundleID = c.Request.FormValue("app_bundle_id"); bundleID == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if version = c.Request.FormValue("version"); version == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	switch c.Request.FormValue("type") {
	case android.TypeChannel:
		apkType = "channel"
	case android.TypePatch:
		apkType = "patch"
	}
	folder := path.Join(bundleID, version, apkType)
	isUnzip, _ = strconv.ParseBool(c.Request.FormValue("unzip"))
	// apk file
	f, h, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	defer f.Close()
	filename := h.Filename
	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, f); err != nil {
		c.JSON(nil, err)
		return
	}
	if isUnzip {
		if !strings.HasSuffix(filename, ".zip") {
			c.JSON(nil, ecode.RequestErr)
			return
		}
		if _, filePaths, err = svr.Unzip(c, f, int64(buf.Len()), folder); err != nil {
			log.Error("svr.Unzip error(%v)", err)
			c.JSON(nil, err)
			return
		}
	} else {
		var (
			folder   = path.Join(bundleID, version, apkType)
			filePath string
		)
		if filePath, err = svr.ApkUpload(c, buf, folder, filename); err != nil {
			c.JSON(nil, err)
			return
		}
		filePaths = append(filePaths, filePath)
	}
	c.JSON(filePaths, err)
}

// apklUploadCDN upload to CDN.
func apklUploadCDN(c *bm.Context) {
	var (
		params                                    = c.Request.Form
		bundleID, version, apkType, filename, url string
		err                                       error
	)
	// params
	if bundleID = params.Get("app_bundle_id"); bundleID == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if version = params.Get("version"); version == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	switch params.Get("type") {
	case android.TypeChannel:
		apkType = "channel"
	case android.TypePatch:
		apkType = "patch"
	}
	if filename = params.Get("filename"); filename == "" {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	folder := path.Join(bundleID, version, apkType)
	if url, err = svr.ApkPutOss(c, folder, filename); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(url, err)
}
