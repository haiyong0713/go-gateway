package view

import (
	"context"
	"fmt"
	"testing"

	"go-gateway/app/app-svr/app-view/interface/model"

	"go-common/component/metadata/device"
	api "go-gateway/app/app-svr/app-view/interface/api/view"
	"go-gateway/app/app-svr/app-view/interface/model/view"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

//func Test_View(t *testing.T) {
//	Convey("View", t, func() {
//		v, err := s.ViewHttp(&blademaster.Context{}, 1, 880035114, 0, 0, 10000, 0, 0, 0, "", "", "", "", "", "", "", "", "", "", "", "", "0", "", "", "", "")
//		Println(v, err)
//	})
//}

func Test_ViewPage(t *testing.T) {
	Convey("ViewPage", t, func() {
		v, err := s.ViewPage(context.TODO(), 880035114, 0, 0, "", "", "", false, "", "", "", nil, "", "", "", 0)
		Println(v, err)
	})
}

func Test_AddShare(t *testing.T) {
	Convey("AddShare", t, func() {
		_, _, _, _, err := s.AddShare(context.TODO(), 880035114, 0, 629000090, "app", "127.0.0.1", "abcd")
		So(err, ShouldBeNil)
	})
}

func Test_Shot(t *testing.T) {
	Convey("Shot", t, func() {
		shot, _ := s.Shot(context.TODO(), 400044161, 10285289, "", 0)
		fmt.Printf("===%+v===", shot)
	})
}

func Test_Like(t *testing.T) {
	Convey("Like", t, func() {
		s.Like(context.TODO(), 1, 1, 0, 0, "", "", "", "", "", "", "", "")
	})
}

func Test_AddCoin(t *testing.T) {
	Convey("AddCoin", t, func() {
		s.AddCoin(context.TODO(), 1, 1684013, 2, 0, 1, 0, "", "", "", "", "", "", "", "")
	})
}

func Test_Paster(t *testing.T) {
	Convey("Paster", t, func() {
		s.Paster(context.TODO(), 1, 1, "1", "1", "")
	})
}

func Test_VipPlayURL(t *testing.T) {
	Convey("VipPlayURL", t, func() {
		s.VipPlayURL(context.TODO(), 1, 1, 1684013)
	})
}

func TestMaterialView(t *testing.T) {
	var (
		c      = context.Background()
		params = &view.MaterialParam{
			AID:      1,
			CID:      1,
			Build:    8470,
			Platform: "ios",
			Device:   "phone",
			MobiApp:  "iphone",
		}
	)
	Convey("MaterialView", t, func() {
		material, err := s.Material(c, params)
		for _, v := range material {
			fmt.Printf("-----%+v-------", v)
		}
		Convey("Then err should be nil.", func() {
			So(err, ShouldBeNil)
		})
	})
}

func TestService_VideoDownload(t *testing.T) {
	c := context.Background()
	shortFormVideoDownload := api.ShortFormVideoDownloadReq{
		Aid:      10318733,
		Cid:      10211333,
		Mid:      27515258,
		Buvid:    "ZB4855B33874DAAA4E01A5FE5589C3A1FC06",
		MobiApp:  "iPhone",
		Build:    6190400,
		Device:   "phone",
		Platform: "iPhone",
		Spmid:    "346989d0-1bc9-49c4-ae0b-5daf83a93def",
	}
	req := view.VideoDownloadReq{
		ShortFormVideoDownloadReq: &shortFormVideoDownload,
	}
	Convey("VideoDownload", t, func() {
		data, err := s.VideoDownload(c, &req)
		fmt.Println(data)
		Convey("Then err should be nil.", func() {
			So(err, ShouldBeNil)
		})
	})
}

func TestService_VideoOnlineText(t *testing.T) {
	t1 := s.onlineText(9)
	assert.Equal(t, "<10", t1)
	t2 := s.onlineText(89)
	assert.Equal(t, "80+", t2)
	t3 := s.onlineText(209)
	assert.Equal(t, "200+", t3)
	t4 := s.onlineText(3189)
	assert.Equal(t, "3000+", t4)
	t5 := s.onlineText(70012)
	assert.Equal(t, "7万+", t5)
	t6 := s.onlineText(93011)
	assert.Equal(t, "9.3万+", t6)
	t7 := s.onlineText(1210983)
	assert.Equal(t, "10万+", t7)
}

func TestService_VideoOnline(t *testing.T) {
	res, err := s.VideoOnline(context.Background(), &view.VideoOnlineParam{
		Mid:   123,
		Buvid: "abc",
		Aid:   47527433,
		Cid:   83244511,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
}

func TestHandleIosVersionProblem(t *testing.T) {
	var (
		dev = &device.Device{
			RawMobiApp: "iphone",
			Device:     "pad",
			Build:      64300100,
		}
		playerCards = &api.VideoGuide{}
	)
	Convey("handleIosVersionProblem", t, func() {
		extendDefaultOperationCardOnIOS643(model.PlatNew(dev.RawMobiApp, dev.Device), dev.Build, playerCards)
		assert.NotEmpty(t, playerCards.OperationCardNew)
	})
}

func TestService_BizExtra(t *testing.T) {
	arg := "%7B%22ad_play_page%22%3A1%7D"
	result := decodeBizExtra(arg)
	assert.Equal(t, 1, result.AdPlayPage)
}

func TestReOrderChronosPkgListByRank(t *testing.T) {
	in := []*view.PackageInfo{
		{
			Rank: 1,
		},
		{
			Rank: 3,
		},
		{
			Rank: 2,
		},
		{
			Rank: 2,
		},
	}
	out := reOrderChronosPkgListByRank(in)
	for i := 0; i < len(in)-1; i++ {
		assert.Equal(t, out[i+1].Rank <= out[i].Rank, true)
	}
}

func TestSplitMessageByDotAndConvertToInt(t *testing.T) {
	res := map[int64]struct{}{
		1: {},
		2: {},
		3: {},
		4: {},
		5: {},
	}
	reply := splitMessageByDotAndConvertToInt("1,2,3,4,5")
	for k := range reply {
		_, ok := res[k]
		assert.Equal(t, ok, true)
	}
}
