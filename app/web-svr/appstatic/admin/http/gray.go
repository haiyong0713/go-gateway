package http

import (
	"io/ioutil"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/appstatic/admin/model"
)

// cdnPreload .
func cdnPreload(c *bm.Context) {
	// var (
	// 	err error
	// )
	// req := &struct {
	// 	Url string `form:"url" validate:"required"`
	// }{}
	// if err := c.Bind(req); err != nil {
	// 	return
	// }
	// if err = apsSvc.BossCdnPreload(c, req.Url); err != nil {
	// 	res := map[string]interface{}{}
	// 	res["message"] = err.Error()
	// 	c.JSONMap(res, ecode.RequestErr)
	// 	return
	// }
	c.JSON(nil, nil)
}

// cdnPublishCheck .
func cdnPublishCheck(c *bm.Context) {
	// var (
	// 	err error
	// )
	// req := &struct {
	// 	ID int64 `form:"id" validate:"required"`
	// }{}
	// if err := c.Bind(req); err != nil {
	// 	return
	// }
	// if err = apsSvc.BossCdnPublishCheck(c, req.ID); err != nil {
	// 	res := map[string]interface{}{}
	// 	res["message"] = err.Error()
	// 	c.JSONMap(res, ecode.RequestErr)
	// 	return
	// }
	c.JSON(nil, nil)
}

// cdnStatus .
func cdnStatus(c *bm.Context) {
	// req := &struct {
	// 	Urls []string `form:"urls,split" validate:"required"`
	// }{}
	// if err := c.Bind(req); err != nil {
	// 	return
	// }
	// res, err := apsSvc.BossCdnStatus(c, req.Urls)
	// if err != nil {
	// 	res := map[string]interface{}{}
	// 	res["message"] = err.Error()
	// 	c.JSONMap(res, ecode.RequestErr)
	// 	return
	// }
	// c.JSON(res, nil)
	c.JSON(nil, nil)
}

// grayIndex .
func grayIndex(c *bm.Context) {
	var (
		err  error
		gray *model.ResourceGray
	)
	req := &struct {
		ID int64 `form:"id" json:"id" validate:"required"`
	}{}
	if err := c.Bind(req); err != nil {
		return
	}
	res := map[string]interface{}{}
	if gray, err = apsSvc.Gray(req.ID); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(gray, nil)
}

// addGray addGray
func addGray(c *bm.Context) {
	var (
		err error
	)
	req := &model.ResourceGray{}
	if err := c.Bind(req); err != nil {
		return
	}
	res := map[string]interface{}{}
	if err = apsSvc.AddGray(req); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// saveGray .
func saveGray(c *bm.Context) {
	var (
		err error
	)
	req := &model.ResourceGray{}
	if err := c.Bind(req); err != nil {
		return
	}
	res := map[string]interface{}{}
	if err = apsSvc.SaveGray(req); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(nil, nil)
}

// uploadGray .
func uploadGray(c *bm.Context) {
	var (
		req      = c.Request
		err      error
		file     *model.FileInfo
		contents []byte
	)
	res := map[string]interface{}{}
	fileInput, _, errFile := req.FormFile("file")
	if errFile != nil {
		res["message"] = errFile.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	defer fileInput.Close()
	if contents, err = ioutil.ReadAll(fileInput); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	if file, err = apsSvc.UploadGray(c, contents); err != nil {
		res["message"] = err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(file, nil)
}
