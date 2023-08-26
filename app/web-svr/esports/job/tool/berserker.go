package tool

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

// Reference: 数据盘 >>> https://info.bilibili.co/pages/viewpage.action?pageId=7542783
type BerserkerCfg struct {
	ArchiveScoreInDB bool
	AppKey           string
	AppSecret        string
	CronSpec         string
	Enabled          bool
	Host             string
	KeepBackupFile   bool
}

type berserkerCreteJobRes struct {
	Code         int    `json:"code"`
	Msg          string `json:"msg"`
	JobStatusUrl string `json:"jobStatusUrl"`
}

type berserkerJobStatusRes struct {
	Code      int      `json:"code"`
	Msg       string   `json:"msg"`
	StatusID  int      `json:"statusId"`
	StatusMsg string   `json:"statusMsg"`
	FileUrls  []string `json:"hdfsPath"`
}

func (cfg BerserkerCfg) Validate() error {
	if cfg.Host == "" || cfg.AppKey == "" || cfg.AppSecret == "" || cfg.CronSpec == "" {
		return errors.New("host / appKey / appSecret / CronSpec maybe empty")
	}

	return nil
}

const (
	berserkerJobStatusOfSucceed = 1 << iota
	berserkerJobStatusOfFailed
	berserkerJobStatusOfRunning
	berserkerJobStatusOfPreparing

	BerserkerFieldTypeOfAppKey = 1 << iota
	BerserkerFieldTypeOfAppSecret
	BerserkerFieldTypeOfArchiveScoreInDB
	BerserkerFieldTypeOfCronSpec
	BerserkerFieldTypeOfEnabled
	BerserkerFieldTypeOfHost
	BerserkerFieldTypeOfKeepBackupFile

	berserkerPath4CreateJob = "avenger/api/%v/query"

	BerserkerAppID            = 670
	BerserkerSignMethodOfHMAC = "hmac"
	BerserkerSignMethodOfMD5  = "md5"
	BerserkerVersionOfDefault = "1.0"
)

var (
	innerBerserker atomic.Value
)

func init() {
	berserker4Init := BerserkerCfg{}
	innerBerserker.Store(berserker4Init)
}

func InitBerserker(cfg BerserkerCfg) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	innerBerserker.Store(cfg)

	return nil
}

func berserkerValueByFieldType(fieldType int) interface{} {
	d, ok := innerBerserker.Load().(BerserkerCfg)
	if !ok {
		return ""
	}

	switch fieldType {
	case BerserkerFieldTypeOfAppKey:
		return d.AppKey
	case BerserkerFieldTypeOfAppSecret:
		return d.AppSecret
	case BerserkerFieldTypeOfArchiveScoreInDB:
		return d.ArchiveScoreInDB
	case BerserkerFieldTypeOfCronSpec:
		return d.CronSpec
	case BerserkerFieldTypeOfHost:
		return d.Host
	case BerserkerFieldTypeOfEnabled:
		return d.Enabled
	case BerserkerFieldTypeOfKeepBackupFile:
		return d.KeepBackupFile
	default:
		return ""
	}
}

func berserkerAppKey() string {
	if d, ok := berserkerValueByFieldType(BerserkerFieldTypeOfAppKey).(string); ok {
		return d
	}

	return ""
}

func berserkerAppSecret() string {
	if d, ok := berserkerValueByFieldType(BerserkerFieldTypeOfAppSecret).(string); ok {
		return d
	}

	return ""
}

func berserkerHost() string {
	if d, ok := berserkerValueByFieldType(BerserkerFieldTypeOfHost).(string); ok {
		return d
	}

	return ""
}

func IsArchiveScoreBizEnabled() bool {
	if d, ok := berserkerValueByFieldType(BerserkerFieldTypeOfEnabled).(bool); ok {
		return d
	}

	return false
}

func CanResetArchiveScoreInDB() bool {
	if d, ok := berserkerValueByFieldType(BerserkerFieldTypeOfArchiveScoreInDB).(bool); ok {
		return d
	}

	return false
}

func KeepBackupFile() bool {
	if d, ok := berserkerValueByFieldType(BerserkerFieldTypeOfKeepBackupFile).(bool); ok {
		return d
	}

	return false
}

func FetchBerserkerJobFile(query string) (files []string, err error) {
	params, _ := genSignStr(
		berserkerAppKey(),
		berserkerAppSecret(),
		BerserkerVersionOfDefault,
		BerserkerSignMethodOfMD5,
		time.Now().Format("2006-01-02 15:04:05"),
		query)
	reqUrl := genQueryUrl()

	if _, err = url.Parse(reqUrl); err != nil {
		return
	}

	jobStatusUrl, err := createBerserkerJob(reqUrl, params)
	if err != nil {
		return
	}

	filePaths, err := waitBerserkerSucceed(jobStatusUrl)
	return filePaths, err
}

