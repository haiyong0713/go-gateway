package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"
)

const (
	_addExampleStatSQL = "INSERT INTO dance_example (aid, ts, action) VALUES %s"
)

// Update update stat's fields
func (d *dao) GatherExamples(c context.Context, aid int64, stats []*model.Stat) error {

	if len(stats) == 0 {
		return nil
	}
	values := make([]string, 0)
	params := make([]interface{}, 0)
	for _, v := range stats {
		st, e := json.Marshal(v.StatCore)
		if e != nil {
			return e
		}
		values = append(values, fmt.Sprintf("(?,?,?)"))
		params = append(params, aid, v.TS, st)
	}

	_, e := d.db.Exec(c, fmt.Sprintf(_addExampleStatSQL, strings.Join(values, ",")), params...)
	return e
}
