package main

import (
	"bytes"
	"fmt"
	"go-common/app/tool/cachegen/common"
	"go/ast"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"text/template"
)

//go:generate go install
func main() {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 64*1024)
			buf = buf[:runtime.Stack(buf, false)]
			log.Fatalf("程序解析失败, err: %+v  请企业微信联系 @ouyangkeshou", err)
		}
	}()
	//os.Args = append(os.Args, "/Users/duck/go/src/go-gateway/app/web-svr/activity/interface/api/api.pb.go", "Activity")
	if len(os.Args) < 3 {
		log.Fatalf("缺少必要的启动参数\n启动参数说明:%s pb产出文件 service名字\n示例:%s api.pb.go Activity", os.Args[0], os.Args[0])
	}
	src := os.Args[1]
	service := os.Args[2]
	pkg := os.Getenv("GOPACKAGE")
	if pkg == "" {
		pkg = "api"
	}
	if !strings.HasSuffix(src, ".go") {
		log.Fatalf("%s不是合法的golang代码文件,需要.go结尾的，proto生成文件", src)
	}
	dst := strings.ReplaceAll(src, ".pb.go", ".client.go")
	if !strings.Contains(dst, ".client.go") {
		dst = dst[0:len(dst)-3] + ".client.go"
	}
	srcTyp := service + "Client"
	b, err := ioutil.ReadFile(src)
	if err != nil {
		log.Fatalf("加载%s文件异常，异常信息%+v", src, err)
	}
	s := common.NewSource(string(b))
	c := s.F.Scope.Lookup(srcTyp)
	if (c == nil) || (c.Kind != ast.Typ) {
		log.Fatalf("%s文件中找不到%s的interface定义", src, srcTyp)
	}
	typ := "local" + srcTyp
	lists := c.Decl.(*ast.TypeSpec).Type.(*ast.InterfaceType).Methods.List
	imports := make([]string, 0)
	imports = append(imports, "\"go-common/library/net/rpc/warden\"")
	imports = append(imports, "\"go-gateway/app/web-svr/activity/tools/lib/grpclocal\"")
	imported := make(map[string]struct{})
	for _, list := range lists {
		for _, i := range s.Packages(list) {
			if _, ok := imported[i]; !ok {
				imports = append(imports, i)
				imported[i] = struct{}{}
			}
		}
	}
	t := template.Must(template.New("header").Parse(_headerTemplate))
	var buffer bytes.Buffer
	if err := t.Execute(&buffer, map[string]interface{}{
		"PkgName": pkg,
		"Imports": strings.Join(imports, "\n"),
		"Type":    typ,
		"Service": service,
		"SrcType": srcTyp,
	}); err != nil {
		log.Fatalf("execute template: %s", err)
	}
	code := buffer.String() + "\n"
	t = template.Must(template.New("body").Parse(_bodyTemplate))
	for _, list := range lists {
		params := make([]string, 0)
		vars := make([]string, 0)
		for _, p := range list.Type.(*ast.FuncType).Params.List {
			params = append(params, p.Names[0].Name+" "+s.ExprString(p.Type))
			if s.ExprString(p.Type) != "...grpc.CallOption" {
				vars = append(vars, p.Names[0].Name)
			}
		}
		results := make([]string, 0)
		rets := make([]string, 0)
		rlyCount := 0
		errCount := 0
		for _, r := range list.Type.(*ast.FuncType).Results.List {
			if len(r.Names) > 0 {
				results = append(results, r.Names[0].Name+" "+s.ExprString(r.Type))
				rets = append(rets, r.Names[0].Name)
			} else {
				var name string
				switch s.ExprString(r.Type) {
				case "error":
					{
						if errCount > 0 {
							name = fmt.Sprint("err", errCount)
						} else {
							name = "err"
						}
					}
				default:
					{
						if rlyCount > 0 {
							name = fmt.Sprint("rly", errCount)
						} else {
							name = "rly"
						}
					}
				}
				results = append(results, name+" "+s.ExprString(r.Type))
				rets = append(rets, name)
			}
		}
		var buffer bytes.Buffer
		if err := t.Execute(&buffer, map[string]interface{}{
			"Method":  list.Names[0].Name,
			"Params":  strings.Join(params, ","),
			"Results": strings.Join(results, ","),
			"Rets":    strings.Join(rets, ","),
			"Vars":    strings.Join(vars, ","),
			"Type":    typ,
			"Service": service,
		}); err != nil {
			log.Fatalf("execute template: %s", err)
		}
		code = code + buffer.String() + "\n"
	}
	code = common.FormatCode(code)
	ioutil.WriteFile(dst, []byte(code), 0644)
}
