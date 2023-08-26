package preferenceproto

import (
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"go-gateway/app/app-svr/distribution/distribution/embed"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/pkg/errors"
	statik "github.com/rakyll/statik/fs"
)

type BAPIProto struct {
	Meta           *ProtoMeta
	FileDescriptor *desc.FileDescriptor
}

type BAPIProtos []*BAPIProto

type ProtoMeta struct {
	ProtoPath  string
	ImportPath string
	Content    []byte `json:"-"`
}

type ProtoMetas []*ProtoMeta

type ProtoStore struct {
	Meta      ProtoMetas
	ALLProtos BAPIProtos
}

func NewEmbed() (*ProtoStore, error) {
	protoFS, err := statik.NewWithNamespace(embed.Deivcesetting)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	entrypoint := "/"
	file, err := protoFS.Open(entrypoint)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	rawProto, err := collectRawProto(protoFS, entrypoint, file)
	if err != nil {
		return nil, err
	}
	ps := &ProtoStore{
		Meta: rawProto,
	}
	allProto, err := ps.Meta.ParseALL()
	if err != nil {
		return nil, err
	}
	ps.ALLProtos = allProto
	return ps, nil
}

func collectRawProto(protoFS http.FileSystem, entrypoint string, in http.File) (ProtoMetas, error) {
	stat, err := in.Stat()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !stat.IsDir() {
		return nil, errors.Errorf("%q is not a directory", stat.Name())
	}
	fileInfos, err := in.Readdir(-1)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	rawProto := ProtoMetas{}
	for _, fi := range fileInfos {
		if fi.IsDir() {
			secEntrypoint := path.Join(entrypoint, fi.Name())
			subDir, err := protoFS.Open(secEntrypoint)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			collected, err := collectRawProto(protoFS, secEntrypoint, subDir)
			if err != nil {
				return nil, err
			}
			rawProto = append(rawProto, collected...)
			continue
		}
		if filepath.Ext(fi.Name()) != ".proto" {
			continue
		}
		pm := &ProtoMeta{
			ProtoPath: path.Join(entrypoint, fi.Name()),
		}
		pm.ImportPath = strings.TrimPrefix(pm.ProtoPath, "/")
		f, err := protoFS.Open(pm.ProtoPath)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		content, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		pm.Content = content
		rawProto = append(rawProto, pm)
	}
	return rawProto, nil
}

func (pms ProtoMetas) TrimImportPath(prefix []string) []*ProtoMeta {
	out := make([]*ProtoMeta, 0, len(pms))
	for _, pm := range pms {
		for _, toTrim := range prefix {
			dup := *pm
			dup.ImportPath = strings.TrimPrefix(pm.ImportPath, toTrim)
			out = append(out, &dup)
		}
	}
	return out
}

func (pms ProtoMetas) asProtoFileContents() map[string]string {
	out := make(map[string]string, len(pms))
	for _, pm := range pms {
		out[pm.ImportPath] = string(pm.Content)
	}
	return out
}

func (pms ProtoMetas) groupByImportPath() map[string]ProtoMetas {
	grouped := map[string]ProtoMetas{}
	for _, pm := range pms {
		dir, _ := filepath.Split(pm.ImportPath)
		if _, ok := grouped[dir]; !ok {
			grouped[dir] = ProtoMetas{}
		}
		grouped[dir] = append(grouped[dir], pm)
	}
	return grouped
}

func (pms ProtoMetas) allImportPath() []string {
	out := make([]string, 0, len(pms))
	for _, pm := range pms {
		out = append(out, pm.ImportPath)
	}
	return out
}

func (pms ProtoMetas) ParseALL() (BAPIProtos, error) {
	out := BAPIProtos{}
	parser := protoparse.Parser{
		Accessor:              protoparse.FileContentsFromMap(pms.asProtoFileContents()),
		IncludeSourceCodeInfo: true,
	}
	// parse by directory
	for _, subMetas := range pms.groupByImportPath() {
		parsed, err := parser.ParseFiles(subMetas.allImportPath()...)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if len(parsed) != len(subMetas) {
			return nil, errors.Errorf("inconsistent parsed proto file: %+v, %+v", parsed, subMetas)
		}
		for i := range parsed {
			bp := &BAPIProto{
				Meta:           subMetas[i],
				FileDescriptor: parsed[i],
			}
			out = append(out, bp)
		}
	}
	return out, nil
}

func (bps BAPIProtos) Find(fn func(*BAPIProto) bool) (*BAPIProto, bool) {
	for _, bp := range bps {
		if fn(bp) {
			return bp, true
		}
	}
	return nil, false
}

func (bps BAPIProtos) Filter(fn func(*BAPIProto) bool) BAPIProtos {
	out := BAPIProtos{}
	for _, bp := range bps {
		if fn(bp) {
			out = append(out, bp)
		}
	}
	return out
}

func (bps BAPIProtos) Iter(fn func(*BAPIProto) bool) {
	for _, bp := range bps {
		if !fn(bp) {
			break
		}
	}
}
