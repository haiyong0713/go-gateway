package http

import (
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/archive-push/admin/internal/model"
)

// vendorsAvailable 所有可用厂商
func vendorsAvailable(ctx *bm.Context) {
	ctx.JSON(model.DefaultVendors, nil)
}

// vendorsUserBindable 可进行用户绑定的厂商
func vendorsUserBindable(ctx *bm.Context) {
	res := make([]model.ArchivePushVendor, 0)
	for _, vendor := range model.DefaultVendors {
		if vendor.UserBindable {
			res = append(res, vendor)
		}
	}
	ctx.JSON(res, nil)
}
