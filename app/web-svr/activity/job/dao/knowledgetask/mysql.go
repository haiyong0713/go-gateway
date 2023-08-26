package knowledgetask

import (
	"context"
	xsql "database/sql"
	"fmt"
	"strings"

	"go-common/library/log"
	model "go-gateway/app/web-svr/activity/job/model/knowledge_task"
)

const sql4knowledgeCalcListSQL = `
SELECT id,mid,arc_count,arc_coin,arc_favorite
FROM act_knowledge_calculate
WHERE log_date = ? AND id > ? ORDER BY id ASC LIMIT ?
`

func (d *Dao) RawKnowledgeCalcList(ctx context.Context, logDate string, id, limit int64) (res []*model.KnowledgeTaskCalc, err error) {
	rows, err := d.db.Query(ctx, sql4knowledgeCalcListSQL, logDate, id, limit)
	if err != nil {
		log.Errorc(ctx, "RawKnowledgeCalcList d.db.Query error(%+v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		item := new(model.KnowledgeTaskCalc)
		if err = rows.Scan(&item.Id, &item.Mid, &item.ArcCount, &item.ArcCoin, &item.ArcFavorite); err != nil {
			log.Errorc(ctx, "RawKnowledgeCalcList rows.Scan() error(%+v)", err)
			return
		}
		res = append(res, item)
	}
	if err = rows.Err(); err != nil {
		log.Errorc(ctx, "RawKnowledgeCalcList rows.Err() error(%+v)", err)
		return
	}
	return
}

const sql4InsertUpdateKnowledgeTask = `
INSERT INTO act_knowledge_task(mid,had_arc,coin,favorite) 
VALUES %s ON DUPLICATE KEY UPDATE
had_arc=values(had_arc),
coin=values(coin),
favorite=values(favorite);
`

// InsertUpdateUserKnowTask .
func (d *Dao) InsertUpdateUserKnowTask(ctx context.Context, list []*model.KnowledgeTaskCalc) (err error) {
	var (
		rowsValue []interface{}
		rowsParam []string
	)
	for _, r := range list {
		rowsParam = append(rowsParam, "(?,?,?,?)")
		rowsValue = append(rowsValue, r.Mid, r.ArcCount, r.ArcCoin, r.ArcFavorite)
	}
	sql := fmt.Sprintf(sql4InsertUpdateKnowledgeTask, strings.Join(rowsParam, ","))
	if _, err = d.db.Exec(ctx, sql, rowsValue...); err != nil {
		log.Errorc(ctx, "InsertUpdateUserKnowTask db.Exec() error(%+v)", err)
		return
	}
	return
}

const sql4DeleteKnowledgeCalcSQL = `
DELETE FROM act_knowledge_calculate
WHERE log_date <= ? LIMIT ?
`

func (d *Dao) DeleteKnowledgeCalculate(ctx context.Context, logDate string, limit int) (affected int64, err error) {
	var res xsql.Result
	if res, err = d.db.Exec(ctx, sql4DeleteKnowledgeCalcSQL, logDate, limit); err != nil {
		log.Errorc(ctx, "DeleteKnowledgeCalculate db.Exec() logDate(%s) error(%+v)", logDate, err)
		return
	}
	return res.RowsAffected()
}
