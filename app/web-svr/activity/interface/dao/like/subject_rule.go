package like

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/model/like"

	"github.com/pkg/errors"
)

const (
	_clockInSubIDSQL  = "SELECT id FROM act_subject WHERE type=22 AND state=1 AND lstime<=? AND letime>=?"
	_subRuleSQL       = "SELECT id,sid,category,type_ids,tags,state,ctime,mtime,rule_name,sids,coefficient,task_id,attribute,stime,etime FROM act_subject_rule WHERE sid=? AND state IN (1,2)"
	_subRuleSidsSQL   = "SELECT id,sid,category,type_ids,tags,state,ctime,mtime,rule_name,sids,coefficient,task_id,attribute,stime,etime FROM act_subject_rule WHERE sid IN(%s) AND state IN (1,2)"
	_subRuleUpdateSql = "UPDATE act_subject_rule SET type_ids=?, tags=?, state=?, attribute=?, rule_name=?, sids=?, coefficient=?,aid_source=?,aid_source_type=? WHERE id=?"
	_subRuleInsertSql = "INSERT INTO act_subject_rule(`sid`, `category`, `type_ids`, `tags`, `state`, `attribute`, `rule_name`, `sids`, `coefficient`, `task_id`,`aid_source`,`aid_source_type`) VALUES %s"
)

func (d *Dao) RawClockInSubIDs(c context.Context, queryTime time.Time) ([]int64, error) {
	rows, err := d.db.Query(c, _clockInSubIDSQL, queryTime, queryTime)
	if err != nil {
		err = errors.Wrapf(err, "RawClockInSubIDs d.db.Query(%s)", _clockInSubIDSQL)
		return nil, err
	}
	defer rows.Close()
	var sids []int64
	for rows.Next() {
		var id int64
		if err = rows.Scan(&id); err != nil {
			err = errors.Wrap(err, "RawClockInSubIDs scan()")
			return nil, err
		}
		sids = append(sids, id)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return sids, nil
}

func (d *Dao) RawSubjectRulesBySid(c context.Context, sid int64) ([]*like.SubjectRule, error) {
	rows, err := d.db.Query(c, _subRuleSQL, sid)
	if err != nil {
		err = errors.Wrapf(err, "RawSubjectRulesBySid d.db.Query(%s)", _subRuleSQL)
		return nil, err
	}
	defer rows.Close()
	var rules []*like.SubjectRule
	for rows.Next() {
		r := new(like.SubjectRule)
		if err = rows.Scan(&r.ID, &r.Sid, &r.Type, &r.TypeIds, &r.Tags, &r.State, &r.Ctime, &r.Mtime, &r.RuleName,
			&r.Sids, &r.Coefficient, &r.TaskID, &r.Attribute, &r.Stime, &r.Etime); err != nil {
			err = errors.Wrap(err, "RawSubjectRulesBySid scan()")
			return nil, err
		}
		rules = append(rules, r)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return rules, nil
}

func (d *Dao) RawSubjectRulesBySids(c context.Context, sids []int64) (map[int64][]*like.SubjectRule, error) {
	if len(sids) == 0 {
		return nil, nil
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_subRuleSidsSQL, xstr.JoinInts(sids)))
	if err != nil {
		err = errors.Wrapf(err, "RawSubjectRulesBySids d.db.Query(%s) sids(%v)", _subRuleSidsSQL, sids)
		return nil, err
	}
	defer rows.Close()
	rules := make(map[int64][]*like.SubjectRule)
	for rows.Next() {
		r := new(like.SubjectRule)
		if err = rows.Scan(&r.ID, &r.Sid, &r.Type, &r.TypeIds, &r.Tags, &r.State, &r.Ctime, &r.Mtime, &r.RuleName,
			&r.Sids, &r.Coefficient, &r.TaskID, &r.Attribute, &r.Stime, &r.Etime); err != nil {
			err = errors.Wrap(err, "RawSubjectRulesBySid scan()")
			return nil, err
		}
		rules[r.Sid] = append(rules[r.Sid], r)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return rules, nil
}

// UpdateSubjectRule ...
func (d *Dao) UpdateSubjectRule(c context.Context, r *like.SubjectRule) error {
	_, err := d.db.Exec(c, _subRuleUpdateSql, r.TypeIds, r.Tags, r.State, r.Attribute, r.RuleName, r.Sids,
		r.Coefficient, r.AidSource, r.AidSourceType, r.ID)
	if err != nil {
		log.Error("UpdateSubjectRule d.db.Exec error(%v)", err)
	}
	return err
}

// InsertSubjectRules ...
func (d *Dao) InsertSubjectRules(c context.Context, rs []*like.SubjectRule) error {
	sql := fmt.Sprintf(_subRuleInsertSql, strings.TrimRight(strings.Repeat("("+strings.TrimRight(
		strings.Repeat("?,", 12), ",")+"),", len(rs)), ","))
	args := make([]interface{}, 0, len(rs)*12)
	for _, r := range rs {
		args = append(args, r.Sid, r.Type, r.TypeIds, r.Tags, r.State, r.Attribute, r.RuleName, r.Sids,
			r.Coefficient, r.TaskID, r.AidSource, r.AidSourceType)
	}
	_, err := d.db.Exec(c, sql, args...)
	if err != nil {
		log.Error("InsertSubjectRules d.db.Exec error(%v)", err)
	}
	return err
}
