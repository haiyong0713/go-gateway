package dao

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go-common/library/database/bfs"
	"go-common/library/log"
)

// _maxFileSize max file size is 10M
const (
	_maxFileSize = 10 * 1024 * 1024
	_svg         = "image/svg+xml"
)

var (
	bucket = "esport"
)

// BfsUpload .
func (d *Dao) BfsUpload(c context.Context, bs []byte, fileName string) (location string, err error) {
	var ftype string
	if len(bs) == 0 {
		return "", fmt.Errorf("nil value")
	}
	if strings.HasSuffix(fileName, "svg") {
		ftype = _svg
	} else {
		ftype = http.DetectContentType(bs)
		if ftype != "image/jpeg" && ftype != "image/jpg" && ftype != "image/png" && ftype != "image/webp" {
			log.Error("file type not allow file type(%s)", ftype)
			return "", fmt.Errorf("file type is error")
		}
	}
	if len(bs) > _maxFileSize {
		return "", fmt.Errorf("file is to large")
	}
	if location, err = d.bfsClient.Upload(c, &bfs.Request{
		Filename:    fileName,
		Bucket:      bucket,
		ContentType: ftype,
		File:        bs,
	}); err != nil {
		log.Error("bfs.BfsUpload error(%v)", err)
	}
	return
}
