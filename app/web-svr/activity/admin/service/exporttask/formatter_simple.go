package exporttask

import "context"

type simpleFormatter struct {
	Output []*ExportOutputField
}

func (s *simpleFormatter) Formatter(c context.Context, taskRet []map[string]string) ([][]string, error) {
	dataSet := make([][]string, 0, len(taskRet))
	// 遍历taskRet 处理输出顺序和格式化
	for _, one := range taskRet {
		set := make([]string, 0, len(s.Output))
		for _, filed := range s.Output {
			value := one[filed.Name]
			if filed.Format != nil {
				value = filed.Format(one[filed.Name])
			}
			set = append(set, value)
		}
		// 追加到全局
		dataSet = append(dataSet, set)
	}
	return dataSet, nil
}

func (s *simpleFormatter) Headers() []string {
	header := make([]string, 0, len(s.Output))
	for _, one := range s.Output {
		if one.Title != "" {
			header = append(header, one.Title)
		} else {
			header = append(header, one.Name)
		}
	}
	return header
}
