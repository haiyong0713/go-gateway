package card

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-job/job/conf"
	"go-gateway/app/app-svr/app-show/interface/model"
	"go-gateway/app/app-svr/app-show/interface/model/card"
	"go-gateway/pkg/idsafe/bvid"

	"github.com/pkg/errors"
)

const (
	// daily_selection
	_appColumnSQL = "SELECT id,tab,resource_id,tpl,name,plat_ver FROM app_column WHERE state=1"
	_appPosRecSQL = "SELECT p.id,p.tab,p.resource_id,p.type,p.title,p.cover,p.re_type,p.re_value,p.plat_ver,p.desc,p.tag_id FROM app_pos_rec AS p " +
		"WHERE p.stime<? AND p.etime>? AND p.state=1 ORDER BY p.weight ASC"
	_appContentRSQL = "SELECT c.id,c.module,c.rec_id,c.ctype,c.cvalue,c.ctitle,c.tag_id FROM app_content AS c, app_pos_rec AS r " +
		"WHERE c.rec_id=r.id AND r.state=1 AND r.stime<? AND r.etime>? AND c.module=1"
	_appColumnNperSQL = "SELECT n.id,n.column_id,n.name,n.desc,n.nper,n.nper_time,n.cover,n.plat_ver,n.title,n.re_type,n.re_value FROM app_column_nper AS n " +
		"WHERE n.cron_time<? AND n.state=1 ORDER BY n.nper DESC"
	_appContentNSQL = "SELECT c.id,c.module,c.rec_id,c.ctype,c.cvalue,c.ctitle,c.tag_id FROM app_content AS c, app_column_nper AS n " +
		"WHERE c.rec_id=n.id AND n.state=1 AND n.cron_time<? AND c.module=2"
	_appColumnList = "SELECT c.id,c.name,cn.id,cn.title,cn.plat_ver FROM app_column AS c,app_column_nper AS cn " +
		"WHERE c.id=cn.column_id AND c.state=1 AND cn.state=1 AND cn.cron_time<? ORDER BY cn.nper DESC"
	_cardSetSQL    = `SELECT c.id,c.type,c.value,c.title,c.long_title,c.content FROM card_set AS c WHERE c.deleted=0`
	_eventTopicSQL = `SELECT c.id,c.title,c.desc,c.cover,c.re_type,c.re_value,c.corner,p.sticky,c.show_title FROM event_topic AS c JOIN popular_card AS p ON p.card_value = c.id  WHERE c.deleted=0 AND p.is_delete=0 AND p.card_type="event_topic"`
	// redis use
	_loadCardKey       = "loadCardCache"
	_loadNperKey       = "loadNperCache"
	_loadColumnsKey    = "loadColumnsCache"
	_loadColumnListKey = "loadColumnListCache"
	_loadCardSetKey    = "loadCardSetCache"
	_loadEventTopicKey = "loadEventTopicCache"
	_cardRedisKey      = "card"
	_splitToken        = ":"
	_cardExpire        = 604800
)

// Dao is card dao.
type Dao struct {
	db          *sql.DB
	column      *sql.Stmt
	posRec      *sql.Stmt
	recContent  *sql.Stmt
	nperContent *sql.Stmt
	columnNper  *sql.Stmt
	columnList  *sql.Stmt
	redis       *redis.Pool
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		db:    sql.NewMySQL(c.MySQL.Show),
		redis: redis.NewPool(c.Redis.Recommend.Config),
	}
	d.column = d.db.Prepared(_appColumnSQL)
	d.posRec = d.db.Prepared(_appPosRecSQL)
	d.recContent = d.db.Prepared(_appContentRSQL)
	d.nperContent = d.db.Prepared(_appContentNSQL)
	d.columnNper = d.db.Prepared(_appColumnNperSQL)
	d.columnList = d.db.Prepared(_appColumnList)
	return d
}

