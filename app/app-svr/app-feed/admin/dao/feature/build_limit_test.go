package feature

import (
	"go-gateway/app/app-svr/app-feed/admin/model/feature"
	"testing"

	. "github.com/glycerine/goconvey/convey"
)

func TestDao_CreateBuildLt(t *testing.T) {
	Convey("TestDao_CreateBuildLt", t, WithDao(func(d *Dao) {
		attrs := &feature.BuildLimit{
			TreeID:   treeID,
			Path:     path,
			KeyName:  keyName,
			Config:   "android,ios",
			Creator:  creator,
			Modifier: creator,
			State:    "on",
		}
		res, err := d.SaveBuildLt(c, attrs)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestDao_GetBuildLtByID(t *testing.T) {
	Convey("TestDao_GetBuildLtByID", t, WithDao(func(d *Dao) {
		res, err := d.GetBuildLtByID(c, buildLtID)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestDao_SearchBuildLt(t *testing.T) {
	Convey("TestDao_SearchBuildLt", t, WithDao(func(d *Dao) {
		req := &feature.BuildListReq{
			TreeID:  treeID,
			Path:    path,
			KeyName: keyName,
			Creator: "",
			Pn:      0,
			Ps:      0,
		}
		res, _, err := d.SearchBuildLt(c, req, false, false)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestDao_SaveBuildLt(t *testing.T) {
	Convey("TestDao_SaveBuildLt", t, WithDao(func(d *Dao) {
		attrs := &feature.BuildLimit{
			TreeID:   treeID,
			Path:     path,
			KeyName:  keyName,
			Config:   "",
			Creator:  creator,
			Modifier: creator,
		}
		res, err := d.SaveBuildLt(c, attrs)
		printDetail(res)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestDao_UpdateBuildLt(t *testing.T) {
	Convey("TestDao_SaveBuildLt", t, WithDao(func(d *Dao) {
		attrs := map[string]interface{}{
			"state": StateOn,
		}
		err := d.UpdateBuildLt(c, buildLtID, attrs)
		So(err, ShouldBeNil)
	}))
}
