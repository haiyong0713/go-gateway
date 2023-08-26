package service

import (
	"context"
	"go-gateway/app/web-svr/esports/admin/component"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTeamsInSeason(t *testing.T) {
	if err := component.GlobalDB.DB().Ping(); err != nil {
		t.Fatal(err)
	}
	//init global service
	init := WithService(func(s *Service) {})
	init()

	t.Run("add team to season", TestAddTeamToSeason)
	t.Run("list team in season", TestListTeamsInSeason)
	t.Run("del team from season", TestDelTeamInSeason)
}

func TestAddTeamToSeason(t *testing.T) {
	assert.Equal(t, nil, svf.AddTeamToSeason(context.Background(), 1, 1, 22))
}

func TestDelTeamInSeason(t *testing.T) {
	assert.Equal(t, nil, svf.RemoveTeamFromSeason(context.Background(), 1, 1))
	teams, err := svf.ListTeamInSeason(context.Background(), 1)
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, len(teams))
}

func TestListTeamsInSeason(t *testing.T) {
	teams, err := svf.ListTeamInSeason(context.Background(), 1)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(teams))
	assert.Equal(t, int64(22), teams[0].Rank)
}
