package http

import (
	"go-gateway/app/web-svr/appstatic/admin/model"

	bm "go-common/library/net/http/blademaster"
)

func dolbyWhiteList(c *bm.Context) {
	c.JSON(apsSvc.FetchDolbyWhiteList())
}

func addDolbyWhiteList(c *bm.Context) {
	params := &model.DolbyWhiteList{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(nil, apsSvc.AddDolbyWhiteList(params))
}

func saveDolbyWhiteList(c *bm.Context) {
	params := &model.DolbyWhiteList{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(nil, apsSvc.SaveDolbyWhiteList(params))
}

func deleteDolbyWhiteList(c *bm.Context) {
	params := struct {
		ID int64 `from:"id"`
	}{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(nil, apsSvc.DelDolbyWhiteList(params.ID))
}

func qnBlackList(c *bm.Context) {
	c.JSON(apsSvc.FetchQnBlackList())
}
func addQnBlackList(c *bm.Context) {
	params := &model.QnBlackList{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(nil, apsSvc.AddQnBlackList(params))
}

func saveQnBlackList(c *bm.Context) {
	params := &model.QnBlackList{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(nil, apsSvc.SaveQnBlackList(params))
}

func deleteQnBlackList(c *bm.Context) {
	params := struct {
		ID int64 `from:"id"`
	}{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(nil, apsSvc.DeleteQnBlackList(params.ID))
}

func limitFreeList(c *bm.Context) {
	c.JSON(apsSvc.LimitFreeList())
}

func addLimitFree(c *bm.Context) {
	params := &model.LimitFreeInfo{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(nil, apsSvc.AddLimitFree(params))
}

func editLimitFree(c *bm.Context) {
	params := &model.LimitFreeInfo{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(nil, apsSvc.EditLimitFree(params))
}

func deleteLimitFree(c *bm.Context) {
	params := &struct {
		ID int64 `form:"id"`
	}{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(nil, apsSvc.DeleteLimitFree(params.ID))
}
