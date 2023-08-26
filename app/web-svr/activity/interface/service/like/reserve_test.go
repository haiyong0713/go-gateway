package like

import (
	"context"
	"errors"
	"reflect"
	"testing"

	relationAPI "git.bilibili.co/bapis/bapis-go/account/service/relation"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/currency"
	"go-gateway/app/web-svr/activity/interface/dao/like"
	"go-gateway/app/web-svr/activity/interface/mock"
	likeM "go-gateway/app/web-svr/activity/interface/model/like"

	"github.com/golang/mock/gomock"

	"bou.ke/monkey"
)

// go test -v reserve_test.go service.go single.go stein.go bdf.go subject_rule.go article.go wx_lottery.go act.go question.go currency.go likeact.go reserve.go like.go lottery.go task.go
func TestReserveBiz(t *testing.T) {
	t.Run("test reserve only err", testReserveOnlyErr)
	t.Run("test reserved biz", testReserved)
	t.Run("test spring reserve biz", testSpringReserveBiz)
	t.Run("test async reserve biz", testAsyncReserveBiz)
}

func testAsyncReserveBiz(t *testing.T) {
	t.Run("test insert into db successfully", testInsertIntoDBSuccessfully)
	t.Run("test insert into db failed", testInsertIntoDBFailed)
}

func testInsertIntoDBFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	springCurrencyID := int64(1)
	springReserveSID := int64(6666)
	mid := int64(88)
	num := int32(1)
	followerLimit := int64(88)
	errMsg := "UpUserAmount err"

	mockService := new(Service)
	{
		cfg := new(conf.Config)
		{
			springCfg := new(conf.StarSpring)
			{
				springCfg.CurrID = springCurrencyID
				springCfg.FollowerLimit = followerLimit
				springCfg.ReserveSid = springReserveSID
			}

			cfg.StarSpring = springCfg
		}

		mockService.c = cfg
	}

	dao := new(like.Dao)
	monkey.PatchInstanceMethod(reflect.TypeOf(dao), "ReserveOnly", func(_ *like.Dao, ctx context.Context, sid, mid int64) (res *likeM.HasReserve, err error) {
		return
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(dao), "AddReserve", func(_ *like.Dao, ctx context.Context, reserveInfo *likeM.ActReserve) (id int64, err error) {
		if reserveInfo.Mid != mid || reserveInfo.State != 1 || reserveInfo.Num != num {
			t.Errorf("AddReserve params is invalid, please check")
		}

		if reserveInfo.Mid == mid {
			err = errors.New("Duplicate entry")
		}

		return
	})

	currencyDao := new(currency.Dao)
	monkey.PatchInstanceMethod(reflect.TypeOf(currencyDao), "UpUserAmount", func(_ *currency.Dao, ctx context.Context, id, fromMid, toMid, amount int64, remark string) (err error) {
		if id != springCurrencyID || fromMid != 0 || toMid != mid || amount != followerLimit || remark != "" {
			err = errors.New(errMsg)
			t.Errorf("UpUserAmount params is invalid, please check")
		}

		return
	})

	mockRelationClient := mock.NewMockRelationClient(ctrl)
	mockRelationClient.EXPECT().Stat(
		ctx,
		&relationAPI.MidReq{
			Mid: mid,
		}).Return(&relationAPI.StatReply{Follower: followerLimit}, nil)
	mockService.relClient = mockRelationClient

	err := mockService.AsyncReserve(ctx, springReserveSID, mid, num)
	if err == nil || err != ecode.ActivityRepeatSubmit {
		t.Errorf("error should as (%v), but now %v", ecode.ActivityRepeatSubmit, err)
	}
}

