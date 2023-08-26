package bws

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestBwsmidKey(t *testing.T) {
	convey.Convey("midKey", t, func(convCtx convey.C) {
		var (
			bid = int64(0)
			mid = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := midKey(bid, mid)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestBwskeyKey(t *testing.T) {
	convey.Convey("keyKey", t, func(convCtx convey.C) {
		var (
			bid = int64(0)
			key = ""
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := keyKey(bid, key)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestBwsbwsPointsKey(t *testing.T) {
	convey.Convey("bwsPointsKey", t, func(convCtx convey.C) {
		var (
			id = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := bwsPointsKey(id)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestBwspointsKey(t *testing.T) {
	convey.Convey("pointsKey", t, func(convCtx convey.C) {
		var (
			id = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := pointsKey(id)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestBwsrechargeLevelKey(t *testing.T) {
	convey.Convey("rechargeLevelKey", t, func(convCtx convey.C) {
		var (
			id = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := rechargeLevelKey(id)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestBwsrechargeAwardKey(t *testing.T) {
	convey.Convey("rechargeAwardKey", t, func(convCtx convey.C) {
		var (
			id = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := rechargeAwardKey(id)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestBwsachievesKey(t *testing.T) {
	convey.Convey("achievesKey", t, func(convCtx convey.C) {
		var (
			id = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := achievesKey(id)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestBwsbwsSignKey(t *testing.T) {
	convey.Convey("bwsSignKey", t, func(convCtx convey.C) {
		var (
			id = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := bwsSignKey(id)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestBwspointSignKey(t *testing.T) {
	convey.Convey("pointSignKey", t, func(convCtx convey.C) {
		var (
			pid = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := pointSignKey(pid)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestBwsfieldsListKey(t *testing.T) {
	convey.Convey("fieldsListKey", t, func(convCtx convey.C) {
		var (
			bid = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := fieldsListKey(bid)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.So(p1, convey.ShouldNotBeNil)
			})
		})
	})
}
