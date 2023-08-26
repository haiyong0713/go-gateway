package datamanage

import (
	"context"
	"encoding/csv"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/datamanage"
)

func (s *Service) DataManageDiff(c context.Context, req *datamanage.ReqDataManageUpdate, reader *csv.Reader) (interface{}, error) {
	res := struct {
		Insert  []map[string]interface{}
		Updated []map[string]interface{}
	}{
		Insert:  []map[string]interface{}{},
		Updated: []map[string]interface{}{},
	}
	_, err := s.checkAndUpdate(c, req, reader, func(req *datamanage.ReqDataManageUpdate, data [][]string, columns map[string]int) error {
		for _, record := range data {
			p := make(map[string]interface{})
			for k, i := range columns {
				p[k] = record[i]
			}
			res.Insert = append(res.Insert, p)
		}
		return nil
	}, func(req *datamanage.ReqDataManageUpdate, updated map[string]interface{}, priValue string, raw map[string]string) error {
		res.Updated = append(res.Updated, map[string]interface{}{
			"raw":    raw,
			"update": updated,
		})
		return nil
	})
	if err != nil {
		log.Errorc(c, "DataManageDiff s.checkAndUpdate err[%v]", err)
		return nil, err
	}
	return res, nil
}
