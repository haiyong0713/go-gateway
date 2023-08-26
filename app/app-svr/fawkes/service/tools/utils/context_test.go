package utils

import (
	"context"
	"testing"

	bm "go-common/library/net/http/blademaster"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetOperator(t *testing.T) {
	Convey("test", t, func() {

		var bc = bm.Context{
			Context:   context.Background(),
			Request:   nil,
			Writer:    nil,
			Keys:      nil,
			Error:     nil,
			RoutePath: "",
			Params:    nil,
		}

		nbc := new(bm.Context)
		nbc.Context = context.Background()
		GetUsername(nbc)
		GetUsername(bc)

	})
}
