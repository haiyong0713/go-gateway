package game

import (
	"testing"

	"go-gateway/app/app-svr/app-card/interface/model/card/game"

	"github.com/stretchr/testify/assert"
)

func TestDeriveMaterialsIds(t *testing.T) {
	gp1 := []*game.GameParam{
		{GameId: 1, CreativeId: 1},
		{GameId: 2, CreativeId: 2},
	}
	var gp2 []*game.GameParam
	assert.Equal(t, deriveMaterialsIds(gp1), "1_1,2_2")
	assert.Equal(t, deriveMaterialsIds(gp2), "")

}
