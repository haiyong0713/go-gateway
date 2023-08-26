package card

import (
	"context"
	"encoding/json"
	"go-gateway/app/app-svr/app-show/interface/component"
	"strconv"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-show/interface/conf"
	"go-gateway/app/app-svr/app-show/interface/model/card"
	"go-gateway/app/app-svr/app-show/interface/model/selected"

	"github.com/siddontang/go-mysql/mysql"
)

const (
	// hot card
	_cardSQL = `SELECT c.id,c.title,c.card_type,c.card_value,c.recommand_reason,c.recommand_state,c.priority FROM popular_card AS c
	WHERE c.stime<? AND c.etime>? AND c.check=2 AND c.is_delete=0 ORDER BY c.priority ASC`
	_cardPlatSQL = `SELECT card_id,plat,conditions,build FROM popular_card_plat WHERE is_delete=0`
	// selected series
	_selectedSeriesSQL = "SELECT id, `type`, number, `subject`, stime, etime, status FROM selected_serie WHERE type = ? AND deleted = 0 AND `status` IN (2,4) AND pubtime <= ? ORDER BY number DESC"                                                       // 2=审核通过, 4=灾备数据, pubtime控制发布时间
	_serieConfigSQL    = "SELECT id, `type`, number, `subject`, stime, etime, hint, color, cover, share_title, share_subtitle, `status`,media_id FROM selected_serie WHERE type = ? AND number = ? AND deleted = 0 AND `status` IN (2,4) AND pubtime <= ?" // 获取配置只看通过的，灾备的配置走配置文件
	// selected resources
	_selectedResourceSQL = "SELECT rid, rtype, serie_id, position, rcmd_reason FROM selected_resource WHERE (serie_id = ? AND `status` = 1 AND deleted = 0) ORDER BY position" // status = 1 means passed
)

// Dao is card dao.
type Dao struct {
	db *sql.DB
	// redis
	redis *redis.Pool
	conf  *conf.Config
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		db: component.GlobalShowDB,
		// redis
		redis: redis.NewPool(c.Redis.Recommend.Config),
		conf:  c,
	}
	return d
}

