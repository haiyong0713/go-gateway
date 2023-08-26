package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-common/library/log"
	pb "go-gateway/app/api-gateway/api-manager/api"
	"go-gateway/app/api-gateway/delay"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go-gateway/app/api-gateway/api-manager/internal/model"
	"go-gateway/app/api-gateway/code-generator/generator"
)

const tmplFilePath = "tmplFile"

func (s *Service) GenerateCode(ctx context.Context, apiID int64) (reply model.CodeGeneratorReply, err error) {
	var apiInfo *model.ApiInfo
	apiInfo, err = s.getGenerateApiInfo(ctx, apiID)
	if err != nil {
		return
	}

	var tempFilePath string
	tempFilePath, err = generator.GetTempPath()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			os.RemoveAll(tempFilePath)
		}
	}()
	var goFilePath string
	goFilePath, err = generator.GetTempPath()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			os.RemoveAll(goFilePath)
		}
	}()

	var info *model.CodeGeneratorInfo
	info, err = s.createCodeGeneratorInfo(ctx, apiInfo, tempFilePath, goFilePath)
	if err != nil {
		return
	}

	var builder *generator.Builder
	builder, err = initBuilder(ctx, info)
	if err != nil {
		return
	}

	version := fmt.Sprintf("%s-%s", apiInfo.ApiName, time.Now().Format("20060102150405"))
	var lastID int64
	lastID, err = s.delay.AddRowDB(ctx, apiInfo.ApiName, version)
	if err != nil {
		return
	}
	reply.WorkflowID = lastID
	err = s.codeGenerate.Do(ctx, func(ctx context.Context) {
		//defer os.RemoveAll(tempFilePath)
		//defer os.RemoveAll(goFilePath)
		output, err := codeGenerate(ctx, builder)
		if err != nil {
			log.Errorc(ctx, "code generate return err %v", err)
			fmt.Printf("code generate return err %v\n", err)
			output = fmt.Sprintf("%s\n%v", output, err)
			_ = s.delay.UpdateLog(ctx, lastID, delay.DisplayNameGenCode, delay.DisplayStateFailed, output)
			return
		}
		_ = s.delay.UpdateLog(ctx, lastID, delay.DisplayNameGenCode, delay.DisplayStateSucceeded, output)
		_ = s.delay.MergeStep(ctx, apiInfo.ApiName, goFilePath)
	})

	return
}

func (s *Service) getGenerateApiInfo(ctx context.Context, apiID int64) (apiInfo *model.ApiInfo, err error) {
	apiInfo, err = s.getApiInfo(ctx, apiID)
	if err != nil {
		return
	}

	var wfDetail *delay.WFDetail
	wfDetail, err = s.delay.GetLatestWF(ctx, apiInfo.ApiName)
	if err != nil {
		return
	}
	if wfDetail != nil && wfDetail.State == delay.WFStateNormal {
		err = errors.New("存在未结单的发布流程，请先结单")
		return
	}

	return
}

func (s *Service) getApiInfo(ctx context.Context, apiID int64) (apiInfo *model.ApiInfo, err error) {
	if apiID == -1 {
		return testApiInfo(), nil
	}
	var apiList []*model.ContralApi
	apiList, err = s.dao.ApiByIDs(ctx, []int64{apiID})
	if err != nil {
		return
	}
	if len(apiList) == 0 {
		err = errors.New("can not find api info")
		return
	}
	item := apiList[0]

	apiInfo = &model.ApiInfo{
		ApiName: item.ApiName,
		ApiType: item.ApiType,
		Router:  item.Router,
		Handler: item.Handler,
	}

	apiInfo.Req, err = s.readContentFromUrl(item.Req)
	if err != nil {
		return
	}
	apiInfo.Reply, err = s.readContentFromUrl(item.Reply)
	if err != nil {
		return
	}
	apiInfo.DSLCode, err = s.readContentFromUrl(item.DSLCode)
	if err != nil {
		return
	}
	apiInfo.DSLStruct, err = s.readContentFromUrl(item.DSLStruct)
	if err != nil {
		return
	}
	apiInfo.CustomCode, err = s.readContentFromUrl(item.CustomCode)
	if err != nil {
		return
	}
	return
}

