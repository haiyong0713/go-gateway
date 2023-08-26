package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

const projectName = "api-gateway"

type Builder struct {
	TmplFilePath         string
	GoFilePath           string
	ServiceName          string
	ServiceProtoFile     string
	ServiceInterfaceInfo *ServiceInterfaceInfo
	TaskFlowRawInfo      *TaskFlowRawInfo
	RpcClientMap         map[string]*RpcClientInfo
	HttpRequestMap       map[string]*HttpRequestInfo
}

type ServiceInterfaceInfo struct {
	FuncName         string
	InputType        string
	OutputType       string
	HttpPath         string
	IsAuthUser       bool
	IsVerify         bool
	ServiceProtoFile string
}

type RpcClientInfo struct {
	DiscoveryId string
	ServiceName string
	PbAlias     string
	PbPath      string
	Timeout     string
}

type HttpRequestInfo struct {
	Url               string
	PackageName       string
	BodyType          string
	ResponseType      string
	ProtoFile         string
	ResponseCodeField string
}

type TaskFlowRawInfo struct {
	TaskProtoFile string
	TaskDSLFile   string
	TaskUdfFile   string
}

func (p *Builder) Build(ctx context.Context) (codeGenerator *CodeGenerator, err error) {
	var taskFlowInfo *TaskFlowInfo
	taskFlowInfo, err = p.parseTaskFlow(ctx)
	if err != nil {
		return
	}

	codeInfo := &CodeInfo{
		TmplFilePath:         p.TmplFilePath,
		GoFilePath:           p.GoFilePath,
		ProjectName:          projectName,
		ServiceName:          p.ServiceName,
		ServiceProtoFile:     p.ServiceProtoFile,
		TaskFlowInfo:         taskFlowInfo,
		ServiceInterfaceInfo: p.ServiceInterfaceInfo,
		RpcClientMap:         p.RpcClientMap,
		HttpRequestMap:       p.HttpRequestMap,
	}
	codeGenerator, err = NewCodeGenerator(ctx, codeInfo)
	return
}

func (p *Builder) parseTaskFlow(ctx context.Context) (taskFlowInfo *TaskFlowInfo, err error) {
	filename := p.TaskFlowRawInfo.TaskDSLFile
	var jsonData []byte
	jsonData, err = ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	var taskDSL TaskDSL
	err = json.Unmarshal(jsonData, &taskDSL)
	if err != nil {
		return
	}

	p.parseTaskDSL(ctx, &taskDSL)
	//fmt.Printf("%+v", taskDSL)

	taskFlowInfo = &TaskFlowInfo{
		TaskProtoFile:     p.TaskFlowRawInfo.TaskProtoFile,
		TaskUdfFile:       p.TaskFlowRawInfo.TaskUdfFile,
		TaskList:          taskDSL.TaskList,
		ServiceInputType:  p.ServiceInterfaceInfo.InputType,
		ServiceOutputType: p.ServiceInterfaceInfo.OutputType,
	}

	return
}

func (p *Builder) parseTaskDSL(ctx context.Context, taskDSL *TaskDSL) {
	pbAlias := make(map[string]string)
	for _, rpcClientInfo := range p.RpcClientMap {
		if rpcClientInfo.Timeout == "" {
			rpcClientInfo.Timeout = "500ms"
		}
		pbAlias[rpcClientInfo.PbAlias] = "git.bilibili.co/bapis/bapis-go" + rpcClientInfo.PbPath
	}
	servicePath := strings.ToLower(p.ServiceName)
	pbAlias["servicePb"] = fmt.Sprintf("git.bilibili.co/platform/%s/app/%s/service/api",
		projectName, servicePath)

	taskMap := make(map[string]*TaskInfo)
	for _, taskInfo := range taskDSL.TaskList {
		taskMap[taskInfo.Name] = taskInfo

		taskInfo.ImportPb = make(map[string]string)
		outputTypeSplits := strings.Split(taskInfo.OutputInfo.RowType, ".")
		if len(outputTypeSplits) == 1 {
			continue
		}
		pb := outputTypeSplits[0]
		//fmt.Println(pb)
		taskInfo.ImportPb[pb] = pbAlias[pb]
	}
	for _, taskInfo := range taskDSL.TaskList {
		for index, inputInfo := range taskInfo.InputList {
			//inputInfo.ParamType = "*" + inputInfo.RowType
			splits := strings.Split(inputInfo.RowValue, ".")
			if splits[0] == "$task" {
				inputInfo.PreTask = splits[1]
				inputInfo.Value = splits[1]
				inputInfo.ValueFrom = "task"
				inputInfo.ParamType = taskMap[inputInfo.PreTask].OutputInfo.RowType
			} else if splits[0] == "$service" {
				inputInfo.Value = "." + splits[1]
				inputInfo.ValueFrom = "service"
				inputInfo.ParamType = "servicePb." + p.ServiceInterfaceInfo.InputType
				taskInfo.ImportPb["servicePb"] = pbAlias["servicePb"]
			}
			if inputInfo.Name == "" {
				inputInfo.Name = "input" + strconv.Itoa(index+1)
			}
			//fmt.Println(inputInfo)
		}
		if taskInfo.OutputInfo.Name == "" {
			taskInfo.OutputInfo.Name = "output"
		}

		taskInfo.OutputInfo.ReturnType = taskInfo.OutputInfo.RowType
	}
}
