package dao

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDao_StaffAids(t *testing.T) {
	var mid int64 = 3333
	Convey("RawStaffAids", t, func() {
		data, err := d.RawStaffAids(ctx, mid)
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(data)
		Printf("%s", string(bs))
	})
}
