package http

import (
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/binding"
	"go-gateway/app/web-svr/esports/admin/client"
	"go-gateway/app/web-svr/esports/admin/model"
	v1 "go-gateway/app/web-svr/esports/interface/api/v1"
)

type Paging struct {
	Num      int64 `form:"pn" validate:"min=1" default:"1"`
	Size     int64 `form:"ps" validate:"min=0" default:"10"`
	SeasonID int64 `form:"season_id" validate:"min=1"`
}

func contestSeriesList(ctx *bm.Context) {
	paging := new(Paging)
	if err := ctx.Bind(paging); err != nil {
		ctx.JSON(nil, err)

		return
	}

	data, err := esSvc.ContestSeriesPaging(ctx, paging.SeasonID, paging.Size, paging.Num)
	if err != nil {
		ctx.JSON(nil, err)

		return
	}

	ctx.JSON(data, nil)
}

func deleteContestSeries(ctx *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	d, err := esSvc.DeleteContestSeriesByID(ctx, v.ID)
	if err != nil {
		ctx.JSON(nil, err)

		return
	}

	ctx.JSON(d, nil)
}

func fetchContestSeries(ctx *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	d, err := esSvc.FetchContestSeriesByID(ctx, v.ID)
	if err != nil {
		ctx.JSON(nil, err)

		return
	}

	ctx.JSON(d, nil)
}

func createContestSeries(ctx *bm.Context) {
	series := new(model.ContestSeries)
	if err := ctx.Bind(series); err != nil {
		ctx.JSON(nil, err)

		return
	}

	if err := esSvc.AddContestSeries(ctx, series); err != nil {
		ctx.JSON(nil, err)

		return
	}

	ctx.JSON(nil, nil)
}

func updateContestSeries(ctx *bm.Context) {
	series := new(model.ContestSeries)
	if err := ctx.Bind(series); err != nil {
		ctx.JSON(nil, err)

		return
	}

	if err := esSvc.UpdateContestSeries(ctx, series); err != nil {
		ctx.JSON(nil, err)

		return
	}

	ctx.JSON(nil, nil)
}

func addContestSeriesPointMatchConfig(ctx *bm.Context) {
	v := new(v1.SeriesPointMatchConfig)
	if err := ctx.BindWith(v, binding.JSON); err != nil {
		return
	}
	ctx.JSON(client.EsportsGrpcClient.AddSeriesPointMatchConfig(ctx, v))
}

func getContestSeriesPointMatchConfig(ctx *bm.Context) {
	v := new(v1.GetSeriesPointMatchReq)
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.EsportsGrpcClient.GetSeriesPointMatchConfig(ctx, v))
}

func updateContestSeriesPointMatchConfig(ctx *bm.Context) {
	v := new(v1.SeriesPointMatchConfig)
	if err := ctx.BindWith(v, binding.JSON); err != nil {
		return
	}
	ctx.JSON(client.EsportsGrpcClient.UpdateSeriesPointMatchConfig(ctx, v))
}

func previewPointMatchInfo(ctx *bm.Context) {
	v := new(v1.SeriesPointMatchConfig)
	if err := ctx.BindWith(v, binding.JSON); err != nil {
		return
	}
	ctx.JSON(client.EsportsGrpcClient.PreviewSeriesPointMatchInfo(ctx, v))
}

func getPointMatchInfo(ctx *bm.Context) {
	v := new(v1.GetSeriesPointMatchInfoReq)
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.EsportsGrpcClient.GetSeriesPointMatchInfo(ctx, v))
}

func refreshPointMatchInfo(ctx *bm.Context) {
	v := new(v1.RefreshSeriesPointMatchInfoReq)
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.EsportsGrpcClient.RefreshSeriesPointMatchInfo(ctx, v))
}

func addContestSeriesKnockoutMatchConfig(ctx *bm.Context) {
	v := new(v1.SeriesKnockoutMatchConfig)
	if err := ctx.BindWith(v, binding.JSON); err != nil {
		return
	}
	ctx.JSON(client.EsportsGrpcClient.AddSeriesKnockoutMatchConfig(ctx, v))
}

func getContestSeriesKnockoutMatchConfig(ctx *bm.Context) {
	v := new(v1.GetSeriesKnockoutMatchConfigReq)
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.EsportsGrpcClient.GetSeriesKnockoutMatchConfig(ctx, v))
}

func updateContestSeriesKnockoutMatchConfig(ctx *bm.Context) {
	v := new(v1.SeriesKnockoutMatchConfig)
	if err := ctx.BindWith(v, binding.JSON); err != nil {
		return
	}
	ctx.JSON(client.EsportsGrpcClient.UpdateSeriesKnockoutMatchConfig(ctx, v))
}

func previewKnockoutMatchInfo(ctx *bm.Context) {
	v := new(v1.SeriesKnockoutMatchConfig)
	if err := ctx.BindWith(v, binding.JSON); err != nil {
		return
	}
	ctx.JSON(client.EsportsGrpcClient.PreviewSeriesKnockoutMatchInfo(ctx, v))
}

func getKnockoutMatchInfo(ctx *bm.Context) {
	v := new(v1.GetSeriesKnockoutMatchInfoReq)
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.EsportsGrpcClient.GetSeriesKnockoutMatchInfo(ctx, v))
}

func refreshKnockoutMatchInfo(ctx *bm.Context) {
	v := new(v1.RefreshSeriesKnockoutMatchInfoReq)
	if err := ctx.Bind(v); err != nil {
		return
	}
	ctx.JSON(client.EsportsGrpcClient.RefreshSeriesKnockoutMatchInfo(ctx, v))
}

func getScoreRules(ctx *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	d, err := esSvc.GetScoreRules(ctx, v.ID)
	if err != nil {
		ctx.JSON(nil, err)

		return
	}
	res := &model.PUBGContestSeriesExtraRules{}
	res.ScoreRules = *d
	ctx.JSON(res, nil)
}

func saveScoreRules(ctx *bm.Context) {
	v := new(struct {
		ID         int64   `json:"id" form:"id" validate:"min=1"`
		RankScores []int64 `json:"rank_scores" form:"rank_scores" validate:"required"`
		KillScore  int64   `json:"kill_score" form:"kill_score" validate:"required"`
	})
	if err := ctx.BindWith(v, binding.JSON); err != nil {
		return
	}
	err := esSvc.SaveScoreRules(ctx, v.ID, &model.PUBGContestSeriesScoreRule{
		KillScore:  v.KillScore,
		RankScores: v.RankScores,
	})
	if err != nil {
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(nil, nil)
}
