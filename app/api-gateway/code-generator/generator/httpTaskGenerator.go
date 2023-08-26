package generator

import (
	"context"
	"fmt"
)

type HttpTaskGenerator struct {
	TmplFilePath    string
	GoFilePath      string
	ProjectName     string
	ServiceName     string
	ServicePath     string
	TaskInfo        *TaskInfo
	HasQuery        bool
	HasUrlBody      bool
	HasJsonBody     bool
	HttpRequestInfo *HttpRequestInfo
}

func NewHttpTaskGenerator(ctx context.Context, codeInfo *CodeInfo, taskInfo *TaskInfo) (taskGenerator *HttpTaskGenerator, err error) {
	httpRequestInfo := codeInfo.HttpRequestMap[taskInfo.Url]
	return &HttpTaskGenerator{
		TmplFilePath:    codeInfo.TmplFilePath,
		GoFilePath:      codeInfo.GoFilePath,
		ProjectName:     codeInfo.ProjectName,
		ServiceName:     codeInfo.ServiceName,
		ServicePath:     codeInfo.ServicePath,
		TaskInfo:        taskInfo,
		HasQuery:        false,
		HttpRequestInfo: httpRequestInfo,
	}, nil
}

func (p *HttpTaskGenerator) Generate(ctx context.Context) (err error) {
	fmt.Printf("Generate httpTask %s\n", p.TaskInfo.Name)

	for _, mappingRule := range p.TaskInfo.Query {
		p.HasQuery = true
		err = ProcessMappingRule(mappingRule)
		if err != nil {
			return
		}
	}

	for _, mappingRule := range p.TaskInfo.Header {
		err = ProcessMappingRule(mappingRule)
		if err != nil {
			return
		}
	}

	for _, mappingRule := range p.TaskInfo.UrlBody {
		p.HasUrlBody = true
		err = ProcessMappingRule(mappingRule)
		if err != nil {
			return
		}
	}

	for _, mappingRule := range p.TaskInfo.JsonBody {
		p.HasJsonBody = true
		err = ProcessMappingRule(mappingRule)
		if err != nil {
			return
		}
	}

	src := fmt.Sprintf("%s/server/service/internal/task",
		p.TmplFilePath)
	dest := fmt.Sprintf("%s/%s/app/%s/service/internal/task",
		p.GoFilePath, p.ProjectName, p.ServicePath)

	err = p.createGoFile(
		src,
		"httpTask.go.tmpl",
		dest,
		"httpTask"+p.TaskInfo.Name+".go",
		p,
	)
	return
}

func (p *HttpTaskGenerator) createGoFile(src, srcFileName, dest, destFileName string, data interface{}) (err error) {
	err = CreateGoFile(src, srcFileName, dest, destFileName, data)
	return
}
