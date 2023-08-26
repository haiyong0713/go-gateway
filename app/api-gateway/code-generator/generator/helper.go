package generator

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
)

type ServiceProtoRawInfo struct {
	TmplFilePath string
	ProtoFile    string
	ServiceName  string
	FuncName     string
	Input        string
	Output       string
	InputType    string
	OutputType   string
}

func CreateServiceProtoFile(ctx context.Context, rawInfo *ServiceProtoRawInfo) (err error) {
	rawInfo.InputType, err = getMessageType(ctx, rawInfo.Input)
	if err != nil {
		return
	}
	rawInfo.OutputType, err = getMessageType(ctx, rawInfo.Output)
	if err != nil {
		return
	}
	err = CreateGoFile(
		rawInfo.TmplFilePath+"/proto",
		"api.proto.tmpl",
		filepath.Dir(rawInfo.ProtoFile),
		filepath.Base(rawInfo.ProtoFile),
		rawInfo,
	)
	return
}

type TaskProtoRawInfo struct {
	TmplFilePath string
	ProtoFile    string
	Content      string
}

func CreateTaskProtoFile(ctx context.Context, rawInfo *TaskProtoRawInfo) (err error) {
	err = CreateGoFile(
		rawInfo.TmplFilePath+"/proto",
		"taskModel.proto.tmpl",
		filepath.Dir(rawInfo.ProtoFile),
		filepath.Base(rawInfo.ProtoFile),
		rawInfo,
	)
	return
}

type HttpProtoRawInfo struct {
	TmplFilePath      string
	ProtoFile         string
	PackageName       string
	JsonBody          string
	Response          string
	JsonBodyType      string
	ResponseType      string
	ResponseCodeField string
}

func CreateHttpProtoFile(ctx context.Context, rawInfo *HttpProtoRawInfo) (err error) {
	if rawInfo.JsonBody != "" {
		rawInfo.JsonBodyType, err = getMessageType(ctx, rawInfo.JsonBody)
		if err != nil {
			return
		}
	}
	rawInfo.ResponseType, err = getMessageType(ctx, rawInfo.Response)
	if err != nil {
		return
	}
	err = CreateGoFile(
		rawInfo.TmplFilePath+"/proto",
		"httpModel.proto.tmpl",
		filepath.Dir(rawInfo.ProtoFile),
		filepath.Base(rawInfo.ProtoFile),
		rawInfo,
	)
	if err != nil {
		return
	}

	var fileDescriptor *desc.FileDescriptor
	fileDescriptor, err = parseProto(rawInfo.ProtoFile)
	if err != nil {
		return
	}
	for _, messageDescriptor := range fileDescriptor.GetMessageTypes() {
		if messageDescriptor.GetName() == rawInfo.ResponseType {
			fieldDiscriptor := messageDescriptor.FindFieldByName("Code")
			if fieldDiscriptor != nil {
				rawInfo.ResponseCodeField = "Code"
			}
			break
		}
	}

	return
}

func parseProto(filename string) (fileDescriptor *desc.FileDescriptor, err error) {
	goPath := os.Getenv("GOPATH")
	goRoot := os.Getenv("GOROOT")

	Parser := protoparse.Parser{}
	Parser.ImportPaths = append(Parser.ImportPaths, "/")
	Parser.ImportPaths = append(Parser.ImportPaths, "./")
	Parser.ImportPaths = append(Parser.ImportPaths, goRoot+"/src")
	Parser.ImportPaths = append(Parser.ImportPaths, goPath+"/src")
	//加载并解析 proto文件,得到一组 FileDescriptor
	var descs []*desc.FileDescriptor
	descs, err = Parser.ParseFiles(filename)
	if err != nil {
		err = fmt.Errorf("ParseFiles %s return err %v", filename, err)
		return
	}
	if len(descs) == 0 {
		err = fmt.Errorf("can not get FileDescriptor from %s", filename)
		return
	}
	fileDescriptor = descs[0]
	return
}

func getMessageType(ctx context.Context, message string) (messageType string, err error) {
	reg := regexp.MustCompile(`message\s+(\w*)\s+`)
	if reg == nil {
		err = errors.New("regexp MustCompile err")
		return
	}

	//根据规则提取关键信息
	result := reg.FindAllStringSubmatch(message, -1)
	for _, item := range result {
		messageType = item[1]
		return
	}
	err = errors.New("regexp FindAllStringSubmatch err")
	return
}

func GetDiscoveryIdListFromTaskDSL(ctx context.Context, taskDSL *TaskDSL) (discoveryIdList []string, err error) {
	for _, taskInfo := range taskDSL.TaskList {
		if taskInfo.DiscoveryId != "" {
			discoveryIdList = append(discoveryIdList, taskInfo.DiscoveryId)
		}
	}
	return
}

func GetUrlListFromTaskDSL(ctx context.Context, taskDSL *TaskDSL) (urlList []string, err error) {
	for _, taskInfo := range taskDSL.TaskList {
		if taskInfo.Url != "" {
			urlList = append(urlList, taskInfo.Url)
		}
	}
	return
}

func GetTempPath() (path string, err error) {
	path, err = ioutil.TempDir("/tmp", "temp")
	return
}

func GetTempFile(path string) (fileName string, err error) {
	var file *os.File
	file, err = ioutil.TempFile(path, "temp")
	if err != nil {
		return
	}
	defer file.Close()
	fileName = file.Name()

	return
}
