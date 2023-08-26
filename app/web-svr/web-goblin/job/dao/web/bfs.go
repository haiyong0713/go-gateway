package web

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"go-common/library/database/bfs"

	"github.com/pkg/errors"
)

func (d *Dao) UploadBFS(ctx context.Context, filename string, content []byte) (string, error) {
	location, err := d.bfsClient.Upload(ctx, &bfs.Request{
		Bucket:      d.c.BFS.Bucket,
		Dir:         d.c.BFS.Dir,
		ContentType: "application/json",
		Filename:    filename,
		File:        content,
	})
	if err != nil {
		return "", err
	}
	return location, nil
}

func (d *Dao) ReadURLContent(c context.Context, outURL string) ([]byte, error) {
	var (
		req    *http.Request
		resp   *http.Response
		cancel func()
	)
	req, err := http.NewRequest("GET", outURL, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "ReadURLContent http.NewRequest(%s)", outURL)
	}
	c, cancel = context.WithTimeout(c, time.Duration(d.c.Rule.ReadTimeout))
	defer cancel()
	req = req.WithContext(c)
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return nil, errors.Wrapf(err, "ReadURLContent httpClient.Do(%s)", outURL)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, errors.New(fmt.Sprintf("ReadURLContent url(%s) resp.StatusCode(%v)", outURL, resp.StatusCode))
	}
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "ReadURLContent ioutil.ReadAll error:%v")
	}
	defer resp.Body.Close()
	return res, nil
}
