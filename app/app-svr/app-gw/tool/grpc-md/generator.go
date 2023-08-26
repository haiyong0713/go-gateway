//nolint:biliautomaxprocs
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var (
	member = make(map[string]*App)
)

func run(_ *cli.Context, flag *Flag) error {
	// Step 1: 检查并初始化参数
	if err := check(flag); err != nil {
		return err
	}
	// Step 2: 解析bapi目录，提取文件信息
	bapis := &Bapi{Hmap: make(map[string][]*BFile)}
	if err := listFile(flag.Path, "", bapis); err != nil {
		return err
	}
	// Step 3: 解析proto
	protoDesc, err := parseProto(bapis)
	if err != nil {
		return err
	}
	// Step 4: 遍历desc，解析wdcli.appid.
	parseAppID(protoDesc, bapis)
	// Step 5: 输出信息
	err = print(flag.OutPath, flag.Package)
	if err != nil {
		return err
	}
	return nil
}

func check(flag *Flag) error {
	if flag.Path == "" {
		return errors.Errorf("invalid path: %v", flag.Path)
	}
	if flag.Package == "" {
		flag.Package = "api"
	}
	//if flag.OutPath == "" {
	//	pwd, err := os.Getwd()
	//	if err != nil {
	//		return err
	//	}
	//	flag.OutPath = pwd
	//}
	return nil
}

func listFile(myfolder, prefix string, bapis *Bapi) error {
	files, _ := ioutil.ReadDir(myfolder)
	for _, file := range files {
		if file.IsDir() {
			prefixTmp := file.Name()
			if prefix != "" {
				prefixTmp = prefix + "/" + prefixTmp
			}
			if err := listFile(filepath.Join(myfolder, file.Name()), prefixTmp, bapis); err != nil {
				return err
			}
			continue
		}
		if path.Ext(file.Name()) != ".proto" {
			continue
		}
		bfile := &BFile{
			Name:    file.Name(),
			Content: "",
			Prefix:  prefix,
			Path:    myfolder,
		}
		err := bfile.ReadFile()
		if err != nil {
			return errors.WithStack(err)
		}
		bapis.Hmap[prefix] = append(bapis.Hmap[prefix], bfile)
	}
	return nil
}

func parseProto(bapis *Bapi) ([]*desc.FileDescriptor, error) {
	proto := []*desc.FileDescriptor{}
	p := protoparse.Parser{
		Accessor: protoparse.FileContentsFromMap(bapis.GetAccessorMap()),
	}
	for _, names := range bapis.GetFileNamesGroup() {
		protoTmp, err := p.ParseFiles(names...)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		proto = append(proto, protoTmp...)
	}
	return proto, nil
}

func parseAppID(protoDesc []*desc.FileDescriptor, bapis *Bapi) {
	r := regexp.MustCompile("\n.*option.*wdcli.appid.*\"(.*)\".*")
	for _, desc := range protoDesc {
		p := desc.AsFileDescriptorProto()
		proto, ok := bapis.GetContentByPrefixName(p.GetName())
		if !ok {
			continue
		}
		if sub := r.FindStringSubmatch(proto); len(sub) > 1 {
			appid := sub[1]
			services := p.GetService()
			pkg := &Package{Name: p.GetPackage(), Services: make(map[string]*Service)}
			for _, service := range services {
				serviceMethod := service.GetName()
				if p.GetPackage() != "" {
					serviceMethod = p.GetPackage() + "." + service.GetName()
				}
				methods := []string{}
				for _, method := range service.Method {
					methods = append(methods, "/"+serviceMethod+"/"+method.GetName())
				}
				pkg.Services[service.GetName()] = &Service{
					Name:    service.GetName(),
					Methods: methods,
				}
			}
			member[appid] = &App{
				Name: appid,
				Packages: map[string]*Package{
					p.GetPackage(): pkg,
				},
			}
		}
	}
}

