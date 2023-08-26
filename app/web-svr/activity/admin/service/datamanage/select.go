package datamanage

import (
	"context"
	"github.com/jinzhu/gorm"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/datamanage"
)

func (s *Service) getDB(conn string) *gorm.DB {
	switch conn {
	case "tidb":
		{
			return s.TIDB
		}
	}
	return s.DB
}

func (s *Service) DataManageSelect(c context.Context, req *datamanage.ReqDataManageSelect, where map[string]interface{}) (*datamanage.ResDataManageSelect, error) {
	db := s.getDB(req.Conn).Table(req.Table).Offset(req.Offset).Limit(req.Limit)
	if len(where) > 0 {
		db = db.Where(where)
	}
	var count int
	if err := db.Count(&count).Error; err != nil {
		log.Errorc(c, "DataManageSelect db.Count(&count) err[%v]", err)
		return nil, err
	}
	rawData, columns, err := s.fetchRows(c, db)
	if err != nil {
		log.Errorc(c, "DataManageSelect s.fetchRows(db) err[%v]", err)
		return nil, err
	}
	ignoreField := map[string]struct{}{}
	for _, f := range req.IgnoreField {
		ignoreField[f] = struct{}{}
	}
	data := make([]map[string]string, 0, len(rawData))
	for _, record := range rawData {
		value := map[string]string{}
		for k, v := range record {
			if _, ok := ignoreField[k]; ok {
				continue
			}
			value[k] = v
		}
		data = append(data, value)
	}
	tmpColumns := make([]string, 0, len(columns))
	for _, c := range columns {
		if _, ok := ignoreField[c]; ok {
			continue
		}
		tmpColumns = append(tmpColumns, c)
	}
	return &datamanage.ResDataManageSelect{
		Count:   count,
		List:    data,
		Req:     req,
		Where:   where,
		Columns: tmpColumns,
	}, nil
}

func (s *Service) fetchRows(c context.Context, db *gorm.DB) ([]map[string]string, []string, error) {
	var data []map[string]string
	rows, err := db.Rows()
	if err != nil {
		log.Errorc(c, "fetchRows db.Rows() err[%v]", err)
		return nil, nil, err
	}
	defer rows.Close()
	var columns []string
	columns, err = rows.Columns()
	if err != nil {
		log.Errorc(c, "fetchRows rows.Columns() err[%v]", err)
		return nil, nil, err
	}
	for rows.Next() {
		values := make([]interface{}, 0, len(columns))
		for range columns {
			var s string
			values = append(values, &s)
		}
		if err := rows.Scan(values...); err != nil {
			log.Errorc(c, "fetchRows rows.Scan(values...) err[%v]", err)
			return nil, nil, err
		}
		value := map[string]string{}
		for i, k := range columns {
			value[k] = *(values[i].(*string))
		}
		data = append(data, value)
	}
	return data, columns, nil
}
