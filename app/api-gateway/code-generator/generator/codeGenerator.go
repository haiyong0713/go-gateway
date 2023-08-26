package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	"os"
	"strings"
)

type CodeGenerator struct {
	ServerGenerator     *ServerGenerator
	RpcClientGenerator  *RpcClientGenerator
	HttpClientGenerator *HttpClientGenerator
	HttpModelGenerator  *HttpModelGenerator
	TaskGenerator       *TaskGenerator
	TaskFlowGenerator   *TaskFlowGenerator
	GoFilePath          string
	ProjectName         string
	ServicePath         string
	GeneratorData       *GeneratorData
}

func NewCodeGenerator(ctx context.Context, codeInfo *CodeInfo) (codeGenerator *CodeGenerator, err error) {
	var generatorData = new(GeneratorData)
	param, _ := json.Marshal(codeInfo)
	fmt.Println(string(param))
	generatorData.AppendOutput(string(param))

	var serverGenerator *ServerGenerator
	codeInfo.ServicePath = strings.ToLower(codeInfo.ServiceName)
	serverGenerator, err = NewServerGenerator(ctx, codeInfo, generatorData)
	if err != nil {
		return
	}
	var rpcClientGenerator *RpcClientGenerator
	rpcClientGenerator, err = NewRpcClientGenerator(ctx, codeInfo, generatorData)
	if err != nil {
		return
	}
	var httpClientGenerator *HttpClientGenerator
	httpClientGenerator, err = NewHttpClientGenerator(ctx, codeInfo, generatorData)
	if err != nil {
		return
	}
	var httpModelGenerator *HttpModelGenerator
	httpModelGenerator, err = NewHttpModelGenerator(ctx, codeInfo, generatorData)
	if err != nil {
		return
	}
	var taskGenerator *TaskGenerator
	taskGenerator, err = NewTaskGenerator(ctx, codeInfo, generatorData)
	if err != nil {
		return
	}
	var taskFlowGenerator *TaskFlowGenerator
	taskFlowGenerator, err = NewTaskFlowGenerator(ctx, codeInfo, generatorData)
	if err != nil {
		return
	}
	return &CodeGenerator{
		ServerGenerator:     serverGenerator,
		RpcClientGenerator:  rpcClientGenerator,
		HttpClientGenerator: httpClientGenerator,
		HttpModelGenerator:  httpModelGenerator,
		TaskGenerator:       taskGenerator,
		TaskFlowGenerator:   taskFlowGenerator,
		GoFilePath:          codeInfo.GoFilePath,
		ProjectName:         codeInfo.ProjectName,
		ServicePath:         codeInfo.ServicePath,
		GeneratorData:       generatorData,
	}, nil
}

func (p *CodeGenerator) Generate(ctx context.Context) (err error) {
	log.Infoc(ctx, "generate code start")
	p.GeneratorData.AppendOutput("generate code start")

	err = p.deleteDir()
	if err != nil {
		return
	}

	err = p.ServerGenerator.Generate(ctx)
	if err != nil {
		return
	}
	err = p.RpcClientGenerator.Generate(ctx)
	if err != nil {
		return
	}
	err = p.HttpClientGenerator.Generate(ctx)
	if err != nil {
		return
	}
	err = p.HttpModelGenerator.Generate(ctx)
	if err != nil {
		return
	}
	err = p.TaskGenerator.Generate(ctx)
	if err != nil {
		return
	}
	err = p.TaskFlowGenerator.Generate(ctx)
	if err != nil {
		return
	}
	err = p.processGoMod()
	if err != nil {
		return
	}

	log.Infoc(ctx, "generate code finish")
	p.GeneratorData.AppendOutput("generate code finish")
	return
}

func (p *CodeGenerator) Check(ctx context.Context) (err error) {
	log.Infoc(ctx, "check code")
	p.GeneratorData.AppendOutput("check code")
	err = p.checkProject()
	return
}

func (p *CodeGenerator) deleteDir() (err error) {
	path := fmt.Sprintf("%s/%s/app/%s/service/internal/", p.GoFilePath, p.ProjectName, p.ServicePath)
	fmt.Printf("delete dir %s\n", path)
	return os.RemoveAll(path)
}

func (p *CodeGenerator) processGoMod() (err error) {
	command := fmt.Sprintf("cd %s/%s/app/%s; go mod tidy", p.GoFilePath, p.ProjectName, p.ServicePath)
	p.GeneratorData.AppendOutput(command)
	var output string
	output, err = ExecCommand(command)
	p.GeneratorData.AppendOutput(output)
	if err != nil {
		return
	}
	return
}

func (p *CodeGenerator) checkProject() (err error) {
	command := fmt.Sprintf("cd %s/%s/app/%s/service/cmd; go build", p.GoFilePath, p.ProjectName, p.ServicePath)
	p.GeneratorData.AppendOutput(command)
	var output string
	output, err = ExecCommand(command)
	p.GeneratorData.AppendOutput(output)
	if err != nil {
		return
	}
	file := fmt.Sprintf("%s/%s/app/%s/service/cmd/cmd", p.GoFilePath, p.ProjectName, p.ServicePath)
	fmt.Printf("delete file %s\n", file)
	p.GeneratorData.AppendOutput(fmt.Sprintf("delete file %s", file))
	return os.Remove(file)
}

func (p *CodeGenerator) GetOutPut() (output string) {
	return p.GeneratorData.Output
}

type GeneratorData struct {
	Output string
}

func (p *GeneratorData) AppendOutput(output string) {
	p.Output = p.Output + output + "\n"
}