func print(out, pkgName string) error {
	buffer := bytes.Buffer{}
	buffer.WriteString(fmt.Sprintf("package %s\n\n", pkgName))
	buffer.WriteString(`import (
	"strings"
	"regexp"
)` + "\n\n")
	buffer.Write(StrStruct())
	// write function
	buffer.WriteString(StrGetByAppID())
	buffer.WriteString(StrGetByAppIDs())
	buffer.WriteString(StrPrefixMatch())
	buffer.WriteString(StrFuzzyMatch())
	buffer.WriteString(StrGetAllAppID())
	// write var grpcMd
	buffer.WriteString("var (\n")
	buffer.WriteString("\tgrpcMd = map[string]*App{\n")
	for appid, pkg := range member {
		buffer.WriteString("\t\"")
		buffer.WriteString(appid)
		buffer.WriteString(`": `)
		buffer.Write(pkg.GetInitCode())
		buffer.WriteString(",\n")
	}
	buffer.WriteString("\t}\n)\n\n")
	// print stdout
	if out == "" {
		fmt.Println(buffer.String())
		return nil
	}
	exists, err := PathExists(out)
	if err != nil {
		return errors.WithStack(err)
	}
	if !exists {
		if err := os.Mkdir(out, os.ModePerm); err != nil {
			return errors.WithStack(err)
		}
	}
	// print file
	//nolint:gosec
	err = ioutil.WriteFile(filepath.Join(out, "grpc.md.go"), buffer.Bytes(), 0644)
	if err != nil {
		return errors.WithStack(err)
	}
	// go fmt file
	//nolint:errcheck,gosec
	exec.Command("go", "fmt", filepath.Join(out, "grpc.md.go")).Run()
	return nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func StrStruct() []byte {
	buf := bytes.Buffer{}
	// type App struct
	buf.WriteString(`type App struct {
	Name     string
	Packages map[string]*Package
}` + "\n\n")
	buf.WriteString(`func (a *App) Copy() *App {
	ret := &App{
		Name:     a.Name,
		Packages: make(map[string]*Package),
	}
	for k, v := range a.Packages {
		ret.Packages[k] = v.Copy()
	}
	return ret
}` + "\n\n")
	// type Package struct
	buf.WriteString(`type Package struct {
	Name     string
	Services map[string]*Service
}` + "\n\n")
	buf.WriteString(`func (p *Package) Copy() *Package {
	ret := &Package{
		Name:     p.Name,
		Services: make(map[string]*Service),
	}
	for k, v := range p.Services {
		ret.Services[k] = v.Copy()
	}
	return ret
}` + "\n\n")
	// type Service struct
	buf.WriteString(`type Service struct {
	Name    string
	Methods []string
}` + "\n\n")
	buf.WriteString(`func (s *Service) Copy() *Service {
	ret := &Service{
		Name: s.Name,
	}
	for _, v := range s.Methods {
		ret.Methods = append(ret.Methods, v)
	}
	return ret
}` + "\n\n")
	return buf.Bytes()
}

func StrGetByAppID() string {
	return `func GetByAppID(appid string) (*App, bool) {
	app, ok := grpcMd[appid]
	if !ok {
		return nil, false
	}
	return app.Copy(), true
}` + "\n\n"
}

func StrGetByAppIDs() string {
	return `func GetByAppIDs(appids []string) (map[string]*App, bool) {
	ret := map[string]*App{}
	for _, appid := range appids {
		app, ok := grpcMd[appid]
		if !ok {
			continue
		}
		ret[appid] = app.Copy()
	}
	if len(ret) == 0 {
		return nil, false
	}
	return ret, true
}` + "\n\n"
}

func StrPrefixMatch() string {
	return `func PrefixMatch(prefix string) (map[string]*App, bool) {
	ret := map[string]*App{}
	for key := range grpcMd {
		if strings.HasPrefix(key, prefix) {
			ret[key] = grpcMd[key].Copy()
		}
	}
	if len(ret) == 0 {
		return nil, false
	}
	return ret, true
}` + "\n\n"
}

func StrFuzzyMatch() string {
	return `func FuzzyMatch(candidate string) (map[string]*App, bool, error) {
	ret := map[string]*App{}
	r, err := regexp.Compile(candidate)
	if err != nil {
		return nil, false, err
	}
	for key := range grpcMd {
		if r.Match([]byte(key)) {
			ret[key] = grpcMd[key].Copy()
		}
	}
	if len(ret) == 0 {
		return nil, false, nil
	}
	return ret, true, nil
}` + "\n\n"
}

func StrGetAllAppID() string {
	return `func GetAllAppID() []string {
	ret := []string{}
	for appid := range grpcMd {
		ret = append(ret, appid)
	}
	return ret
}` + "\n\n"
}