func createBerserkerJob(reqUrl string, params interface{}) (jobStatusUrl string, err error) {
	jobInfo := new(berserkerCreteJobRes)
	bs, _ := json.Marshal(params)

	err = berserkerDo(reqUrl, http.MethodPost, strings.NewReader(string(bs)), jobInfo)
	if err != nil {
		return
	}

	if jobInfo.Code == http.StatusOK {
		jobStatusUrl = jobInfo.JobStatusUrl
		if jobStatusUrl == "" {
			err = errors.Errorf("berserker: create job failed(job status url is empty), res(%v)", jobInfo)
		}

		return
	}

	err = errors.Errorf("berserker: create job failed, res(%v)", jobInfo)

	return
}

func waitBerserkerSucceed(jobStatusUrl string) (filePaths []string, err error) {
	filePaths = make([]string, 0)
	jobStatus := new(berserkerJobStatusRes)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*60)
	ticker := time.NewTicker(time.Second * 30)
	defer func() {
		ticker.Stop()
		cancel()
	}()

	for {
		select {
		case <-ctx.Done():
			err = errors.Errorf("berserker: wait job succeed timeout, jobStatusUrl(%v)", jobStatusUrl)

			return
		case <-ticker.C:
			doErr := berserkerDo(jobStatusUrl, http.MethodGet, nil, jobStatus)
			if doErr != nil {
				err = doErr

				return
			}

			if jobStatus.Code != http.StatusOK {
				err = errors.Errorf("berserker: wait job succeed failed(biz code is not 200), res(%v)", jobStatus)

				return
			}

			switch jobStatus.StatusID {
			case berserkerJobStatusOfSucceed:
				filePaths = jobStatus.FileUrls

				return
			case berserkerJobStatusOfFailed:
				err = errors.Errorf("berserker: wait job succeed failed(job is failed), res(%v)", jobStatus)

				return
			default:
				continue
			}
		}
	}
}

func berserkerDo(reqUrl, httpMethod string, params io.Reader, d interface{}) (err error) {
	c := http.Client{}
	{
		c.Timeout = time.Minute * 1
	}

	req, _ := http.NewRequest(httpMethod, reqUrl, params)
	req.Header.Set("Content-Type", ContentTypeOfJson)
	res, err := c.Do(req)
	if err != nil {
		return
	}

	if res == nil {
		err = errors.New(fmt.Sprintf("berserker req(%v), response(%v) is not excepted", req, res))

		return
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		err = errors.New(
			fmt.Sprintf(
				"berserker req(%v), http statu code is not 200(%v)",
				req,
				res.Status))
	}

	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	if unmarshalErr := json.Unmarshal(bs, d); unmarshalErr != nil {
		err = errors.New(
			fmt.Sprintf(
				"berserker req(%v), unmarshal resp(%v) failed, err: %v",
				req,
				string(bs),
				unmarshalErr.Error()))

		return
	}

	return
}

// e.g: http://berserker.bilibili.co/avenger/api/37/query?appKey=ac59b99d3f6d0ec1dc82a9a2aefb01a5&timestamp=2018-01-04 11:11:12
//
//	&signMethod=md5&sign=7CA05690169E57F4B5398B05C297B13D
//	&query=select filed1, filed2 form db1.table1 limit 10
func genQueryUrl() string {
	path := fmt.Sprintf(berserkerPath4CreateJob, BerserkerAppID)

	return fmt.Sprintf("%v/%v", berserkerHost(), path)
}

func genSignStr(appKey, secret, version, method, timestamp, query string) (map[string]interface{}, string) {
	var signStr string
	params := make(map[string]interface{}, 0)

	m := make(map[string]string, 3)
	{
		m["appKey"] = appKey
		m["timestamp"] = timestamp // time.Now().Format("2006/01/02 15:04:05")
		m["version"] = version
	}

	keyArr := make([]string, 0)
	for k := range m {
		keyArr = append(keyArr, k)
	}

	sort.Strings(keyArr)

	for _, key := range keyArr {
		for k, v := range m {
			if key == k {
				signStr = fmt.Sprintf("%v%v%v", signStr, k, v)
				params[k] = v
			}
		}
	}

	switch method {
	case BerserkerSignMethodOfMD5:
		signStr = fmt.Sprintf("%v%v%v", secret, signStr, secret)
		h := md5.New()
		h.Write([]byte(signStr))
		newStr := h.Sum(nil)
		signStr = fmt.Sprintf("%X", newStr)
		params["sign"] = signStr
		params["signMethod"] = BerserkerSignMethodOfMD5
		params["query"] = query

	case BerserkerSignMethodOfHMAC:
		// TODO: maybe need sdk!!!
	}

	return params, signStr
}
