package generator

import (
	"context"
	"fmt"
)

type GrpcTaskGenerator struct {
	TmplFilePath  string
	GoFilePath    string
	ProjectName   string
	ServiceName   string
	ServicePath   string
	TaskInfo      *TaskInfo
	RpcClientInfo *RpcClientInfo
}

func NewGrpcTaskGenerator(ctx context.Context, codeInfo *CodeInfo, taskInfo *TaskInfo, rpcClientInfo *RpcClientInfo) (taskGenerator *GrpcTaskGenerator, err error) {
	return &GrpcTaskGenerator{
		TmplFilePath:  codeInfo.TmplFilePath,
		GoFilePath:    codeInfo.GoFilePath,
		ProjectName:   codeInfo.ProjectName,
		ServiceName:   codeInfo.ServiceName,
		ServicePath:   codeInfo.ServicePath,
		TaskInfo:      taskInfo,
		RpcClientInfo: rpcClientInfo,
	}, nil
}

func (p *GrpcTaskGenerator) Generate(ctx context.Context) (err error) {
	fmt.Printf("Generate grpcTask %s\n", p.TaskInfo.Name)

	err = p.createGoFile(
		p.TmplFilePath+"/server/service/internal/task",
		"rpcTask.go.tmpl",
		p.GoFilePath+"/"+p.ProjectName+"/app/"+p.ServicePath+"/service/internal/task",

		"rpcTask"+p.TaskInfo.Name+".go",
		p,
	)
	return
}

func (p *GrpcTaskGenerator) createGoFile(src, srcFileName, dest, destFileName string, data interface{}) (err error) {
	err = CreateGoFile(src, srcFileName, dest, destFileName, data)
	return
}
