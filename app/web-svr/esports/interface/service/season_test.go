package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// go test -v season_test.go auto_subscribe.go favorite.go guess.go grpc.go live.go \
// match.go match_active.go pointdata.go s10.go s10_score_analysis.go s10_tab.go \
// s9.go search.go season.go service.go service_test.go -vet off
func TestTeamsInSeasonBiz(t *testing.T) {
	go svr.AsyncUpdateOngoingSeasonTeamInMemoryCache()
	t.Run("test get teams in season", testGetTeamInSeason)
	t.Run("test teams in season memory cache", testTeamsInSeasonMemoryCache)
}

func testGetTeamInSeason(t *testing.T) {
	//mysql> select * from es_team_in_seasons;
	//	    +-----+-----+------+
	//	    | sid | tid | rank |
	//		+-----+-----+------+
	//    	|   1 |   1 |   10 |
	//		+-----+-----+------+
	//		1 row in set (0.00 sec)
	teams, err := svr.GetTeamsInSeason(context.Background(), 1)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(teams))
	assert.Equal(t, int64(10), teams[0].Rank)
	assert.Equal(t, int64(1), teams[0].SeasonId)
	assert.Equal(t, int64(1), teams[0].TeamId)
}

func testTeamsInSeasonMemoryCache(t *testing.T) {
	//sleep 15 second to wait memory init
	time.Sleep(15 * time.Second)
	//season 1 is ongoing, should exists in memory
	teamsInSeasonMap := svr.teamsInSeasonMap
	_, exist := teamsInSeasonMap[1]
	assert.Equal(t, true, exist)
}
