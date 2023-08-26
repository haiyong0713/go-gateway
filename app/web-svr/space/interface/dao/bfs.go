package dao

import (
	"context"
	"fmt"
	"image"
	"io/ioutil"
	"net/http"
	"time"

	// image decode
	_ "image/jpeg"
	_ "image/png"

	"go-common/library/database/bfs"

	"github.com/pkg/errors"
)

func (d *Dao) DecodeImageSize(c context.Context, imageURL string) (w, h int64, name string, err error) {
	req, err := http.NewRequest(http.MethodGet, imageURL, nil)
	if err != nil {
		return 0, 0, "", errors.Wrapf(err, "DecodeImageSize http.NewRequest(%s)", imageURL)
	}
	ctx, cancel := context.WithTimeout(c, time.Duration(d.c.Bfs.ReadTimeout))
	defer cancel()
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, 0, "", errors.Wrapf(err, "DecodeImageSize httpClient.Do(%s)", imageURL)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return 0, 0, "", errors.New(fmt.Sprintf("DecodeImageSize url(%s) resp.StatusCode(%v)", imageURL, resp.StatusCode))
	}
	defer resp.Body.Close()
	res, name, err := image.Decode(resp.Body)
	if err != nil {
		return 0, 0, "", errors.New(fmt.Sprintf("DecodeImageSize url(%s) image.Decode(%v)", imageURL, err))
	}
	return int64(res.Bounds().Dx()), int64(res.Bounds().Dy()), name, nil
}

func (d *Dao) ReadURLContent(c context.Context, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "ReadURLContent http.NewRequest(%s)", url)
	}
	ctx, cancel := context.WithTimeout(c, time.Duration(d.c.Bfs.ReadTimeout))
	defer cancel()
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "ReadURLContent httpClient.Do(%s)", url)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, errors.New(fmt.Sprintf("ReadURLContent url(%s) resp.StatusCode(%v)", url, resp.StatusCode))
	}
	res, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, "ReadURLContent ioutil.ReadAll error:%v")
	}
	return res, nil
}

func (d *Dao) UploadBFS(ctx context.Context, content []byte, useBfsDir bool) (string, error) {
	req := &bfs.Request{
		Bucket:      d.c.Bfs.Bucket,
		ContentType: http.DetectContentType(content),
		File:        content,
	}
	if useBfsDir {
		req.Dir = d.c.Bfs.Dir
	}
	location, err := d.bfsClient.Upload(ctx, req)
	if err != nil {
		return "", err
	}
	return location, nil
}
