package model

import (
	"strings"
	"time"
)

const (
	ApiTypeGrpc = 0
	ApiTypeHttp = 1
)

type ApiList struct {
	Infos []*ApiRawInfo `json:"infos"`
	Count int64         `json:"count"`
}

type ApiRawInfo struct {
	ID          int64     `json:"id"`
	DiscoveryID string    `json:"discovery_id"` //grpc
	Protocol    int8      `json:"protocol"`     //0-grpc 1-http
	ApiService  string    `json:"api_service"`  //grpc
	Method      string    `json:"method"`       //http
	ApiPath     string    `json:"api_path"`     //grpc(package/rpcName)或http
	ApiHeader   string    `json:"api_header"`   //http
	ApiParams   string    `json:"api_params"`   //http
	FormBody    string    `json:"form_body"`    //http
	JsonBody    string    `json:"json_body"`    //http
	Output      string    `json:"output"`       //grpc或http
	State       int8      `json:"state"`
	Description string    `json:"description"`
	Ctime       time.Time `json:"ctime"`
	Mtime       time.Time `json:"mtime"`
}

type ProtoInfo struct {
	ID          int64     `json:"id"`
	FilePath    string    `json:"file_path"`    //GoPath+xxx.proto
	GoPath      string    `json:"go_path"`      //提供给代码生产服务
	DiscoveryID string    `json:"discovery_id"` //wdcli.appid或者根据GoPath生成(将'/'替换为'.')
	Alias       string    `json:"alias"`        //FilePath将'/'替换为'.',提供给代码生产服务
	Package     string    `json:"package"`
	File        string    `json:"file"`
	Ctime       time.Time `json:"ctime"`
	Mtime       time.Time `json:"mtime"`
}

type AddApiReq struct {
	Method      string `json:"method" validate:"required"`
	ApiPath     string `json:"api_path" validate:"required"`
	ApiHeader   string `json:"api_header" validate:"required"`
	ApiParams   string `json:"api_params"`
	FormBody    string `json:"form_body"`
	JsonBody    string `json:"json_body"`
	Output      string `json:"output" validate:"required"`
	Description string `json:"description" validate:"required"`
}

func (a *AddApiReq) ToRawInfo() *ApiRawInfo {
	return &ApiRawInfo{
		Protocol:    ApiTypeHttp,
		Method:      a.Method,
		ApiPath:     a.ApiPath,
		ApiHeader:   a.ApiHeader,
		ApiParams:   a.ApiParams,
		FormBody:    a.FormBody,
		JsonBody:    a.JsonBody,
		Output:      a.Output,
		Description: a.Description,
	}
}

func (a *AddApiReq) Check() bool {
	switch a.Method {
	case "GET":
		if a.ApiParams == "" {
			return false
		}
	case "POST":
		switch {
		case !strings.Contains(a.ApiHeader, "application/json") && !strings.Contains(a.ApiHeader, "application/x-www-form-urlencoded"):
			return false
		case strings.Contains(a.ApiHeader, "application/json") && a.JsonBody == "":
			return false
		case strings.Contains(a.ApiHeader, "application/x-www-form-urlencoded") && a.FormBody == "":
			return false
		}
	default:
		return false
	}
	return true
}