// Columns
func (d *Dao) Columns(ctx context.Context) (map[int8][]*card.Column, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", cardActionKey("loadColumnsCache", "column")))
	if err != nil {
		return nil, err
	}
	var res map[int8][]*card.Column
	if err = json.Unmarshal(reply, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// PosRecs
func (d *Dao) PosRecs(ctx context.Context) (map[int8]map[int][]*card.Card, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", cardActionKey("loadCardCache", "card")))
	if err != nil {
		return nil, err
	}
	var res map[int8]map[int][]*card.Card
	if err = json.Unmarshal(reply, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// RecContents
func (d *Dao) RecContents(ctx context.Context) (map[int][]*card.Content, map[int][]int64, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", cardActionKey("loadCardCache", "content")))
	if err != nil {
		return nil, nil, err
	}
	res := map[int][]*card.Content{}
	if err = json.Unmarshal(reply, &res); err != nil {
		return nil, nil, err
	}
	reply, err = redis.Bytes(conn.Do("GET", cardActionKey("loadCardCache", "aids")))
	if err != nil {
		return nil, nil, err
	}
	aids := map[int][]int64{}
	if err = json.Unmarshal(reply, &aids); err != nil {
		return nil, nil, err
	}
	return res, aids, nil
}

// NperContents
func (d *Dao) NperContents(ctx context.Context) (map[int][]*card.Content, map[int][]int64, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", cardActionKey("loadNperCache", "content")))
	if err != nil {
		return nil, nil, err
	}
	res := map[int][]*card.Content{}
	if err = json.Unmarshal(reply, &res); err != nil {
		return nil, nil, err
	}
	reply, err = redis.Bytes(conn.Do("GET", cardActionKey("loadNperCache", "aids")))
	if err != nil {
		return nil, nil, err
	}
	aids := map[int][]int64{}
	if err = json.Unmarshal(reply, &aids); err != nil {
		return nil, nil, err
	}
	return res, aids, nil
}

// ColumnNpers
func (d *Dao) ColumnNpers(ctx context.Context) (map[int8][]*card.ColumnNper, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", cardActionKey("loadNperCache", "columnNper")))
	if err != nil {
		return nil, err
	}
	var res map[int8][]*card.ColumnNper
	if err = json.Unmarshal(reply, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// ColumnList
func (d *Dao) ColumnList(ctx context.Context) ([]*card.ColumnList, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", cardActionKey("loadColumnListCache", "columnList")))
	if err != nil {
		return nil, err
	}
	var res []*card.ColumnList
	if err := json.Unmarshal(reply, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// Card channel card
func (d *Dao) Card(ctx context.Context, now time.Time) (res []*card.PopularCard, err error) {
	rows, err := d.db.Query(ctx, _cardSQL, now, now)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		c := &card.PopularCard{}
		var valueStr string
		if err = rows.Scan(&c.ID, &c.Title, &c.Type, &valueStr, &c.Reason, &c.ReasonType, &c.Pos); err != nil {
			return
		}
		c.Value, _ = strconv.ParseInt(valueStr, 10, 64)
		res = append(res, c)
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return
}

// CardPlat channel card  plat
func (d *Dao) CardPlat(ctx context.Context) (res map[int64]map[int8][]*card.PopularCardPlat, err error) {
	res = map[int64]map[int8][]*card.PopularCardPlat{}
	rows, err := d.db.Query(ctx, _cardPlatSQL)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		c := &card.PopularCardPlat{}
		if err = rows.Scan(&c.CardID, &c.Plat, &c.Condition, &c.Build); err != nil {
			return
		}
		if r, ok := res[c.CardID]; !ok {
			res[c.CardID] = map[int8][]*card.PopularCardPlat{
				c.Plat: {c},
			}
		} else {
			r[c.Plat] = append(r[c.Plat], c)
		}
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return
}

// CardSet card set
func (d *Dao) CardSet(ctx context.Context) (map[int64]*operate.CardSet, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", cardActionKey("loadCardSetCache", "cardSet")))
	if err != nil {
		return nil, err
	}
	var res map[int64]*operate.CardSet
	if err = json.Unmarshal(reply, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// EventTopic event_topic all
func (d *Dao) EventTopic(ctx context.Context) (map[int64]*operate.EventTopic, error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", cardActionKey("loadEventTopicCache", "eventTopic")))
	if err != nil {
		return nil, err
	}
	var res map[int64]*operate.EventTopic
	if err = json.Unmarshal(reply, &res); err != nil {
		return nil, err
	}
	return res, nil
}

// Series picks all the series of the given type
func (d *Dao) Series(ctx context.Context, sType string) (res []*selected.SerieFilter, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, _selectedSeriesSQL, sType, time.Now().Format(mysql.TimeFormat)); err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		c := &selected.SerieFilter{}
		if err = rows.Scan(&c.ID, &c.Type, &c.Number, &c.Subject, &c.Stime, &c.Etime, &c.Status); err != nil {
			return
		}
		res = append(res, c)
	}
	if err = rows.Err(); err != nil {
		return
	}
	return
}

// SerieConfig picks the config of one selected serie
func (d *Dao) SerieConfig(ctx context.Context, sType string, number int64) (res *selected.SerieConfig, err error) {
	res = &selected.SerieConfig{}
	err = d.db.QueryRow(ctx, _serieConfigSQL, sType, number, time.Now().Format(mysql.TimeFormat)).
		Scan(&res.ID, &res.Type, &res.Number, &res.Subject, &res.Stime, &res.Etime, &res.Hint, &res.Color, &res.Cover, &res.ShareTitle, &res.ShareSubtitle, &res.Status, &res.MediaID)
	if err == sql.ErrNoRows { // transform 500 to 404
		err = ecode.NothingFound
	}
	return
}

// SelectedRes picks all the resources of the given serie
func (d *Dao) SelectedRes(ctx context.Context, serieID int64) (res []*selected.SelectedRes, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(ctx, _selectedResourceSQL, serieID); err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		c := &selected.SelectedRes{}
		if err = rows.Scan(&c.RID, &c.Rtype, &c.SerieID, &c.Position, &c.RcmdReason); err != nil {
			return
		}
		res = append(res, c)
	}
	if err = rows.Err(); err != nil {
		return
	}
	return
}

func (d *Dao) Close() {
	if d.db != nil {
		d.db.Close()
	}
	if d.redis != nil {
		d.redis.Close()
	}
}
