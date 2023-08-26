package delay

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDao_CompressPkg(t *testing.T) {
	var (
		ctx  = context.TODO()
		api  = "testDemo"
		name = "testDemo2.tar.gz"
	)
	Convey("test", t, func() {
		err := testDao.Compress("/Users/carlos/Downloads/api-gateway", fmt.Sprintf("/tmp/%s", name))
		So(err, ShouldBeNil)
		res, err := testDao.ReadTarPackage(fmt.Sprintf("/tmp/%s", name))
		So(err, ShouldBeNil)
		url, err := testDao.Upload(ctx, "test", name, res)
		So(err, ShouldBeNil)
		err = testDao.AddRowDB(ctx, api, "abc")
		So(err, ShouldBeNil)
		err = testDao.UpdateLog(ctx, 1, "")
		So(err, ShouldBeNil)
		err = testDao.UpdateBoss(ctx, 1, url)
		So(err, ShouldBeNil)
	})
}
