package dao

import (
	"github.com/glycerine/goconvey/convey"
	"go-gateway/app/app-svr/archive-push/admin/internal/model"
	"testing"
)

func Test_SearchAuditLog(t *testing.T) {
	convey.Convey("SearchAuditLog", t, func() {
		queryParams := &model.AuditLogSearchParams{Business: model.BusinessIDAuthor, Order: "ctime", Type: 2}
		res, err := testD.SearchAuditLog(queryParams)
		convey.ShouldBeNil(err)
		convey.ShouldNotBeNil(res)
	})
}
