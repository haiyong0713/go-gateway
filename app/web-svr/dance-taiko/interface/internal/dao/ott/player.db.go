package ott

import (
	"context"
	"fmt"

	"go-common/library/xstr"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"

	"github.com/pkg/errors"
)

const (
	_selectScoreByGames  = "SELECT `mid`, `score` FROM `dance_players` WHERE game_id IN (%s) AND `is_deleted`=0 ORDER BY `score` DESC"
	_selectPlayersByGame = "SELECT `mid`, `score` FROM `dance_players` WHERE game_id=? AND `is_deleted`=0"
	_addPlayer           = "INSERT INTO `dance_players` (`mid`,`game_id`) VALUES(?,?)"
)

func (d *dao) SelectPlayersByGames(c context.Context, gameIds []int64) ([]*model.PlayerHonor, error) {
	if len(gameIds) == 0 {
		return nil, nil
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_selectScoreByGames, xstr.JoinInts(gameIds)))
	if err != nil {
		return nil, errors.Wrapf(err, "SelectPlayersByGames gameIds(%v)", gameIds)
	}
	defer rows.Close()
	var res = make([]*model.PlayerHonor, 0)
	for rows.Next() {
		var playerHonor = new(model.PlayerHonor)
		if err := rows.Scan(&playerHonor.Mid, &playerHonor.Score); err != nil {
			return nil, errors.Wrapf(err, "SelectPlayersByGames gameIds(%v)", gameIds)
		}
		res = append(res, playerHonor)
	}
	return res, nil
}

func (d *dao) RawPlayers(c context.Context, gameId int64) ([]*model.PlayerHonor, error) {
	rows, err := d.db.Query(c, _selectPlayersByGame, gameId)
	if err != nil {
		return nil, errors.Wrapf(err, "RawPlayers gameId(%d)", gameId)
	}
	defer rows.Close()
	var res = make([]*model.PlayerHonor, 0)
	for rows.Next() {
		var player = new(model.PlayerHonor)
		if err := rows.Scan(&player.Mid, &player.Score); err != nil {
			return nil, errors.Wrapf(err, "RawPlayers gameId(%d)", gameId)
		}
		if player != nil && player.Mid > 0 {
			res = append(res, player)
		}
	}
	return res, nil
}

func (d *dao) AddPlayer(c context.Context, gameId, mid int64) error {
	if _, err := d.db.Exec(c, _addPlayer, mid, gameId); err != nil {
		return errors.Wrapf(err, "AddPlayer gameId(%d) mid(%d)", gameId, mid)
	}
	return nil
}
