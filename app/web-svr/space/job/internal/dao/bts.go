package dao

import (
	"context"

	"go-gateway/app/web-svr/space/interface/model"
)

//go:generate kratos tool btsgen
type _bts interface {
	// bts: -nullcache=&model.TopPhotoArc{Aid:-1} -check_null_code=$==nil||$.Aid==-1 -singleflight=true
	TopPhotoArc(c context.Context, mid int64) (*model.TopPhotoArc, error)
}
