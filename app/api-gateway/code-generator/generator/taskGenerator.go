package generator

import (
	"context"
	"fmt"
	"path/filepath"
)

type TaskGenerator struct {
	TmplFilePath  string
	GoFilePath    string
	ProjectName   string
	ServiceName   string
	ServicePath   string
	CodeInfo      *CodeInfo
	TaskFlowInfo  *TaskFlowInfo
	GeneratorData *GeneratorData
}

func NewTaskGenerator(ctx context.Context, codeInfo *CodeInfo, generatorData *GeneratorData) (taskGenerator *TaskGenerator, err error) {
	return &TaskGenerator{
		TmplFilePath:  codeInfo.TmplFilePath,
		GoFilePath:    codeInfo.GoFilePath,
		ProjectName:   codeInfo.ProjectName,
		ServiceName:   codeInfo.ServiceName,
		ServicePath:   codeInfo.ServicePath,
		CodeInfo:      codeInfo,
		TaskFlowInfo:  codeInfo.TaskFlowInfo,
		GeneratorData: generatorData,
	}, nil
}

func (p *TaskGenerator) Generate(ctx context.Context) (err error) {
	fmt.Println("Generate Task List")
	p.GeneratorData.AppendOutput("Generate Task List")
	err = p.generateTaskPb()
	if err != nil {
		return
	}

	err = p.generateTypeMapping()
	if err != nil {
		return
	}

	err = p.generateHelper()
	if err != nil {
		return
	}

	err = p.generateUdfTask()
	if err != nil {
		return
	}

	err = p.generateTaskList(ctx)
	if err != nil {
		return
	}

	return
}

func (p *TaskGenerator) generateTaskList(ctx context.Context) (err error) {
	for _, taskInfo := range p.TaskFlowInfo.TaskList {
		if taskInfo.Type == "grpc" {
			rpcClientInfo := p.CodeInfo.RpcClientMap[taskInfo.DiscoveryId]
			var grpcTaskGenerator *GrpcTaskGenerator
			grpcTaskGenerator, err = NewGrpcTaskGenerator(ctx, p.CodeInfo, taskInfo, rpcClientInfo)
			if err != nil {
				return
			}
			err = grpcTaskGenerator.Generate(ctx)
			if err != nil {
				return
			}
		} else if taskInfo.Type == "mapping" {
			var mappingTaskGenerator *MappingTaskGenerator
			mappingTaskGenerator, err = NewMappingTaskGenerator(ctx, p.CodeInfo, taskInfo)
			if err != nil {
				return
			}
			err = mappingTaskGenerator.Generate(ctx)
			if err != nil {
				return
			}
		} else if taskInfo.Type == "http" {
			var httpTaskGenerator *HttpTaskGenerator
			httpTaskGenerator, err = NewHttpTaskGenerator(ctx, p.CodeInfo, taskInfo)
			if err != nil {
				return
			}
			err = httpTaskGenerator.Generate(ctx)
			if err != nil {
				return
			}
		}
	}

	return
}

func (p *TaskGenerator) generateTaskPb() (err error) {
	dest := fmt.Sprintf("%s/%s/app/%s/service/internal/task",
		p.GoFilePath, p.ProjectName, p.ServicePath)
	err = p.copyFile(
		filepath.Dir(p.TaskFlowInfo.TaskProtoFile),
		filepath.Base(p.TaskFlowInfo.TaskProtoFile),
		dest,
		"taskModel.proto",
	)
	if err != nil {
		return
	}

	command := fmt.Sprintf("cd %s; kratos tool protoc --grpc taskModel.proto", dest)
	p.GeneratorData.AppendOutput(command)
	var output string
	output, err = ExecCommand(command)
	p.GeneratorData.AppendOutput(output)
	return
}

func (p *TaskGenerator) generateUdfTask() (err error) {
	if p.TaskFlowInfo.TaskUdfFile == "" {
		return
	}
	dest := fmt.Sprintf("%s/%s/app/%s/service/internal/task",
		p.GoFilePath, p.ProjectName, p.ServicePath)
	err = p.copyFile(
		filepath.Dir(p.TaskFlowInfo.TaskUdfFile),
		filepath.Base(p.TaskFlowInfo.TaskUdfFile),
		dest,
		"udfTask.go",
	)
	if err != nil {
		return
	}

	return
}

func (p *TaskGenerator) generateTypeMapping() (err error) {
	err = p.createGoFile(
		p.TmplFilePath+"/server/service/internal/task",
		p.GoFilePath+"/"+p.ProjectName+"/app/"+p.ServicePath+"/service/internal/task",
		"typeMapping.go",
		p,
	)

	return
}

func (p *TaskGenerator) generateHelper() (err error) {
	err = p.createGoFile(
		p.TmplFilePath+"/server/service/internal/task",
		p.GoFilePath+"/"+p.ProjectName+"/app/"+p.ServicePath+"/service/internal/task",
		"helper.go",
		p,
	)

	return
}

func (p *TaskGenerator) copyFile(src, srcFile, dest, destFile string) (err error) {
	err = CopyFile(src, srcFile, dest, destFile)
	return
}

func (p *TaskGenerator) createGoFile(src, dest, fileName string, data interface{}) (err error) {
	err = CreateGoFile(src, fileName+".tmpl", dest, fileName, data)
	return
}
