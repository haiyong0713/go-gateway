package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
)

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil || os.IsExist(err)
}

func FileIsDir(filePath string) bool {
	f, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return f.IsDir()
}

func ReadFile(filePath string) (string, error) {
	body, err := ioutil.ReadFile(filePath)
	return string(body), err
}

func WriteFile(filePath, content string) (err error) {
	//nolint:gosec
	err = ioutil.WriteFile(filePath, []byte(content), 0644)
	return
}

func ReplaceFileText(filePath, output, old, new string, n int) (err error) {
	var (
		body []byte
	)
	if body, err = ioutil.ReadFile(filePath); err != nil {
		return
	}
	rewriteStr := strings.Replace(string(body), old, new, n)
	//nolint:gosec
	err = ioutil.WriteFile(output, []byte(rewriteStr), 0644)
	return
}

func FileCopy(filePath, dest string) (err error) {
	if FileExists(dest) {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("文件已存在 - path: %s", dest))
		return
	}
	if err = os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return
	}
	in, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

// MultipartFileCopy 将文件拷贝到dir文件夹下，文件名为header中的文件名。
// 如果dir不存在则创建，如果dir下存在同名文件则报错。
func MultipartFileCopy(file multipart.File, header *multipart.FileHeader, dir string) (filePath string, err error) {
	dest := path.Join(dir, header.Filename)
	var destFile *os.File
	if FileExists(dest) {
		err = ecode.Error(ecode.ServerErr, fmt.Sprintf("文件已存在 - path: %s", dest))
		return
	}
	if err = os.MkdirAll(dir, 0777); err != nil {
		return
	}
	if destFile, err = os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0777); err != nil {
		return
	}
	defer destFile.Close()
	if _, err = io.Copy(destFile, file); err != nil {
		return
	}
	defer file.Close()
	return dest, err
}

// DirSizeB get file size by path return bytes
func DirSizeB(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

// CopyDir Recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
func CopyDir(source string, destination string) (err error) {

	// get properties of source dir
	sourceFile, err := os.Stat(source)
	if err != nil {
		return err
	}

	if !sourceFile.IsDir() {
		return ecode.String("Source is not a directory")
	}

	// ensure destination dir does not already exist
	_, err = os.Open(destination)
	if !os.IsNotExist(err) {
		return ecode.String("Destination already exists")
	}

	// create destination dir
	err = os.MkdirAll(destination, sourceFile.Mode())
	if err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(source)

	for _, entry := range entries {
		sourcePath := source + "/" + entry.Name()
		destinationPath := destination + "/" + entry.Name()
		if entry.IsDir() {
			err = CopyDir(sourcePath, destinationPath)
			if err != nil {
				log.Error("%v", err)
			}
		} else {
			// perform copy
			err = CopyFile(sourcePath, destinationPath)
			if err != nil {
				log.Error("%v", err)
			}
		}
	}
	return
}

// CopyFile Copies original source to destination destination.
func CopyFile(source string, destination string) (err error) {
	originalFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func(originalFile *os.File) {
		err := originalFile.Close()
		if err != nil {
			return
		}
	}(originalFile)
	destinationFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer func(destinationFile *os.File) {
		err := destinationFile.Close()
		if err != nil {
			return
		}
	}(destinationFile)
	_, err = io.Copy(destinationFile, originalFile)
	if err == nil {
		info, err1 := os.Stat(source)
		if err1 != nil {
			err = os.Chmod(destination, info.Mode())
		}
	}
	return
}

var suffixes = []string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}

func HumanFileSize(size float64) string {
	if size <= 0 {
		return "0B"
	}
	base := math.Log(size) / math.Log(1024)
	getSize := round(math.Pow(1024, base-math.Floor(base)), .5, 2)
	getSuffix := suffixes[int(math.Floor(base))]
	return strconv.FormatFloat(getSize, 'f', -1, 64) + " " + getSuffix
}

func round(val float64, roundOn float64, places int) (newVal float64) {
	var round float64
	pow := math.Pow(10, float64(places))
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	newVal = round / pow
	return
}
