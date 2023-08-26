package guess

import (
	"context"

	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/guess"
)

//go:generate kratos tool btsgen
type _bts interface {
	// cache
	UserStat(c context.Context, mid int64, stakeType int64, business int64) (*api.UserGuessDataReply, error)
	// cache
	GuessMain(c context.Context, mainID int64) (*guess.MainGuess, error)
	// cache
	UserGuessList(c context.Context, mid int64, business int64) ([]*guess.UserGuessLog, error)
	// cache
	MDResult(c context.Context, id int64, business int64) (*guess.MainRes, error)
	// cache
	MDsResult(c context.Context, ids []int64, business int64) (map[int64]*guess.MainRes, error)
	// cache
	OidMIDs(c context.Context, oid int64, business int64) ([]*guess.MainID, error)
	// cache
	OidsMIDs(c context.Context, oids []int64, business int64) (map[int64][]*guess.MainID, error)
	// cache
	UserGuess(c context.Context, mainIDs []int64, mid int64) (map[int64]*guess.UserGuessLog, error)
}
