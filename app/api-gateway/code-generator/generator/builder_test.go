package generator

import (
	"context"
	"fmt"
	"testing"
)

// http://api.bilibili.com/x/v2/dm/ajax?aid=588097598
// uat archiveId:10113211
func TestBuild(t *testing.T) {
	serviceInterfaceInfo := &ServiceInterfaceInfo{
		FuncName:   "SayHello",
		InputType:  "HelloReq",
		OutputType: "HelloResp",
		HttpPath:   "/sayHello",
		IsAuthUser: false,
		IsVerify:   false,
	}

	taskFlowRawInfo := &TaskFlowRawInfo{
		TaskProtoFile: "../test/model.proto",
		TaskDSLFile:   "../test/taskDSL5.json",
		TaskUdfFile:   "../test/udfTask5.go.example",
	}

	rpcClientInfo := &RpcClientInfo{
		DiscoveryId: "archive.service",
		ServiceName: "Archive",
		PbAlias:     "archivePb",
		PbPath:      "/archive/service",
		//Timeout:     "1000ms",
	}
	rpcClientMap := make(map[string]*RpcClientInfo)
	rpcClientMap["archive.service"] = rpcClientInfo

	httpRequestInfo := &HttpRequestInfo{
		Url:               "http://api.bilibili.com/x/v2/dm/ajax",
		PackageName:       "httpModel1",
		BodyType:          "HttpBody",
		ResponseType:      "HttpResult",
		ProtoFile:         "../test/httpModel1.proto",
		ResponseCodeField: "Code",
	}
	httpRequestMap := make(map[string]*HttpRequestInfo)
	httpRequestMap["http://api.bilibili.com/x/v2/dm/ajax"] = httpRequestInfo

	builder := Builder{
		TmplFilePath:         "tmplFile",
		GoFilePath:           "/Users/kula/Documents/project/go",
		ServiceName:          "Demo",
		ServiceProtoFile:     "../test/api.proto",
		TaskFlowRawInfo:      taskFlowRawInfo,
		ServiceInterfaceInfo: serviceInterfaceInfo,
		RpcClientMap:         rpcClientMap,
		HttpRequestMap:       httpRequestMap,
	}

	ctx := context.Background()
	generator, err := builder.Build(ctx)
	if err != nil {
		fmt.Println(err)
	}

	err = generator.Generate(ctx)
	if err != nil {
		fmt.Println(err)
		fmt.Println("-------")
		fmt.Println(generator.GetOutPut())
		return
	}

	err = generator.Check(ctx)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("-------")
	fmt.Println(generator.GetOutPut())
}
