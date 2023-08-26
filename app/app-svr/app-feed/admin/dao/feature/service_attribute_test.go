package feature

import (
	"testing"

	"go-gateway/app/app-svr/app-feed/admin/model/feature"

	. "github.com/glycerine/goconvey/convey"
)

func TestDao_GetSvrAttrByTreeID(t *testing.T) {
	Convey("TestDao_GetSvrAttrByTreeID", t, WithDao(func(d *Dao) {
		res, err := d.GetSvrAttrByTreeID(c, treeID)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestDao_SaveSvrAttr(t *testing.T) {
	Convey("TestDao_SaveSvrAttr", t, WithDao(func(d *Dao) {
		attrs := &feature.ServiceAttribute{
			ID:          0,
			TreeID:      treeID,
			MobiApps:    "android,ios",
			Modifier:    creator,
			ModifierUID: creatorUID,
		}
		err := d.SaveSvrAttr(c, attrs)
		So(err, ShouldBeNil)
	}))
}

func TestDao_GetSvrAttrPlats(t *testing.T) {
	Convey("TestDao_GetSvrAttrPlats", t, WithDao(func(d *Dao) {
		res, err := d.GetSvrAttrPlats(c, treeID)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}
