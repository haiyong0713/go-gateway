package archive

import (
	"context"
	"fmt"
	"testing"

	"go-gateway/app/app-svr/app-view/interface/model/view"

	"go-gateway/app/app-svr/archive/service/api"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_View3(t *testing.T) {
	Convey("View3", t, func() {
		res, err := d.View3(context.TODO(), 880035114)
		fmt.Println(res, err)
	})
}

func Test_Description(t *testing.T) {
	Convey("Description", t, func() {
		d.Description(context.TODO(), 280027822)
	})
}

func Test_DescriptionV2(t *testing.T) {
	v := view.ViewStatic{
		Arc: &api.Arc{Aid: 1},
	}
	a := &view.View{ViewStatic: &v}
	Convey("Description", t, func() {
		d.DescriptionV2(context.TODO(), 600042219, a)
	})
}

func Test_Argument(t *testing.T) {
	Convey("Argument", t, func() {
		d.Argument(context.TODO(), 2)
	})
}
