package mod

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go-gateway/app/app-svr/fawkes/service/conf"
)

func TestMain(m *testing.M) {
	var err error
	if err = conf.Init(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestBusFile_SetURL(t *testing.T) {
	Convey("test", t, func() {
		a := &BusFile{
			ID:          0,
			VersionID:   0,
			Name:        "",
			ContentType: "",
			Size:        0,
			Md5:         "",
			URL:         "/appstaticboss/xxxx/xxx/x/x/",
			IsPatch:     false,
			FromVer:     0,
			Version:     nil,
			Config:      nil,
			Gray:        nil,
		}
		a.SetURL(conf.Conf.Mod)
	})
}
