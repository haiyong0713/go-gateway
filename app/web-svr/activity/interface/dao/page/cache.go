package page

import (
	"context"
	model "go-gateway/app/web-svr/activity/interface/model/page"
)

//go:generate kratos tool btsgen
type _bts interface {
	// bts: -struct_name=Dao -nullcache=&model.ActPage{ID:0} -check_null_code=$==nil||$.ID==0 -sync=true
	GetPageByID(c context.Context, id int64) (*model.ActPage, error)
}
