package http

import (
	xecode "go-gateway/app/app-svr/hkt-note/ecode"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
)

func upload(c *bm.Context) {
	var (
		err  error
		file multipart.File
		body []byte
	)
	if file, _, err = c.Request.FormFile("file"); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	defer file.Close()
	if body, err = ioutil.ReadAll(file); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	fileType := http.DetectContentType(body)
	if fileType != "image/jpeg" && fileType != "image/jpg" && fileType != "image/png" {
		c.JSON(nil, xecode.ImageTypeError)
	}
	mid, ok := c.Get("mid")
	if !ok {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	c.JSON(imgSvr.BfsNoteUpload(c, mid.(int64), fileType, body))
}

func download(c *bm.Context) {
	req := new(struct {
		ImageId int64 `form:"image_id" validate:"required"`
	})
	err := c.Bind(req)
	if err != nil {
		return
	}
	mid, ok := c.Get("mid")
	if !ok {
		c.JSON(nil, ecode.NoLogin)
		return
	}
	data, fileType, err := imgSvr.BfsImage(c, mid.(int64), req.ImageId)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.Bytes(http.StatusOK, fileType, data)
}

func downloadPub(c *bm.Context) {
	req := new(struct {
		ImageId int64  `form:"image_id" validate:"required"`
		Mid     int64  `form:"mid" validate:"required"`
		Token   string `form:"token" validate:"required"`
	})
	err := c.Bind(req)
	if err != nil {
		return
	}
	if req.Token != publicToken {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	data, fileType, err := imgSvr.BfsImage(c, req.Mid, req.ImageId)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.Bytes(http.StatusOK, fileType, data)
}
