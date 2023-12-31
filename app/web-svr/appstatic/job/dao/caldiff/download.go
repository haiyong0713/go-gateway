package caldiff

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"go-common/library/log"
)

const (
	errFormat = "Func:[%s] - Step:[%s] - Error:[%v]"
)

// DownloadFile downloads one file from url to local
func (d *Dao) DownloadFile(ctx context.Context, api string, fileName string) (bts int64, err error) {
	var (
		res []byte
		f   *os.File
		req *http.Request
	)
	if req, err = d.client.NewRequest(http.MethodGet, api, "", url.Values{}); err != nil {
		log.Error(errFormat, "DownloadFile", fmt.Sprintf("NewRequest file name (%s) api(%s)", fileName, api), err)
		return
	}
	if res, err = d.client.Raw(ctx, req); err != nil {
		log.Error(errFormat, "DownloadFile", fmt.Sprintf("Raw file name (%s) api(%s)", fileName, api), err)
		return
	}
	if f, err = os.Create(fileName); err != nil {
		log.Error(errFormat, "DownloadFile", fmt.Sprintf("Create (%s)", fileName), err)
		return
	}
	if bts, err = io.Copy(f, bytes.NewReader(res)); err != nil {
		log.Error(errFormat, "DownloadFile", fmt.Sprintf("io.Copy(%s)", fileName), err)
	}
	return
}
