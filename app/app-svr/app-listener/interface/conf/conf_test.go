package conf

import (
	"testing"

	"go-common/component/metadata/device"
)

const data = `[Res]
  [Res.Text]
    AddFav = "添加收藏"
  [Res.Icon]
    PickHeaderBtn = "testIcon"
`

func TestSet(t *testing.T) {
	conf := new(AppConfig)
	err := conf.Set(data)
	if err != nil {
		t.Fatal(err)
	}
	if conf.Res.Text.AddFav != `添加收藏` || conf.Res.Icon.PickHeaderBtn != `testIcon` {
		t.Errorf("not match")
	}
}

const featureData = `[Feature]
  [Feature.MusicFavTabShow]
    Android = 6530000
    IPhone = 65300000
`

func TestFeatureGate(t *testing.T) {
	d := &device.Device{Build: 65300000, RawMobiApp: "iphone", Device: "phone"}
	conf := new(AppConfig)
	err := conf.Set(featureData)
	if err != nil {
		t.Fatal(err)
	}
	if conf.Feature.MusicFavTabShow.Enabled(d) != true {
		t.Errorf("Feature MusicFavTabShow expected enabled")
	}
}
