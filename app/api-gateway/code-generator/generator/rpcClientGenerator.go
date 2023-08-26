package generator

import (
	"context"
	"fmt"
)

type RpcClientGenerator struct {
	TmplFilePath  string
	GoFilePath    string
	ProjectName   string
	ServiceName   string
	ServicePath   string
	RpcClientMap  map[string]*RpcClientInfo
	NeedRpcClient bool
	GeneratorData *GeneratorData
}

func NewRpcClientGenerator(ctx context.Context, codeInfo *CodeInfo, generatorData *GeneratorData) (rpcClientGenerator *RpcClientGenerator, err error) {
	needRpcClient := true
	if codeInfo.RpcClientMap == nil || len(codeInfo.RpcClientMap) == 0 {
		needRpcClient = false
	}
	return &RpcClientGenerator{
		TmplFilePath:  codeInfo.TmplFilePath,
		GoFilePath:    codeInfo.GoFilePath,
		ProjectName:   codeInfo.ProjectName,
		ServiceName:   codeInfo.ServiceName,
		ServicePath:   codeInfo.ServicePath,
		RpcClientMap:  codeInfo.RpcClientMap,
		NeedRpcClient: needRpcClient,
		GeneratorData: generatorData,
	}, nil
}

func (p *RpcClientGenerator) Generate(ctx context.Context) (err error) {
	fmt.Println("Generate RpcClient")
	p.GeneratorData.AppendOutput("Generate RpcClient")
	err = p.generateRpcClient()
	if err != nil {
		return
	}

	err = p.generateRpcClientConfig()
	if err != nil {
		return
	}
	return
}

func (p *RpcClientGenerator) generateRpcClient() (err error) {
	err = p.createGoFile(
		p.TmplFilePath+"/server/service/internal/rpcClient",
		p.GoFilePath+"/"+p.ProjectName+"/app/"+p.ServicePath+"/service/internal/rpcClient",
		"rpcClient.go",
		p,
	)
	return
}

func (p *RpcClientGenerator) generateRpcClientConfig() (err error) {
	err = p.createGoFile(
		p.TmplFilePath+"/server/service/configs",
		p.GoFilePath+"/"+p.ProjectName+"/app/"+p.ServicePath+"/service/configs",
		"rpcClient.toml",
		p,
	)
	return
}

func (p *RpcClientGenerator) createGoFile(src, dest, fileName string, data interface{}) (err error) {
	err = CreateGoFile(src, fileName+".tmpl", dest, fileName, data)
	return
}
