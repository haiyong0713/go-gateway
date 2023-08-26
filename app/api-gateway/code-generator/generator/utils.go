package generator

import (
	"fmt"
	//	tmplFile "go-gateway/app/api-gateway/code-generator/tmpl"
	"io/ioutil"
	"os"
	"os/exec"
)

func ExecCommand(command string) (res string, err error) {
	sh := os.Getenv("SHELL")
	if sh == "" {
		sh = "/bin/bash"
	}
	fmt.Printf("exec command: %s -c %s\n\n", sh, command)
	cmd := exec.Command(sh, "-c", command)
	var output []byte
	if output, err = cmd.CombinedOutput(); err != nil {
		res = string(output)
		return
	}
	res = string(output)
	fmt.Println(res)
	return
}

func CreateGoFile(src, srcFile, dest, destFile string, data interface{}) (err error) {
	/*
		var fileData []byte
		fileData, err = tmplFile.Fs.ReadFile(src + "/" + srcFile)
		if err != nil {
			fmt.Println(src + "/" + srcFile)
			return
		}
		var tmpl *template.Template
		tmpl, err = template.New("").Parse(string(fileData))
		//tmpl, err := template.ParseFiles(src + "/" + srcFile)
		if err != nil {
			return
		}

		var file *os.File
		err = os.MkdirAll(dest, os.ModePerm)
		if err != nil {
			return
		}
		if file, err = os.Create(dest + "/" + destFile); err != nil {
			return
		}
		defer file.Close()
		err = tmpl.Execute(file, data)
		if err != nil {
			return
		}
	*/
	return
}

func CopyFile(src, srcFile, dest, destFile string) (err error) {
	err = os.MkdirAll(dest, os.ModePerm)
	if err != nil {
		return
	}
	var input []byte
	input, err = ioutil.ReadFile(src + "/" + srcFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ioutil.WriteFile(dest+"/"+destFile, input, 0644)
	if err != nil {
		fmt.Println("Error creating", dest+"/"+destFile)
		fmt.Println(err)
		return
	}
	return
}

func WriteFile(content, path, file string) (err error) {
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return
	}
	err = ioutil.WriteFile(path+"/"+file, []byte(content), 0644)
	if err != nil {
		fmt.Println("Error creating", path+"/"+file)
		fmt.Println(err)
		return
	}
	return
}

func CreateTempPath() (path string, err error) {
	path, err = ioutil.TempDir("/tmp", "temp")
	if err != nil {
		return
	}
	return
}

func CreateTempFile(path string) (fileName string, err error) {
	var file *os.File
	file, err = ioutil.TempFile(path, "temp")
	if err != nil {
		return
	}
	defer file.Close()
	fileName = file.Name()
	return
}
