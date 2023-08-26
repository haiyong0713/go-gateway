package like

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	xsql "go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/component"
	"go-gateway/app/web-svr/activity/job/model/like"

	"github.com/pkg/errors"
)

const _selProductRoleSQL = "SELECT id,category_id,category_type,role,product,tags,tags_type,vote_num FROM act_selection_productrole WHERE is_deleted=0 ORDER BY id"

// SelProductRoles
func (d *Dao) SelProductRoles(ctx context.Context) (res []*like.ProductRoleDB, err error) {
	res = make([]*like.ProductRoleDB, 0)
	rows, err := d.db.Query(ctx, _selProductRoleSQL)
	if err != nil {
		log.Error("SelProductRoles.Query() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		row := &like.ProductRoleDB{}
		if err = rows.Scan(&row.ID, &row.CategoryID, &row.CategoryType, &row.Role, &row.Product, &row.Tags, &row.TagsType, &row.VoteNum); err != nil {
			log.Error("SelProductRoles row.Scan() error(%v)", err)
			return
		}
		res = append(res, row)
	}
	if err = rows.Err(); err != nil {
		log.Error("SelProductRoles rows.Err() error(%v)", err)
	}
	return
}

const _selPrCategorySQL = "Select id,category_id,category_type,role,product,tags,tags_type,vote_num FROM act_selection_productrole WHERE is_deleted=0 AND category_id=?"

func (d *Dao) SelProductRoleByCategory(ctx context.Context, categoryID int64) (res []*like.ProductRoleDB, err error) {
	res = make([]*like.ProductRoleDB, 0)
	rows, err := d.db.Query(ctx, _selPrCategorySQL, categoryID)
	if err != nil {
		log.Error("SelProductRoleByCategory.Query() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		row := &like.ProductRoleDB{}
		if err = rows.Scan(&row.ID, &row.CategoryID, &row.CategoryType, &row.Role, &row.Product, &row.Tags, &row.TagsType, &row.VoteNum); err != nil {
			log.Error("SelProductRoleByCategory row.Scan() error(%v)", err)
			return
		}
		res = append(res, row)
	}
	if err = rows.Err(); err != nil {
		log.Error("SelProductRoleByCategory rows.Err() error(%v)", err)
	}
	return
}

const _productroleCntSQL = "SELECT COUNT(1) AS cnt FROM act_selection_vote_log WHERE productrole_id=? AND vote_date=?  AND status=0"

// PrDayCnt .
func (d *Dao) PrDayCnt(c context.Context, productRoleID int64, voteDate string) (count int, err error) {
	row := d.db.QueryRow(c, _productroleCntSQL, productRoleID, voteDate)
	if err = row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Errorc(c, "PrDayCnt:QueryRow(%d) date(%s) error(%+v)", productRoleID, voteDate, err)
		}
	}
	return
}

const _prDayMidCntSQL = "SELECT COUNT(DISTINCT mid) AS cnt FROM act_selection_vote_log WHERE vote_date=?  AND status=0"

// PrDayMidCnt .
func (d *Dao) PrDayMidCnt(c context.Context, voteDate string) (count int, err error) {
	row := d.db.QueryRow(c, _prDayMidCntSQL, voteDate)
	if err = row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Errorc(c, "PrDayMidCnt:QueryRow(%s) error(%+v)", voteDate, err)
		}
	}
	return
}

const _prMidsSQL = "SELECT DISTINCT mid FROM act_selection_vote_log  WHERE status=0"

// VoteMids .
func (d *Dao) VoteMids(c context.Context) (res []int64, err error) {
	var rows *xsql.Rows
	if rows, err = d.db.Query(c, _prMidsSQL); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Errorc(c, "VoteMids:QueryRow(%s) error(%+v)", _prMidsSQL, err)
			return
		}
	}
	defer rows.Close()
	for rows.Next() {
		var mid int64
		if err = rows.Scan(&mid); err != nil {
			log.Errorc(c, "VoteMids:scan(%s) error(%+v)", _prMidsSQL, err)
			return
		}
		res = append(res, mid)
	}
	if err = rows.Err(); err != nil {
		log.Errorc(c, "VoteMids:rows.Err(%s) error(%+v)", _prMidsSQL, err)
	}
	return
}

