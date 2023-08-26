package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"go-common/library/log"
	"go-common/library/railgun"

	"go-gateway/app/api-gateway/api-manager/internal/model"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type rpc struct {
	funcName string
	req      string
	reqMsg   []byte
	reply    string
	repMsg   []byte
}

func (s *Service) initRg() {
	s.protoRg = railgun.NewRailGun("proto定时更新", nil, railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: "0 0/10 * * * ?"}),
		railgun.NewCronProcessor(nil, func(ctx context.Context) (policy railgun.MsgPolicy) {
			log.Warnc(ctx, "start Rg")
			s.getAllDis()
			_ = s.gitClone()
			if err := s.analysisProto(); err != nil {
				log.Errorc(ctx, "s.analysisProto error: %+v", err)
			}
			return railgun.MsgPolicyNormal
		}))
	s.protoRg.Start()
}

func (s *Service) gitClone() (err error) {
	_, err = git.PlainClone(s.gitCfg.DirPath, false, &git.CloneOptions{
		URL:      s.gitCfg.Url,
		Progress: os.Stdout,
		Auth: &http.BasicAuth{
			Username: s.gitCfg.UserName,
			Password: s.gitCfg.Token,
		},
	})
	if err != nil {
		log.Error("git clone err:%+v", err)
	}
	return
}

// 获取指定目录下的所有文件,包含子目录下的文件
func GetAllFiles(dirPth string) (files []string, err error) {
	var dirs []string
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	PthSep := string(os.PathSeparator)
	//suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	for _, fi := range dir {
		if fi.IsDir() { // 目录, 递归遍历
			dirs = append(dirs, dirPth+PthSep+fi.Name())
			//GetAllFiles(dirPth + PthSep + fi.Name())
		} else {
			// 过滤指定格式
			ok := strings.HasSuffix(fi.Name(), ".proto") && !strings.Contains(fi.Name(), "ecode")
			if ok {
				files = append(files, dirPth+PthSep+fi.Name())
			}
		}
	}

	// 读取子目录下文件
	for _, table := range dirs {
		temp, _ := GetAllFiles(table)
		for _, temp1 := range temp {
			if strings.HasSuffix(temp1, ".proto") {
				files = append(files, temp1)
			}
		}
	}

	return files, nil
}

//nolint:unparam
func regexpReplace(reg, src, temp string) (match bool, res string) {
	var result []byte
	pattern := regexp.MustCompile(reg)
	for _, subMatches := range pattern.FindAllStringSubmatchIndex(src, -1) {
		result = pattern.ExpandString(result, temp, src, subMatches)
		match = true
	}
	return match, string(result)
}

func deleteExtraSpace(s string) string {
	//删除字符串中的多余空格，有多个空格时，仅保留一个空格
	s1 := strings.Replace(s, "  ", " ", -1)     //替换tab为空格
	regStr := "\\s{2,}"                         //两个及两个以上空格的正则表达式
	reg, _ := regexp.Compile(regStr)            //编译正则表达式
	s2 := make([]byte, len(s1))                 //定义字符数组切片
	copy(s2, s1)                                //将字符串复制到切片
	spcIndex := reg.FindStringIndex(string(s2)) //在字符串中搜索
	for len(spcIndex) > 0 {                     //找到适配项
		s2 = append(s2[:spcIndex[0]+1], s2[spcIndex[1]:]...) //删除多余空格
		spcIndex = reg.FindStringIndex(string(s2))           //继续在字符串中搜索
	}
	return string(s2)
}

func (s *Service) getPath(path string) string {
	return fmt.Sprintf("%s%s", s.gitCfg.DirPath, path)
}

