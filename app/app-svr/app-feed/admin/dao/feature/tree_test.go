package feature

import (
	"testing"

	. "github.com/glycerine/goconvey/convey"
)

func TestDao_FetchRoleTree(t *testing.T) {
	Convey("TestDao_CreateBuildLt", t, WithDao(func(d *Dao) {
		cookie := ""
		res, err := d.FetchRoleTree(c, cookie)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}
