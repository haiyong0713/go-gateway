//nolint:biliautomaxprocs
package main

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type Flag struct {
	Path    string
	Package string
	OutPath string
	Stdout  bool
}

//nolint:deadcode
type File struct {
	PackagePath string `json:"package_path"`
	Name        string `json:"name"`
}

type Bapi struct {
	Hmap map[string][]*BFile
}

func (b *Bapi) GetAccessorMap() map[string]string {
	ret := make(map[string]string)
	for _, files := range b.Hmap {
		for _, f := range files {
			key := join(f.Prefix, f.Name)
			if f.Prefix == "" {
				key = f.Name
			}
			ret[key] = f.Content
		}
	}
	return ret
}

func (b *Bapi) GetFileNamesGroup() [][]string {
	ret := [][]string{}
	for _, group := range b.Hmap {
		fileNames := []string{}
		for _, f := range group {
			name := join(f.Prefix, f.Name)
			if f.Prefix == "" {
				name = f.Name
			}
			fileNames = append(fileNames, name)
		}
		ret = append(ret, fileNames)
	}
	return ret
}

func (f *Bapi) GetContentByPrefixName(name string) (string, bool) {
	prefix, file := splitPath(name)
	bfiles, ok := f.Hmap[prefix]
	if !ok {
		return "", false
	}
	for _, f := range bfiles {
		if f.Name == file {
			return f.Content, true
		}
	}
	return "", false
}

type BFile struct {
	Name    string
	Content string
	Prefix  string
	Path    string
}

func (f *BFile) ReadFile() error {
	if f.Name == "" {
		return errors.Errorf("invalid filename: %v", f.Name)
	}
	if f.Path == "" {
		return errors.Errorf("invalid path: %v", f.Path)
	}
	content, err := ioutil.ReadFile(filepath.Join(f.Path, f.Name))
	if err != nil {
		return errors.WithStack(err)
	}
	if strings.HasPrefix(f.Prefix, "third_party") {
		f.Prefix = strings.TrimPrefix(f.Prefix, "third_party/")
	}
	f.Content = string(content)
	return nil
}

type App struct {
	Name     string
	Packages map[string]*Package
}

func (app *App) GetInitCode() []byte {
	buf := bytes.Buffer{}
	buf.WriteString("&App {Name:\"")
	buf.WriteString(app.Name)
	buf.WriteString("\",\n Packages: map[string]*Package {\n")
	for key, pkg := range app.Packages {
		buf.WriteString(`"`)
		buf.WriteString(key)
		buf.WriteString(`":`)
		buf.Write(pkg.GetInitCode())
		buf.WriteString(",\n")
	}
	buf.WriteString("},\n}")
	return buf.Bytes()
}

type Package struct {
	Name     string
	Services map[string]*Service
}

func (p *Package) GetInitCode() []byte {
	buf := bytes.Buffer{}
	buf.WriteString("&Package {Name:\"")
	buf.WriteString(p.Name)
	buf.WriteString("\",\n Services: map[string]*Service {\n")
	for key, svc := range p.Services {
		buf.WriteString(`"`)
		buf.WriteString(key)
		buf.WriteString(`":`)
		buf.Write(svc.GetInitCode())
		buf.WriteString(",\n")
	}
	buf.WriteString("},\n}")
	return buf.Bytes()
}

type Service struct {
	Name    string
	Methods []string
}

func (s *Service) GetInitCode() []byte {
	buf := bytes.Buffer{}
	buf.WriteString("&Service {Name: \"")
	buf.WriteString(s.Name)
	buf.WriteString("\",\n Methods: []string {\n")
	for _, method := range s.Methods {
		buf.WriteString(`"`)
		buf.WriteString(method)
		buf.WriteString(`"`)
		buf.WriteString(",\n")
	}
	buf.WriteString("},\n}")
	return buf.Bytes()
}

func join(elem ...string) string {
	if len(elem) == 0 {
		return ""
	}
	if len(elem) == 1 {
		return elem[0]
	}
	ret := elem[0]
	for i := 1; i < len(elem); i++ {
		ret = ret + "/" + elem[i]
	}
	return ret
}

/*
*

	 	split by "/"
		return dir, file
*/
func splitPath(path string) (string, string) {
	nodes := strings.Split(path, "/")
	if len(nodes) == 1 {
		return "", nodes[0]
	}
	dir := nodes[0]
	for i := 1; i < len(nodes)-1; i++ {
		dir = dir + "/" + nodes[i]
	}
	file := nodes[len(nodes)-1]
	return dir, file
}
