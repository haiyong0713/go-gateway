package bws

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/bws"

	"github.com/pkg/errors"
)

const (
	_addAchieveURI   = "/x/internal/activity/bws/achieve/add"
	_achievementsSQL = "SELECT id,achieve_point FROM act_bws_achievements WHERE bid=? AND del=0"
	_userAchievesSQL = "SELECT aid,`key` FROM act_bws_user_achieves WHERE bid=? AND `key` IN(%s) AND del=0"
)

// Achievements .
func (d *Dao) Achievements(c context.Context, bid int64) (data map[int64]*bws.Achieve, err error) {
	var rows *sql.Rows
	if rows, err = d.db.Query(c, _achievementsSQL, bid); err != nil {
		log.Error("Achievements: db.Exec(%d) error(%v)", bid, err)
		return
	}
	defer rows.Close()
	data = make(map[int64]*bws.Achieve)
	for rows.Next() {
		r := new(bws.Achieve)
		if err = rows.Scan(&r.ID, &r.AchievePoint); err != nil {
			log.Error("Achievements:row.Scan() error(%v)", err)
			return
		}
		data[r.ID] = r
	}
	err = rows.Err()
	return
}

// UserAchieves .
func (d *Dao) UserAchieves(c context.Context, bid int64, keys []string) (data map[string][]*bws.UserAchieve, err error) {
	var rows *sql.Rows
	var rowStrings []string
	rowArgs := []interface{}{bid}
	for _, key := range keys {
		rowStrings = append(rowStrings, "?")
		rowArgs = append(rowArgs, key)
	}
	if rows, err = d.db.Query(c, fmt.Sprintf(_userAchievesSQL, strings.Join(rowStrings, ",")), rowArgs...); err != nil {
		log.Error("UserAchieves: db.Exec(%d) error(%v)", bid, err)
		return
	}
	defer rows.Close()
	data = make(map[string][]*bws.UserAchieve)
	for rows.Next() {
		r := new(bws.UserAchieve)
		if err = rows.Scan(&r.Aid, &r.Key); err != nil {
			log.Error("UserAchieves:row.Scan() error(%v)", err)
			return
		}
		data[r.Key] = append(data[r.Key], r)
	}
	err = rows.Err()
	return
}

// AddAchieve .
func (d *Dao) AddAchieve(c context.Context, mid, bid int64) (err error) {
	var res struct {
		Code int `json:"code"`
	}
	params := url.Values{}
	params.Set("bid", strconv.FormatInt(bid, 10))
	params.Set("mid", strconv.FormatInt(mid, 10))
	if err = d.httpClient.Post(c, d.addAchieveURL, "", params, &res); err != nil {
		log.Error("AddAchieve:d.httpClient.Post mid(%d) bid(%d) error(%v)", mid, bid, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.addAchieveURL+"?"+params.Encode())
	}
	return
}

// AchieveRank .
func (d *Dao) AchieveRank(c context.Context, bid int64) (list []int64, err error) {
	var (
		key  string
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	key = fmt.Sprintf("bws_a_r_%d", bid)
	if list, err = redis.Int64s(conn.Do("ZREVRANGE", key, 0, -1)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("AchieveRank conn.Do(ZREVRANGE,%s,%d) error(%v)", key, -1, err)
		}
	}
	return
}
