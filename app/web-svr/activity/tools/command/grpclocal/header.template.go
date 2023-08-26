package main

var (
	_headerTemplate = `
// Code generated by grpclocal. DO NOT EDIT.
package {{.PkgName}}

import (
	{{.Imports}}
)
type {{.Type}} struct {
}
var (
	local{{.Service}}Server {{.Service}}Server
	_ {{.SrcType}} = &{{.Type}}{}
)
func InitLocal{{.Service}}Server(svc {{.Service}}Server){
	local{{.Service}}Server = svc
}
func NewLocalClient(cfg *warden.ClientConfig, opts ...grpc.DialOption) ({{.SrcType}}, error){
	return &{{.Type}}{},nil
}
`
)