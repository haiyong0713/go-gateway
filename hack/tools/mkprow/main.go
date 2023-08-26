package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"go-gateway/hack/tools"

	"github.com/ghodss/yaml"
)

type ProwTemplateType int

const (
	HeadProwTemplateType ProwTemplateType = iota
	AlwaysProwTemplateType
	SubdivisionProwTemplateType
	ProjectProwTemplateType
	ImageProwTemplateType
	ProwJobTemplateType
	PreJenkinsTemplateType
	PostJenkinsTemplateType
	PostHeadTemplateType
)

var (
	targets = []string{
		"app/app-svr/app-player/interface",
		"app/web-svr/web/interface",
		"app/app-svr/app-interface/interface",
		"app/app-svr/app-view/interface",
		"app/app-svr/app-resource/interface",
		"app/app-svr/app-channel/interface",
		"app/app-svr/app-dynamic/interface",
		"app/app-svr/archive/service",
		"app/app-svr/app-feed/interface",
	}
)

// Image save job execute image
type Image struct {
	Image []struct {
		Name  string `yaml:"name"`
		Image string `yaml:"image"`
	} `yaml:"image"`
}

// read template file
func readTemplate(t ProwTemplateType) ([]byte, error) {
	var path string
	switch t {
	case HeadProwTemplateType:
		path = filepath.Join(tools.ProwJobTemplatePath, "head.yaml")
	case AlwaysProwTemplateType:
		path = filepath.Join(tools.ProwJobTemplatePath, "always", "template.yaml")
	case SubdivisionProwTemplateType:
		path = filepath.Join(tools.ProwJobTemplatePath, "auto", "department", "template.yaml")
	case ProjectProwTemplateType:
		path = filepath.Join(tools.ProwJobTemplatePath, "auto", "project", "template.yaml")
	case ImageProwTemplateType:
		path = filepath.Join(tools.ProwJobTemplatePath, "image.yaml")
	case PreJenkinsTemplateType:
		path = filepath.Join(tools.ProwJobTemplatePath, "auto", "jenkins", "template.yaml")
	case PostHeadTemplateType:
		path = filepath.Join(tools.ProwJobTemplatePath, "posthead.yaml")
	case PostJenkinsTemplateType:
		path = filepath.Join(tools.ProwJobTemplatePath, "postsubmit", "template.yaml")
	case ProwJobTemplateType:
		path = tools.ProwJobFilePath
	default:
		return nil, fmt.Errorf("invalid template type: %s", t)
	}
	return ioutil.ReadFile(path)
}

func fillArgs(temp []byte, paths []string, option tools.DirOptions) []byte {
	if paths != nil {
		temp = bytes.ReplaceAll(temp, []byte("<<dir>>"), []byte(strings.Join(paths[1:], "/")))
		temp = bytes.ReplaceAll(temp, []byte("<<department>>"), []byte(paths[1]))
		temp = bytes.ReplaceAll(temp, []byte("<<dir_alias>>"), []byte(strings.Join(paths, "-")))
		temp = bytes.ReplaceAll(temp, []byte("<<Optional>>"), []byte(strconv.FormatBool(!option.UnitTestRestrictive)))
		temp = bytes.ReplaceAll(temp, []byte("<<UNIT_TEST_ALL>>"), []byte(fmt.Sprintf(`"%s"`, strconv.FormatBool(option.UnitTestAll))))
		temp = bytes.ReplaceAll(temp, []byte("<<branch>>"), []byte("^master$"))
	}
	temp = bytes.ReplaceAll(temp, []byte("<<group>>"), []byte(tools.GitlabGroupName))
	temp = bytes.ReplaceAll(temp, []byte("<<repo>>"), []byte(tools.GitlabRepositoryName))

	var i Image
	yamlFile, err := readTemplate(ImageProwTemplateType)
	if err != nil {
		log.Fatal("yamlFile.Get err", err)
	}
	err = yaml.Unmarshal(yamlFile, &i)
	for _, im := range i.Image {
		temp = bytes.ReplaceAll(temp, []byte(im.Name), []byte(im.Image))
	}
	return temp
}

