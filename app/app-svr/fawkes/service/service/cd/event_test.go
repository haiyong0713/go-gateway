package cd

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"

	"github.com/smartystreets/goconvey/convey"

	"go-gateway/app/app-svr/fawkes/service/conf"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	cdmdl "go-gateway/app/app-svr/fawkes/service/model/cd"
)

var srv *Service

func init() {
	err := conf.Init()
	if err != nil {
		panic(err)
	}
	srv = New(conf.Conf)
	time.Sleep(time.Second)
}

func Test_splitChannels(t *testing.T) {
	convey.Convey("test split channels", t, func() {
		var channels []*appmdl.Channel

		for i := 0; i < 130; i++ {
			c := &appmdl.Channel{
				AppKey: "w19e",
				ID:     int64(i),
				AID:    int64(i) + 100,
				Group:  nil,
			}
			channels = append(channels, c)
		}
		splitChannels(channels, 50)
	})
}

func Test_SortGroup(t *testing.T) {
	convey.Convey("test split channels", t, func() {

		gs := make([]*appmdl.GroupChannels, 0, 2)
		gs = append(gs, &appmdl.GroupChannels{
			Group: &appmdl.ChannelGroupInfo{
				Name:     "1",
				Priority: 1,
			},
		})
		gs = append(gs, &appmdl.GroupChannels{
			Group: &appmdl.ChannelGroupInfo{
				Name:     "3",
				Priority: 3,
			},
		})
		gs = append(gs, &appmdl.GroupChannels{
			Group: &appmdl.ChannelGroupInfo{
				Name:     "2",
				Priority: 2,
			},
		})
		var g Groups
		g = gs
		sort.Sort(g)
		log.Info("%v", g)
	})
}

func TestService_packGenerateUpdateAction(t *testing.T) {
	var generateSuccessList []*cdmdl.Generate
	generateSuccessList = append(generateSuccessList, &cdmdl.Generate{ID: 1}, &cdmdl.Generate{ID: 2}, &cdmdl.Generate{ID: 3})

	convey.Convey("", t, func() {
		cache := fanout.New("test")
		for i, v := range generateSuccessList {
			id := v.ID
			cache.Do(context.Background(), func(ctx context.Context) {
				fmt.Printf("i %v, V.id %v, id %v \n", i, &generateSuccessList[i].ID, id)
			})
		}
		_ = cache.Close()
	})
}

func TestService_packGenerateUpdateAction1(t *testing.T) {
	var generateSuccessList []*cdmdl.Generate

	var generateSuccessList1 []cdmdl.Generate

	generateSuccessList = append(generateSuccessList, &cdmdl.Generate{ID: 1}, &cdmdl.Generate{ID: 2}, &cdmdl.Generate{ID: 3})
	generateSuccessList1 = append(generateSuccessList1, cdmdl.Generate{ID: 1}, cdmdl.Generate{ID: 2}, cdmdl.Generate{ID: 3})

	convey.Convey("", t, func() {
		for _, v := range generateSuccessList {
			fmt.Printf("out list id type %T point %p point %p  point %p  point %p  \n", v, v, generateSuccessList[0], generateSuccessList[1], generateSuccessList[2])
			go func() {
				fmt.Printf("list id type %T point %p \n", v, v)
			}()
		}

		for _, v := range generateSuccessList1 {
			fmt.Printf("out list1 id type %T point %p point %p  point %p  point %p  \n", v, &v, &generateSuccessList1[0], &generateSuccessList1[1], &generateSuccessList1[2])
			go func() {
				fmt.Printf("list1 id type %T point %p \n", v, &v)
			}()
		}
	})
}

func TestService_packGenerateUpdateAction2(t *testing.T) {
	var a, b, c = 1, 2, 3

	var pointList []*int
	var List []int

	pointList = append(pointList, &a, &b, &c)
	List = append(List, a, b, c)

	convey.Convey("", t, func() {

		fmt.Printf("pointList p %p p %p p %p p %p\n", pointList, &pointList[0], &pointList[1], &pointList[2])
		fmt.Printf("List p %p p %p p %p p %p\n", List, &List[0], &List[1], &List[2])
		fmt.Printf("List p %p\n", List)
	})

}

func TestService_PubPackGreyData(t *testing.T) {
	convey.Convey("", t, func() {
		producer := srv.fkDao.NewProducer(context.Background(), conf.Conf.Databus.Topics.PackGreyDataPub.Group, conf.Conf.Databus.Topics.PackGreyDataPub.Name)
		convey.So(producer, convey.ShouldNotBeNil)
	})
}