func testApiInfo() (apiInfo *model.ApiInfo) {
	apiInfo = &model.ApiInfo{
		ApiName:    "sayHello",
		ApiType:    "http",
		Router:     "/test",
		Req:        "message HelloReq {\n  string name = 1 [(gogoproto.moretags) = 'form:\"name\" validate:\"required\"'];\n  int64 aid = 2 [(gogoproto.moretags) = 'form:\"aid\"'];\n}",
		Reply:      "message HelloResp {\n  string Content = 1 [(gogoproto.jsontag) = 'content'];\n}",
		DSLCode:    "{\n  \"taskList\": [\n    {\n      \"name\": \"SayWorld1\",\n      \"type\": \"udf\",\n      \"input\": [\n        {\n          \"value\": \"$service.request\"\n        }\n      ],\n      \"output\": {\n        \"type\": \"WorldResp\"\n      }\n    },\n    {\n      \"name\": \"SayWorld2\",\n      \"type\": \"udf\",\n      \"input\": [\n        {\n          \"value\": \"$task.SayWorld1\"\n        }\n      ],\n      \"output\": {\n        \"type\": \"servicePb.HelloResp\",\n        \"isResponse\": true\n      }\n    }\n  ]\n}",
		DSLStruct:  "message WorldResp {\n    string Content = 1;\n    int64 aid = 2;\n}\n\nmessage HttpResult {\n    int64 Code = 1 [(gogoproto.jsontag) = 'code'];\n    string Message = 2 [(gogoproto.jsontag) = 'message'];\n    repeated string Data = 3 [(gogoproto.jsontag) = 'data'];\n}",
		CustomCode: "package task\n\nimport (\n\t\"context\"\n\t\"fmt\"\n\tservicePb \"git.bilibili.co/platform/api-gateway/app/sayhello/service/api\"\n)\n\n\nfunc (p *TaskFlow) SayWorld1(ctx context.Context, req *servicePb.HelloReq) (reply *WorldResp, err error) {\n\treply = &WorldResp{\n\t\tContent: fmt.Sprintf(\"hello SayWorld1 %s\", req.Name),\n\t}\n\treturn\n}\n\nfunc (p *TaskFlow) SayWorld2(ctx context.Context, req *WorldResp) (reply *servicePb.HelloResp, err error) {\n\treply = &servicePb.HelloResp{\n\t\tContent: fmt.Sprintf(\"hello SayWorld2 %s\", req.Content),\n\t}\n\treturn\n}",
	}
	return
}

func (s *Service) readContentFromUrl(url string) (content string, err error) {
	var resp *http.Response
	resp, err = http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("get %s return code %d", url, resp.StatusCode)
		return
	}
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	content = string(body)
	return
}

func (s *Service) createCodeGeneratorInfo(ctx context.Context, apiInfo *model.ApiInfo, tempFilePath, goFilePath string) (info *model.CodeGeneratorInfo, err error) {
	jsonData := apiInfo.DSLCode
	var taskDSL generator.TaskDSL
	err = json.Unmarshal([]byte(jsonData), &taskDSL)
	if err != nil {
		return
	}
	var grpcInfoMap map[string]*pb.ApiInfo
	grpcInfoMap, err = s.getGrpcInfoMap(ctx, &taskDSL)
	if err != nil {
		return
	}
	var httpInfoMap map[string]*pb.ApiInfo
	httpInfoMap, err = s.getHttpInfoMap(ctx, &taskDSL)
	if err != nil {
		return
	}

	info = new(model.CodeGeneratorInfo)
	info.TmplFilePath = tmplFilePath
	info.GoFilePath = goFilePath
	info.TempFilePath = tempFilePath
	info.ServiceName = strings.Title(apiInfo.ApiName)
	info.ApiInfo = apiInfo
	info.GrpcInfoMap = grpcInfoMap
	info.HttpInfoMap = httpInfoMap
	return
}