func walk(root string, ch chan []byte) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("unable to create test dir tree: %v\n", err)
		}

		if !info.IsDir() || path == "." {
			return nil
		}

		if tools.InBlacklist(path) {
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}

		if !tools.IsProjectRootDirectory(path) {
			return filepath.SkipDir
		}

		paths := strings.Split(path, string(filepath.Separator))
		pathsLen := len(paths)

		var t []byte

		if pathsLen == tools.SubdivisionDirectoryLevel {
			t, err = readTemplate(SubdivisionProwTemplateType)
		} else if pathsLen > tools.SubdivisionDirectoryLevel && tools.IsProject(path) {
			t, err = readTemplate(ProjectProwTemplateType)
		} else {
			return nil
		}

		if err != nil {
			return err
		}

		owner, err := tools.ReadOwner(filepath.Join(path, tools.OWNERSFileName))
		if err != nil {
			return err
		}

		ch <- fillArgs(t, paths, owner.Options)

		if pathsLen > tools.SubdivisionDirectoryLevel {
			return filepath.SkipDir
		}
		return nil
	})
}

func walkPreJenkins(root string, ch chan []byte) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("unable to create test dir tree: %v\n", err)
		}

		if !info.IsDir() || path == "." {
			return nil
		}

		if tools.InBlacklist(path) {
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}

		if !tools.IsProjectRootDirectory(path) {
			return filepath.SkipDir
		}

		paths := strings.Split(path, string(filepath.Separator))
		pathsLen := len(paths)

		var t []byte

		if pathsLen == tools.JenkinsServiceLevel {
			t, err = readTemplate(PreJenkinsTemplateType)
		} else {
			return nil
		}

		if err != nil {
			return err
		}

		if found(path, targets) {
			ch <- fillArgs(t, paths, tools.DirOptions{})
		}
		return nil
	})
}

func walkPostJenkins(root string, ch chan []byte) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("unable to create test dir tree: %v\n", err)
		}

		if !info.IsDir() || path == "." {
			return nil
		}

		if tools.InBlacklist(path) {
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}

		if !tools.IsProjectRootDirectory(path) {
			return filepath.SkipDir
		}

		paths := strings.Split(path, string(filepath.Separator))
		pathsLen := len(paths)

		var t []byte

		if pathsLen == tools.JenkinsServiceLevel {
			t, err = readTemplate(PostJenkinsTemplateType)
		} else {
			return nil
		}

		if err != nil {
			return err
		}

		if found(path, targets) {
			ch <- fillArgs(t, paths, tools.DirOptions{})
		}
		return nil
	})
}

func hasCMD(str string) bool {
	f, err := os.Stat(filepath.Join(str, "cmd"))
	if err != nil {
		return false
	} else {
		return f.IsDir()
	}
}

func main() {
	log.SetOutput(os.Stdout)

	var content []byte
	temp := make(chan []byte)
	done := make(chan int)

	// append head
	head, _ := readTemplate(HeadProwTemplateType)
	content = append(content, head...)
	// append always
	task, _ := readTemplate(AlwaysProwTemplateType)
	content = append(content, task...)

	content = fillArgs(content, nil, tools.DirOptions{})

	go func() {
		for {
			select {
			case c := <-temp:
				content = append(content, c...)
			case <-done:
				break
			}
		}
	}()

	err := walk(".", temp)
	if err != nil {
		os.Exit(1)
	}

	err = walkPreJenkins(".", temp)
	if err != nil {
		os.Exit(1)
	}

	post, _ := readTemplate(PostHeadTemplateType)
	content = append(content, post...)
	content = fillArgs(content, nil, tools.DirOptions{})

	err = walkPostJenkins(".", temp)
	if err != nil {
		os.Exit(1)
	}
	done <- 0

	oldContent, err := readTemplate(ProwJobTemplateType)
	if err != nil {
		log.Fatalf("read prow job template failed: %v", err)
	}

	err = ioutil.WriteFile(tools.ProwJobFilePath, content, 0644)
	if err != nil {
		os.Exit(1)
	}
	if bytes.Compare(oldContent, content) != 0 {
		os.Exit(-1)
	}
}

func found(path string, targets []string) bool {
	for _, target := range targets {
		if target == path {
			return true
		}
	}
	return false
}
