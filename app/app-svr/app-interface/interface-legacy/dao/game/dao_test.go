package game

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	dao *Dao
)

func TestMain(m *testing.M) {
	flag.Set("conf", "../../cmd/app-interface-test.toml")
	if err := conf.Init(); err != nil {
		panic(err)
	}
	dao = New(conf.Conf)
	os.Exit(m.Run())
	// time.Sleep(time.Second)
}

func TestDao_FolderVideo(t *testing.T) {
	Convey("folder video", t, func() {
		_, err := dao.RecentGame(context.Background(), 111001927, 1, 4, "android")
		So(err, ShouldBeNil)
	})
}

func TestDao_RecentGameSub(t *testing.T) {
	Convey("RecentGameSub", t, func() {
		_, err := dao.RecentGameSub(context.Background(), 111001927, 1, 4, "android")
		So(err, ShouldBeNil)
	})
}

func TestGameItem(t *testing.T) {
	res, err := dao.FetchGameTip(context.Background(), 50505, 6540000, 1, "aaaaa")
	assert.NoError(t, err)
	for _, v := range res {
		t.Logf("game tip is (%+v)", v)
		assert.NotEqual(t, v.ID, 0)
		assert.NotEqual(t, v.Content, "")
		assert.NotEqual(t, v.Url, "")
		assert.NotEqual(t, v.Icon, "")
	}
}