// proto可能存在这种情况:一个proto里有多个service,需要将rpc根据service分类
// imports 存储引入的用户自定义的proto 用于解析message内部的message或者enum
func (s *Service) readFile(path string) (pkg, discoveryID string, res map[string][]*rpc, content []byte, imports []string) {
	file, err := os.OpenFile(s.getPath(path), os.O_RDWR, 0666)
	if err != nil {
		log.Error("Open file:%s error:%+v", path, err)
		return
	}
	defer file.Close()

	buf := bufio.NewReader(file)
	res = make(map[string][]*rpc)
	imports = append(imports, path)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Warn("File:%s read ok!", path)
				break
			} else {
				log.Error("Read file:%s error:%+v", path, err)
				return
			}
		}
		contentTmp := make([]byte, len(line)) //定义字符数组切片
		copy(contentTmp, line)
		content = append(content, contentTmp...)
		line = strings.TrimSpace(line) //删除行首和行尾的空格
		line = deleteExtraSpace(line)

		if str := s.getPackage(line); str != "" {
			pkg = str
		}

		if str := s.getDiscoveryID(line); str != "" {
			discoveryID = str
		}

		if s.isInnerImport(line, path) {
			line = strings.Replace(line, "import ", "", 1)
			imports = append(imports, line)
			continue
		}

		//遍历整个文件 检索出所有的service
		funcName := `service([\s]*)(?P<var1>[a-zA-Z]+)([\s]*)\{`
		servicePattern := `$var1`
		ok, serviceName := regexpReplace(funcName, line, servicePattern)
		if !ok {
			continue
		}
		rpcs := s.getRPCs(path, serviceName)
		if len(rpcs) != 0 {
			res[serviceName] = rpcs
		}
	}
	return
}

// path格式为 /crm/service/skyeye/api.proto
func (s *Service) isInnerImport(line string, path string) bool {
	var strLen = 4
	strs := strings.Split(path, "/")
	if len(strs) < strLen {
		return false
	}
	path = fmt.Sprintf("%s/%s/%s", strs[0], strs[1], strs[2])
	return strings.Contains(line, path)
}

func (s *Service) getPackage(line string) string {
	funcName := `package([\s]*)(?P<var1>[0-9a-zA-Z.]+)([\s]*);`
	pkgPattern := `$var1`
	ok, pkg := regexpReplace(funcName, line, pkgPattern)
	if ok {
		return pkg
	}
	return ""
}

func (s *Service) getDiscoveryID(line string) string {
	disName := `wdcli.appid\)([\s]*)=([\s]*)"(?P<var1>[0-9a-zA-Z.]+)"([\s]*);`
	disPattern := `$var1`
	ok, dis := regexpReplace(disName, line, disPattern)
	if ok {
		return dis
	}
	return ""
}

func (s *Service) getRPCs(path, serviceName string) (res []*rpc) {
	file, _ := os.OpenFile(s.getPath(path), os.O_RDWR, 0666)
	defer file.Close()

	buf := bufio.NewReader(file)
	var start bool
	for {
		line, err := buf.ReadString('\n')
		if err == io.EOF {
			break
		}
		line = strings.TrimSpace(line)
		line = deleteExtraSpace(line)
		sName := fmt.Sprintf(`service([\s]*)%s([\s]*)\{`, serviceName)
		subStr := `$var1`
		if ok, _ := regexpReplace(sName, line, subStr); ok {
			start = true
			continue
		}
		if !start || strings.HasPrefix(line, "//") || line == "" {
			continue
		}
		if strings.Contains(line, "}") {
			break
		}
		if strings.HasPrefix(line, "rpc ") {
			line = strings.Replace(line, "	", "", -1) //删除tab
			line = strings.Replace(line, " ", "", -1) //删除空格

			funcName := `rpc(?P<var1>[0-9a-zA-Z]+)\(`
			subStr := `$var1`
			ok1, res1 := regexpReplace(funcName, line, subStr)
			if !ok1 {
				continue
			}

			reqName := res1 + `\((?P<var1>[0-9a-zA-Z]+)\)`
			ok2, res2 := regexpReplace(reqName, line, subStr)
			if !ok2 {
				continue
			}

			repName := `returns\((?P<var1>[0-9a-zA-Z]+)\)`
			ok3, res3 := regexpReplace(repName, line, subStr)
			if !ok3 {
				continue
			}
			item := &rpc{
				funcName: res1,
				req:      res2,
				reply:    res3,
			}
			res = append(res, item)
		}
	}
	return
}

