package dynamic

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	dynmal "go-gateway/app/app-svr/app-dynamic/interface/model/dynamic"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDao_MultiStats(t *testing.T) {
	var (
		mid  = int64(1)
		bus  = make(map[string][]*dynmal.LikeBusiItem)
		item []*dynmal.LikeBusiItem
	)
	item = append(item, &dynmal.LikeBusiItem{
		MsgID: 400024925,
	})
	item = append(item, &dynmal.LikeBusiItem{
		MsgID: 880105765,
	})
	bus["archive"] = item
	Convey("MultiStats", t, func() {
		res, err := d.MultiStats(context.TODO(), mid, bus)
		ress, _ := json.Marshal(res)
		fmt.Printf("%s", ress)
		So(err, ShouldBeNil)
	})
}
