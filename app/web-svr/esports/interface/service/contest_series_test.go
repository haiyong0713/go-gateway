package service

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/interface/dao"
	"go-gateway/app/web-svr/esports/interface/model"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getRandInt64(max, not int64) (v int64) {
	for {
		v = rand.NewSource(time.Now().UnixNano()).Int63() % max
		if v != not {
			return
		}
		time.Sleep(time.Nanosecond)
	}
}

// go test -v -count=1 service_test.go contest_series_test.go service.go contest_series.go auto_subscribe.go s10_tab.go s10_score_analysis.go s10.go grpc.go match.go season.go match_active.go live.go favorite.go guess.go pointdata.go
func TestSeriesPointMatchInfo(t *testing.T) {
	teams := make(map[int64]*model.Team, 0)
	for id := int64(1); id <= int64(20); id++ {
		idInner := id
		teams[idInner] = &model.Team{
			ID:    idInner,
			Title: fmt.Sprintf("Team-%v", idInner),
		}
	}
	contests := make(map[int64]*model.Contest, 0)
	for id := 1; id <= 100; id++ {
		homeId := getRandInt64(20, 0)
		awayId := getRandInt64(20, homeId)
		contests[int64(id)] = &model.Contest{
			Etime:     1614829665,
			HomeID:    homeId,
			AwayID:    awayId,
			HomeScore: getRandInt64(5, 0),
			AwayScore: getRandInt64(5, 0),
		}
	}

	req := &v1.SeriesPointMatchConfig{
		SeasonId:           1,
		SeriesId:           1,
		ScoreIncrWin:       1,
		ScoreDecrLose:      -1,
		SmallScoreIncrWin:  1,
		SmallScoreDecrLose: -1,
		UseTeamGroup:       true,
		Teams:              make([]*v1.SeriesPointMatchTeamConfig, 0),
	}
	for _, t := range teams {
		req.Teams = append(req.Teams, &v1.SeriesPointMatchTeamConfig{
			Tid:      t.ID,
			Group:    fmt.Sprintf("group-%v", t.ID%5),
			Priority: getRandInt64(5, 0),
		})
	}
	res, err := svr.innerGeneratePreviewPointMatchInfo(context.Background(), req, teams, contests)
	assert.Equal(t, nil, err)
	bs1, _ := json.Marshal(req)
	fmt.Println(string(bs1))
	bs2, _ := json.Marshal(res)
	fmt.Println(string(bs2))

	err = svr.dao.SetSeriesPointMatchInfo(context.Background(), res)
	assert.Equal(t, nil, err)
	res2, _, err := svr.dao.GetSeriesPointMatchInfo(context.Background(), req.SeriesId)
	assert.Equal(t, nil, err)
	bs3, _ := json.Marshal(res2)
	fmt.Println(string(bs3))
	assert.Equal(t, bs2, bs3)
}

func TestSeriesRefreshing(t *testing.T) {
	ctx := context.Background()
	ok, err := svr.dao.MarkSeriesRefreshing(ctx, 1, dao.SeriesTypPoint)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, ok)
	ok, err = svr.dao.MarkSeriesRefreshing(ctx, 1, dao.SeriesTypPoint)
	assert.Equal(t, nil, err)
	assert.Equal(t, false, ok)
	time.Sleep(time.Second * 2)
	ok, err = svr.dao.MarkSeriesRefreshing(ctx, 1, dao.SeriesTypPoint)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, ok)

}
