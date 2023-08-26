package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"

	bm "go-common/library/net/http/blademaster"
)

var _blankFiles = map[string]interface{}{
	"build/root/go_common_job.yaml": nil,
}

var (
	_httpClient *bm.Client
	_warnStat   map[string]*int
)

func checkFileContent(file, appkey, secret string) (err error) {
	var (
		f        *os.File
		respData *FilterData
		warnNum  = _warnStat[file]
	)

	if f, err = os.Open(file); err != nil {
		return errors.Errorf("Open file(%s), err(%v)", file, err)
	}
	defer f.Close()
	input := bufio.NewScanner(f)
	for i := 0; input.Scan(); i++ {
		lineStr := strings.TrimSpace(string(input.Bytes()))
		if lineStr == "" {
			continue
		}
		if respData, err = filterWord(context.TODO(), lineStr); err != nil {
			return errors.Errorf("filter file(%s), err(%v)", file, err)
		}
		if respData.Level >= _filterLevel {
			hitStr := strings.Join(respData.Hit, ", ")
			log.Printf("ERROR: file: %s, sensitive word: %s, line: %d, %s", file, hitStr, i+1, lineStr)
			*warnNum++
		}
	}
	return
}

func scanFiles(files <-chan string, maxGoroutines int, appkey, secret string) {
	log.Print("start scan changed files!")
	filesLen := len(files)
	if filesLen < maxGoroutines {
		maxGoroutines = filesLen
	}
	var wg sync.WaitGroup
	wg.Add(maxGoroutines)
	for i := 0; i < maxGoroutines; i++ {
		go func() {
			defer wg.Done()
			for file := range files {
				if err := checkFileContent(file, appkey, secret); err != nil {
					log.Printf("check file content error: %v", err)
				}
			}
		}()
	}
	wg.Wait()
	log.Print("finish scan changed files!")
}

type options struct {
	flagDebug     bool
	appKey        string
	secret        string
	maxGoroutines int
	targetFiles   []string
}

func genOptions() options {
	o := options{}

	flag.BoolVar(&o.flagDebug, "debug", false, "set true, if need print debug info")
	flag.StringVar(&o.appKey, "appkey", "", "appkey")
	flag.StringVar(&o.secret, "secret", "", "secret")
	flag.IntVar(&o.maxGoroutines, "maxGoroutines", 5, "max scan file numbers at same time")
	flag.Parse()

	if o.appKey == "" || o.secret == "" {
		o.appKey = os.Getenv("SECURITY_APPKEY")
		o.secret = os.Getenv("SECURITY_SECRET")
	}
	o.targetFiles = flag.Args()

	return o
}

func (o options) validate() error {
	if o.appKey == "" || o.secret == "" {
		return errors.New("can not get environment variable!")
	}
	return nil
}

func main() {
	log.SetOutput(os.Stdout)

	o := genOptions()
	if err := o.validate(); err != nil {
		log.Fatalf("validate failed: %v", err)
	}

	NewHttpClient(o.appKey, o.secret)
	_warnStat = make(map[string]*int)

	checkFiles := make(chan string, len(o.targetFiles))
	for _, file := range o.targetFiles {
		if _, ok := _blankFiles[file]; !ok {
			checkFiles <- file
			num := 0
			_warnStat[file] = &num
		}
	}
	close(checkFiles)

	scanFiles(checkFiles, o.maxGoroutines, o.appKey, o.secret)

	totalWarn := 0
	for _, file := range o.targetFiles {
		if _, ok := _blankFiles[file]; !ok {
			totalWarn += *_warnStat[file]
		}
	}
	if totalWarn > 0 {
		log.Fatalf("你修改的文件中包含敏感词汇共 %d 个，请确认并修改！", totalWarn)
	}
}
