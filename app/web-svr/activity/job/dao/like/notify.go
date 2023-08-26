package like

import (
	"context"
	"github.com/pkg/errors"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/like"
	"time"
)

const (
	_SQLQueryNotifyBySID   = "SELECT id, sid, notify_type, title, receiver, threshold, author, rule_id, template_id, ext FROM act_subject_notify WHERE sid = ? AND state = 1 AND notify_time = 0"
	_SQLUpdateNotifyFinish = "UPDATE act_subject_notify SET notify_time = ? WHERE id = ?"
)

func (dao *Dao) NotifyList(c context.Context, sid int64) ([]*like.ActSubjectNotify, error) {
	rows, err := dao.db.Query(c, _SQLQueryNotifyBySID, sid)
	if err != nil {
		log.Errorc(c, "NotifyList:d.db.Query(%d) err[%v]", sid, err)
		return nil, errors.Wrapf(err, "NotifyList:d.db.Query(%d) err[%v]", sid, err)
	}
	defer rows.Close()
	res := make([]*like.ActSubjectNotify, 0, 100)
	for rows.Next() {
		n := new(like.ActSubjectNotify)
		if err = rows.Scan(&n.ID, &n.Sid, &n.NotifyType, &n.Title, &n.Receiver, &n.Threshold, &n.Author, &n.RuleID, &n.TemplateID, &n.Ext); err != nil {
			log.Errorc(c, "NotifyList:row.Scan row err[%v]", err)
			return nil, errors.Wrapf(err, "NotifyList:row.Scan row err[%v]", err)
		}
		if len(n.Ext) == 0 {
			n.Ext = []byte(`{}`)
		}
		res = append(res, n)
	}
	if err = rows.Err(); err != nil {
		log.Errorc(c, "NotifyList:rowsErr(%v)", err)
		return nil, errors.Wrapf(err, "NotifyList:rowsErr(%v)", err)
	}
	return res, nil
}

func (dao *Dao) NotifyMarkFinish(c context.Context, notifyID int64) (int64, error) {
	row, err := dao.db.Exec(c, _SQLUpdateNotifyFinish, time.Now().Unix(), notifyID)
	if err != nil {
		log.Errorc(c, "NotifyMarkFinish:dao.db.Exec(%v)", err)
		return 0, errors.Wrap(err, "NotifyMarkFinish dao.db.Exec")
	}
	return row.RowsAffected()
}
