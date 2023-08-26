package datamanage

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/datamanage"
	"io"
	"os"
	"strings"
	"time"
)

const (
	batchInsert = 1000
	batchUpdate = 1000
)

func (s *Service) checkAndUpdate(c context.Context, req *datamanage.ReqDataManageUpdate, reader *csv.Reader,
	actionInsert func(*datamanage.ReqDataManageUpdate, [][]string, map[string]int) error,
	actionUpdate func(*datamanage.ReqDataManageUpdate, map[string]interface{}, string, map[string]string) error) (interface{}, error) {
	record, err := reader.Read()
	if err != nil {
		log.Errorc(c, "checkAndUpdate reader.Read() err[%v]", err)
		return nil, err
	}
	if strings.HasPrefix(record[0], "\xEF\xBB\xBF") {
		record[0] = record[0][3:]
	}
	ignoreField := map[string]struct{}{}
	for _, f := range req.IgnoreField {
		ignoreField[f] = struct{}{}
	}
	columns := make(map[string]int)
	for i, k := range record {
		k = strings.TrimSpace(k)
		if _, ok := ignoreField[k]; ok {
			continue
		}
		columns[k] = i
	}
	if _, ok := columns[req.Primary]; !ok {
		return nil, ecode.Error(ecode.RequestErr, "primary 字段不存在")
	}
	insert := make([][]string, 0, batchInsert)
	update := make([][]string, 0, batchUpdate)
	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if req.Trim {
			for k, v := range record {
				record[k] = strings.TrimSpace(v)
			}
		}
		if record[columns[req.Primary]] == "" || record[columns[req.Primary]] == "0" {
			insert = append(insert, record)
			if len(insert) >= batchInsert {
				if err := actionInsert(req, insert, columns); err != nil {
					log.Errorc(c, "checkAndUpdate actionInsert(%v, %v, %v) err[%v]", req, insert, columns, err)
					return nil, err
				}
				insert = make([][]string, 0, batchInsert)
			}
		} else {
			update = append(update, record)
			if len(update) >= batchUpdate {
				var err error
				if insert, err = s.updateData(c, req, update, columns, insert, actionInsert, actionUpdate); err != nil {
					log.Errorc(c, "checkAndUpdate s.updateData(%v, %v, %v, %v, actionInsert, actionUpdate) err[%v]", req, update, columns, insert, err)
					return nil, err
				}
				update = make([][]string, 0, batchUpdate)
			}
		}
	}
	if len(update) > 0 {
		if insert, err = s.updateData(c, req, update, columns, insert, actionInsert, actionUpdate); err != nil {
			log.Errorc(c, "checkAndUpdate s.updateData(%v, %v, %v, %v, actionInsert, actionUpdate) err[%v]", req, update, columns, insert, err)
			return nil, err
		}
	}
	if len(insert) > 0 {
		if err := actionInsert(req, insert, columns); err != nil {
			log.Errorc(c, "checkAndUpdate actionInsert(%v, %v, %v) err[%v]", req, insert, columns, err)
			return nil, err
		}
	}
	return nil, nil
}

func (s *Service) DataManageUpdate(c context.Context, req *datamanage.ReqDataManageUpdate, reader *csv.Reader) (interface{}, error) {
	var f *os.File
	defer func() {
		if f != nil {
			f.Close()
		}
	}()
	return s.checkAndUpdate(c, req, reader, s.insertData, func(req *datamanage.ReqDataManageUpdate, updated map[string]interface{}, priValue string, raw map[string]string) error {
		if f == nil {
			basedir := "/data/log/backup"
			if os.Getenv("DEPLOYMENT_ID") == "" {
				basedir = "/tmp/log/backup"
			}
			os.MkdirAll(basedir, os.ModePerm)
			var err error
			f, err = os.Create(fmt.Sprintf("%s/%s.%d.data", basedir, req.Table, time.Now().Unix()))
			if err != nil {
				return err
			}
		}
		b, _ := json.Marshal(raw)
		f.Write(b)
		f.Write([]byte("\n"))
		return s.getDB(req.Conn).Table(req.Table).Where(map[string]interface{}{
			req.Primary: priValue,
		}).Updates(updated).Error
	})
}

func (s *Service) insertData(req *datamanage.ReqDataManageUpdate, data [][]string, columns map[string]int) error {
	cls := make([]string, 0, len(columns))
	for k := range columns {
		cls = append(cls, k)
	}
	values := make([]interface{}, 0, len(data)*len(columns))
	placeholder := make([]string, 0, len(data))
	for _, record := range data {
		for _, k := range cls {
			values = append(values, record[columns[k]])
		}
		placeholder = append(placeholder, fmt.Sprintf("(%s)", strings.TrimRight(strings.Repeat("?,", len(columns)), ",")))
	}
	sql := fmt.Sprintf("INSERT INTO %s(`%s`) VALUES%s", req.Table, strings.Join(cls, "`,`"), strings.Join(placeholder, ","))
	return s.getDB(req.Conn).Exec(sql, values...).Error
}

func (s *Service) updateData(c context.Context, req *datamanage.ReqDataManageUpdate, data [][]string, columns map[string]int, insert [][]string,
	actionInsert func(*datamanage.ReqDataManageUpdate, [][]string, map[string]int) error,
	actionUpdate func(*datamanage.ReqDataManageUpdate, map[string]interface{}, string, map[string]string) error) ([][]string, error) {
	query := make([]string, 0, batchUpdate)
	for _, record := range data {
		query = append(query, record[columns[req.Primary]])
	}
	db := s.getDB(req.Conn).Table(req.Table).Where(fmt.Sprintf("%s in (?)", req.Primary), query)
	rawData, _, err := s.fetchRows(c, db)
	if err != nil {
		log.Errorc(c, "updateData s.fetchRows(c, db) err[%v]", err)
		return nil, err
	}
	priData := make(map[string]map[string]string)
	for _, record := range rawData {
		priData[record[req.Primary]] = record
	}
	for _, record := range data {
		if rawRecord, ok := priData[record[columns[req.Primary]]]; ok {
			updated := map[string]interface{}{}
			for k, i := range columns {
				if record[i] != rawRecord[k] {
					updated[k] = record[i]
				}
			}
			if len(updated) > 0 {
				if err := actionUpdate(req, updated, rawRecord[req.Primary], rawRecord); err != nil {
					log.Errorc(c, "updateData actionUpdate(%v, %v, %v) err[%v]", req, updated, rawRecord[req.Primary], err)
					return nil, err
				}
			}
		} else {
			insert = append(insert, record)
			if len(insert) >= batchInsert {
				if err := actionInsert(req, insert, columns); err != nil {
					log.Errorc(c, "updateData actionInsert(%v, %v, %v) err[%v]", req, insert, columns, err)
					return nil, err
				}
				insert = make([][]string, 0, batchInsert)
			}
		}
	}
	return insert, nil
}
