package dao

import (
	"context"
	"database/sql"
	"fmt"

	"go-common/library/xstr"
	"go-gateway/app/app-svr/up-archive/job/internal/model"

	"github.com/pkg/errors"
)

const (
	// nolint:gosec
	_arcPassedSQL = "SELECT aid,mid,pubtime,copyright,attribute,attribute_v2,redirect_url,up_from FROM archive WHERE mid=? AND state>=0"
	_arcsSQL      = "SELECT aid,mid,pubtime,copyright,attribute,attribute_v2,redirect_url,up_from FROM archive WHERE aid IN (%s) AND state>=0"
	_arcSQL       = "SELECT aid,mid,pubtime,copyright,attribute,attribute_v2,redirect_url,up_from FROM archive WHERE aid=? AND state>=0"
	_midSQL       = "SELECT mid FROM archive_result_mid_tmp WHERE mid>? ORDER BY mid LIMIT ?"
)

func (d *dao) RawArcPassed(ctx context.Context, mid int64) ([]*model.UpArc, error) {
	rows, err := d.resultDB.Query(ctx, _arcPassedSQL, mid)
	if err != nil {
		return nil, errors.Wrapf(err, "RawArcPassed Query mid:%d", mid)
	}
	defer rows.Close()
	var res []*model.UpArc
	for rows.Next() {
		item := &model.UpArc{}
		if err = rows.Scan(&item.Aid, &item.Mid, &item.PubTime, &item.CopyRight, &item.Attribute, &item.AttributeV2, &item.RedirectURL, &item.UpFrom); err != nil {
			return nil, errors.Wrapf(err, "RawArcPassed rows.Scan mid:%d", mid)
		}
		res = append(res, item)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "RawArcPassed rows.Err() mid:%d", mid)
	}
	return res, nil
}

func (d *dao) RawArcs(ctx context.Context, aids []int64) ([]*model.UpArc, error) {
	rows, err := d.resultDB.Query(ctx, fmt.Sprintf(_arcsSQL, xstr.JoinInts(aids)))
	if err != nil {
		return nil, errors.Wrapf(err, "RawArcs Query aids:%v", aids)
	}
	defer rows.Close()
	var res []*model.UpArc
	for rows.Next() {
		item := &model.UpArc{}
		if err = rows.Scan(&item.Aid, &item.Mid, &item.PubTime, &item.CopyRight, &item.Attribute, &item.AttributeV2, &item.RedirectURL, &item.UpFrom); err != nil {
			return nil, errors.Wrapf(err, "RawArcs rows.Scan aids:%v", aids)
		}
		res = append(res, item)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "RawArcs rows.Err() aids:%v", aids)
	}
	return res, nil
}

func (d *dao) RawArc(ctx context.Context, aid int64) (*model.UpArc, error) {
	res := &model.UpArc{}
	if err := d.resultDB.QueryRow(ctx, _arcSQL, aid).Scan(&res.Aid, &res.Mid, &res.PubTime, &res.CopyRight, &res.Attribute, &res.AttributeV2, &res.RedirectURL, &res.UpFrom); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "RawArc QueryRow aid:%d", aid)
	}
	return res, nil
}

func (d *dao) RawUpper(ctx context.Context, mid, limit int64) ([]int64, error) {
	rows, err := d.tempDB.Query(ctx, _midSQL, mid, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []int64
	for rows.Next() {
		var mid int64
		if err = rows.Scan(&mid); err != nil {
			return nil, err
		}
		res = append(res, mid)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
