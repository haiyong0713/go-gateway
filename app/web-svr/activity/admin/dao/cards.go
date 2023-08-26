package dao

import (
	"context"
	"database/sql"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/cards"
)

const (
	_createMidCards = "CREATE TABLE IF NOT EXISTS act_cards_nums_%d LIKE act_cards_nums"
	_addMidCards    = "INSERT INTO act_cards(name,lottery_id,cards_num,cards,sid,reserve_id) VALUE (?,?,?,?,?,?)"
)

func (d *Dao) AddCards(ctx context.Context, card *cards.Cards) (id int64, err error) {
	var (
		result sql.Result
	)
	if result, err = d.db.Exec(ctx, _addMidCards, card.Name, card.LotteryID, card.CardsNum, card.Cards, card.SID, card.ReserveID); err != nil {
		log.Errorc(ctx, "lottery@Add d.db.Exec() INSERT failed. error(%v)", err)
		return
	}
	if id, err = result.LastInsertId(); err != nil {
		log.Error("lottery@Add result.LastInsertId() failed. error(%v)", err)
		return
	}
	return
}

func (d *Dao) CreateMidCards(ctx context.Context, id int64) (err error) {
	if _, err = d.db.Exec(ctx, fmt.Sprintf(_createMidCards, id)); err != nil {
		log.Errorc(ctx, "lottery@CreateMidCards CREATE TABLE failed. error(%v)", err)
	}
	return
}
