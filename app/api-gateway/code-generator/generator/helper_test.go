package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestCreateServiceProtoFile(t *testing.T) {
	rawInfo := &ServiceProtoRawInfo{
		TmplFilePath: "tmplFile",
		ProtoFile:    "../test/helper_test/service.proto",
		ServiceName:  "Demo",
		FuncName:     "SayHello",
		Input:        "message HelloReq {\n  string name = 1 [(gogoproto.moretags) = 'form:\"name\" validate:\"required\"'];\n  int64 aid = 2 [(gogoproto.moretags) = 'form:\"aid\"'];\n}",
		Output:       "message HelloResp {\n  string Content = 1 [(gogoproto.jsontag) = 'content'];\n}",
	}

	ctx := context.Background()
	err := CreateServiceProtoFile(ctx, rawInfo)
	if err != nil {
		fmt.Println(err)
	}
}

func TestCreateTaskProtoFile(t *testing.T) {
	rawInfo := &TaskProtoRawInfo{
		TmplFilePath: "tmplFile",
		ProtoFile:    "../test/helper_test/taskModel.proto",
		Content:      "message WorldResp {\n    string Content = 1;\n    int64 aid = 2;\n}",
	}

	ctx := context.Background()
	err := CreateTaskProtoFile(ctx, rawInfo)
	if err != nil {
		fmt.Println(err)
	}
}

func TestCreateHttpProtoFile(t *testing.T) {
	rawInfo := &HttpProtoRawInfo{
		TmplFilePath: "tmplFile",
		ProtoFile:    "../test/helper_test/httpModel1.proto",
		PackageName:  "httpModel1",
		JsonBody:     "message HttpBody {\n    string Body = 1 [(gogoproto.jsontag) = 'body'];\n}",
		Response:     "message HttpResult {\n    int64 Code = 1 [(gogoproto.jsontag) = 'code'];\n    string Message = 2 [(gogoproto.jsontag) = 'message'];\n    repeated string Data = 3 [(gogoproto.jsontag) = 'data'];\n}",
	}

	ctx := context.Background()
	err := CreateHttpProtoFile(ctx, rawInfo)
	if err != nil {
		fmt.Println(err)
	}
	jsonData, _ := json.Marshal(rawInfo)
	fmt.Println(string(jsonData))
}

func TestGetDiscoveryIdListFromTaskDSL(t *testing.T) {
	filename := "../test/taskDSL2.json"
	jsonData, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	var taskDSL TaskDSL
	err = json.Unmarshal(jsonData, &taskDSL)
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx := context.Background()
	discoveryIdList, _ := GetDiscoveryIdListFromTaskDSL(ctx, &taskDSL)
	fmt.Println(discoveryIdList)
}

func TestGetUrlListFromTaskDSL(t *testing.T) {
	filename := "../test/taskDSL5.json"
	jsonData, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	var taskDSL TaskDSL
	err = json.Unmarshal(jsonData, &taskDSL)
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx := context.Background()
	urlList, _ := GetUrlListFromTaskDSL(ctx, &taskDSL)
	fmt.Println(urlList)
}
