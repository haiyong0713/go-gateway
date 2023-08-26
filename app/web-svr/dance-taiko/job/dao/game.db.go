package dao

import (
	"context"
	"fmt"
	"strings"

	"go-gateway/app/web-svr/dance-taiko/job/model"

	"github.com/pkg/errors"
)

const (
	_selectPlayersByGame = "SELECT `mid`, `score` FROM `dance_players` WHERE game_id=? AND `is_deleted`=0"
	_updatePlayerScore   = "INSERT INTO `dance_players` (`game_id`, `mid`, `score`) VALUES %s ON DUPLICATE KEY UPDATE `score` = VALUES(`score`)"
	_playerScoreValues   = "(%d, %d, %d)"
	_gamesByStatus       = "SELECT id,aid,cid,status,stime from dance_game WHERE status=? AND deleted=0"
)

func (d *Dao) RawPlayers(c context.Context, gameId int64) ([]*model.PlayerHonor, error) {
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
		if player.Mid > 0 {
			res = append(res, player)
		}
	}
	return res, nil
}

func (d *Dao) PlayersMap(c context.Context, gameId int64) (map[int64]*model.PlayerHonor, error) {
	playerList, err := d.RawPlayers(c, gameId)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]*model.PlayerHonor)
	for _, player := range playerList {
		res[player.Mid] = player
	}
	return res, nil
}

func (d *Dao) UpdatePlayers(c context.Context, gameId int64, players []*model.PlayerHonor) error {
	var values []string
	for _, player := range players {
		values = append(values, fmt.Sprintf(_playerScoreValues, gameId, player.Mid, player.Score))
	}
	if _, err := d.db.Exec(c, fmt.Sprintf(_updatePlayerScore, strings.Join(values, ","))); err != nil {
		return errors.Wrapf(err, "UpdatePlayers gameId(%d) players(%v)", gameId, players)
	}
	return nil
}

func (d *Dao) GamesByStatus(c context.Context, status string) ([]model.OttGame, error) {
	rows, err := d.db.Query(c, _gamesByStatus, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res = make([]model.OttGame, 0)
	for rows.Next() {
		g := model.OttGame{}

		// SELECT id,aid,cid,status,stime from dance_game WHERE status=? AND deleted=0
		if err := rows.Scan(&g.GameId, &g.Aid, &g.Cid, &g.Status, &g.Stime); err != nil {
			return nil, err
		}
		res = append(res, g)
	}
	return res, nil
}
