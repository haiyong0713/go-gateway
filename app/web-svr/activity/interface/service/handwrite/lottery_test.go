package handwrite

import (
	"context"
	"testing"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/dao/handwrite"

	. "github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAddLotteryTimes(t *testing.T) {
	Convey("test AddLotteryTimes success", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockAcc := accapi.NewMockAccountClient(mockCtl)
		mockHandwriteDao := handwrite.NewMockDao(mockCtl)
		defer mockCtl.Finish()

		mockHandwriteDao.EXPECT().AddTimeLock(Any(), Any()).Return(nil)
		mockAcc.EXPECT().Profile3(Any(), Any()).Return(&accapi.ProfileReply{Profile: &accapi.Profile{TelStatus: 1, Silence: 0}}, nil)
		// mockHandwriteDao.EXPECT().AddTimesRecord(Any(), Any(), Any()).Return(nil)

		s.handwrite = mockHandwriteDao
		s.accClient = mockAcc
		_, err := s.AddLotteryTimes(context.Background(), 1)
		So(err, ShouldNotBeNil)

	}))
	Convey("test AddLotteryTimes error1", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockAcc := accapi.NewMockAccountClient(mockCtl)
		mockHandwriteDao := handwrite.NewMockDao(mockCtl)
		defer mockCtl.Finish()

		mockHandwriteDao.EXPECT().AddTimeLock(Any(), Any()).Return(ecode.ActivityWriteHandActivityMemberErr)
		// mockAcc.EXPECT().Profile3(Any(), Any()).Return(&accapi.ProfileReply{Profile: &accapi.Profile{TelStatus: 1, Silence: 0}}, nil)
		// mockHandwriteDao.EXPECT().AddTimesRecord(Any(), Any(), Any()).Return(nil)

		s.handwrite = mockHandwriteDao
		s.accClient = mockAcc
		_, err := s.AddLotteryTimes(context.Background(), 1)
		So(err, ShouldNotBeNil)

	}))
	Convey("test AddLotteryTimes error2", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockAcc := accapi.NewMockAccountClient(mockCtl)
		mockHandwriteDao := handwrite.NewMockDao(mockCtl)
		defer mockCtl.Finish()

		mockHandwriteDao.EXPECT().AddTimeLock(Any(), Any()).Return(nil)
		mockAcc.EXPECT().Profile3(Any(), Any()).Return(&accapi.ProfileReply{Profile: &accapi.Profile{TelStatus: 0, Silence: 0}}, nil)
		// mockHandwriteDao.EXPECT().AddTimesRecord(Any(), Any(), Any()).Return(nil)

		s.handwrite = mockHandwriteDao
		s.accClient = mockAcc
		_, err := s.AddLotteryTimes(context.Background(), 1)
		So(err, ShouldNotBeNil)
	}))
	Convey("test AddLotteryTimes error3", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockAcc := accapi.NewMockAccountClient(mockCtl)
		mockHandwriteDao := handwrite.NewMockDao(mockCtl)
		defer mockCtl.Finish()

		mockHandwriteDao.EXPECT().AddTimeLock(Any(), Any()).Return(nil)
		mockAcc.EXPECT().Profile3(Any(), Any()).Return(&accapi.ProfileReply{Profile: &accapi.Profile{TelStatus: 1, Silence: 0}}, nil)
		// mockHandwriteDao.EXPECT().AddTimesRecord(Any(), Any(), Any()).Return(ecode.ActivityWriteHandActivityMemberErr)

		s.handwrite = mockHandwriteDao
		s.accClient = mockAcc
		_, err := s.AddLotteryTimes(context.Background(), 1)
		So(err, ShouldNotBeNil)
	}))
}

func TestCoins(t *testing.T) {
	Convey("test Coin success", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockHandwriteDao := handwrite.NewMockDao(mockCtl)
		defer mockCtl.Finish()

		mockHandwriteDao.EXPECT().GetAddTimesRecord(Any(), Any(), Any()).Return("true", nil)

		s.handwrite = mockHandwriteDao
		_, err := s.Coin(context.Background(), 1)
		So(err, ShouldNotBeNil)

	}))
	Convey("test Coin err", t, WithService(func(s *Service) {
		mockCtl := NewController(t)
		mockHandwriteDao := handwrite.NewMockDao(mockCtl)
		defer mockCtl.Finish()

		mockHandwriteDao.EXPECT().GetAddTimesRecord(Any(), Any(), Any()).Return("true", nil)

		s.handwrite = mockHandwriteDao
		_, err := s.Coin(context.Background(), 1)
		So(err, ShouldNotBeNil)

	}))
}
