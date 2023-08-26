package dao

import (
	"bytes"
	"go-common/library/log"
	"io"
	"net/http"
)

func (d *Dao) Download(url string) (res *bytes.Buffer, cf func(), err error) {
	var resp *http.Response
	resp, err = http.Get(url)
	if err != nil {
		log.Error("archive-push-admin.Download.Get Error (%v)", err)
		return
	}
	res = new(bytes.Buffer)
	if _, err = io.Copy(res, resp.Body); err != nil {
		log.Error("archive-push-admin.Download.io.Copy Error (%v)", err)
		return
	}
	cf = func() {
		resp.Body.Close()
	}
	return
}
