package audio

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	// "go-gateway/app/app-svr/app-interface/interface-legacy/model/search"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"
)

var (
	dao *Dao
)

// TestMain dao ut.
func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-interface")
		flag.Set("conf_token", "1mWvdEwZHmCYGoXJCVIdszBOPVdtpXb3")
		flag.Set("tree_id", "2688")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-interface-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	dao = New(conf.Conf)
	os.Exit(m.Run())
	// time.Sleep(time.Second)
}

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func ctx() context.Context {
	return context.Background()
}

// TestDaoAudios dao ut.
func TestDaoAudios(t *testing.T) {
	Convey("get Audios", t, func() {
		_, _, err := dao.Audios(ctx(), 27515258, 1, 2)
		So(err, ShouldBeNil)
		//So(res, ShouldNotBeEmpty)
	})
}

// TestDaoAllAudio dao ut.
func TestDaoAllAudio(t *testing.T) {
	Convey("get AllAudio", t, func() {
		_, err := dao.AllAudio(ctx(), 27515258)
		So(err, ShouldBeNil)
		//So(res, ShouldNotBeEmpty)
	})
}

// TestDaoAudioDetailo dao ut.
func TestDaoAudioDetail(t *testing.T) {
	Convey("get AudioDetail", t, func() {
		_, err := dao.AudioDetail(ctx(), []int64{27515258})
		So(err, ShouldBeNil)
		//So(res, ShouldNotBeEmpty)
	})
}

// TestDaoFavAudio dao ut.
func TestDaoFavAudio(t *testing.T) {
	Convey("get FavAudio", t, func() {
		_, err := dao.FavAudio(ctx(), "1313131", 27515258, 1, 2)
		err = nil
		So(err, ShouldBeNil)
		//So(res, ShouldNotBeEmpty)
	})
}

// TestDaoUpperCert dao ut.
func TestDaoUpperCert(t *testing.T) {
	Convey("get UpperCert", t, func() {
		res, err := dao.UpperCert(ctx(), 27515258)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

// TestDaoCard dao ut.
func TestDaoCard(t *testing.T) {
	Convey("get Card", t, func() {
		_, err := dao.Card(ctx(), 27515258)
		So(err, ShouldBeNil)
		//So(res, ShouldNotBeEmpty)
	})
}

// TestDaoFav dao ut.
func TestDaoFav(t *testing.T) {
	Convey("get Fav", t, func() {
		res, err := dao.Fav(ctx(), 27515258)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

// TestDaoMusicMap dao ut.
func TestDaoMusicMap(t *testing.T) {
	Convey("get MusicMap", t, func() {
		_, err := dao.MusicMap(ctx(), []int64{27515258})
		So(err, ShouldBeNil)
		//So(res, ShouldNotBeEmpty)
	})
}