func testInsertIntoDBSuccessfully(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	springCurrencyID := int64(1)
	springReserveSID := int64(6666)

	mid := int64(1)
	num := int32(1)
	followerLimit := int64(1)
	errMsg := "UpUserAmount err"
	errMsg4SaveDB := "actReserve is not expected"
	errMsg4AddCacheReserveOnly := "AddCacheReserveOnly is not expected"
	primaryKey := int64(8888)

	mockService := new(Service)
	{
		cfg := new(conf.Config)
		{
			springCfg := new(conf.StarSpring)
			{
				springCfg.CurrID = springCurrencyID
				springCfg.FollowerLimit = followerLimit
				springCfg.ReserveSid = springReserveSID
			}

			cfg.StarSpring = springCfg
		}

		mockService.c = cfg
	}
	monkey.PatchInstanceMethod(reflect.TypeOf(mockService), "AwardSubject", func(_ *Service, ctx context.Context, sidInMock, midInMock int64) (err error) {
		if sidInMock != springReserveSID || midInMock != mid {
			t.Errorf("AwardSubject params is not expected")
		}

		return
	})

	mockService.cache = fanout.New("cache", fanout.Worker(1), fanout.Buffer(1024))

	dao := new(like.Dao)
	monkey.PatchInstanceMethod(reflect.TypeOf(dao), "ReserveOnly", func(_ *like.Dao, ctx context.Context, sid, mid int64) (res *likeM.HasReserve, err error) {
		return
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(dao), "IncrSubjectStat", func(_ *like.Dao, ctx context.Context, sidInMock int64, numInMock int32) (err error) {
		if sidInMock != springReserveSID || numInMock != num {
			t.Errorf("IncrSubjectStat params is not expected")
		}

		return
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(dao), "IncrCacheReserveTotal", func(_ *like.Dao, ctx context.Context, sidInMock int64, numInMock int32) (err error) {
		if sidInMock != springReserveSID || numInMock != num {
			t.Errorf("IncrCacheReserveTotal params is not expected")
		}

		return
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(dao), "AddReserve", func(_ *like.Dao, ctx context.Context, reserveInfo *likeM.ActReserve) (id int64, err error) {
		if reserveInfo.Mid != mid || reserveInfo.State != 1 || reserveInfo.Num != num {
			err = errors.New(errMsg4SaveDB)
			t.Errorf("AddReserve params is invalid, please check")
		}

		id = primaryKey

		return
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(dao), "AddCacheReserveOnly", func(_ *like.Dao, ctx context.Context, id int64, val *likeM.HasReserve, mid int64) (err error) {
		if id != springReserveSID || val.ID != primaryKey || val.Num != num || val.State != 1 {
			err = errors.New(errMsg4AddCacheReserveOnly)
			t.Errorf("AddCacheReserveOnly params is invalid, please check")
		}

		return
	})

	currencyDao := new(currency.Dao)
	monkey.PatchInstanceMethod(reflect.TypeOf(currencyDao), "UpUserAmount", func(_ *currency.Dao, ctx context.Context, id, fromMid, toMid, amount int64, remark string) (err error) {
		if id != springCurrencyID || fromMid != 0 || toMid != mid || amount != followerLimit || remark != "" {
			err = errors.New(errMsg)
			t.Errorf("UpUserAmount params is invalid, please check")
		}

		return
	})

	mockRelationClient := mock.NewMockRelationClient(ctrl)
	mockRelationClient.EXPECT().Stat(
		ctx,
		&relationAPI.MidReq{
			Mid: mid,
		}).Return(&relationAPI.StatReply{Follower: followerLimit}, nil)
	mockService.relClient = mockRelationClient

	err := mockService.AsyncReserve(ctx, springReserveSID, mid, num)
	if err != nil {
		t.Errorf("error should as nil, but now %v", err)
	}
}

func testSpringReserveBiz(t *testing.T) {
	t.Run("test relation server err", testRelationServerErr)
	t.Run("test follower limit err", testFollowerLimitError)
	t.Run("test update user amount err", testUpdateUserAmountErr)
}

func testUpdateUserAmountErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	springCurrencyID := int64(1)
	springReserveSID := int64(6666)
	mid := int64(1)
	num := int32(1)
	followerLimit := int64(1)
	errMsg := "UpUserAmount err"

	mockService := new(Service)
	{
		cfg := new(conf.Config)
		{
			springCfg := new(conf.StarSpring)
			{
				springCfg.CurrID = springCurrencyID
				springCfg.FollowerLimit = followerLimit
				springCfg.ReserveSid = springReserveSID
			}

			cfg.StarSpring = springCfg
		}

		mockService.c = cfg
	}

	dao := new(like.Dao)
	monkey.PatchInstanceMethod(reflect.TypeOf(dao), "ReserveOnly", func(_ *like.Dao, ctx context.Context, sid, mid int64) (res *likeM.HasReserve, err error) {
		return
	})
	currencyDao := new(currency.Dao)
	monkey.PatchInstanceMethod(reflect.TypeOf(currencyDao), "UpUserAmount", func(_ *currency.Dao, ctx context.Context, id, fromMid, toMid, amount int64, remark string) (err error) {
		if id != springCurrencyID || fromMid != 0 || toMid != mid || amount != followerLimit || remark != "" {
			t.Errorf("UpUserAmount params is invalid, please check")
		}

		err = errors.New(errMsg)

		return
	})

	mockRelationClient := mock.NewMockRelationClient(ctrl)
	mockRelationClient.EXPECT().Stat(
		ctx,
		&relationAPI.MidReq{
			Mid: mid,
		}).Return(&relationAPI.StatReply{Follower: followerLimit}, nil)
	mockService.relClient = mockRelationClient

	err := mockService.AsyncReserve(ctx, springReserveSID, mid, num)
	if err == nil || err.Error() != errMsg {
		t.Errorf("error should as %v, but now %v", ecode.ActivityRepeatSubmit, err)
	}
}

func testRelationServerErr(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	springReserveSID := int64(6666)
	mid := int64(1)
	num := int32(1)
	followerLimit := int64(1)
	invalidFollowerLimit := int64(2)
	errMsg := "relation server err"

	mockService := new(Service)
	{
		cfg := new(conf.Config)
		{
			springCfg := new(conf.StarSpring)
			{
				springCfg.FollowerLimit = followerLimit
				springCfg.ReserveSid = springReserveSID
			}

			cfg.StarSpring = springCfg
		}

		mockService.c = cfg
	}

	dao := new(like.Dao)
	monkey.PatchInstanceMethod(reflect.TypeOf(dao), "ReserveOnly", func(_ *like.Dao, ctx context.Context, sid, mid int64) (res *likeM.HasReserve, err error) {
		return
	})

	mockRelationClient := mock.NewMockRelationClient(ctrl)
	mockRelationClient.EXPECT().Stat(
		ctx,
		&relationAPI.MidReq{
			Mid: mid,
		}).Return(&relationAPI.StatReply{Follower: invalidFollowerLimit}, errors.New(errMsg))
	mockService.relClient = mockRelationClient

	err := mockService.AsyncReserve(ctx, springReserveSID, mid, num)
	if err == nil || err.Error() != errMsg {
		t.Errorf("error should as %v, but now %v", ecode.ActivityRepeatSubmit, err)
	}
}

func testFollowerLimitError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	springReserveSID := int64(6666)
	mid := int64(1)
	num := int32(1)
	followerLimit := int64(1)
	invalidFollowerLimit := int64(2)

	mockService := new(Service)
	{
		cfg := new(conf.Config)
		{
			springCfg := new(conf.StarSpring)
			{
				springCfg.FollowerLimit = followerLimit
				springCfg.ReserveSid = springReserveSID
			}

			cfg.StarSpring = springCfg
		}

		mockService.c = cfg
	}

	dao := new(like.Dao)
	monkey.PatchInstanceMethod(reflect.TypeOf(dao), "ReserveOnly", func(_ *like.Dao, ctx context.Context, sid, mid int64) (res *likeM.HasReserve, err error) {
		return
	})

	mockRelationClient := mock.NewMockRelationClient(ctrl)
	mockRelationClient.EXPECT().Stat(
		ctx,
		&relationAPI.MidReq{
			Mid: mid,
		}).Return(&relationAPI.StatReply{Follower: invalidFollowerLimit}, nil)
	mockService.relClient = mockRelationClient

	err := mockService.AsyncReserve(ctx, springReserveSID, mid, num)
	if err == nil || err != ecode.ActivityUpFanLimit {
		t.Errorf("error should as %v, but now %v", ecode.ActivityRepeatSubmit, err)
	}
}

func testReserved(t *testing.T) {
	mockService := new(Service)
	dao := new(like.Dao)

	monkey.PatchInstanceMethod(reflect.TypeOf(dao), "ReserveOnly", func(_ *like.Dao, ctx context.Context, sid, mid int64) (res *likeM.HasReserve, err error) {
		res = new(likeM.HasReserve)
		{
			res.ID = 1
			res.State = 1
		}

		return
	})

	ctx := context.Background()
	sid := int64(1)
	mid := int64(1)
	num := int32(1)

	err := mockService.AsyncReserve(ctx, sid, mid, num)
	if err == nil || err != ecode.ActivityRepeatSubmit {
		t.Errorf("error should as %v, but now %v", ecode.ActivityRepeatSubmit, err)
	}
}

func testReserveOnlyErr(t *testing.T) {
	mockService := new(Service)
	dao := new(like.Dao)
	errMsg := "ReserveOnly_err"

	monkey.PatchInstanceMethod(reflect.TypeOf(dao), "ReserveOnly", func(_ *like.Dao, ctx context.Context, sid, mid int64) (res *likeM.HasReserve, err error) {
		if sid == 1 && mid == 1 {
			err = errors.New(errMsg)
		}

		return
	})

	ctx := context.Background()
	sid := int64(1)
	mid := int64(1)
	num := int32(1)

	err := mockService.AsyncReserve(ctx, sid, mid, num)
	if err == nil || err.Error() != errMsg {
		newErr := errors.New(errMsg)

		t.Errorf("error should as %v, but now %v", newErr, err)
	}
}