func (d *Dao) Columns(ctx context.Context) (map[int8][]*card.Column, error) {
	rows, err := d.column.Query(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := map[int8][]*card.Column{}
	for rows.Next() {
		c := &card.Column{}
		if err = rows.Scan(&c.ID, &c.Tab, &c.RegionID, &c.Tpl, &c.Name, &c.PlatVer); err != nil {
			return nil, err
		}
		for _, limit := range c.ColumnPlatChange() {
			tmpc := &card.Column{}
			*tmpc = *c
			tmpc.Plat = limit.Plat
			tmpc.Build = limit.Build
			tmpc.Condition = limit.Condition
			tmpc.PlatVer = ""
			tmpc.ColumnGotoChannge()
			res[tmpc.Plat] = append(res[tmpc.Plat], tmpc)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) AddCacheColumns(ctx context.Context, tmp map[int8][]*card.Column) error {
	if len(tmp) == 0 {
		return nil
	}
	conn := d.redis.Get(ctx)
	defer conn.Close()
	bs, err := json.Marshal(tmp)
	if err != nil {
		return errors.WithStack(err)
	}
	key := cardActionKey(_loadColumnsKey, "column")
	if _, err = conn.Do("SETEX", key, _cardExpire, bs); err != nil {
		return err
	}
	return nil
}

func (d *Dao) PosRecs(ctx context.Context, now time.Time) (map[int8]map[int][]*card.Card, error) {
	rows, err := d.posRec.Query(ctx, now, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := map[int8]map[int][]*card.Card{}
	for rows.Next() {
		c := &card.Card{}
		if err = rows.Scan(&c.ID, &c.Tab, &c.RegionID, &c.Type, &c.Title, &c.Cover, &c.Rtype, &c.Rvalue, &c.PlatVer, &c.Desc, &c.TagID); err != nil {
			return nil, err
		}
		for _, limit := range c.CardPlatChange() {
			tmpc := &card.Card{}
			*tmpc = *c
			tmpc.Plat = limit.Plat
			tmpc.Build = limit.Build
			tmpc.Condition = limit.Condition
			tmpc.PlatVer = ""
			tmpc.CardGotoChannge()
			if cards, ok := res[tmpc.Plat]; ok {
				cards[tmpc.RegionID] = append(cards[tmpc.RegionID], tmpc)
			} else {
				res[tmpc.Plat] = map[int][]*card.Card{
					tmpc.RegionID: {tmpc},
				}
			}
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) RecContents(ctx context.Context, now time.Time) (map[int][]*card.Content, map[int][]int64, error) {
	rows, err := d.recContent.Query(ctx, now, now)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	res := map[int][]*card.Content{}
	aids := map[int][]int64{}
	for rows.Next() {
		c := &card.Content{}
		if err = rows.Scan(&c.ID, &c.Module, &c.RecID, &c.Type, &c.Value, &c.Title, &c.TagID); err != nil {
			return nil, nil, err
		}
		res[c.RecID] = append(res[c.RecID], c)
		if c.Type == model.CardGotoAv {
			if c.Value != "" {
				aidInt, err := getAvID(c.Value)
				if err != nil {
					log.Error("日志告警 RecContents aidInt parse err(%+v)", err)
					continue
				}
				if aidInt > 0 {
					aids[c.RecID] = append(aids[c.RecID], aidInt)
				}
			}
		}
	}
	if err = rows.Err(); err != nil {
		return nil, nil, err
	}
	return res, aids, nil
}

func (d *Dao) AddCacheCard(ctx context.Context, hdm map[int8]map[int][]*card.Card, itm map[int][]*card.Content, aids map[int][]int64) error {
	conn := d.redis.Get(ctx)
	defer conn.Close()

	bs, err := json.Marshal(hdm)
	if err != nil {
		return errors.WithStack(err)
	}
	key := cardActionKey(_loadCardKey, "card")
	var keys []string
	keys = append(keys, key)
	argsMDs := redis.Args{}
	argsMDs = argsMDs.Add(key).Add(bs)

	bs, err = json.Marshal(itm)
	if err != nil {
		return errors.WithStack(err)
	}
	key = cardActionKey(_loadCardKey, "content")
	keys = append(keys, key)
	argsMDs = argsMDs.Add(key).Add(bs)

	bs, err = json.Marshal(aids)
	if err != nil {
		return errors.WithStack(err)
	}
	key = cardActionKey(_loadCardKey, "aids")
	keys = append(keys, key)
	argsMDs = argsMDs.Add(key).Add(bs)

	if err = conn.Send("MSET", argsMDs...); err != nil {
		return err
	}
	for _, v := range keys {
		if err = conn.Send("EXPIRE", v, _cardExpire); err != nil {
			return err
		}
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush() error(%v)", err)
		return err
	}
	for i := 0; i < len(keys)+1; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return err
		}
	}
	return nil
}

func (d *Dao) NperContents(ctx context.Context, now time.Time) (map[int][]*card.Content, map[int][]int64, error) {
	res := map[int][]*card.Content{}
	rows, err := d.nperContent.Query(ctx, now)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	aids := map[int][]int64{}
	for rows.Next() {
		c := &card.Content{}
		if err = rows.Scan(&c.ID, &c.Module, &c.RecID, &c.Type, &c.Value, &c.Title, &c.TagID); err != nil {
			return nil, nil, err
		}
		res[c.RecID] = append(res[c.RecID], c)
		if c.Type == model.CardGotoAv {
			if c.Value != "" {
				aidInt, err := getAvID(c.Value)
				if err != nil {
					log.Error("日志告警 NperContents aidInt parse err(%+v)", err)
					continue
				}
				if aidInt > 0 {
					aids[c.RecID] = append(aids[c.RecID], aidInt)
				}
			}
		}
	}
	if err = rows.Err(); err != nil {
		return nil, nil, err
	}
	return res, aids, nil
}

func (d *Dao) ColumnNpers(ctx context.Context, now time.Time) (map[int8][]*card.ColumnNper, error) {
	rows, err := d.columnNper.Query(ctx, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := map[int8][]*card.ColumnNper{}
	for rows.Next() {
		c := &card.ColumnNper{}
		if err = rows.Scan(&c.ID, &c.ColumnID, &c.Name, &c.Desc, &c.Nper, &c.NperTime, &c.Cover, &c.PlatVer, &c.Title, &c.Rtype, &c.Rvalue); err != nil {
			return nil, err
		}
		for _, limit := range c.ColumnNperPlatChange() {
			tmpc := &card.ColumnNper{}
			*tmpc = *c
			tmpc.Plat = limit.Plat
			tmpc.Build = limit.Build
			tmpc.Condition = limit.Condition
			tmpc.PlatVer = ""
			tmpc.ColumnNperGotoChange()
			res[tmpc.Plat] = append(res[tmpc.Plat], tmpc)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) AddCacheNper(ctx context.Context, hdm map[int8][]*card.ColumnNper, itm map[int][]*card.Content, aids map[int][]int64) error {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	argsMDs := redis.Args{}
	var keys []string
	bs, err := json.Marshal(hdm)
	if err != nil {
		return errors.WithStack(err)
	}
	key := cardActionKey(_loadNperKey, "columnNper")
	keys = append(keys, key)
	argsMDs = argsMDs.Add(key).Add(bs)

	bs, err = json.Marshal(itm)
	if err != nil {
		return errors.WithStack(err)
	}
	key = cardActionKey(_loadNperKey, "content")
	keys = append(keys, key)
	argsMDs = argsMDs.Add(key).Add(bs)

	bs, err = json.Marshal(aids)
	if err != nil {
		return errors.WithStack(err)
	}
	key = cardActionKey(_loadNperKey, "aids")
	keys = append(keys, key)
	argsMDs = argsMDs.Add(key).Add(bs)

	if err = conn.Send("MSET", argsMDs...); err != nil {
		return err
	}
	for _, v := range keys {
		if err = conn.Send("EXPIRE", v, _cardExpire); err != nil {
			return err
		}
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush() error(%v)", err)
		return err
	}
	for i := 0; i < len(keys)+1; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return err
		}
	}
	return nil
}

func (d *Dao) ColumnList(ctx context.Context, now time.Time) ([]*card.ColumnList, error) {
	rows, err := d.columnList.Query(ctx, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []*card.ColumnList
	for rows.Next() {
		c := &card.ColumnList{}
		if err = rows.Scan(&c.Ceid, &c.Cname, &c.Cid, &c.Name, &c.PlatVer); err != nil {
			return nil, err
		}
		res = append(res, c)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) AddCacheColumnList(ctx context.Context, child []*card.ColumnList) error {
	if len(child) == 0 {
		return nil
	}
	conn := d.redis.Get(ctx)
	defer conn.Close()
	bs, err := json.Marshal(child)
	if err != nil {
		return errors.WithStack(err)
	}
	key := cardActionKey(_loadColumnListKey, "columnList")
	if _, err = conn.Do("SETEX", key, _cardExpire, bs); err != nil {
		return err
	}
	return nil
}

func (d *Dao) CardSet(ctx context.Context) (map[int64]*operate.CardSet, error) {
	rows, err := d.db.Query(ctx, _cardSetSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make(map[int64]*operate.CardSet)
	for rows.Next() {
		var (
			c     = &operate.CardSet{}
			value string
		)
		if err = rows.Scan(&c.ID, &c.Type, &value, &c.Title, &c.LongTitle, &c.Content); err != nil {
			return nil, err
		}
		if value != "" {
			c.Value, err = strconv.ParseInt(value, 10, 64)
			if err != nil {
				log.Error("日志告警 CardSet aidInt parse err(%+v)", err)
				continue
			}
		}
		res[c.ID] = c
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) AddCacheCardSet(ctx context.Context, cards map[int64]*operate.CardSet) error {
	if len(cards) == 0 {
		return nil
	}
	conn := d.redis.Get(ctx)
	defer conn.Close()
	bs, err := json.Marshal(cards)
	if err != nil {
		return errors.WithStack(err)
	}
	key := cardActionKey(_loadCardSetKey, "cardSet")
	if _, err = conn.Do("SETEX", key, _cardExpire, bs); err != nil {
		return err
	}
	return nil
}

func (d *Dao) EventTopic(ctx context.Context) (map[int64]*operate.EventTopic, error) {
	rows, err := d.db.Query(ctx, _eventTopicSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make(map[int64]*operate.EventTopic)
	for rows.Next() {
		c := &operate.EventTopic{}
		if err = rows.Scan(&c.ID, &c.Title, &c.Desc, &c.Cover, &c.ReType, &c.ReValue, &c.Corner, &c.Sticky, &c.ShowTitle); err != nil {
			return nil, err
		}
		res[c.ID] = c
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func (d *Dao) AddCacheEventTopic(ctx context.Context, eventTopic map[int64]*operate.EventTopic) error {
	if len(eventTopic) == 0 {
		return nil
	}
	conn := d.redis.Get(ctx)
	defer conn.Close()
	bs, err := json.Marshal(eventTopic)
	if err != nil {
		return errors.WithStack(err)
	}
	key := cardActionKey(_loadEventTopicKey, "eventTopic")
	if _, err = conn.Do("SETEX", key, _cardExpire, bs); err != nil {
		return err
	}
	return nil
}

// Close Dao
func (d *Dao) Close() {
	if d.db != nil {
		_ = d.db.Close()
	}
	if d.redis != nil {
		_ = d.redis.Close()
	}
}

func getAvID(input string) (int64, error) {
	aid, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		aid, err = bvid.BvToAv(input)
		if err != nil {
			return 0, fmt.Errorf("视频ID(%s)非法！", input)
		}
	}
	return aid, nil
}

func cardActionKey(source string, param string) string {
	var builder strings.Builder
	builder.WriteString(_cardRedisKey)
	builder.WriteString(_splitToken)
	builder.WriteString(source)
	builder.WriteString(_splitToken)
	builder.WriteString(param)
	return builder.String()
}
