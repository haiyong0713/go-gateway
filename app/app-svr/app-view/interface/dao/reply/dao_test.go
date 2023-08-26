package reply

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"
	"testing"

	"go-gateway/app/app-svr/app-view/interface/conf"
	"go-gateway/app/app-svr/app-view/interface/model"

	api "git.bilibili.co/bapis/bapis-go/community/interface/reply"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func init() {
	dir, _ := filepath.Abs("../../cmd/app-view-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	d = New(conf.Conf)
}

func TestDao_GetReplyListPreface(t *testing.T) {
	Convey("GetReplyListPreface", t, func() {
		_, err := d.GetReplyListPreface(context.Background(), 1, 2, "")
		So(err, ShouldBeNil)
	})
}

func TestDao_GetReplyListsPreface(t *testing.T) {
	Convey("GetReplyListsPreface", t, func() {
		var sub []*api.ReplyListPrefaceReq
		sub = append(sub, &api.ReplyListPrefaceReq{Mid: 1, Oid: 2, Type: model.ReplyTypeAv, Buvid: ""})
		res, err := d.GetReplyListsPreface(context.Background(), &api.ReplyListsPrefaceReq{Subjects: sub})
		ress, _ := json.Marshal(res)
		fmt.Printf("res:%s", ress)
		fmt.Printf("err:%+v", err)
		So(err, ShouldBeNil)
	})
}
