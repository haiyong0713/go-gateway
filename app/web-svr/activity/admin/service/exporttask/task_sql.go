package exporttask

import (
	"context"
	sql2 "database/sql"
	"fmt"
	"go-common/library/database/sql"
	"go-common/library/log"
)

type formatter interface {
	Formatter(c context.Context, taskRet []map[string]string) ([][]string, error)
	Headers() []string
}

type appended interface {
	Append(c context.Context, taskRet []map[string]string) []map[string]string
}

type taskExportSQL struct {
	SQL            string
	Tablet         func(map[string]string) int
	Builder        func(string, map[string]string) string
	PrimaryKey     string
	PrimaryDefault interface{}
	Args           []string
	Append         []appended
	formatter
}

func (e *taskExportSQL) GetData(c context.Context, db *sql.DB, data map[string]string, do func([]map[string]string) error) error {
	data["primary_key"] = fmt.Sprint(e.PrimaryDefault)

	// 根据分表规则预处理task sql
	querySQL := e.SQL
	if e.Builder != nil {
		querySQL = e.Builder(querySQL, data)
	} else if e.Tablet != nil {
		querySQL = fmt.Sprintf(querySQL, e.Tablet(data))
	}

	// 限制最大请求次数，避免异常死循环
	var maxQuery = 100000

	// 分批依次拉取数据
	for {
		maxQuery--
		// 计算sql 参数
		args := make([]interface{}, 0, len(e.Args))
		for _, arg := range e.Args {
			args = append(args, data[arg])
		}
		// 执行批次sql查询
		log.Infoc(c, "doTask s.export.Query querySQL[%s] args[%v]", querySQL, args)
		rows, err := db.Query(c, querySQL, args...)
		if err != nil {
			log.Errorc(c, "doTask s.export.Query error[%v]", err)
			return err
		}
		// 获取列名信息
		var columns []string
		columns, err = rows.Columns()
		if err != nil {
			log.Errorc(c, "doTask rows.Columns error[%v]", err)
			return err
		}
		// 读取数据
		tmpSet := make([]map[string]string, 0, 1000)
		defer rows.Close()
		for rows.Next() {
			r := make([]interface{}, 0, len(columns))
			for range columns {
				var p sql2.NullString
				r = append(r, &p)
			}
			if err = rows.Scan(r...); err != nil {
				log.Errorc(c, "doTask row.Scan() error(%v)", err)
				return err
			}
			one := make(map[string]string)
			for i, k := range columns {
				one[k] = (*(r[i].(*sql2.NullString))).String
			}
			tmpSet = append(tmpSet, one)
		}
		if err = rows.Err(); err != nil {
			log.Errorc(c, "doTask rows.Err() error(%v)", err)
			return err
		}

		// 无数据，或者达到批次上限
		if len(tmpSet) == 0 || maxQuery <= 0 {
			break
		}

		// 执行数据后续操作
		err = do(tmpSet)
		if err != nil {
			log.Errorc(c, "doTask custom do error(%v)", err)
			return err
		}

		// 获取下轮便利主键信息
		data["primary_key"] = tmpSet[len(tmpSet)-1][e.PrimaryKey]
	}
	return nil
}

func (e *taskExportSQL) Do(c context.Context, db *sql.DB, data map[string]string, writer *readerWriter) error {
	return e.GetData(c, db, data, func(set []map[string]string) error {
		if len(e.Append) > 0 {
			for _, apd := range e.Append {
				set = apd.Append(c, set)
			}
		}
		ret, err := e.Formatter(c, set)
		if err != nil {
			return err
		}
		writer.Put(ret)
		return nil
	})
}

func (e *taskExportSQL) Header(c context.Context, data map[string]string) ([]string, error) {
	return e.Headers(), nil
}
