package kfc

import (
	"context"
	"fmt"

	"go-gateway/app/web-svr/activity/interface/model/kfc"
)

func kfcKey(id int64) string {
	return fmt.Sprintf("b_kfc_c_%d", id)
}

func kfcCodeKey(code string) string {
	return fmt.Sprintf("b_kfc_code_%s", code)
}

//go:generate kratos tool btsgen
type _bts interface {
	// bts: -sync=true
	KfcCoupon(c context.Context, id int64) (*kfc.BnjKfcCoupon, error)
}

//go:generate kratos tool mcgen
type _mc interface {
	// mc: -key=kfcKey
	CacheKfcCoupon(c context.Context, id int64) (*kfc.BnjKfcCoupon, error)
	// mc: -key=kfcKey -expire=d.mcKfcExpire -encode=pb
	AddCacheKfcCoupon(c context.Context, id int64, val *kfc.BnjKfcCoupon) error
	// mc: -key=kfcKey
	DelCacheKfcCoupon(c context.Context, id int64) error
	// mc: -key=kfcCodeKey
	CacheKfcCode(c context.Context, code string) (int64, error)
	// mc: -key=kfcCodeKey -expire=d.mcKfcCodeExpire -encode=raw
	AddCacheKfcCode(c context.Context, code string, val int64) error
}
