package archive

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestRedisMSetWithExp(t *testing.T) {
	var (
		c     = context.TODO()
		kvMap = make(map[string][]byte)
		exp   = int64(200)
	)

	for i := 0; i < 3; i++ {
		value, _ := json.Marshal("test" + strconv.Itoa(i))
		kvMap[strconv.Itoa(i)] = value
	}

	convey.Convey("redisMSetWithExp", t, func(ctx convey.C) {
		err := d.redisMSetWithExp(c, kvMap, exp)
		ctx.Convey("Then err should be nil.st should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
