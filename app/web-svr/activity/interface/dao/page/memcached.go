package page

import (
	"context"
	"fmt"
	model "go-gateway/app/web-svr/activity/interface/model/page"
)

//go:generate kratos tool mcgen

type _mc interface {
	// mc: -key=pageKey -struct_name=Dao
	CacheGetPageByID(c context.Context, id int64) (*model.ActPage, error)
	// mc: -key=pageKey -expire=d.pageExpire -encode=json -struct_name=Dao -check_null_code=$==nil||$.ID==0
	AddCacheGetPageByID(c context.Context, id int64, value *model.ActPage) error
}

func pageKey(id int64) string {
	return fmt.Sprintf("act_page_%d", id)
}
