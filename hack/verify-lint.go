package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type checkPolicy int

const (
	_nonePolicy checkPolicy = iota
	_bugPolicy
	_smellPolicy
	_allPolicy
)

var (
	_err     = errors.New("lint failed")
	_pathCfg = []rule{
		{path: "app", policy: _allPolicy},
	}
)

type rule struct {
	path   string
	policy checkPolicy
}

func check(path string) checkPolicy {
	for _, v := range _pathCfg {
		if strings.HasPrefix(path, v.path) {
			return v.policy
		}
	}
	return _nonePolicy
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("未指定项目路径")
		os.Exit(1)
	}
	path := os.Args[1]
	lintPath := getLintPath()
	var (
		e1, e2 error
		run    bool
		err    error
		result string
	)
	if check(path) == _allPolicy || check(path) == _bugPolicy {
		fmt.Printf("check bug path: %s\n", path)
		result, err = runLint(path, lintPath)
		run = true
		checkErr("check bug run lint", err, result)
		e1 = checkBug(result)
	}
	if check(path) == _allPolicy || check(path) == _smellPolicy {
		if !run {
			result, err = runLint(path, lintPath)
		}
		fmt.Printf("check smell path: %s\n", path)
		checkErr("check smell run lint", err, result)
		e2 = checkSmell(result)
	}
	if e1 != nil || e2 != nil {
		os.Exit(1)
	}
	fmt.Printf("%s 检测通过\n", path)
}

func getLintPath() string {
	path := "/tmp/lint"
	if runtime.GOOS != "darwin" {
		checkErr("curl", simpleCmd("curl", []string{"-o", path, "http://172.22.34.51:7000/lint/linux-amd64/golangci-lint"}), "")
		checkErr("chmod", simpleCmd("chmod", []string{"+x", path}), "")
		return path
	}
	return "lint"
}

func checkBug(text string) error {
	var (
		errorArr []string
	)
	for _, row := range strings.Split(text, "\n") {
		if strings.Contains(row, "::error") {
			errorArr = append(errorArr, row)
		}
	}
	if len(errorArr) > 0 {
		fmt.Printf("请解决 %d 个错误\n", len(errorArr))
		for _, row := range errorArr {
			fmt.Println(row)
		}
		fmt.Println()
		return _err
	}
	return nil
}

func checkSmell(text string) error {
	var (
		infoArr []string
	)
	for _, row := range strings.Split(text, "\n") {
		if strings.Contains(row, "::info") {
			infoArr = append(infoArr, row)
		}
	}
	if len(infoArr) > 0 {
		fmt.Printf("请解决 %d 个异味\n", len(infoArr))
		for _, row := range infoArr {
			fmt.Println(row)
		}
		fmt.Println()
		return _err
	}
	return nil
}

func checkErr(msg string, err error, extra string) {
	if err != nil {
		fmt.Printf("msg: %s err: %+v\n%s\n", msg, err, extra)
		os.Exit(1)
	}
}

func runLint(path, lintPath string) (string, error) {
	args := []string{"run", "-c", "./.golangci.yml", path + "/...", "--timeout=30m", "--out-format=github-actions", "--issues-exit-code=0", "--max-issues-per-linter=0", "--max-same-issues=0", "--allow-serial-runners=true", "--allow-parallel-runners=true"}
	return runCmd(lintPath, args)
}

func runCmd(cmd string, args []string) (stdout string, err error) {
	buf := &bytes.Buffer{}
	commond := exec.Command(cmd, args...)
	commond.Env = os.Environ()
	commond.Stdout = buf
	commond.Stderr = os.Stderr
	err = commond.Run()
	stdout = buf.String()
	return
}

func simpleCmd(cmd string, args []string) (err error) {
	commond := exec.Command(cmd, args...)
	commond.Env = os.Environ()
	commond.Stdout = os.Stdout
	commond.Stderr = os.Stderr
	return commond.Run()
}
