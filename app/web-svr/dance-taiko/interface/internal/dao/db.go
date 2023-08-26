package dao

import (
	"context"
	"time"

	"go-common/library/conf/paladin"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"
)

const (
	_insertGame       = "INSERT IGNORE INTO dance_game(game_id,aid,status) VALUES (?,?,?)"
	_currentGame      = "SELECT id,game_id,aid,status FROM dance_game WHERE deleted=0 ORDER BY id DESC LIMIT 1"
	_startGame        = "UPDATE dance_game SET status=?,stime=? WHERE game_id=?"
	_updateGameStatus = "UPDATE dance_game SET status=? WHERE game_id=?"
)

func NewDB() (db *sql.DB, cf func(), err error) {
	var (
		cfg sql.Config
		ct  paladin.TOML
	)
	if err = paladin.Get("db.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Client").UnmarshalTOML(&cfg); err != nil {
		return
	}
	db = sql.NewMySQL(&cfg)
	cf = func() { db.Close() }
	return
}

func (d *dao) CreateGame(c context.Context, aid int64, gameId int64) error {
	if _, err := d.db.Exec(c, _insertGame, gameId, aid, model.GameJoining); err != nil {
		log.Error("insert game into mysql fail error: %v", err)
		return err
	}
	return nil
}

func (d *dao) CurrentGame(c context.Context) (game *model.Game, err error) {
	game = &model.Game{}
	row := d.db.QueryRow(c, _currentGame)
	err = row.Scan(&game.ID, &game.GameID, &game.AID, &game.Status)
	if err != nil {
		log.Error("get latest game from mysql fail error: %v", err)
		return nil, err
	}
	return
}

func (d *dao) UpdateGameStatus(c context.Context, gameId int64, status string) error {
	if _, err := d.db.Exec(c, _updateGameStatus, status, gameId); err != nil {
		log.Error("update game status fail error: %v", err)
		return err
	}
	return nil
}

func (d *dao) StartGame(c context.Context, gameId int64, status string) error {
	sTime := time.Now().UnixNano() / int64(time.Millisecond)
	if _, err := d.db.Exec(c, _startGame, status, sTime, gameId); err != nil {
		log.Error("update game status fail error: %v", err)
		return err
	}
	if err := d.RedisSetSTime(c, gameId, sTime); err != nil {
		log.Error("update redis game status fail error: %v", err)
		return err
	}
	return nil
}
