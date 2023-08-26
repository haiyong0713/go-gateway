package tianma

import (
	"bufio"
	"context"
	"crypto/md5"
	"fmt"
	"github.com/pkg/errors"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/tianma"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

// 更新某一条推荐的file信息
func (s *Service) UpdateMidFileInfo(id int64, item *tianma.PosRecItem, username string) (err error) {
	// if username == "" {
	// 	err = ecode.Error(ecode.NoLogin, "未登录")
	// 	return
	// }

	err = s.dao.UpdatePosRecItemById(id, item)

	if err != nil {
		log.Error("s.dao.UpdatePosRecItemById id(%v) item(%v) err(%v)", id, item, err)
	}

	return
}

// 通过 URL 下载某个文件到本地
func (s *Service) DownloadFileByUrl(url string) (target string, err error) {
	//nolint:gosec
	netResp, err := http.Get(url)
	if err != nil {
		log.Error("http.Get(%v) error %v", url, err)
		return
	}
	defer netResp.Body.Close()

	// 创建一个文件用于保存
	fileName := fmt.Sprintf("%x", md5.Sum([]byte(url)))
	target = path.Join(s.c.Boss.LocalDir, _midFileNetDir, fileName)
	out, err := os.Create(target)
	if err != nil {
		log.Error("os.Create %v error %v", target, err)
		return
	}
	defer out.Close()

	// 然后将响应流和文件流对接起来
	_, err = io.Copy(out, netResp.Body)
	if err != nil {
		log.Error("io.Copy error %v", err)
		return
	}

	log.Info("DownloadFileByUrl succeed url(%s) target(%s)", url, target)

	return
}

// 通过 boss 的 key 下载某个文件到本地
func (s *Service) DownloadFileByBossKey(key string) (target string, err error) {
	target = path.Join(s.c.Boss.LocalDir, _midFileBossDir, key)

	err = s.BossDownloadLocalFile(key, target)
	if err != nil {
		log.Error("s.BossDownloadLocalFile error %v", err)
		return
	}

	log.Info("DownloadFileByBossKey succeed key(%s) target(%s)", key, target)

	return
}

// 监控是否有推荐对应的mid人群包没有被下载或者上传
func (s *Service) MidFileMonitor() (err error) {
	log.Info("MidFileMonitor loop start at time(%v)", time.Now())
	for {
		log.Info("MidFileMonitor check at time(%v)", time.Now())
		err = s.CheckStatusThenHandle()
		if err != nil {
			log.Error("s.CheckStatusThenHandle error %v", err)
		}

		err = s.MonitorDmpOrHdfs()
		if err != nil {
			log.Error("s.MonitorDmpOrHdfs error %v", err)
		}
		time.Sleep(10 * time.Second)
	}
}

// 判断httppath 是否可访问
func (s *Service) IsHttpPathAccessible(httpPath string) (flag bool, err error) {
	//nolint:gosec
	res, err := http.Get(httpPath)
	if err != nil || res.StatusCode != 200 {
		return false, err
	}

	buf, err := ioutil.ReadAll(io.LimitReader(res.Body, 10))
	res.Body.Close()
	if err != nil || len(buf) == 0 {
		return false, err
	}

	return true, nil
}

// 判断hdfsPath 是否可访问
func (s *Service) IsHdfsPathAccessible(hdfsPath string) (flag bool, err error) {
	httpPath := s.TransfHdfsToHttPath(hdfsPath, false)
	flag, err = s.IsHttpPathAccessible(httpPath)
	return
}

const (
	FILE_TYPE_DMO_ID = 1
	FILE_TYPE_URL    = 2
)

// 监控是否配置商业dmp id 或 是否配置hdfs路径
func (s *Service) MonitorDmpOrHdfs() (err error) {
	//查询出所有商业dmp Id 和 hdfs地址
	log.Info("MonitorDmpOrHdfs loop start at time(%v)", time.Now())
	var fileTypeAdditionArr = []int{1, 2}
	posRecList, err := s.dao.SearchPosRecListByTypeAddition(0, fileTypeAdditionArr)
	if err != nil {
		log.Error("service.MonitorDmpOrHdfs SearchPosRecListByTypeAddition error %v", err)
		return
	}

	for _, posRec := range posRecList {
		var httpPath = ""
		if posRec.FileTypeAddition == FILE_TYPE_DMO_ID { //商业dmp id
			hdfsPath, err1 := s.GetUrlByDmpId(posRec.FilePathAddition)
			if err1 != nil {
				log.Error("service.MonitorDmpOrHdfs GetUrlByDmpId error %v", err)
				continue
			}
			httpPath = s.TransfHdfsToHttPath(hdfsPath, true)
		} else if posRec.FileTypeAddition == FILE_TYPE_URL { //存储url方式(http链接 或 hdfs链接)
			if strings.HasPrefix(posRec.FilePathAddition, "http") || strings.HasPrefix(posRec.FilePathAddition, "ftp") {
				httpPath = posRec.FilePathAddition
			} else {
				httpPath = s.TransfHdfsToHttPath(posRec.FilePathAddition, false)
			}
		}

		//判断http链接是否有效
		ret, _ := s.IsHttpPathAccessible(httpPath)
		if ret {
			posRecItem := &tianma.PosRecItem{
				FileStatus: 1,
				FilePath:   httpPath,
			}
			//nolint:errcheck
			s.dao.UpdatePosRecItemById(posRec.Id, posRecItem)
		} else {
			posRecItem := &tianma.PosRecItem{
				FileStatus: 9, // hdfs下载失败
			}
			if err = s.dao.UpdatePosRecItemById(posRec.Id, posRecItem); err != nil {
				log.Error("service.MonitorDmpOrHdfs UpdatePosRecItemById Id(%+v) err(%+v)", posRec.Id, err)
			}
		}
	}

	return
}

const DMP_EXPORT_URI = "/dmp/api/user_api/v1/group/export_group/uri"

// 根据商业dmp id 获取人群包下载路径
func (s *Service) GetUrlByDmpId(dmpId string) (hdfsPath string, err error) {
	params := url.Values{}
	params.Set("gid", dmpId)
	url := s.dao.CmmngHost + DMP_EXPORT_URI
	req, err := s.dao.HttpClient.NewRequest(http.MethodGet, url, "", params)
	if err != nil {
		log.Error("service.GetUrlByDmpId NewRequest err %v", err)
		return
	}

	res := struct {
		Status string `json:"status"`
		Result []struct {
			HdfsPath string `json:"hdfs_path"`
		} `json:"result"`
	}{}
	//nolint:errcheck
	s.dao.HttpClient.Do(context.Background(), req, &res)
	if len(res.Result) == 0 {
		err = errors.New("HdfsPath is empty")
		log.Error("service.GetUrlByDmpId HdfsPath empty")
		return
	}

	hdfsPath = res.Result[0].HdfsPath
	return
}

const DOWNLOAD_HDFS = "/avenger/download/hdfs"
const DMP_FILE_NAME = "part-00000" //商业dmp 人群包固定文件后缀 000000_0

// 根据hdfs下载路径，拼接http下载地址
func (s *Service) TransfHdfsToHttPath(hdfsPath string, needSuffix bool) (httpPath string) {
	httpPath = s.dao.BerserkerHost + DOWNLOAD_HDFS + "?" + "path="
	if strings.HasPrefix(hdfsPath, "/") {
		httpPath += hdfsPath
	} else {
		httpPath += "/" + hdfsPath
	}
	if needSuffix {
		httpPath += "/" + DMP_FILE_NAME
	}

	return
}

// 检查目前状态为：
// 1 URL：改为 2；下载，完成后改为 3，写入本地路径；失败改成 1
// 3 在本地：改为 4；上传，完成后改为 5，写入 key；失败改成 3
// 7 在 boss：改为 8；下载，计算行数，完成后改为 5；失败改成 7
func (s *Service) CheckStatusThenHandle() (err error) {
	// 查出所有 status=1，需要从 URL 下载到本地
	posRecList1, err := s.dao.SearchPosRecListByStatus(1)
	if err != nil {
		log.Error("s.dao.SearchPosRecListByStatus error %v", err)
		return
	}

	for _, posRec1 := range posRecList1 {
		//nolint:errcheck,biligowordcheck
		go s.downloadStep(posRec1)
	}

	// 查出所有 status=3，需要从本地上传到 boss
	posRecList3, err := s.dao.SearchPosRecListByStatus(3)
	if err != nil {
		log.Error("s.dao.SearchPosRecListByStatus error %v", err)
		return
	}

	for _, posRec3 := range posRecList3 {
		//nolint:errcheck,biligowordcheck
		go s.uploadStep(posRec3)
	}

	// 查出所有 status=7，需要从 boss 下载到本地，然后计算行数
	posRecList7, err := s.dao.SearchPosRecListByStatus(7)
	if err != nil {
		log.Error("s.dao.SearchPosRecListByStatus error %v", err)
		return
	}

	for _, posRec7 := range posRecList7 {
		//nolint:errcheck,biligowordcheck
		go s.downloadFromBossAndCountLineNum(posRec7)
	}

	return
}

func (s *Service) getLocationFromPath(filePath string, index int) (location string) {
	locations := strings.Split(filePath, "|")
	if len(locations) <= index {
		return
	}
	location = locations[index]
	return
}

func (s *Service) downloadStep(posRec *tianma.PosRecItem) (err error) {
	// 变更为 2，标识开始下载
	err = s.UpdateMidFileInfo(posRec.Id, &tianma.PosRecItem{
		FileStatus: 2,
	}, "server")
	if err != nil {
		log.Error("s.UpdateMidFileInfo error %v", err)
		return
	}
	defer func() {
		if err != nil {
			// 失败，回滚回 status=1，下一次任务继续
			err = s.UpdateMidFileInfo(posRec.Id, &tianma.PosRecItem{
				FileStatus: 1,
			}, "server")
			if err != nil {
				log.Error("s.UpdateMidFileInfo recover error %v", err)
			}
		}
	}()

	url := s.getLocationFromPath(posRec.FilePath, 0)

	// 开始下载
	target, err := s.DownloadFileByUrl(url)
	if err != nil {
		log.Error("s.DownloadFileByUrl error %v", err)
		return
	}

	// 下载成功，更改状态为 3，并写入本地存储的路径
	err = s.UpdateMidFileInfo(posRec.Id, &tianma.PosRecItem{
		FilePath:   url + "|" + target,
		FileStatus: 3,
	}, "server")
	if err != nil {
		log.Error("s.UpdateMidFileInfo id(%v) FilePath(%v) error(%v)", posRec.Id, target, err)
		return
	}

	// 获取文件行数，并写入，不关键
	lineNum, err := s.getFileLineNum(target)
	if err != nil {
		log.Error("s.getFileLineNum filePath(%v) error(%v)", target, err)
		return
	}
	err = s.UpdateMidFileInfo(posRec.Id, &tianma.PosRecItem{
		FileRows: lineNum,
	}, "server")
	if err != nil {
		log.Error("s.UpdateMidFileInfo error %v", err)
	}

	return
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func (s *Service) uploadStep(posRec *tianma.PosRecItem) (err error) {
	// 检查是否存储了本地路径
	localPath := s.getLocationFromPath(posRec.FilePath, 1)
	if localPath == "" {
		// 表示没有存储本地路径，恢复为1
		err = s.UpdateMidFileInfo(posRec.Id, &tianma.PosRecItem{
			FileStatus: 1,
		}, "server")
		if err != nil {
			log.Error("s.UpdateMidFileInfo error %v", err)
		}
		return
	}

	// 检查文件是否在本地
	if ext, _ := exists(localPath); !ext {
		// 如果文件丢了，修改状态为 1 等待重新下载
		err = s.UpdateMidFileInfo(posRec.Id, &tianma.PosRecItem{
			FileStatus: 1,
		}, "server")
		if err != nil {
			log.Error("s.UpdateMidFileInfo error %v", err)
		}
		return
	}

	// 更改为 4，标识正在上传到 boss
	err = s.UpdateMidFileInfo(posRec.Id, &tianma.PosRecItem{
		FileStatus: 4,
	}, "server")
	if err != nil {
		log.Error("s.UpdateMidFileInfo error %v", err)
		return
	}
	defer func() {
		if err != nil {
			// 失败，状态恢复为 3，下一次任务继续
			err = s.UpdateMidFileInfo(posRec.Id, &tianma.PosRecItem{
				FileStatus: 3,
			}, "server")
			if err != nil {
				log.Error("s.UpdateMidFileInfo recover error %v", err)
			}
		}
	}()

	// 开始上传
	target, err := s.BossUploadLocalFile(localPath)
	if err != nil {
		log.Error("s.DownloadFileByUrl error %v", err)
		return
	}

	// 上传成功，更改状态为 5，并写入 boss 的 key
	err = s.UpdateMidFileInfo(posRec.Id, &tianma.PosRecItem{
		FilePath:   target.Key,
		FileStatus: 5,
	}, "server")
	if err != nil {
		log.Error("s.UpdateMidFileInfo id(%v) FilePath(%v) error(%v)", posRec.Id, target.Key, err)
		return
	}

	// 上传成功，本地文件可以删除了
	err = os.Remove(localPath)
	if err != nil {
		log.Error("os.Remove localPath %v error %v", localPath, err)
		err = nil
	}

	return
}

// 单独下载 boos 文件并计算和更新行数，只适用于有 file_status=7 的推荐
func (s *Service) downloadFromBossAndCountLineNum(posRec *tianma.PosRecItem) (err error) {
	// 变更为 8，标识开始下载
	err = s.UpdateMidFileInfo(posRec.Id, &tianma.PosRecItem{
		FileStatus: 8,
	}, "server")
	if err != nil {
		log.Error("s.UpdateMidFileInfo error %v", err)
		return
	}
	defer func() {
		if err != nil {
			// 失败，回滚回 status=7，下一次任务继续
			err = s.UpdateMidFileInfo(posRec.Id, &tianma.PosRecItem{
				FileStatus: 7,
			}, "server")
			if err != nil {
				log.Error("s.UpdateMidFileInfo recover error %v", err)
			}
		}
	}()

	// 开始下载
	target := path.Join(s.c.Boss.LocalDir, _midFileBossDir, posRec.FilePath)
	err = s.BossDownloadLocalFile(posRec.FilePath, target)
	if err != nil {
		log.Error("s.BossDownloadLocalFile error %v", err)
		return
	}

	// 获取文件行数，并写入，不关键
	lineNum, err := s.getFileLineNum(target)
	if err != nil {
		log.Error("s.getFileLineNum filePath(%v) error(%v)", target, err)
		return
	}

	// 变更为 status=5，可以开始使用
	err = s.UpdateMidFileInfo(posRec.Id, &tianma.PosRecItem{
		FileRows:   lineNum,
		FileStatus: 5,
	}, "server")
	if err != nil {
		log.Error("s.UpdateMidFileInfo error %v", err)
	}

	// 计算完成了，本地文件可以删除了
	err = os.Remove(target)
	if err != nil {
		log.Error("os.Remove filePath %v error %v", target, err)
		err = nil
	}

	return
}

// 获取本地文件行数
func (s *Service) getFileLineNum(filePath string) (num int64, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Error("os.Open() error %v", filePath)
		return
	}
	defer file.Close()

	fd := bufio.NewReader(file)
	for {
		_, err := fd.ReadString('\n')
		if err != nil {
			break
		}
		num++
	}

	return
}
