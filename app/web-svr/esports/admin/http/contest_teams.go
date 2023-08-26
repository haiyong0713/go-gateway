package http

import (
	xecode "go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/esports/admin/model"
)

func contestTeams(ctx *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	d, err := esSvc.GetContestTeams(ctx, v.ID)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	teamIds := make([]int64, 0)
	for _, team := range d {
		teamIds = append(teamIds, team.TeamId)
	}
	ctx.JSON(xstr.JoinInts(teamIds), nil)
}

func contestTeamsSave(ctx *bm.Context) {
	v := new(struct {
		ID      int64  `form:"id" validate:"min=1"`
		Sid     int64  `form:"sid" validate:"min=1"`
		TeamIds string `form:"team_ids"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	err := esSvc.ContestTeamsAdd(ctx, v.ID, v.Sid, v.TeamIds)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(nil, nil)
}

func contestTeamsUpdate(ctx *bm.Context) {
	v := new(struct {
		ID      int64  `form:"id" validate:"min=1"`
		Sid     int64  `form:"sid" validate:"min=1"`
		TeamIds string `form:"team_ids"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	err := esSvc.ContestTeamsUpdate(ctx, &model.Contest{
		ID:            v.ID,
		GameStage:     "",
		Stime:         0,
		Etime:         0,
		HomeID:        0,
		AwayID:        0,
		HomeScore:     0,
		AwayScore:     0,
		LiveRoom:      0,
		Aid:           0,
		Collection:    0,
		GameState:     0,
		Dic:           "",
		Status:        0,
		Sid:           v.Sid,
		Mid:           0,
		Special:       0,
		SuccessTeam:   0,
		SpecialName:   "",
		SpecialTips:   "",
		SpecialImage:  "",
		Playback:      "",
		CollectionURL: "",
		LiveURL:       "",
		DataType:      0,
		Data:          "",
		Adid:          0,
		MatchID:       0,
		GuessType:     0,
		GameStage1:    "",
		GameStage2:    "",
		SeriesID:      0,
		PushSwitch:    0,
		TeamIds:       v.TeamIds,
	}, nil)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(nil, err)
}

func contestTeamsCheck(ctx *bm.Context) {
	v := new(struct {
		Sid     int64  `form:"sid" validate:"min=1"`
		TeamIds string `form:"team_ids" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	teamIds, err := xstr.SplitInts(v.TeamIds)
	if err != nil {
		ctx.JSON(xecode.RequestErr, err)
		return
	}
	d, err := esSvc.ContestTeamsCheck(ctx, v.Sid, teamIds)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(d, nil)
}

func contestTeamScores(ctx *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	d, err := esSvc.GetContestTeamsOrderBySurvivalRank(ctx, v.ID)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(d, nil)
}

func contestTeamScoresSave(ctx *bm.Context) {
	v := new(struct {
		ID         int64               `json:"id" form:"id" validate:"min=1"`
		TeamScores []*model.TeamScores `json:"team_scores" form:"team_scores" validate:"required"`
	})
	if err := ctx.BindWith(v, binding.JSON); err != nil {
		return
	}
	err := esSvc.SaveContestTeamsBySurvivalRank(ctx, v.ID, v.TeamScores)
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(nil, nil)
}