func (s *Service) getGrpcInfoMap(ctx context.Context, taskDSL *generator.TaskDSL) (grpcInfoMap map[string]*pb.ApiInfo, err error) {
	var discoveryIdList []string
	discoveryIdList, err = generator.GetDiscoveryIdListFromTaskDSL(ctx, taskDSL)
	if err != nil {
		return
	}
	if len(discoveryIdList) == 0 {
		return
	}
	grpcInfoMap, err = s.dao.GetProtoByDis(ctx, discoveryIdList)
	return
}

func (s *Service) getHttpInfoMap(ctx context.Context, taskDSL *generator.TaskDSL) (httpInfoMap map[string]*pb.ApiInfo, err error) {
	var urlList []string
	urlList, err = generator.GetUrlListFromTaskDSL(ctx, taskDSL)
	if err != nil {
		return
	}
	if len(urlList) == 0 {
		return
	}
	httpInfoMap, err = s.dao.GetHttpApisByPath(ctx, urlList)
	return
}

func initBuilder(ctx context.Context, info *model.CodeGeneratorInfo) (builder *generator.Builder, err error) {
	var serviceInterfaceInfo *generator.ServiceInterfaceInfo
	serviceInterfaceInfo, err = initServiceInterfaceInfo(ctx, info)
	if err != nil {
		return
	}
	var taskFlowRawInfo *generator.TaskFlowRawInfo
	taskFlowRawInfo, err = initTaskFlowRawInfo(ctx, info)
	if err != nil {
		return
	}
	var rpcClientMap map[string]*generator.RpcClientInfo
	rpcClientMap, err = initRpcClientMap(info)
	if err != nil {
		return
	}
	var httpRequestMap map[string]*generator.HttpRequestInfo
	httpRequestMap, err = initHttpRequestMap(ctx, info)
	if err != nil {
		return
	}

	builder = new(generator.Builder)
	builder.TmplFilePath = info.TmplFilePath
	builder.GoFilePath = info.GoFilePath
	builder.ServiceName = info.ServiceName
	builder.ServiceProtoFile = serviceInterfaceInfo.ServiceProtoFile
	builder.ServiceInterfaceInfo = serviceInterfaceInfo
	builder.TaskFlowRawInfo = taskFlowRawInfo
	builder.RpcClientMap = rpcClientMap
	builder.HttpRequestMap = httpRequestMap
	return
}

func initServiceInterfaceInfo(ctx context.Context, info *model.CodeGeneratorInfo) (serviceInterfaceInfo *generator.ServiceInterfaceInfo, err error) {
	var fileName string
	fileName, err = generator.CreateTempFile(info.TempFilePath)
	if err != nil {
		return
	}

	rawInfo := &generator.ServiceProtoRawInfo{
		TmplFilePath: info.TmplFilePath,
		ProtoFile:    fileName,
		ServiceName:  info.ServiceName,
		FuncName:     strings.Title(info.ApiInfo.ApiName),
		Input:        info.ApiInfo.Req,
		Output:       info.ApiInfo.Reply,
	}

	err = generator.CreateServiceProtoFile(ctx, rawInfo)
	if err != nil {
		return
	}
	handlerMap := make(map[string]bool)
	handlerMap["auth"] = false
	handlerMap["verify"] = false
	for _, item := range strings.Split(info.ApiInfo.Handler, ",") {
		handlerMap[item] = true
	}

	serviceInterfaceInfo = &generator.ServiceInterfaceInfo{
		FuncName:         rawInfo.FuncName,
		InputType:        rawInfo.InputType,
		OutputType:       rawInfo.OutputType,
		IsAuthUser:       handlerMap["auth"],
		IsVerify:         handlerMap["verify"],
		ServiceProtoFile: rawInfo.ProtoFile,
	}

	if info.ApiInfo.ApiType == "http" {
		serviceInterfaceInfo.HttpPath = info.ApiInfo.Router + "/" + info.ApiInfo.ApiName
	}

	return
}

