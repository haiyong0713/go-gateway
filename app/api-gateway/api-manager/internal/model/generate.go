package model

import (
	pb "go-gateway/app/api-gateway/api-manager/api"
	"go-gateway/app/api-gateway/code-generator/generator"
)

type CodeGeneratorInfo struct {
	Builder      *generator.Builder
	TmplFilePath string
	GoFilePath   string
	TempFilePath string
	ServiceName  string
	ApiInfo      *ApiInfo
	GrpcInfoMap  map[string]*pb.ApiInfo
	HttpInfoMap  map[string]*pb.ApiInfo
}

type CodeGeneratorReq struct {
	ApiID int64 `form:"api_id" validate:"required"`
}

type CodeGeneratorReply struct {
	WorkflowID int64 `json:"workflow_id"`
}

type ApiInfo struct {
	ApiName    string
	ApiType    string
	Router     string
	Handler    string
	Req        string
	Reply      string
	DSLCode    string
	DSLStruct  string
	CustomCode string
}