func (s *Service) getMessage(allContent []string, name string) (rs []byte) {
	var (
		count int
		index int
	)
	for i, line := range allContent {
		lineTmp := line
		line = strings.TrimSpace(line)
		line = deleteExtraSpace(line)
		if strings.HasPrefix(line, "//") || line == "" {
			continue
		}
		funcName := `message([\s]*)(?P<var1>[0-9a-zA-Z]+)([\s]*)\{`
		s3 := `$var1`
		ok1, res1 := regexpReplace(funcName, line, s3)
		if !ok1 || res1 != name {
			continue
		}
		rs = append(rs, []byte(lineTmp)...)
		count = 1
		index = i + 1
		if strings.Contains(lineTmp, "{") && strings.Contains(lineTmp, "}") {
			return
		}
		break
	}
	for count > 0 && index < len(allContent) {
		line := allContent[index]
		rs = append(rs, []byte(line)...)
		index++
		funcName := `(?P<var1>[\{\}])`
		s3 := `$var1`
		ok1, res1 := regexpReplace(funcName, line, s3)
		if !ok1 {
			continue
		}
		if res1 == "{" {
			count++
		}
		if res1 == "}" {
			count--
		}
	}
	return
}

func (s *Service) analysisProto() (err error) {
	//获取所有的*.proto文件的绝对路径
	files, err := GetAllFiles(s.gitCfg.DirPath)
	if err != nil {
		log.Error("analysisProto error:%+v", err)
		return
	}

	//去除路径前缀 /tmp/bapi/account/service/api.proto ==> /account/service/api.proto
	for i := range files {
		files[i] = strings.Replace(files[i], s.gitCfg.DirPath, "", 1)
		//获取整个proto里的rpc接口 content用于落库 allContent用于搜索message
		pkg, discoveryID, rs, content, _ := s.readFile(files[i])
		if len(rs) == 0 {
			log.Warn("no rpc discoveryID:%v", discoveryID)
			continue
		}

		pathStrs := strings.Split(files[i], "/")
		var pathLen = 3
		if len(pathStrs) < pathLen {
			log.Warn("wrong path len discoveryID:%v", discoveryID)
			continue
		}
		alias := strings.Join(pathStrs[1:len(pathStrs)-1], "_")
		goPath := strings.Join(pathStrs[:len(pathStrs)-1], "/")

		if discoveryID == "" {
			discoveryID = strings.Join(pathStrs[1:len(pathStrs)-1], ".")
		}
		fmt.Println(pkg, discoveryID, alias, goPath)

		allContent := s.mergeFiles([]string{files[i]})

		s.mutex.Lock()
		proto, ok := s.allProtos[discoveryID]
		s.mutex.Unlock()
		if ok && proto != nil && proto.File == string(content) {
			continue
		}

		for svc, rpcs := range rs {
			for _, rpc := range rpcs {
				rpc.reqMsg = s.getMessage(allContent, rpc.req)
				rpc.repMsg = s.getMessage(allContent, rpc.reply)
				r := &model.ApiRawInfo{
					DiscoveryID: discoveryID,
					ApiService:  svc,
					ApiPath:     fmt.Sprintf("/%s.%s/%s", pkg, svc, rpc.funcName),
					JsonBody:    string(rpc.reqMsg),
					Output:      string(rpc.repMsg),
				}
				err = s.dao.AddApi(context.Background(), r)
				if err != nil {
					time.Sleep(time.Second)
				}
			}
		}

		pro := &model.ProtoInfo{
			FilePath:    files[i],
			GoPath:      goPath,
			DiscoveryID: discoveryID,
			Alias:       alias,
			Package:     pkg,
			File:        string(content),
		}
		s.mutex.Lock()
		s.allProtos[discoveryID] = pro
		s.mutex.Unlock()
		_ = s.dao.AddProto(context.Background(), pro)
	}
	return
}

// 将多个proto文件合并,用于搜索出全部的message
func (s *Service) mergeFiles(files []string) []string {
	res := make([]string, 0)
	for _, path := range files {
		f, err := os.OpenFile(s.getPath(path), os.O_RDWR, 0666)
		if err != nil {
			log.Error("Open file:%s error:%+v", path, err)
			_ = f.Close()
			return nil
		}

		buf := bufio.NewReader(f)
		for {
			line, err := buf.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					log.Warn("File:%s read ok!", path)
					_ = f.Close()
					break
				} else {
					log.Error("Read file:%s error:%+v", path, err)
					_ = f.Close()
					return nil
				}
			}
			res = append(res, line)
		}
	}
	return res
}
