package usermodel

import (
	"context"

	"go-common/library/database/sql"
	"go-common/library/log"

	model "go-gateway/app/app-svr/app-interface/interface-legacy/model/family"
)

const (
	_fyRelsOfParentSQL   = "SELECT id,parent_mid,child_mid,timelock_state,daily_duration FROM `family_relation` WHERE `parent_mid`=? and `state`=?"
	_fyRelsOfChildSQL    = "SELECT id,parent_mid,child_mid,timelock_state,daily_duration FROM `family_relation` WHERE `child_mid`=? and `state`=?"
	_fyUnbindSQL         = "UPDATE family_relation SET state=? WHERE id=?"
	_fyBindSQL           = "INSERT INTO family_relation (parent_mid,child_mid,state,daily_duration) VALUES (?,?,?,?)"
	_fyUpdateTimelockSQL = "UPDATE family_relation SET timelock_state=?,daily_duration=? WHERE id=?"
	_fyLatestRelSQL      = "SELECT daily_duration FROM family_relation WHERE parent_mid=? AND child_mid=? ORDER BY ID DESC LIMIT 1"
)

func (d *dao) RawFamilyRelsOfParent(ctx context.Context, pmid int64) ([]*model.FamilyRelation, error) {
	rows, err := d.db.Query(ctx, _fyRelsOfParentSQL, pmid, model.RelStateBind)
	if err != nil {
		log.Error("Fail to query family_relation of parent, pmid=%+v error=%+v", pmid, err)
		return nil, err
	}
	defer rows.Close()
	var rels []*model.FamilyRelation
	for rows.Next() {
		rel := &model.FamilyRelation{}
		if err = rows.Scan(&rel.ID, &rel.ParentMid, &rel.ChildMid, &rel.TimelockState, &rel.DailyDuration); err != nil {
			log.Error("Fail to scan family_relation of parent, pmid=%+v error=%+v", pmid, err)
			return nil, err
		}
		rels = append(rels, rel)
	}
	if err := rows.Err(); err != nil {
		log.Error("Fail to rows family_relation of parent, pmid=%+v error=%+v", pmid, err)
		return nil, err
	}
	return rels, nil
}

func (d *dao) RawFamilyRelsOfChild(ctx context.Context, cmid int64) (*model.FamilyRelation, error) {
	row := d.db.QueryRow(ctx, _fyRelsOfChildSQL, cmid, model.RelStateBind)
	rel := &model.FamilyRelation{}
	if err := row.Scan(&rel.ID, &rel.ParentMid, &rel.ChildMid, &rel.TimelockState, &rel.DailyDuration); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Error("Fail to scan family_relation of child, cmid=%+v error=%+v", cmid, err)
		return nil, err
	}
	return rel, nil
}

func (d *dao) unbindFamily(ctx context.Context, id int64) error {
	if _, err := d.db.Exec(ctx, _fyUnbindSQL, model.RelStateUnbind, id); err != nil {
		log.Error("Fail to unbind family_relation, id=%+v error=%+v", id, err)
		return err
	}
	return nil
}

func (d *dao) bindFamily(ctx context.Context, pmid, cmid, duration int64) error {
	if _, err := d.db.Exec(ctx, _fyBindSQL, pmid, cmid, model.RelStateBind, duration); err != nil {
		log.Error("Fail to bind family_relation, pmid=%+v cmid=%+v error=%+v", pmid, cmid, err)
		return err
	}
	return nil
}

func (d *dao) updateTimelock(ctx context.Context, id, state, duration int64) error {
	if _, err := d.db.Exec(ctx, _fyUpdateTimelockSQL, state, duration, id); err != nil {
		log.Error("Fail to update timelock, id=%+v error=%+v", id, err)
		return err
	}
	return nil
}

func (d *dao) LatestFamilyRel(ctx context.Context, pmid, cmid int64) (*model.FamilyRelation, error) {
	row := d.db.QueryRow(ctx, _fyLatestRelSQL, pmid, cmid)
	rel := &model.FamilyRelation{}
	if err := row.Scan(&rel.DailyDuration); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Error("Fail to scan latest family_relation, pmid=%+v cmid=%+v error=%+v", pmid, cmid, err)
		return nil, err
	}
	return rel, nil
}
