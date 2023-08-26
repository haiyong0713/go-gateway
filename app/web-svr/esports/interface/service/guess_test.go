package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"go-gateway/app/web-svr/esports/interface/model"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx = context.Background()
)

func TestService_GuessDetail(t *testing.T) {
	Convey("TestService_GuessDetail", t, WithService(func(s *Service) {
		rs, err := s.GuessDetail(ctx, 37, 27515412)
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(rs)
		fmt.Println(string(bs))
	}))
}

func TestService_GuessDetailCoin(t *testing.T) {
	Convey("TestService_GuessDetailCoin", t, WithService(func(s *Service) {
		rs, err := s.GuessDetailCoin(ctx, 27515412, 5)
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(rs)
		fmt.Println(string(bs))
	}))
}

func TestService_AddGuessDetail(t *testing.T) {
	Convey("TestService_AddGuessDetail", t, WithService(func(s *Service) {
		param := &model.AddGuessParam{
			MID:      27515412,
			OID:      39,
			MainID:   18,
			DetailID: 35,
			Count:    3,
		}
		err := s.AddGuessDetail(ctx, param)
		So(err, ShouldBeNil)
	}))
}

func TestService_GuessGame(t *testing.T) {
	Convey("TestService_GuessGame", t, WithService(func(s *Service) {
		rs, err := s.GuessCollGS(ctx, 3)
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(rs)
		fmt.Println(string(bs))
	}))
}

func TestService_GuessCollCal(t *testing.T) {
	Convey("TestService_GuessGame", t, WithService(func(s *Service) {
		param := &model.ParamContest{
			Stime: "2019-01-01 00:00:00",
			Etime: "2019-11-02 00:00:00",
			Pn:    1,
			Ps:    10,
		}
		rs, err := s.GuessCollCalendar(ctx, param)
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(rs)
		fmt.Println(string(bs))
	}))
}

func TestService_GuessCollQues(t *testing.T) {
	Convey("TestService_AddGuessDetail", t, WithService(func(s *Service) {
		param := &model.ParamContest{
			Pn: 1,
			Ps: 1,
		}
		res, _, err := s.GuessCollQues(ctx, param, 27515430)
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}

func TestService_GuessCollStatis(t *testing.T) {
	Convey("TestService_GuessCollStatis", t, WithService(func(s *Service) {
		res, err := s.GuessCollStatis(ctx, 27515412)
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}

func TestService_GuessCollRecord(t *testing.T) {
	Convey("TestService_GuessCollStatis", t, WithService(func(s *Service) {
		p := &model.GuessCollRecoParam{
			Mid:  27515403,
			Type: 1,
			Pn:   1,
			Ps:   10,
		}
		res, err := s.GuessCollRecord(ctx, p)
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}

func TestService_GuessDetailValue(t *testing.T) {
	Convey("TestService_GuessDetailValue", t, WithService(func(s *Service) {
		dbContests, err := s.dao.EpContests(ctx, []int64{594})
		if err != nil {
			panic(err)
		}
		rs := s.GuessDetailValue(ctx, dbContests[594], 27515412)
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(rs)
		fmt.Println(string(bs))
	}))
}

func TestService_GuessTeamRecent(t *testing.T) {
	Convey("TestService_GuessTeamRecent", t, WithService(func(s *Service) {
		dbContests, err := s.GuessTeamRecent(ctx, &model.ParamEsGuess{
			HomeID: 2,
			AwayID: 3,
			CID:    1,
		})
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(dbContests)
		fmt.Println(string(bs))
	}))
}

func TestService_GuessMoreMatch(t *testing.T) {
	Convey("TestService_GuessTeamRecent", t, WithService(func(s *Service) {
		dbContests, err := s.GuessMoreMatch(ctx, 2, 3, 4, 1)
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(dbContests)
		fmt.Println(string(bs))
	}))
}

func TestService_GuessMatchRecord(t *testing.T) {
	Convey("TestService_GuessCollStatis", t, WithService(func(s *Service) {
		res, err := s.GuessMatchRecord(ctx, 27515412, 39)
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(res)
		fmt.Println(string(bs))
	}))
}
