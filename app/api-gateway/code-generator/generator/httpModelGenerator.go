package generator

import (
	"context"
	"fmt"
	"path/filepath"
)

type HttpModelGenerator struct {
	TmplFilePath   string
	GoFilePath     string
	ProjectName    string
	ServiceName    string
	ServicePath    string
	HttpRequestMap map[string]*HttpRequestInfo
	GeneratorData  *GeneratorData
}

func NewHttpModelGenerator(ctx context.Context, codeInfo *CodeInfo, generatorData *GeneratorData) (httpModelGenerator *HttpModelGenerator, err error) {
	return &HttpModelGenerator{
		TmplFilePath:   codeInfo.TmplFilePath,
		GoFilePath:     codeInfo.GoFilePath,
		ProjectName:    codeInfo.ProjectName,
		ServiceName:    codeInfo.ServiceName,
		ServicePath:    codeInfo.ServicePath,
		HttpRequestMap: codeInfo.HttpRequestMap,
		GeneratorData:  generatorData,
	}, nil
}

func (p *HttpModelGenerator) Generate(ctx context.Context) (err error) {
	fmt.Println("Generate HttpModel")
	p.GeneratorData.AppendOutput("Generate HttpModel")
	for _, httpRequestInfo := range p.HttpRequestMap {
		err = p.generateHttpModel(httpRequestInfo)
		if err != nil {
			return
		}
	}

	return
}

func (p *HttpModelGenerator) generateHttpModel(httpRequestInfo *HttpRequestInfo) (err error) {
	dest := fmt.Sprintf("%s/%s/app/%s/service/internal/model/%s",
		p.GoFilePath, p.ProjectName, p.ServicePath, httpRequestInfo.PackageName)
	protoFile := httpRequestInfo.PackageName + ".proto"
	err = p.copyFile(
		filepath.Dir(httpRequestInfo.ProtoFile),
		filepath.Base(httpRequestInfo.ProtoFile),
		dest,
		protoFile,
	)
	if err != nil {
		return
	}

	command := fmt.Sprintf("cd %s; kratos tool protoc --grpc %s", dest, protoFile)
	p.GeneratorData.AppendOutput(command)
	var output string
	output, err = ExecCommand(command)
	p.GeneratorData.AppendOutput(output)
	if err != nil {
		return
	}
	return
}

func (p *HttpModelGenerator) copyFile(src, srcFile, dest, destFile string) (err error) {
	err = CopyFile(src, srcFile, dest, destFile)
	return
}
