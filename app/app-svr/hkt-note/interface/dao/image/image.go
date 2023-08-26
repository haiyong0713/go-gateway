package image

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"go-common/library/database/bfs"
	"go-common/library/ecode"
	"go-common/library/log"
	xecode "go-gateway/app/app-svr/hkt-note/ecode"
	"go-gateway/app/app-svr/hkt-note/interface/model/note"
	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"

	"github.com/pkg/errors"
)

const (
	_wmText = "%d 笔记专用"
)

func (d *Dao) ImgAdd(c context.Context, location string, mid int64) (int64, error) {
	req := &notegrpc.ImgAddReq{
		Mid:      mid,
		Location: location,
	}
	reply, err := d.noteClient.ImgAdd(c, req)
	if err != nil {
		return 0, errors.Wrapf(err, "ImgAdd req(%+v)", req)
	}
	if reply == nil {
		log.Error("NoteError grpc ImgAdd req(%+v) reply empty", req)
		return 0, ecode.NothingFound
	}
	return reply.ImageId, nil
}

func (d *Dao) Img(c context.Context, mid, imageId int64) (string, error) {
	req := &notegrpc.ImgReq{
		Mid:     mid,
		ImageId: imageId,
	}
	reply, err := d.noteClient.Img(c, req)
	if err != nil {
		return "", errors.Wrapf(err, "Img req(%+v)", req)
	}
	if reply == nil {
		log.Error("NoteError grpc Img req(%+v) reply empty", req)
		return "", ecode.NothingFound
	}
	return reply.Location, nil
}

// Upload bfs img upload.
func (d *Dao) NoteImgUpload(c context.Context, mid int64, fileType string, file []byte) (*note.ImageRes, error) {
	req := &bfs.Request{
		Bucket:      "note",
		ContentType: fileType,
		File:        file,
		WMKey:       "note",
		WMText:      fmt.Sprintf(_wmText, mid),
	}
	location, err := d.bfsClient.Upload(c, req)
	if err != nil {
		err = errors.Wrapf(err, "Upload req(%+v)", req)
		return nil, err
	}
	index := strings.Index(location, "/bfs/")
	if index < 0 {
		err = errors.Wrapf(xecode.ImageURLInvalid, "Upload req(%+v) location(%s) no /bfs err", req, location)
		return nil, err
	}
	location = location[index:]

	return &note.ImageRes{Location: location}, nil
}

func (d *Dao) NoteImgDownload(c context.Context, mid int64, location string) (res []byte, fileType string, err error) {
	if _, fileType, err = fileNameAndType(location); err != nil {
		return nil, "", err
	}
	imageURL := fmt.Sprintf("%s%s", d.c.Hosts.BfsHost, location)
	var req *http.Request
	if req, err = http.NewRequest("GET", imageURL, nil); err != nil {
		err = errors.Wrapf(err, "Download mid(%d) location(%s)", mid, location)
		return nil, "", err
	}
	req.Header.Set("Content-Type", fileType)
	resp, err := d.httpClient.Do(req)
	if err != nil {
		err = errors.Wrapf(err, "Download mid(%d) location(%s)", mid, location)
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = errors.Wrapf(ecode.Int(resp.StatusCode), "Download mid(%d) location(%s)", mid, location)
		return nil, "", err
	}
	if res, err = ioutil.ReadAll(resp.Body); err != nil {
		err = errors.Wrapf(err, "Download mid(%d) location(%s)", mid, location)
		return nil, "", err
	}
	return res, fileType, nil
}

func fileNameAndType(location string) (fileName, fileType string, err error) {
	names := strings.Split(location, "/note/")
	if len(names) != 2 { //nolint:gomnd
		err = xecode.ImageURLInvalid
		return
	}
	if fileName = names[1]; fileName == "" {
		err = xecode.ImageURLInvalid
		return
	}
	fileType = strings.ToLower(strings.TrimPrefix(path.Ext(fileName), "."))
	switch fileType {
	case "png":
		fileType = note.FileTypePNG
	case "jpg":
		fileType = note.FileTypeJPG
	case "jpeg":
		fileType = note.FileTypeJPEG
	default:
		err = xecode.ImageTypeError
	}
	return
}
