package generator

import (
	"context"
	"fmt"
)

type HttpClientGenerator struct {
	TmplFilePath  string
	GoFilePath    string
	ProjectName   string
	ServiceName   string
	ServicePath   string
	GeneratorData *GeneratorData
}

func NewHttpClientGenerator(ctx context.Context, codeInfo *CodeInfo, generatorData *GeneratorData) (httpClientGenerator *HttpClientGenerator, err error) {
	return &HttpClientGenerator{
		TmplFilePath:  codeInfo.TmplFilePath,
		GoFilePath:    codeInfo.GoFilePath,
		ProjectName:   codeInfo.ProjectName,
		ServiceName:   codeInfo.ServiceName,
		ServicePath:   codeInfo.ServicePath,
		GeneratorData: generatorData,
	}, nil
}

func (p *HttpClientGenerator) Generate(ctx context.Context) (err error) {
	fmt.Println("Generate HttpClient")
	p.GeneratorData.AppendOutput("Generate HttpClient")
	err = p.generateHttpClient()
	if err != nil {
		return
	}

	err = p.generateHttpClientConfig()
	if err != nil {
		return
	}
	return
}

func (p *HttpClientGenerator) generateHttpClient() (err error) {
	err = p.createGoFile(
		p.TmplFilePath+"/server/service/internal/httpClient",
		p.GoFilePath+"/"+p.ProjectName+"/app/"+p.ServicePath+"/service/internal/httpClient",
		"httpClient.go",
		p,
	)
	return
}

func (p *HttpClientGenerator) generateHttpClientConfig() (err error) {
	err = p.createGoFile(
		p.TmplFilePath+"/server/service/configs",
		p.GoFilePath+"/"+p.ProjectName+"/app/"+p.ServicePath+"/service/configs",
		"httpClient.toml",
		p,
	)
	return
}

func (p *HttpClientGenerator) createGoFile(src, dest, fileName string, data interface{}) (err error) {
	err = CreateGoFile(src, fileName+".tmpl", dest, fileName, data)
	return
}