func initTaskFlowRawInfo(ctx context.Context, info *model.CodeGeneratorInfo) (taskFlowRawInfo *generator.TaskFlowRawInfo, err error) {
	var protoFile string
	protoFile, err = generator.CreateTempFile(info.TempFilePath)
	if err != nil {
		return
	}

	rawInfo := &generator.TaskProtoRawInfo{
		TmplFilePath: info.TmplFilePath,
		ProtoFile:    protoFile,
		Content:      info.ApiInfo.DSLStruct,
	}
	err = generator.CreateTaskProtoFile(ctx, rawInfo)
	if err != nil {
		return
	}

	var taskDSLFile string
	taskDSLFile, err = generator.CreateTempFile(info.TempFilePath)
	if err != nil {
		return
	}
	err = generator.WriteFile(info.ApiInfo.DSLCode, filepath.Dir(taskDSLFile), filepath.Base(taskDSLFile))
	if err != nil {
		return
	}

	var taskUdfFile string
	if info.ApiInfo.CustomCode != "" {
		taskUdfFile, err = generator.CreateTempFile(info.TempFilePath)
		if err != nil {
			return
		}
		err = generator.WriteFile(info.ApiInfo.CustomCode, filepath.Dir(taskUdfFile), filepath.Base(taskUdfFile))
		if err != nil {
			return
		}
	}
	taskFlowRawInfo = &generator.TaskFlowRawInfo{
		TaskProtoFile: rawInfo.ProtoFile,
		TaskDSLFile:   taskDSLFile,
		TaskUdfFile:   taskUdfFile,
	}
	return
}

func initRpcClientMap(info *model.CodeGeneratorInfo) (rpcClientMap map[string]*generator.RpcClientInfo, err error) {
	rpcClientMap = make(map[string]*generator.RpcClientInfo)
	for discoveryId, grpcInfo := range info.GrpcInfoMap {
		if len(grpcInfo.ServiceName) > 1 {
			err = fmt.Errorf("can not process service num > 1 with discoveryId %s", discoveryId)
			return
		}
		rpcClientInfo := &generator.RpcClientInfo{
			DiscoveryId: discoveryId,
			ServiceName: grpcInfo.ServiceName[0],
			PbAlias:     grpcInfo.PbAlias,
			PbPath:      grpcInfo.PbPath,
		}
		rpcClientMap[discoveryId] = rpcClientInfo

	}
	return
}

func initHttpRequestMap(ctx context.Context, info *model.CodeGeneratorInfo) (httpRequestMap map[string]*generator.HttpRequestInfo, err error) {
	httpRequestMap = make(map[string]*generator.HttpRequestInfo)
	packageBaseName := "httpModel"
	count := 1
	for url, httpInfo := range info.HttpInfoMap {
		packageName := packageBaseName + strconv.Itoa(count)
		var httpProtoFile string
		httpProtoFile, err = generator.CreateTempFile(info.TempFilePath)
		if err != nil {
			return
		}

		rawInfo := &generator.HttpProtoRawInfo{
			TmplFilePath: info.TmplFilePath,
			ProtoFile:    httpProtoFile,
			PackageName:  packageName,
			JsonBody:     httpInfo.Input,
			Response:     httpInfo.Output,
		}
		err = generator.CreateHttpProtoFile(ctx, rawInfo)
		if err != nil {
			return
		}

		httpRequestInfo := &generator.HttpRequestInfo{
			Url:               url,
			PackageName:       packageName,
			BodyType:          rawInfo.JsonBodyType,
			ResponseType:      rawInfo.ResponseType,
			ProtoFile:         httpProtoFile,
			ResponseCodeField: rawInfo.ResponseCodeField,
		}
		httpRequestMap[url] = httpRequestInfo
		count++
	}
	return
}

func codeGenerate(ctx context.Context, builder *generator.Builder) (output string, err error) {
	var codeGenerator *generator.CodeGenerator
	codeGenerator, err = builder.Build(ctx)
	if err != nil {
		return
	}
	err = codeGenerator.Generate(ctx)
	if err != nil {
		output = codeGenerator.GetOutPut()
		return
	}

	err = codeGenerator.Check(ctx)
	if err != nil {
		output = codeGenerator.GetOutPut()
		return
	}

	output = codeGenerator.GetOutPut()
	return
}
