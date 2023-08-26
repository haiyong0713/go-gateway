package dao

import (
	"context"
	"testing"

	arcmdl "git.bilibili.co/bapis/bapis-go/archive/service"
	mdlesp "go-gateway/app/web-svr/esports/job/model"

	"github.com/smartystreets/goconvey/convey"
)

func TestEsportsContests(t *testing.T) {
	var (
		c     = context.Background()
		stime = int64(1539590040)
		etime = int64(1539590040)
	)
	convey.Convey("Contests", t, func(ctx convey.C) {
		res, err := d.Contests(c, stime, etime)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
		})
	})
}

func TestEsportsTeams(t *testing.T) {
	var (
		c      = context.Background()
		homeID = int64(1)
		awayID = int64(2)
	)
	convey.Convey("Teams", t, func(ctx convey.C) {
		res, err := d.Teams(c, homeID, awayID)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
		})
	})
}

func TestEsportsArcs(t *testing.T) {
	var (
		c     = context.Background()
		id    = int64(1)
		limit = int(50)
	)
	convey.Convey("Arcs", t, func(ctx convey.C) {
		res, err := d.Arcs(c, id, limit)
		ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestEsportsUpArcScore(t *testing.T) {
	var (
		c        = context.Background()
		partArcs = []*mdlesp.Arc{}
		arcs     map[int64]*arcmdl.Arc
	)
	convey.Convey("UpArcScore", t, func(ctx convey.C) {
		err := d.UpArcScore(c, partArcs, arcs)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestEsportsscore(t *testing.T) {
	var (
		arc = &arcmdl.Arc{}
	)
	convey.Convey("score", t, func(ctx convey.C) {
		res := d.score(arc)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}

func TestEsportsContPoints(t *testing.T) {
	convey.Convey("ContPoints", t, func(convCtx convey.C) {
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.ContPoints(context.Background())
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				println(err, res)
			})
		})
	})
}

func TestSeriesSeason(t *testing.T) {
	convey.Convey("SeriesSeason", t, func(ctx convey.C) {
		res, err := d.SeriesSeason(context.Background(), 2)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(len(res), convey.ShouldBeGreaterThanOrEqualTo, 0)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestLolPlayerSerie(t *testing.T) {
	convey.Convey("LolPlayerSerie", t, func(ctx convey.C) {
		res, err := d.LolPlayerSerie(context.Background(), 1513, 8953)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(res, convey.ShouldNotBeNil)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestLolTeamSerie(t *testing.T) {
	convey.Convey("LolTeamSerie", t, func(ctx convey.C) {
		res, err := d.LolTeamSerie(context.Background(), 1513, 1569)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(res, convey.ShouldNotBeNil)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDotaPlayerSerie(t *testing.T) {
	convey.Convey("DotaPlayerSerie", t, func(ctx convey.C) {
		res, err := d.DotaPlayerSerie(context.Background(), 1723, 20213)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(res, convey.ShouldNotBeNil)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDotaTeamSerie(t *testing.T) {
	convey.Convey("DotaTeamSerie", t, func(ctx convey.C) {
		res, err := d.DotaTeamSerie(context.Background(), 1723, 125796)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(res, convey.ShouldNotBeNil)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestAddLolPlayer(t *testing.T) {
	var players []*mdlesp.LolPlayer
	convey.Convey("AddLolPlayer", t, func(ctx convey.C) {
		player := &mdlesp.LolPlayer{PlayerID: 123, TeamID: 111, Win: 10, KDA: 20}
		players = append(players, player)
		err := d.AddLolPlayer(context.Background(), players)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUpLolPlayer(t *testing.T) {
	convey.Convey("UpLolPlayer", t, func(ctx convey.C) {
		player := &mdlesp.LolPlayer{ID: 123, PlayerID: 123, TeamID: 111, Win: 30, KDA: 40}
		err := d.UpLolPlayer(context.Background(), player)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestAddLolTeam(t *testing.T) {
	var teams []mdlesp.LolTeam
	convey.Convey("AddLolTeam", t, func(ctx convey.C) {
		team := mdlesp.LolTeam{TeamID: 111, Win: 10, KDA: 20}
		teams = append(teams, team)
		err := d.AddLolTeam(context.Background(), teams)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUpLolTeam(t *testing.T) {
	convey.Convey("UpLolTeam", t, func(ctx convey.C) {
		team := mdlesp.LolTeam{ID: 123, TeamID: 111, Win: 30, KDA: 40}
		err := d.UpLolTeam(context.Background(), team)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestAddDotaPlayer(t *testing.T) {
	var players []mdlesp.DotaPlayer
	convey.Convey("AddDotaPlayer", t, func(ctx convey.C) {
		player := mdlesp.DotaPlayer{PlayerID: 123, TeamID: 111, Win: 10, KDA: 20}
		players = append(players, player)
		err := d.AddDotaPlayer(context.Background(), players)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUPDotaPlayer(t *testing.T) {
	convey.Convey("UPDotaPlayer", t, func(ctx convey.C) {
		player := mdlesp.DotaPlayer{ID: 123, PlayerID: 123, TeamID: 111, Win: 30, KDA: 40}
		err := d.UpDotaPlayer(context.Background(), player)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestAddDotaTeam(t *testing.T) {
	var teams []mdlesp.DotaTeam
	convey.Convey("AddDotaTeam", t, func(ctx convey.C) {
		team := mdlesp.DotaTeam{TeamID: 111, Win: 10, KDA: 20}
		teams = append(teams, team)
		err := d.AddDotaTeam(context.Background(), teams)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestUpDotaTeam(t *testing.T) {
	convey.Convey("UpDotaTeam", t, func(ctx convey.C) {
		team := mdlesp.DotaTeam{ID: 123, TeamID: 111, Win: 30, KDA: 40}
		err := d.UpDotaTeam(context.Background(), team)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoLolGames(t *testing.T) {
	convey.Convey("LolGames", t, func(ctx convey.C) {
		res, err := d.LolGames(context.Background(), 524686)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(res, convey.ShouldNotBeNil)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoDotaGames(t *testing.T) {
	convey.Convey("DotaGames", t, func(ctx convey.C) {
		res, err := d.DotaGames(context.Background(), 52344)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(res, convey.ShouldNotBeNil)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoOwGames(t *testing.T) {
	convey.Convey("OwGames", t, func(ctx convey.C) {
		_, err := d.OwGames(context.Background(), 52344)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoAddLolGame(t *testing.T) {
	var players []*mdlesp.LolGame
	convey.Convey("AddDotaTeam", t, func(ctx convey.C) {
		player := &mdlesp.LolGame{ID: 202830, MatchID: 524686, Finished: true, Position: 1}
		players = append(players, player)
		err := d.AddLolGame(context.Background(), players)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoAddDotaGame(t *testing.T) {
	var players []*mdlesp.DotaGame
	convey.Convey("AddDotaGame", t, func(ctx convey.C) {
		player := &mdlesp.DotaGame{ID: 35990, MatchID: 52344, Finished: true, Position: 1}
		players = append(players, player)
		err := d.AddDotaGame(context.Background(), players)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoAddOwGame(t *testing.T) {
	var players []*mdlesp.OwGame
	convey.Convey("AddOwGame", t, func(ctx convey.C) {
		player := &mdlesp.OwGame{ID: 3718, MatchID: 542674, Finished: true, Position: 1}
		players = append(players, player)
		err := d.AddOwGame(context.Background(), players)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoUpLolGame(t *testing.T) {
	convey.Convey("UpLolGame", t, func(ctx convey.C) {
		player := &mdlesp.LolGame{ID: 202830, MatchID: 524686, Finished: true, Position: 1}
		err := d.UpLolGame(context.Background(), player)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoUpDotaGame(t *testing.T) {
	convey.Convey("UpDotaGame", t, func(ctx convey.C) {
		player := &mdlesp.DotaGame{ID: 35990, MatchID: 52344, Finished: true, Position: 1}
		err := d.UpDotaGame(context.Background(), player)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoUpOwGame(t *testing.T) {
	convey.Convey("UpOwGame", t, func(ctx convey.C) {
		player := &mdlesp.OwGame{ID: 3718, MatchID: 542674, Finished: true, Position: 1}
		err := d.UpOwGame(context.Background(), player)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoAddLolCham(t *testing.T) {
	var infos []*mdlesp.LdInfo
	convey.Convey("AddLolCham", t, func(ctx convey.C) {
		info := &mdlesp.LdInfo{ID: 101, Name: "cham", ImageURL: ""}
		infos = append(infos, info)
		err := d.AddLolCham(context.Background(), infos)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoAddDotaHero(t *testing.T) {
	var infos []*mdlesp.LdInfo
	convey.Convey("AddDotaHero", t, func(ctx convey.C) {
		info := &mdlesp.LdInfo{ID: 102, Name: "hero", ImageURL: ""}
		infos = append(infos, info)
		err := d.AddDotaHero(context.Background(), infos)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoAddOwHero(t *testing.T) {
	var infos []*mdlesp.LdInfo
	convey.Convey("AddOwHero", t, func(ctx convey.C) {
		info := &mdlesp.LdInfo{ID: 103, Name: "hero", ImageURL: ""}
		infos = append(infos, info)
		err := d.AddOwHero(context.Background(), infos)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoAddLolItem(t *testing.T) {
	var infos []*mdlesp.LdInfo
	convey.Convey("AddLolItem", t, func(ctx convey.C) {
		info := &mdlesp.LdInfo{ID: 201, Name: "item", ImageURL: ""}
		infos = append(infos, info)
		err := d.AddLolItem(context.Background(), infos)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoAddDotaItem(t *testing.T) {
	var infos []*mdlesp.LdInfo
	convey.Convey("AddDotaItem", t, func(ctx convey.C) {
		info := &mdlesp.LdInfo{ID: 202, Name: "item", ImageURL: ""}
		infos = append(infos, info)
		err := d.AddDotaItem(context.Background(), infos)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoAddOwMap(t *testing.T) {
	var infos []*mdlesp.LdInfo
	convey.Convey("AddLolItem", t, func(ctx convey.C) {
		info := &mdlesp.LdInfo{ID: 203, Name: "map", ImageURL: ""}
		infos = append(infos, info)
		err := d.AddOwMap(context.Background(), infos)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoAddLolMatchPlayer(t *testing.T) {
	var infos []*mdlesp.LdInfo
	convey.Convey("AddLolMatchPlayer", t, func(ctx convey.C) {
		info := &mdlesp.LdInfo{ID: 301, Name: "matchPlayer", ImageURL: ""}
		infos = append(infos, info)
		err := d.AddLolMatchPlayer(context.Background(), infos)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoAddDotaMatchPlayer(t *testing.T) {
	var infos []*mdlesp.LdInfo
	convey.Convey("AddDotaMatchPlayer", t, func(ctx convey.C) {
		info := &mdlesp.LdInfo{ID: 302, Name: "matchPlayer", ImageURL: ""}
		infos = append(infos, info)
		err := d.AddDotaMatchPlayer(context.Background(), infos)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoAddOwMatchPlayer(t *testing.T) {
	var infos []*mdlesp.LdInfo
	convey.Convey("AddDotaMatchPlayer", t, func(ctx convey.C) {
		info := &mdlesp.LdInfo{ID: 303, Name: "matchPlayer", ImageURL: ""}
		infos = append(infos, info)
		err := d.AddOwMatchPlayer(context.Background(), infos)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoAddLolAbility(t *testing.T) {
	var infos []*mdlesp.LdInfo
	convey.Convey("AddLolAbility", t, func(ctx convey.C) {
		info := &mdlesp.LdInfo{ID: 401, Name: "ability", ImageURL: ""}
		infos = append(infos, info)
		err := d.AddLolAbility(context.Background(), infos)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoAddDotaAbility(t *testing.T) {
	var infos []*mdlesp.LdInfo
	convey.Convey("AddLolAbility", t, func(ctx convey.C) {
		info := &mdlesp.LdInfo{ID: 402, Name: "ability", ImageURL: ""}
		infos = append(infos, info)
		err := d.AddDotaAbility(context.Background(), infos)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDao_AutoAdd(t *testing.T) {
	var (
		aid      int64 = 999999
		mid      int64 = 123
		tid      int64 = 12
		tags           = "4,5,6"
		keywords       = "2,3,4"
		gameIDs        = []int64{10, 11, 12}
		matchIDs       = []int64{20, 21, 22}
		teamIDs        = []int64{20, 21, 22}
	)
	convey.Convey("AutoAdd", t, func(ctx convey.C) {
		err := d.AutoAdd(context.Background(), aid, mid, tid, tags, keywords, gameIDs, matchIDs, teamIDs, 2020, 1)
		ctx.Convey("Then res should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
