package dao

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"

	"go-gateway/app/web-svr/dance-taiko/interface/ecode"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"
)

const (
	_games = "SELECT id,game_id,aid,status,stime FROM dance_game WHERE game_id=? AND deleted=0"
)

// RawGames get games from db
func (d *dao) RawGame(c context.Context, gid int64) (*model.Game, error) {
	g := new(model.Game)
	err := d.db.QueryRow(c, _games, gid).Scan(&g.ID, &g.GameID, &g.AID, &g.Status, &g.Stime)
	if err == sql.ErrNoRows {
		return nil, ecode.GameIDErr
	}
	if err != nil {
		return nil, errors.Wrapf(err, "RawGame ID %d", gid)
	}
	return g, nil
}
