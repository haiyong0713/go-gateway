package ott

import (
	"context"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"
	"time"

	"go-common/library/log"

	"github.com/pkg/errors"
)

const (
	_selectGamesByCid = "SELECT `id` FROM `dance_game` WHERE `cid`=? AND `stime` > ? AND `deleted`=0"
	_addGame          = "INSERT INTO `dance_game`(`aid`, `cid`, `status`) VALUES(?,?,?)"
	_startGame        = "UPDATE `dance_game` SET `status`=?, `stime`=? WHERE `id`=?"
	_selectGameById   = "SELECT `id`,`aid`,`cid`,`status`,`stime` FROM `dance_game` WHERE `id`=?"
	_finishGame       = "UPDATE `dance_game` SET `status`=? WHERE `id`=?"
)

func (d *dao) SelectGamesByCid(c context.Context, cid int64) ([]int64, error) {
	rows, err := d.db.Query(c, _selectGamesByCid, cid, getFirstDateOfWeek())
	if err != nil {
		return nil, errors.Wrapf(err, "SelectGameByCid cid(%d)", cid)
	}
	defer rows.Close()
	var res []int64
	for rows.Next() {
		var gameId int64
		if err := rows.Scan(&gameId); err != nil {
			log.Warn("SelectGameByCid cid(%d) err(%v)", cid, err)
			continue
		}
		res = append(res, gameId)
	}
	return res, nil
}

func (d *dao) CreateGame(c context.Context, aid, cid int64) (int64, error) {
	reply, err := d.db.Exec(c, _addGame, aid, cid, model.GameJoining)
	if err != nil {
		return 0, errors.Wrapf(err, "AddGame aid(%d) cid(%d)", aid, cid)
	}
	gameId, err := reply.LastInsertId()
	if err != nil {
		return 0, errors.Wrapf(err, "AddGame aid(%d) cid(%d)", aid, cid)
	}
	return gameId, nil
}

func (d *dao) StartGame(c context.Context, id int64) error {
	stime := time.Now().UnixNano() / int64(time.Millisecond)
	_, err := d.db.Exec(c, _startGame, model.GamePlaying, stime, id)
	if err != nil {
		return errors.Wrapf(err, "StartGame id(%d)", id)
	}
	return nil
}

func (d *dao) rawGame(c context.Context, id int64) (*model.OttGame, error) {
	var res = new(model.OttGame)
	if err := d.db.QueryRow(c, _selectGameById, id).Scan(&res.GameId, &res.Aid, &res.Cid, &res.Status, &res.Stime); err != nil {
		return nil, errors.Wrapf(err, "RawGame id(%d)", id)
	}
	return res, nil
}

func (d *dao) FinishGame(c context.Context, id int64) error {
	if _, err := d.db.Exec(c, _finishGame, model.GameFinished, id); err != nil {
		return errors.Wrapf(err, "FinishGame id(%d)", id)
	}
	return nil
}
