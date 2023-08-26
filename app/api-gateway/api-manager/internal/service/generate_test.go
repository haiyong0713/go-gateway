package service

import (
	"context"
	"fmt"
	pb "go-gateway/app/api-gateway/api-manager/api"
	"go-gateway/app/api-gateway/api-manager/internal/model"
	"go-gateway/app/api-gateway/code-generator/generator"
	"testing"
)

func TestUdfTask(t *testing.T) {
	ctx := context.Background()
	var err error
	apiInfo := &model.ApiInfo{
		ApiName:    "sayHello",
		ApiType:    "http",
		Router:     "/test",
		Req:        "message HelloReq {\n  string name = 1 [(gogoproto.moretags) = 'form:\"name\" validate:\"required\"'];\n  int64 aid = 2 [(gogoproto.moretags) = 'form:\"aid\"'];\n}",
		Reply:      "message HelloResp {\n  string Content = 1 [(gogoproto.jsontag) = 'content'];\n}",
		DSLCode:    "{\n  \"taskList\": [\n    {\n      \"name\": \"SayWorld1\",\n      \"type\": \"udf\",\n      \"input\": [\n        {\n          \"value\": \"$service.request\"\n        }\n      ],\n      \"output\": {\n        \"type\": \"WorldResp\"\n      }\n    },\n    {\n      \"name\": \"SayWorld2\",\n      \"type\": \"udf\",\n      \"input\": [\n        {\n          \"value\": \"$task.SayWorld1\"\n        }\n      ],\n      \"output\": {\n        \"type\": \"servicePb.HelloResp\",\n        \"isResponse\": true\n      }\n    }\n  ]\n}",
		DSLStruct:  "message WorldResp {\n    string Content = 1;\n    int64 aid = 2;\n}\n\nmessage HttpResult {\n    int64 Code = 1 [(gogoproto.jsontag) = 'code'];\n    string Message = 2 [(gogoproto.jsontag) = 'message'];\n    repeated string Data = 3 [(gogoproto.jsontag) = 'data'];\n}",
		CustomCode: "package task\n\nimport (\n\t\"context\"\n\t\"fmt\"\n\tservicePb \"git.bilibili.co/platform/api-gateway/app/sayhello/service/api\"\n)\n\n\nfunc (p *TaskFlow) SayWorld1(ctx context.Context, req *servicePb.HelloReq) (reply *WorldResp, err error) {\n\treply = &WorldResp{\n\t\tContent: fmt.Sprintf(\"hello SayWorld1 %s\", req.Name),\n\t}\n\treturn\n}\n\nfunc (p *TaskFlow) SayWorld2(ctx context.Context, req *WorldResp) (reply *servicePb.HelloResp, err error) {\n\treply = &servicePb.HelloResp{\n\t\tContent: fmt.Sprintf(\"hello SayWorld2 %s\", req.Content),\n\t}\n\treturn\n}",
	}
	info := &model.CodeGeneratorInfo{
		TmplFilePath: "tmplFile",
		GoFilePath:   "/Users/kula/Documents/project/go",
		TempFilePath: "/tmp",
		ServiceName:  "SayHello",
		ApiInfo:      apiInfo,
	}

	var builder *generator.Builder
	builder, err = initBuilder(ctx, info)
	if err != nil {
		fmt.Println(err)
		return
	}
	var output string
	output, err = codeGenerate(ctx, builder)
	fmt.Println(err)
	fmt.Println(output)
}

func TestGrpcTask(t *testing.T) {
	ctx := context.Background()
	var err error
	apiInfo := &model.ApiInfo{
		ApiName:    "sayHello",
		ApiType:    "http",
		Router:     "/test",
		Req:        "message HelloReq {\n  string name = 1 [(gogoproto.moretags) = 'form:\"name\" validate:\"required\"'];\n  int64 aid = 2 [(gogoproto.moretags) = 'form:\"aid\"'];\n}",
		Reply:      "message HelloResp {\n  string Content = 1 [(gogoproto.jsontag) = 'content'];\n}",
		DSLCode:    "{\n  \"taskList\": [\n    {\n      \"name\": \"SayWorld1\",\n      \"type\": \"mapping\",\n      \"input\": [\n        {\n          \"name\": \"req\",\n          \"value\": \"$service.request\"\n        }\n      ],\n      \"output\": {\n        \"type\": \"archivePb.ArcRequest\"\n      },\n      \"mappingRule\": [\n        {\n          \"src\": \"req.Aid\",\n          \"dest\": \"Aid\"\n        }\n      ]\n    },\n    {\n      \"name\": \"SayWorld2\",\n      \"type\": \"grpc\",\n      \"input\": [\n        {\n          \"value\": \"$task.SayWorld1\"\n        }\n      ],\n      \"output\": {\n        \"type\": \"archivePb.ArcReply\"\n      },\n      \"discoveryId\": \"archive.service\",\n      \"rpcInterface\": \"Arc\"\n    },\n    {\n      \"name\": \"SayWorld3\",\n      \"type\": \"udf\",\n      \"input\": [\n        {\n          \"value\": \"$task.SayWorld2\"\n        }\n      ],\n      \"output\": {\n        \"type\": \"servicePb.HelloResp\",\n        \"isResponse\": true\n      }\n    }\n  ]\n}",
		DSLStruct:  "message WorldResp {\n    string Content = 1;\n    int64 aid = 2;\n}\n\nmessage HttpResult {\n    int64 Code = 1 [(gogoproto.jsontag) = 'code'];\n    string Message = 2 [(gogoproto.jsontag) = 'message'];\n    repeated string Data = 3 [(gogoproto.jsontag) = 'data'];\n}",
		CustomCode: "package task\n\nimport (\n    \"context\"\n    \"encoding/json\"\n    \"fmt\"\n    archivePb \"git.bilibili.co/bapis/bapis-go/archive/service\"\n    servicePb \"git.bilibili.co/platform/api-gateway/app/sayhello/service/api\"\n)\n\nfunc (p *TaskFlow) SayWorld3(ctx context.Context, req *archivePb.ArcReply) (reply *servicePb.HelloResp, err error) {\n    data, _ := json.Marshal(req)\n    reply = &servicePb.HelloResp{\n        Content: fmt.Sprintf(\"hello SayWorld3: %s\", string(data)),\n    }\n    return\n}",
	}
	grpcInfoMap := make(map[string]*pb.ApiInfo)
	grpcInfo := &pb.ApiInfo{
		ServiceName: []string{"Archive"},
		PbAlias:     "archivePb",
		PbPath:      "/archive/service",
	}
	grpcInfoMap["archive.service"] = grpcInfo
	info := &model.CodeGeneratorInfo{
		TmplFilePath: "tmplFile",
		GoFilePath:   "/Users/kula/Documents/project/go",
		TempFilePath: "/tmp",
		ServiceName:  "SayHello",
		ApiInfo:      apiInfo,
		GrpcInfoMap:  grpcInfoMap,
	}

	var builder *generator.Builder
	builder, err = initBuilder(ctx, info)
	if err != nil {
		fmt.Println(err)
		return
	}
	var output string
	output, err = codeGenerate(ctx, builder)
	fmt.Println(err)
	fmt.Println(output)
}

