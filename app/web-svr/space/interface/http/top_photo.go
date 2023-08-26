package http

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/space/interface/api/v1"
	"go-gateway/app/web-svr/space/interface/model"
)

var (
	gRand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func topPhotoArc(c *bm.Context) {
	v := new(struct {
		Mid int64 `for:"mid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(spcSvc.TopPhotoArc(c, v.Mid))
}

func setTopPhotoArc(c *bm.Context) {
	v := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
		Aid int64 `form:"aid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	_, err := spcSvc.SetTopPhoto(c, &v1.SetTopPhotoReq{ID: v.Aid, Mid: v.Mid, Type: v1.TopPhotoType_ARCHIVE})
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, nil)
}

// topPhotoIndex .
func topPhotoIndex(c *bm.Context) {
	var (
		midStr interface{}
		ok     bool
		mid    int64
	)
	v := new(struct {
		Vmid int64 `form:"vmid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	if midStr, ok = c.Get("mid"); ok {
		mid = midStr.(int64)
	}
	c.JSON(spcSvc.MemWebTopPhotoIndex(c, mid, v.Vmid))
}

// topPhotoMallIndex .
func topPhotoMallIndex(c *bm.Context) {
	malls, err := spcSvc.GetPhotoMallList(c)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(malls, nil)
}

// uploadTopPhoto .
func uploadTopPhoto(c *bm.Context) {
	v := new(struct {
		Photo string `form:"topphoto" validate:"required"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	v.Photo = strings.TrimSpace(v.Photo)
	photo, err := spcSvc.WebUploadTopPhoto(c, mid, v.Photo)
	if err != nil {
		c.JSON(nil, err)
		return
	}
	res := map[string]interface{}{}
	res["img_url"] = "http://i" + strconv.Itoa(gRand.Intn(3)) + ".hdslb.com/" + photo.ImgPath
	c.JSON(res, nil)
}

// setTopPhoto .
func setTopPhoto(c *bm.Context) {
	var (
		err error
	)
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err = c.Bind(v); err != nil {
		return
	}
	midStr, _ := c.Get("mid")
	mid := midStr.(int64)
	c.JSON(nil, spcSvc.SetWebTopphoto(c, mid, v.ID))
}

// clearCacheTopPhoto
func clearCacheTopPhoto(c *bm.Context) {
	v := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	c.JSON(nil, spcSvc.ClearTopPhotoCache(c, v.Mid, model.UploadTopPhotoWeb))
}

func purgeCacheTopPhoto(c *bm.Context) {
	Param := new(model.PurgeCacheParam)
	if err := c.Bind(Param); err != nil {
		return
	}
	c.JSON(nil, spcSvc.PurgeCache(c, Param))
}