const _prMidCntSQL = "SELECT COUNT(DISTINCT vote_date) AS cnt FROM act_selection_vote_log WHERE mid=? AND status=0"

// MidCountDays .
func (d *Dao) MidCountDays(c context.Context, mid int64) (count int64, err error) {
	row := d.db.QueryRow(c, _prMidCntSQL, mid)
	if err = row.Scan(&count); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Errorc(c, "PrDayMidCnt:QueryRow(%v) error(%+v)", mid, err)
		}
	}
	return
}

const (
	_upSelectionPrSQL   = "UPDATE `act_selection_productrole` SET `vote_num` = vote_num-1 WHERE `id` = ? AND vote_num > 0"
	_upSelectionVoteSQL = "UPDATE `act_selection_vote_log` SET `status` = 1 WHERE `mid` = ? AND `category_id` = ? AND `productrole_id` = ? AND `vote_date` = ? AND `status` = 0"
)

// UpSelectionVoteRisk .
func (d *Dao) UpSelectionVoteRisk(ctx context.Context, mid, categoryID, prID int64, voteDate string) (err error) {
	var (
		tx       *xsql.Tx
		sqlRes   sql.Result
		affected int64
	)
	tx, err = d.db.Begin(ctx)
	if err != nil {
		return
	}
	sqlRes, err = tx.Exec(_upSelectionVoteSQL, mid, categoryID, prID, voteDate) // 检测状态是否成功
	if err == nil {
		if affected, err = sqlRes.RowsAffected(); err == nil && affected > 0 {
			_, err = tx.Exec(_upSelectionPrSQL, prID)
		} else {
			err = errors.Wrap(err, "gaiaRiskProc gaiaRisk UpSelectionVoteRisk affected")
		}
	}
	if err == nil {
		err = tx.Commit()
	} else {
		tmpErr := err
		err = tx.Rollback()
		if err == nil {
			err = tmpErr
		}
	}
	return
}

func keyPrArc(id int64) string {
	return fmt.Sprintf("pr_a_%d", id)
}

// AddCacheAssistance.
func (d *Dao) AddCacheAssistance(c context.Context, productRoleID int64, prHots []*like.ProductRoleHot) (err error) {
	total := len(prHots)
	key := keyPrArc(productRoleID)
	conn := component.GlobalRedis.Conn(c)
	defer conn.Close()
	count := 0
	if err = conn.Send("DEL", key); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", key, err)
		return
	}
	count++
	for _, object := range prHots {
		productRole := &like.ProductRoleArc{
			Aid:     object.Aid,
			PubDate: object.PubDate,
		}
		bs, _ := json.Marshal(productRole)
		if err = conn.Send("ZADD", key, combine(object.HotNum, total), bs); err != nil {
			log.Error("conn.Send(ZADD, %s, %s) error(%v)", key, string(bs), err)
			return
		}
		count++
	}
	if err = conn.Send("EXPIRE", key, 86400); err != nil {
		log.Error("conn.Send(Expire, %s, %d) error(%v)", key, 86400, err)
		return
	}
	count++
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < count; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return

}

func combine(hostNum int64, count int) int64 {
	return hostNum<<16 | int64(count)
}

// ResetProductRoleVoteNum .
func (d *Dao) ResetProductRoleVoteNum(c context.Context, categoryID int) (err error) {
	var res struct {
		Code int `json:"code"`
	}
	params := url.Values{}
	params.Set("category_id", strconv.Itoa(categoryID))
	if err = d.httpClient.Get(c, d.selectionResetURL, "", params, &res); err != nil {
		log.Error("ResetProductRoleVoteNum:d.httpClient.Get error(%v)", err)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.selectionResetURL+"?"+params.Encode())
	}
	return
}
