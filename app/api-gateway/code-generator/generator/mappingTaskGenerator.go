package generator

import (
	"context"
	"fmt"
	"strings"
)

type MappingTaskGenerator struct {
	TmplFilePath string
	GoFilePath   string
	ProjectName  string
	ServiceName  string
	ServicePath  string
	TaskInfo     *TaskInfo
}

func NewMappingTaskGenerator(ctx context.Context, codeInfo *CodeInfo, taskInfo *TaskInfo) (taskGenerator *MappingTaskGenerator, err error) {
	return &MappingTaskGenerator{
		TmplFilePath: codeInfo.TmplFilePath,
		GoFilePath:   codeInfo.GoFilePath,
		ProjectName:  codeInfo.ProjectName,
		ServiceName:  codeInfo.ServiceName,
		ServicePath:  codeInfo.ServicePath,
		TaskInfo:     taskInfo,
	}, nil
}

func (p *MappingTaskGenerator) Generate(ctx context.Context) (err error) {
	fmt.Printf("Generate mappingTask %s\n", p.TaskInfo.Name)

	inputName := make(map[string]bool)
	for _, taskInputInfo := range p.TaskInfo.InputList {
		inputName[taskInputInfo.Name] = true
	}
	for _, mappingRule := range p.TaskInfo.MappingRule {
		splits := strings.Split(mappingRule.Src, ".")
		srcObject := splits[0]
		if _, ok := inputName[srcObject]; ok {
			mappingRule.SrcObject = srcObject
		}
		err = ProcessMappingRule(mappingRule)
		if err != nil {
			return
		}
	}

	err = p.createGoFile(
		p.TmplFilePath+"/server/service/internal/task",
		"mappingTask.go.tmpl",
		p.GoFilePath+"/"+p.ProjectName+"/app/"+p.ServicePath+"/service/internal/task",

		"mappingTask"+p.TaskInfo.Name+".go",
		p,
	)
	return
}

func (p *MappingTaskGenerator) createGoFile(src, srcFileName, dest, destFileName string, data interface{}) (err error) {
	err = CreateGoFile(src, srcFileName, dest, destFileName, data)
	return
}
