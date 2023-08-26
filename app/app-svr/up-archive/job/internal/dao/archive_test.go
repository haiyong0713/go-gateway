package dao

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDao_RawArcPassed(t *testing.T) {
	var mid int64 = 15555180
	Convey("RawArcPassed", t, func() {
		data, err := d.RawArcPassed(ctx, mid)
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(data)
		Printf("%s", string(bs))
	})
}

func TestDao_RawArcs(t *testing.T) {
	var aids = []int64{10097272, 10097274}
	Convey("RawArcs", t, func() {
		data, err := d.RawArcs(ctx, aids)
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(data)
		Printf("%s", string(bs))
	})
}

func TestDao_RawArc(t *testing.T) {
	var aid = int64(920080390)
	Convey("RawArc", t, func() {
		data, err := d.RawArc(ctx, aid)
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(data)
		Printf("%s", string(bs))
	})
}
