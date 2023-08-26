package dynamic

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDao_SVideo(t *testing.T) {
	var (
		offset     = "338983154289229463"
		needOffset = 0
		uid        = int64(88895133)
	)
	Convey("SVideo", t, func() {
		res, err := d.SVideo(context.TODO(), offset, needOffset, uid)
		ress, _ := json.Marshal(res)
		fmt.Printf("%s", ress)
		So(err, ShouldBeNil)
	})
}

// {"items":[{"rid":880105765,"uid":27515255,"dynamic_id":338965300105985685},{"rid":400024925,"uid":27515255,"dynamic_id":338964277905866388}],"has_more":1,"offset":"337237168474832450"}.
