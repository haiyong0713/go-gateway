package generator

import (
	"context"
	"encoding/json"
	"fmt"
)

type TaskFlowGenerator struct {
	TmplFilePath  string
	GoFilePath    string
	ProjectName   string
	ServiceName   string
	ServicePath   string
	TaskFlowInfo  *TaskFlowInfo
	RpcClientMap  map[string]*RpcClientInfo
	FlowNodeList  []*FlowNode
	ImportPb      map[string]string
	GeneratorData *GeneratorData
}

func NewTaskFlowGenerator(ctx context.Context, codeInfo *CodeInfo, generatorData *GeneratorData) (taskFlowGenerator *TaskFlowGenerator, err error) {
	flowCreator, err := newFlowCreator(codeInfo.TaskFlowInfo.TaskList)
	if err != nil {
		return
	}
	generatorData.AppendOutput(flowCreator.Dag.String())

	err = flowCreator.run()
	if err != nil {
		return
	}
	var flowNodeList []*FlowNode
	flowNodeList, err = flowCreator.getFlowNodeList()
	if err != nil {
		return
	}
	jsonData, _ := json.Marshal(flowNodeList)
	generatorData.AppendOutput(string(jsonData))

	importPb := make(map[string]string)
	for _, taskInfo := range codeInfo.TaskFlowInfo.TaskList {
		for pb, pbPath := range taskInfo.ImportPb {
			importPb[pb] = pbPath
		}
	}
	return &TaskFlowGenerator{
		TmplFilePath:  codeInfo.TmplFilePath,
		GoFilePath:    codeInfo.GoFilePath,
		ProjectName:   codeInfo.ProjectName,
		ServiceName:   codeInfo.ServiceName,
		ServicePath:   codeInfo.ServicePath,
		TaskFlowInfo:  codeInfo.TaskFlowInfo,
		RpcClientMap:  codeInfo.RpcClientMap,
		FlowNodeList:  flowNodeList,
		ImportPb:      importPb,
		GeneratorData: generatorData,
	}, nil
}

func (p *TaskFlowGenerator) Generate(ctx context.Context) (err error) {
	fmt.Println("Generate TaskFlow")
	p.GeneratorData.AppendOutput("Generate TaskFlow")
	if err = p.generateTaskNode(); err != nil {
		return
	}

	if err = p.generateTaskFlow(); err != nil {
		return
	}

	return
}

func (p *TaskFlowGenerator) generateTaskNode() (err error) {
	err = p.createGoFile(
		p.TmplFilePath+"/server/service/internal/task",
		p.GoFilePath+"/"+p.ProjectName+"/app/"+p.ServicePath+"/service/internal/task",
		"taskNode.go",
		p,
	)
	return
}

func (p *TaskFlowGenerator) generateTaskFlow() (err error) {
	err = p.createGoFile(
		p.TmplFilePath+"/server/service/internal/task",
		p.GoFilePath+"/"+p.ProjectName+"/app/"+p.ServicePath+"/service/internal/task",
		"taskFlow.go",
		p,
	)
	return
}

func (p *TaskFlowGenerator) createGoFile(src, dest, fileName string, data interface{}) (err error) {
	err = CreateGoFile(src, fileName+".tmpl", dest, fileName, data)
	return
}
