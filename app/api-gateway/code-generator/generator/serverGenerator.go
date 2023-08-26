package generator

import (
	"context"
	"fmt"
	"path/filepath"
)

type ServerGenerator struct {
	TmplFilePath         string
	GoFilePath           string
	ProjectName          string
	ServiceName          string
	ServicePath          string
	ServiceProtoFile     string
	ServiceInterfaceInfo *ServiceInterfaceInfo
	TaskFlowInfo         *TaskFlowInfo
	GeneratorData        *GeneratorData
}

func NewServerGenerator(ctx context.Context, codeInfo *CodeInfo, generatorData *GeneratorData) (serverGenerator *ServerGenerator, err error) {
	return &ServerGenerator{
		TmplFilePath:         codeInfo.TmplFilePath,
		GoFilePath:           codeInfo.GoFilePath,
		ProjectName:          codeInfo.ProjectName,
		ServiceName:          codeInfo.ServiceName,
		ServicePath:          codeInfo.ServicePath,
		ServiceProtoFile:     codeInfo.ServiceProtoFile,
		ServiceInterfaceInfo: codeInfo.ServiceInterfaceInfo,
		TaskFlowInfo:         codeInfo.TaskFlowInfo,
		GeneratorData:        generatorData,
	}, nil
}

func (p *ServerGenerator) Generate(ctx context.Context) (err error) {
	fmt.Println("Generate Server")
	p.GeneratorData.AppendOutput("Generate Server")
	err = p.generateMain()
	if err != nil {
		return
	}

	err = p.generateApp()
	if err != nil {
		return
	}

	err = p.generateWire()
	if err != nil {
		return
	}

	err = p.generateWireGen()
	if err != nil {
		return
	}

	err = p.generateGrpcServer()
	if err != nil {
		return
	}

	err = p.generateHttpServer()
	if err != nil {
		return
	}

	if err = p.generateHttpRouter(); err != nil {
		return
	}

	err = p.generateService()
	if err != nil {
		return
	}

	err = p.generateServiceImpl()
	if err != nil {
		return
	}

	err = p.generateGoMod()
	if err != nil {
		return
	}

	err = p.generateServicePb()
	return
}

func (p *ServerGenerator) generateMain() (err error) {
	err = p.createGoFile(
		p.TmplFilePath+"/server/service/cmd",
		p.GoFilePath+"/"+p.ProjectName+"/app/"+p.ServicePath+"/service/cmd",
		"main.go",
		p,
	)
	return
}

func (p *ServerGenerator) generateApp() (err error) {
	err = p.createGoFile(
		p.TmplFilePath+"/server/service/internal/di",
		p.GoFilePath+"/"+p.ProjectName+"/app/"+p.ServicePath+"/service/internal/di",
		"app.go",
		p,
	)
	return
}

func (p *ServerGenerator) generateWire() (err error) {
	err = p.createGoFile(
		p.TmplFilePath+"/server/service/internal/di",
		p.GoFilePath+"/"+p.ProjectName+"/app/"+p.ServicePath+"/service/internal/di",
		"wire.go",
		p,
	)
	return
}

func (p *ServerGenerator) generateWireGen() (err error) {
	err = p.createGoFile(
		p.TmplFilePath+"/server/service/internal/di",
		p.GoFilePath+"/"+p.ProjectName+"/app/"+p.ServicePath+"/service/internal/di",
		"wire_gen.go",
		p,
	)
	return
}

func (p *ServerGenerator) generateGrpcServer() (err error) {
	err = p.createGoFile(
		p.TmplFilePath+"/server/service/internal/server/grpc",
		p.GoFilePath+"/"+p.ProjectName+"/app/"+p.ServicePath+"/service/internal/server/grpc",
		"server.go",
		p,
	)
	return
}

func (p *ServerGenerator) generateHttpServer() (err error) {

	err = p.createGoFile(
		p.TmplFilePath+"/server/service/internal/server/http",
		p.GoFilePath+"/"+p.ProjectName+"/app/"+p.ServicePath+"/service/internal/server/http",
		"server.go",
		p,
	)
	return
}

func (p *ServerGenerator) generateHttpRouter() (err error) {
	err = p.createGoFile(
		p.TmplFilePath+"/server/service/internal/server/http",
		p.GoFilePath+"/"+p.ProjectName+"/app/"+p.ServicePath+"/service/internal/server/http",
		"serverRouter.go",
		p,
	)
	return
}

func (p *ServerGenerator) generateService() (err error) {
	err = p.createGoFile(
		p.TmplFilePath+"/server/service/internal/service",
		p.GoFilePath+"/"+p.ProjectName+"/app/"+p.ServicePath+"/service/internal/service",
		"service.go",
		p,
	)
	return
}

func (p *ServerGenerator) generateServiceImpl() (err error) {
	err = p.createGoFile(
		p.TmplFilePath+"/server/service/internal/service",
		p.GoFilePath+"/"+p.ProjectName+"/app/"+p.ServicePath+"/service/internal/service",
		"serviceImpl.go",
		p,
	)
	return
}

func (p *ServerGenerator) generateGoMod() (err error) {
	err = p.createGoFile(
		p.TmplFilePath,
		p.GoFilePath+"/"+p.ProjectName+"/app/"+p.ServicePath,
		"go.mod",
		p,
	)
	return
}

func (p *ServerGenerator) generateServicePb() (err error) {
	err = p.copyFile(
		filepath.Dir(p.ServiceProtoFile),
		filepath.Base(p.ServiceProtoFile),
		p.GoFilePath+"/"+p.ProjectName+"/app/"+p.ServicePath+"/service/api",
		"api.proto",
	)
	if err != nil {
		return
	}

	command := fmt.Sprintf("cd %s/%s/app/%s/service/api; kratos tool protoc --grpc --bm api.proto",
		p.GoFilePath, p.ProjectName, p.ServicePath)
	p.GeneratorData.AppendOutput(command)
	var output string
	output, err = ExecCommand(command)
	p.GeneratorData.AppendOutput(output)
	return
}

func (p *ServerGenerator) copyFile(src, srcFile, dest, destFile string) (err error) {
	err = CopyFile(src, srcFile, dest, destFile)
	return
}

func (p *ServerGenerator) createGoFile(src, dest, fileName string, data interface{}) (err error) {
	err = CreateGoFile(src, fileName+".tmpl", dest, fileName, data)
	return
}