func TestHttpTask(t *testing.T) {
	ctx := context.Background()
	var err error
	apiInfo := &model.ApiInfo{
		ApiName:    "sayHello",
		ApiType:    "http",
		Router:     "/test",
		Req:        "message HelloReq {\n  string name = 1 [(gogoproto.moretags) = 'form:\"name\" validate:\"required\"'];\n  int64 aid = 2 [(gogoproto.moretags) = 'form:\"aid\"'];\n}",
		Reply:      "message HelloResp {\n  string Content = 1 [(gogoproto.jsontag) = 'content'];\n}",
		DSLCode:    "{\n  \"taskList\": [\n    {\n      \"name\": \"SayWorld1\",\n      \"type\": \"http\",\n      \"input\": [\n        {\n          \"name\": \"req\",\n          \"value\": \"$service.request\"\n        }\n      ],\n      \"output\": {\n        \"type\": \"HttpResult\"\n      },\n      \"url\": \"http://api.bilibili.com/x/v2/dm/ajax\",\n      \"method\": \"GET\",\n      \"query\": [\n        {\n          \"src\": \"req.Aid\",\n          \"dest\": \"aid\",\n          \"mapFunc\": \"int64Tostring\"\n        }\n      ]\n    },\n    {\n      \"name\": \"SayWorld2\",\n      \"type\": \"udf\",\n      \"input\": [\n        {\n          \"value\": \"$task.SayWorld1\"\n        }\n      ],\n      \"output\": {\n        \"type\": \"servicePb.HelloResp\",\n        \"isResponse\": true\n      }\n    }\n  ]\n}\n",
		DSLStruct:  "message WorldResp {\n    string Content = 1;\n    int64 aid = 2;\n}\n\nmessage HttpResult {\n    int64 Code = 1 [(gogoproto.jsontag) = 'code'];\n    string Message = 2 [(gogoproto.jsontag) = 'message'];\n    repeated string Data = 3 [(gogoproto.jsontag) = 'data'];\n}",
		CustomCode: "package task\n\nimport (\n    \"context\"\n    \"encoding/json\"\n    \"fmt\"\n    servicePb \"git.bilibili.co/platform/api-gateway/app/sayhello/service/api\"\n)\n\nfunc (p *TaskFlow) SayWorld2(ctx context.Context, req *HttpResult) (reply *servicePb.HelloResp, err error) {\n    var byteData []byte\n    byteData, err = json.Marshal(req.Data)\n    reply = &servicePb.HelloResp{\n        Content: fmt.Sprintf(\"hello SayWorld2 %s\", string(byteData)),\n    }\n    return\n}",
	}
	httpInfoMap := make(map[string]*pb.ApiInfo)
	httpInfo := &pb.ApiInfo{
		Input:  "message HttpBody {\n    string Body = 1 [(gogoproto.jsontag) = 'body'];\n}",
		Output: "message HttpResult {\n    int64 Code = 1 [(gogoproto.jsontag) = 'code'];\n    string Message = 2 [(gogoproto.jsontag) = 'message'];\n    repeated string Data = 3 [(gogoproto.jsontag) = 'data'];\n}",
	}
	httpInfoMap["http://api.bilibili.com/x/v2/dm/ajax"] = httpInfo
	info := &model.CodeGeneratorInfo{
		TmplFilePath: "tmplFile",
		GoFilePath:   "/Users/kula/Documents/project/go",
		TempFilePath: "/tmp",
		ServiceName:  "SayHello",
		ApiInfo:      apiInfo,
		HttpInfoMap:  httpInfoMap,
	}

	var builder *generator.Builder
	builder, err = initBuilder(ctx, info)
	if err != nil {
		fmt.Println(err)
		return
	}
	var output string
	output, err = codeGenerate(ctx, builder)
	fmt.Println(err)
	fmt.Println(output)
}
